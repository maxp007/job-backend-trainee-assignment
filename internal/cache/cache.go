package cache

import (
	"context"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"job-backend-trainee-assignment/internal/logger"
	"sync"
	"time"
)

var defaultKeyExpirationTime = 30 * time.Second

type ICacher interface {
	CheckKeyExistence(ctx context.Context, key string) (keyExists bool, err error)
	AddKey(ctx context.Context, key string) (err error)
}

type CacheConfig struct {
	KeyExpirationTime time.Duration
}

// RedisCache is a struct implementing ICacher interface
type RedisCache struct {
	logger logger.ILogger
	cfg    *CacheConfig
	redis  *redis.Pool
	mu     sync.Mutex
}

func NewRedisCache(log logger.ILogger, redisPool *redis.Pool, cfg *CacheConfig) (*RedisCache, error) {

	if log == nil {
		return nil, fmt.Errorf("provided logger param is nil")
	}

	if cfg == nil {
		log.Error("Provided Config param is nil")
		cfg = &CacheConfig{KeyExpirationTime: defaultKeyExpirationTime}
	}

	if redisPool == nil {
		log.Error("Provided redisPool param is nil")
		return nil, fmt.Errorf("provided redisPool param is nil")
	}

	return &RedisCache{
		logger: log,
		redis:  redisPool,
		cfg:    cfg,
		mu:     sync.Mutex{},
	}, nil
}
