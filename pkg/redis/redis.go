package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

type Pool interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Exists(ctx context.Context, keys ...string) *redis.IntCmd
}

const (
	defaultMaxPoolSize  = 10
	defaultConnAttempts = 10
	defaultConnTimeout  = 1 * time.Second
)

type Redis struct {
	*redis.Client
}

func NewRedis(url string, opts ...Option) *Redis {
	rdbOpts := &redis.Options{
		Addr:            url,
		PoolSize:        defaultMaxPoolSize,
		MaxRetries:      defaultConnAttempts,
		MaxRetryBackoff: defaultConnTimeout,
	}
	for _, option := range opts {
		option(rdbOpts)
	}
	return &Redis{redis.NewClient(rdbOpts)}
}

func (r *Redis) Close() {
	if r.Client != nil {
		_ = r.Client.Close()
	}
}
