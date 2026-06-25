package database

import (
	"context"
	"crypto/tls"
	"os"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedis(addr, password string, db int) *RedisClient {
	opt := &redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	}

	if os.Getenv("REDIS_TLS") == "true" {
		opt.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	r := redis.NewClient(opt)
	return &RedisClient{client: r}
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}

func (r *RedisClient) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Client returns the underlying *redis.Client
func (r *RedisClient) Client() *redis.Client {
	return r.client
}