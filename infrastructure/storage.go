package infrastructure

import (
	"fmt"
	"github.com/go-redis/redis"
)

type Storage interface {
	Set(key string, value interface{}) error
	Get(key string) (string, error)
	ZIncrBy(key string, increment float64, member string) error
	ZRevRangeWithScores(key string, start, stop int64) ([]RedisScoredMember, error)
	Scan(cursor uint64, match string, count int64) ([]string, uint64, error)
	FlushDB() error
}

type RedisStorage struct {
	client *redis.Client
}
type RedisScoredMember struct {
	Score  float64
	Member string
}

func (rsm *RedisScoredMember) GetScore() float64 {
	return rsm.Score
}

func (rsm *RedisScoredMember) GetMember() string {
	return rsm.Member
}

func NewRedisStorage(addr string, password string, db int) (*RedisStorage, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
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

func (rs *RedisStorage) ZRevRangeWithScores(key string, start, stop int64) ([]RedisScoredMember, error) {
	zs, err := rs.client.ZRevRangeWithScores(key, start, stop).Result()
	if err != nil {
		return nil, err
	}

	var members []RedisScoredMember

	for _, z := range zs {
		member := RedisScoredMember{Score: z.Score, Member: z.Member.(string)}
		members = append(members, member)

	}
	return members, nil
}

func (rs *RedisStorage) Scan(cursor uint64, match string, count int64) ([]string, uint64, error) {
	return rs.client.Scan(cursor, match, count).Result()
}

func (rs *RedisStorage) FlushDB() error {
	return rs.client.FlushDB().Err()
}
