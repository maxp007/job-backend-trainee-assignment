// +build integration

package cache

import (
	"context"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"job-backend-trainee-assignment/internal/logger"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"
)

type CacheTestCase struct {
	CaseName         string
	GivenKey         string
	ExpectedResult   bool
	ExpectedErr      error
	MethodToCallName string
	Timeout          time.Duration
}

//test intended to check redis cacheStruct + connector + redis instance behaviour
func TestCacheCommon_TestCacheIntegrationWithRedis(t *testing.T) {
	t.Log("TestCacheCommon_TestCacheIntegrationWithRedis")
	v := viper.New()
	v.AddConfigPath(".")
	v.AddConfigPath("../../")
	v.SetConfigName("config")
	v.AutomaticEnv()

	err := v.ReadInConfig()
	require.NoErrorf(t, err, "failed to read config file at: %s, err %v", "config", err)

	var cacheHost string
	if v.GetString("CACHE_HOST") != "" {
		cacheHost = v.GetString("CACHE_HOST")
	} else {
		cacheHost = v.GetString("cache_params.CACHE_HOST")
	}

	redisConnTimeout := v.GetDuration("cache_params.conn_timeout") * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), redisConnTimeout)
	defer cancel()

	cacheLogger := logger.NewLogger(os.Stdout, "RedisCache\t", logger.L_INFO)
	redisPool, poolCloseFunc, err := ConnectToRedisWithTimeout(ctx, cacheLogger, &ConnConfig{
		Host:          cacheHost,
		DBName:        v.GetInt("cache_params.db_name"),
		Port:          v.GetString("cache_params.port"),
		Pass:          v.GetString("cache_params.pass"),
		RetryInterval: v.GetDuration("cache_params.conn_retry_interval") * time.Second,
		MaxConn:       v.GetInt("cache_params.max_conn"),
		MaxIdleConn:   v.GetInt("cache_params.max_idle_conn"),
		IdleTimeout:   v.GetDuration("cache_params.idle_timeout") * time.Second,
	})
	require.NoErrorf(t, err, "must be able to connect to redis cache,err: %v", err)
	defer poolCloseFunc()

	cacheConfig := &CacheConfig{
		KeyExpirationTime: v.GetDuration("cache_params.key_expire_time") * time.Second,
		KeyLookupTimeout:  v.GetDuration("cache_params.cache_lookup_timeout") * time.Second,
		KeySetTimeout:     v.GetDuration("cache_params.cache_set_timeout") * time.Second,
	}

	redisCache, err := NewRedisCache(cacheLogger, redisPool, cacheConfig)
	require.NoErrorf(t, err, "must be able to create new redis cache instance, err: %v", err)

	caseTimeout := v.GetDuration("testing_params.test_case_timeout") * time.Second
	ctx, cancel = context.WithTimeout(context.Background(), caseTimeout)
	defer cancel()

	err = redisCache.AddKey(ctx, "SOMEKEY")
	require.NoErrorf(t, err, "must be able to add key to cache, err: %v", err)

	t.Run("check common behaviour", func(t *testing.T) {
		ctx, cancelCase := context.WithTimeout(context.Background(), caseTimeout)
		defer cancelCase()
		testCases := []CacheTestCase{
			{
				CaseName:         "positive path,test for key adding. Lookup",
				GivenKey:         "SomeKey",
				ExpectedResult:   false,
				MethodToCallName: "CheckKeyExistence",
				ExpectedErr:      nil,
				Timeout:          time.Second,
			},
			{
				CaseName:         "positive path,test for key adding. Set key",
				GivenKey:         "SomeKey",
				MethodToCallName: "AddKey",
				ExpectedErr:      nil,
				Timeout:          time.Second,
			},
			{
				CaseName:         "positive path,test for key adding. Lookup previously set key",
				GivenKey:         "SomeKey",
				ExpectedResult:   true,
				MethodToCallName: "CheckKeyExistence",
				ExpectedErr:      nil,
				Timeout:          time.Second,
			},
			{
				CaseName:         "negative path,test for key adding with early timeout",
				GivenKey:         "SomeOtherKey",
				MethodToCallName: "AddKey",
				ExpectedErr:      ErrContextDeadlineExceeded,
				Timeout:          time.Nanosecond,
			},
			{
				CaseName:         "negative path, test for key lookup with early timeout",
				GivenKey:         "SomeOtherKey",
				ExpectedResult:   false,
				MethodToCallName: "CheckKeyExistence",
				ExpectedErr:      ErrContextDeadlineExceeded,
				Timeout:          time.Nanosecond,
			},
		}

		for caseIdx, testCase := range testCases {
			t.Logf("testing case [%d] %s", caseIdx, testCase.CaseName)
			ctx, cancelTestCase := context.WithTimeout(ctx, testCase.Timeout)

			if testCase.MethodToCallName == "AddKey" {
				err := redisCache.AddKey(ctx, testCase.GivenKey)
				assert.ErrorIsf(t, err, testCase.ExpectedErr, "CheckKeyExistence returned error must be equal to expected, got %v", err)

			} else if testCase.MethodToCallName == "CheckKeyExistence" {
				isKeyExist, err := redisCache.CheckKeyExistence(ctx, testCase.GivenKey)
				assert.ErrorIsf(t, err, testCase.ExpectedErr, "CheckKeyExistence returned error must be equal to expected, got %v", err)
				assert.Equal(t, testCase.ExpectedResult, isKeyExist, "expected result and given must match")
			} else {
				require.Fail(t, "test Case MethodToCall Field must contain one of redisCache methods names")
			}
			cancelTestCase()
		}
	})
}

//test intended to check concurrent redis cacheStruct + connector + redis instance behaviour
func TestCacheConcurrent_TestCacheIntegrationWithRedis(t *testing.T) {
	t.Log("TestCacheConcurrent_TestCacheIntegrationWithRedis")
	v := viper.New()
	v.AddConfigPath(".")
	v.AddConfigPath("../../")
	v.SetConfigName("config")
	v.AutomaticEnv()

	err := v.ReadInConfig()
	require.NoErrorf(t, err, "failed to read config file at: %s, err %v", "config", err)

	var cacheHost string
	if v.GetString("CACHE_HOST") != "" {
		cacheHost = v.GetString("CACHE_HOST")
	} else {
		cacheHost = v.GetString("cache_params.CACHE_HOST")
	}

	redisConnTimeout := v.GetDuration("cache_params.conn_timeout") * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), redisConnTimeout)
	defer cancel()

	cacheLogger := logger.NewLogger(os.Stdout, "RedisLogger\t", logger.L_INFO)
	redisPool, poolCloseFunc, err := ConnectToRedisWithTimeout(ctx, cacheLogger, &ConnConfig{
		Host:          cacheHost,
		DBName:        v.GetInt("cache_params.db_name"),
		Port:          v.GetString("cache_params.port"),
		Pass:          v.GetString("cache_params.pass"),
		RetryInterval: v.GetDuration("cache_params.conn_retry_interval") * time.Second,
		MaxConn:       v.GetInt("cache_params.max_conn"),
		MaxIdleConn:   v.GetInt("cache_params.max_idle_conn"),
		IdleTimeout:   v.GetDuration("cache_params.idle_timeout") * time.Second,
	})
	require.NoErrorf(t, err, "must be able to connect to redis cache,err: %v", err)

	defer poolCloseFunc()

	cacheConfig := &CacheConfig{
		KeyExpirationTime: v.GetDuration("cache_params.key_expire_time") * time.Second,
		KeyLookupTimeout:  v.GetDuration("cache_params.cache_lookup_timeout") * time.Second,
		KeySetTimeout:     v.GetDuration("cache_params.cache_set_timeout") * time.Second,
	}

	redisCache, err := NewRedisCache(cacheLogger, redisPool, cacheConfig)
	require.NoErrorf(t, err, "must be able to create new redis cache instance, err: %v", err)

	caseTimeout := v.GetDuration("testing_params.test_case_timeout") * time.Second

	t.Run("check concurrent key adding, then concurrently lookup", func(t *testing.T) {
		wg := &sync.WaitGroup{}
		maxWorkers := 64
		workersSyncChan := make(chan struct{}, maxWorkers)
		mu := sync.Mutex{}
		iterationsToPerform := 400

		ctx, cancelCase := context.WithTimeout(context.Background(), 10*caseTimeout)
		defer cancelCase()

		for i := 0; i < iterationsToPerform; i++ {
			wg.Add(1)
			go func(iCopy int) {
				defer func() {
					mu.Unlock()
					<-workersSyncChan
					wg.Done()
				}()
				workersSyncChan <- struct{}{}
				mu.Lock()
				err := redisCache.AddKey(ctx, strconv.Itoa(iCopy))
				require.NoError(t, err, "Must be able to concurrently add key to redis cache")
			}(i)
		}
		wg.Wait()
		for i := 0; i < iterationsToPerform; i++ {
			wg.Add(1)

			go func(iCopy int) {
				defer func() {
					mu.Unlock()
					<-workersSyncChan
					wg.Done()
				}()
				workersSyncChan <- struct{}{}
				mu.Lock()
				keyVal, err := redisCache.CheckKeyExistence(ctx, strconv.Itoa(iCopy))
				require.Equal(t, true, keyVal, "Specified key must be found redis cache")
				require.NoError(t, err, "Must be able to concurrently add key to redis cache")
			}(i)
		}
		wg.Wait()
	})
}
