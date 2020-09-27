package cache

import (
	"context"
	"github.com/gomodule/redigo/redis"
)

// CheckKeyExistence checks for specified key existence in redis cache
func (c *RedisCache) CheckKeyExistence(ctx context.Context, key string) (keyExists bool, err error) {
	conn, err := c.redis.GetContext(ctx)
	if err != nil {
		c.logger.Error("CheckKeyExistence, failed to Get Conn from pool, err: %v", err)
		return false, ErrFailedToGetConnFromPool
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			c.logger.Error("CheckKeyExistence, failed to Close conn, err: %v", err)
		}
	}()
	isKeyExist, err := redis.Int64(conn.Do("EXISTS", key))
	if err != nil {
		c.logger.Error("CheckKeyExistence, failed to parse redis response, err: %v", err)
		return false, ErrFailedToPerformDoCommand
	}

	if isKeyExist == 1 {
		return true, nil
	}

	return false, nil

}

// AddKey adds specified key into redis cache
func (c *RedisCache) AddKey(ctx context.Context, key string) (err error) {
	c.mu.Lock()
	KeyExpirationTime := c.cfg.KeyExpirationTime
	c.mu.Unlock()
	conn, err := c.redis.GetContext(ctx)
	if err != nil {
		c.logger.Error("AddKey, failed to Get Conn from pool, err: %v", err)
		return ErrFailedToGetConnFromPool
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			c.logger.Error("AddKey, failed to Close conn, err: %v", err)
		}
	}()

	OkResp, err := redis.String(conn.Do("SETEX", key, KeyExpirationTime.Seconds(), 1))
	if err != nil || OkResp != "OK" {
		c.logger.Error("AddKey, failed to parse redis response, err: %v", err)
		return ErrFailedToPerformDoCommand
	}

	return nil
}
