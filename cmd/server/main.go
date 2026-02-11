package main

import (
	"log"
	"time"

	"github.com/Oyagee/antibruteforce/configs"
	"github.com/Oyagee/antibruteforce/internal/api"
	"github.com/Oyagee/antibruteforce/internal/app"
	"github.com/Oyagee/antibruteforce/internal/bucket"
	"github.com/Oyagee/antibruteforce/internal/service"
	"github.com/Oyagee/antibruteforce/internal/storage"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := configs.LoadConfig()
	store := storage.NewInMemoryStorage()
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
		DB:   0,
	})
	defer func() {
		if err := rdb.Close(); err != nil {
			log.Printf("failed to close redis: %v", err)
		}
	}()
	rl := bucket.NewRateLimiter(rdb, 5*time.Minute,
		bucket.Config{Capacity: cfg.CLogin, RefillPerMinute: cfg.RLogin},
		bucket.Config{Capacity: cfg.CPass, RefillPerMinute: cfg.RPass},
		bucket.Config{Capacity: cfg.CIP, RefillPerMinute: cfg.RIP})
	svc := service.New(store, rl)
	router := api.NewRouter(svc)
	srv := app.NewServer(":"+cfg.Port, router)
	if err := srv.Run(); err != nil {
		if err := rdb.Close(); err != nil {
			log.Printf("failed to close redis: %v", err)
		}
		//nolint: gocritic
		log.Fatalf("server error: %v", err)
	}
}
