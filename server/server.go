package server

import (
	"fmt"
	"infracloud-golang/app"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Server struct {
	r  *gin.Engine
	us *app.URLShortener
}

func NewServer(us *app.URLShortener) *Server {
	server := &Server{
		r:  gin.Default(),
		us: us,
	}

	server.r.GET("/:shortURL", server.handleRedirect)
	server.r.POST("/shorten", server.handleShorten)
	server.r.GET("/metrics", server.handleMetrics)
	server.r.GET("/viewAll", server.handleViewAll)
	server.r.DELETE("/deleteAll", server.handleDeleteAll)

	return server
}

func (server *Server) Run(addr string) error {
	if addr == "" {
		return fmt.Errorf("address cannot be empty")
	}
	return server.r.Run(addr)
}

func (server *Server) handleShorten(c *gin.Context) {
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

	shortURL, err := server.us.Shorten(json.OriginalURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to shorten URL: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"short_url": shortURL,
	})
}

func (server *Server) handleRedirect(c *gin.Context) {
	shortURL := c.Param("shortURL")

	if shortURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "shortURL cannot be empty"})
		return
	}

	originalURL, err := server.us.Redirect(shortURL)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}

	c.Redirect(http.StatusMovedPermanently, originalURL)
}

func (server *Server) handleMetrics(c *gin.Context) {
	topDomains, err := server.us.Metrics()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get top domains: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"top_domains": topDomains,
	})
}

func (server *Server) handleViewAll(c *gin.Context) {
	data, err := server.us.ViewAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to view all URLs: %v", err)})
		return
	}

	c.JSON(http.StatusOK, data)
}

func (server *Server) handleDeleteAll(c *gin.Context) {
	err := server.us.DeleteAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete all records: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "All records have been deleted.",
	})
}
