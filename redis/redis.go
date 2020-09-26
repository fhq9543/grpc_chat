package redis

import (
	"context"
	"grpcChat/config"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()
var redisClient *redis.Client

func InitRedis() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})
}

func RedisSet(key, value string) error {
	return redisClient.Set(ctx, key, value, 0).Err()
}

func RedisGet(key string) (string, error) {
	return redisClient.Get(ctx, key).Result()
}

func RedisRPush(key, value string) error {
	return redisClient.RPush(ctx, key, value).Err()
}

func RedisGetList(key string) ([]string, error) {
	return redisClient.LRange(ctx, key, 0, -1).Result()
}

func RedisPopList(key string) []string {
	var err error
	var value string
	res := []string{}
	for {
		value, err = redisClient.LPop(ctx, key).Result()
		if err != nil {
			break
		}
		res = append(res, value)
	}
	return res
}
