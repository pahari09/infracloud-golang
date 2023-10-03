package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
)

const (
	RedisAddr     = "localhost:6379"
	RedisPassword = ""
	RedisDB       = 0
	ServerAddr    = "localhost:8080"
)

type Storage interface {
	Set(key string, value interface{}) error
	Get(key string) (string, error)
	ZIncrBy(key string, increment float64, member string) error
	ZRevRangeWithScores(key string, start, stop int64) ([]redis.Z, error)
	Scan(cursor uint64, match string, count int64) ([]string, uint64, error)
	FlushDB() error
}

type RedisStorage struct {
	client *redis.Client
}

func NewRedisStorage() (*RedisStorage, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     RedisAddr,
		Password: RedisPassword,
		DB:       RedisDB,
	})

	if _, err := client.Ping().Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisStorage{client: client}, nil
}

func (rs *RedisStorage) Set(key string, value interface{}) error {
	return rs.client.Set(key, value, 0).Err()
}

func (rs *RedisStorage) Get(key string) (string, error) {
	return rs.client.Get(key).Result()
}

func (rs *RedisStorage) ZIncrBy(key string, increment float64, member string) error {
	return rs.client.ZIncrBy(key, increment, member).Err()
}

func (rs *RedisStorage) ZRevRangeWithScores(key string, start, stop int64) ([]redis.Z, error) {
	return rs.client.ZRevRangeWithScores(key, start, stop).Result()
}

func (rs *RedisStorage) Scan(cursor uint64, match string, count int64) ([]string, uint64, error) {
	return rs.client.Scan(cursor, match, count).Result()
}

func (rs *RedisStorage) FlushDB() error {
	return rs.client.FlushDB().Err()
}

type URLShortener struct {
	r     *gin.Engine
	store Storage
}

func NewURLShortener(store Storage) (*URLShortener, error) {
	us := &URLShortener{
		r:     gin.Default(),
		store: store,
	}

	us.r.GET("/:shortURL", us.handleRedirect)
	us.r.POST("/shorten", us.handleShorten)
	us.r.GET("/metrics", us.handleMetrics)
	us.r.GET("/viewAll", us.handleViewAll)
	us.r.DELETE("/deleteAll", us.handleDeleteAll)

	return us, nil
}

func (us *URLShortener) Run(addr string) error {
	if addr == "" {
		return fmt.Errorf("address cannot be empty")
	}
	return us.r.Run(addr)
}

func (us *URLShortener) handleShorten(c *gin.Context) {
	var json struct {
		OriginalURL string `json:"originalURL"`
	}

	if err := c.BindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to bind JSON: %v", err)})
		return
	}

	if json.OriginalURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "originalURL cannot be empty"})
		return
	}

	hash := sha256.Sum256([]byte(json.OriginalURL))
	shortURL := hex.EncodeToString(hash[:])[:8]

	err := us.store.Set("url:"+shortURL, json.OriginalURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to set URL in store: %v", err)})
		return
	}

	parsedURL, err := url.Parse(json.OriginalURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to parse URL: %v", err)})
		return
	}

	domain := parsedURL.Hostname()

	err = us.store.ZIncrBy("domains", 1.0, domain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to increment domain count in store: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"short_url": shortURL,
	})
}

func (us *URLShortener) handleRedirect(c *gin.Context) {
	shortURL := c.Param("shortURL")

	if shortURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "shortURL cannot be empty"})
		return
	}

	originalURL, err := us.store.Get("url:" + shortURL)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}

	c.Redirect(http.StatusMovedPermanently, originalURL)
}

func (us *URLShortener) handleMetrics(c *gin.Context) {
	topDomains, err := us.store.ZRevRangeWithScores("domains", 0, 2)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get top domains from store: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"top_domains": topDomains,
	})
}

func (us *URLShortener) handleViewAll(c *gin.Context) {
	data := make(map[string]string)
	cursor := uint64(0)

	for {
		keys, retCursor, err := us.store.Scan(cursor, "url:*", 10)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to scan keys from store: %v", err)})
			return
		}

		for _, key := range keys {
			value, err := us.store.Get(key)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get value from store: %v", err)})
				return
			}
			data[key] = value
		}

		if retCursor == 0 {
			break
		}
		cursor = retCursor
	}

	c.JSON(http.StatusOK, data)
}

func (us *URLShortener) handleDeleteAll(c *gin.Context) {
	err := us.store.FlushDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to flush database: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "All records have been deleted.",
	})
}

func main() {
	store, err := NewRedisStorage()

	if err != nil {
		fmt.Println(fmt.Sprintf("Failed to create storage: %v", err))
		return
	}

	us, err := NewURLShortener(store)

	if err != nil {
		fmt.Println(fmt.Sprintf("Failed to create URL shortener: %v", err))
		return
	}

	if err := us.Run(ServerAddr); err != nil {
		fmt.Println(fmt.Sprintf("Failed to start server: %v", err))
	}
}
