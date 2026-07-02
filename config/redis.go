package config

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

func InitRedis() (*redis.Client, error) {
	redisConfig := Config.Redis

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisConfig.Addr,
		DB:       redisConfig.DB,
		Password: redisConfig.Password,
	})

	response, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	log.Println("redis answer: ", response)
	return redisClient, nil
}
