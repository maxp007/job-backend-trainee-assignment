package cache

import (
	"context"
	"fmt"
	"github.com/gomodule/redigo/redis"
)

type ChanResult struct {
	keyExist bool
	err      error
}

// CheckKeyExistence checks for specified key existence in redis cache
func (c *RedisCache) CheckKeyExistence(ctx context.Context, key string) (keyExists bool, err error) {
	c.mu.Lock()
	KeyLookupTimeout := c.cfg.KeyLookupTimeout
	c.mu.Unlock()

	ctx, cancel := context.WithTimeout(ctx, KeyLookupTimeout)
	defer cancel()

	conn, err := c.redis.GetContext(ctx)
	if err != nil {
		if ctx.Err() != nil {
			c.logger.Error("CheckKeyExistence, failed to Get Conn from pool, err: %v", err)
			return false, fmt.Errorf("failed to get conn from pool, context err: %w", ErrContextDeadlineExceeded)
		}

		c.logger.Error("CheckKeyExistence, failed to Get Conn from pool, err: %v", err)
		return false, fmt.Errorf("failed to get conn from pool, err: %w, err: %v", ErrFailedToGetConnFromPool, err)
	}

	resCh := make(chan ChanResult, 1)
	go func() {
		resVal, locErr := redis.Int(conn.Do("EXISTS", key))
		if locErr != nil {
			errOnClose := conn.Close()
			if errOnClose != nil {
				c.logger.Error("CheckKeyExistence, failed to Close conn, err: %v", errOnClose)
			}
			c.logger.Error("CheckKeyExistence, failed to parse redis response, err: %v", locErr)
			locErr = fmt.Errorf("failed to check key existence, err:%v, err:%w", locErr, ErrFailedToPerformDoCommand)
			resCh <- ChanResult{keyExist: false, err: locErr}
			return
		}

		var locRes bool
		if resVal == 1 {
			locRes = true
		} else {
			locRes = false
		}
		errOnClose := conn.Close()
		if errOnClose != nil {
			c.logger.Error("CheckKeyExistence, failed to Close conn, err: %v", errOnClose)
		}
		resCh <- ChanResult{keyExist: locRes, err: locErr}
	}()

	select {
	case <-ctx.Done():
		{
			c.logger.Error("CheckKeyExistence, failed to lookup key, context timeout, err: %v", ErrKeyLookUpTimeout)
			return false, fmt.Errorf("key lookup context timeout, err: %w", ErrContextDeadlineExceeded)
		}
	case res := <-resCh:
		{
			if res.err != nil {
				c.logger.Error("CheckKeyExistence, failed to lookup key,error occured, err: %v", err)
			}

			return res.keyExist, res.err
		}
	}

}

// AddKey adds specified key into redis cache
func (c *RedisCache) AddKey(ctx context.Context, key string) error {
	c.mu.Lock()
	KeyExpirationTime := c.cfg.KeyExpirationTime
	KeySetTimeout := c.cfg.KeySetTimeout
	c.mu.Unlock()

	keyCopy := key
	ctx, cancel := context.WithTimeout(ctx, KeySetTimeout)
	defer cancel()

	conn, err := c.redis.GetContext(ctx)
	if err != nil {
		if ctx.Err() != nil {
			c.logger.Error("AddKey, failed to Get Conn from pool, err: %v", err)
			return fmt.Errorf("failed to get conn from pool, context err: %w", ErrContextDeadlineExceeded)
		}

		c.logger.Error("AddKey, failed to Get Conn from pool, err: %v", err)
		return fmt.Errorf("failed to get conn from pool, err: %w, err: %v", ErrFailedToGetConnFromPool, err)
	}

	errCh := make(chan error, 1)
	go func() {
		OkResp, locErr := redis.String(conn.Do("SETEX", keyCopy, KeyExpirationTime.Seconds(), 1))
		if locErr != nil || OkResp != "OK" {
			errOnClose := conn.Close()
			if errOnClose != nil {
				c.logger.Error("AddKey, failed to Close conn, err: %v", errOnClose)
			}
			c.logger.Error("AddKey, failed to parse redis response, err: %v", locErr)
			locErr = fmt.Errorf("failed to set, err:%v, err:%w", err, ErrFailedToPerformDoCommand)
			errCh <- locErr
			return
		}
		errOnClose := conn.Close()
		if errOnClose != nil {
			c.logger.Error("AddKey, failed to Close conn, err: %v", errOnClose)
		}
		errCh <- locErr
	}()

	select {
	case <-ctx.Done():
		{
			c.logger.Error("AddKey, failed to set key, context timeout, err: %vm ctxErr:%v", ErrKeySetTimeout, ctx.Err())
			return fmt.Errorf("key set context timeout, err: %w", ErrContextDeadlineExceeded)
		}
	case errRes := <-errCh:
		{
			if errRes != nil {
				c.logger.Error("AddKey, failed to set key, error occurred, err: %v", err)
			}

			return errRes
		}
	}
}
