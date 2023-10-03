package main

import (
	"fmt"
	"infracloud-golang/app"
	"infracloud-golang/infrastructure"
)

const (
	RedisAddr     = "localhost:6379"
	RedisPassword = ""
	RedisDB       = 0
	ServerAddr    = "localhost:8080"
)

func main() {
	store, err := infrastructure.NewRedisStorage(RedisAddr, RedisPassword, RedisDB)

	if err != nil {
		fmt.Println(fmt.Sprintf("Failed to create storage: %v", err))
		return
	}

	us := app.NewURLShortener(store)

	server := app.NewServer(us)

	if err := server.Run(ServerAddr); err != nil {
		fmt.Println(fmt.Sprintf("Failed to start server: %v", err))
	}
}
