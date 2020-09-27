package cache

import (
	"context"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"job-backend-trainee-assignment/internal/logger"
	"time"
)

type ConnConfig struct {
	Host          string
	Port          string
	DBName        int
	Pass          string
	MaxConn       int
	MaxIdleConn   int
	IdleTimeout   time.Duration
	RetryInterval time.Duration
}

func ConnectToRedisWithTimeout(ctx context.Context, log logger.ILogger, cfg *ConnConfig) (redisPool *redis.Pool, closeFunc func(), err error) {
	if log == nil {
		return nil, nil, fmt.Errorf("provided log param is nil")
	}

	if cfg == nil {
		log.Error("provided config param is nil")
		return nil, nil, fmt.Errorf("provided config param is nil")
	}

	if cfg.Pass == "" {
		log.Error("provided config param Pass is not filled")
		return nil, nil, fmt.Errorf("config param  %s is not filled or have default value", "Pass")
	}
	if cfg.Host == "" || cfg.Port == "" {
		log.Error("provided config params Host or Port is not filled")
		return nil, nil, fmt.Errorf("config param %s or %s is not filled or have default value", "Host", "Port")
	}

	connWaitChan := make(chan struct{})

	pool := &redis.Pool{
		DialContext: func(ctx context.Context) (redis.Conn, error) {
			conn, err := redis.DialContext(ctx, "tcp", fmt.Sprintf("%s:%s", cfg.Host, cfg.Port))
			if err != nil {
				log.Error("DialContext error, err: %v", err)
				return nil, err
			}

			_, err = conn.Do("AUTH", cfg.Pass)
			if err != nil {
				closeErr := conn.Close()
				if closeErr != nil {
					log.Error("AUTH error, err: %v", err)
				}
				return nil, err
			}

			_, err = conn.Do("SELECT", cfg.DBName)
			if err != nil {
				closeErr := conn.Close()
				if closeErr != nil {
					log.Error("SELECT db error, err: %v", err)
				}
				return nil, err
			}

			return conn, nil
		},
		TestOnBorrow: func(c redis.Conn, _ time.Time) error {
			_, err := c.Do("PING")
			return err
		},
		MaxIdle:     cfg.MaxIdleConn,
		MaxActive:   cfg.MaxConn,
		IdleTimeout: cfg.IdleTimeout,
		Wait:        true,
	}

	go func() {
		for {
			c, err := pool.GetContext(ctx)
			if err != nil {
				log.Error("ConnectToRedisWithTimeout, %v,\n trying to connect", err.Error())
				time.Sleep(cfg.RetryInterval)
				continue
			}

			pongResponse, err := redis.String(c.Do("PING"))
			if err != nil || pongResponse != "PONG" {
				log.Error("ConnectToRedisWithTimeout, %v,\n trying to connect", err)
				time.Sleep(cfg.RetryInterval)
				continue
			}

			if ctx.Err() != nil {
				break
			}

			connWaitChan <- struct{}{}
			break
		}
	}()

	select {
	case <-ctx.Done():
		{
			log.Error("ConnectToRedisWithTimeout, connection timeout exceeded")
			return pool, nil, fmt.Errorf("db connection context timeout")
		}
	case <-connWaitChan:
		{
			log.Info("ConnectToRedisWithTimeout, connection established")
			return pool, func() {
				err = pool.Close()
				if err != nil {
					log.Info("ConnectToRedisWithTimeout, on pool close error,err: %s", err.Error())
				}
			}, nil
		}
	}
}
