// +build integration

package app

import (
	"context"
	"errors"
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"job-backend-trainee-assignment/internal/cache"
	"job-backend-trainee-assignment/internal/db_connector"
	"job-backend-trainee-assignment/internal/exchanger"
	"job-backend-trainee-assignment/internal/logger"
	"job-backend-trainee-assignment/internal/test_helpers"
	"os"
	"sync"
	"testing"
	"time"
)

const filePathPrefix = "../../"
const TimeoutTimeMultiplier = 50

func TestBillingApp_TestBalanceDataPersistence(t *testing.T) {
	t.Log("TestBillingApp_TestBalanceDataPersistence")
	v := viper.New()
	v.AddConfigPath(".")
	v.AddConfigPath("../../")
	v.SetConfigName("config")
	v.AutomaticEnv()

	err := v.ReadInConfig()
	require.NoErrorf(t, err, "failed to read config file at: %s, err %v", "config", err)

	var pgHost string
	if v.GetString("DATABASE_HOST") != "" {
		pgHost = v.GetString("DATABASE_HOST")
	} else {
		pgHost = v.GetString("db_params.DATABASE_HOST")
	}

	dbConfig := &db_connector.Config{
		DriverName:    v.GetString("db_params.driver_name"),
		DBUser:        v.GetString("db_params.user"),
		DBPass:        v.GetString("db_params.password"),
		DBName:        v.GetString("db_params.db_name"),
		DBPort:        v.GetString("db_params.port"),
		DBHost:        pgHost,
		SSLMode:       v.GetString("db_params.ssl_mode"),
		RetryInterval: v.GetDuration("db_params.conn_retry_interval") * time.Second,
	}

	dbConnTimeout := v.GetDuration("db_params.conn_timeout") * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), dbConnTimeout)
	defer cancel()
	dummyLogger := &logger.DummyLogger{}

	db, dbCloseFunc, err := db_connector.DBConnectWithTimeout(ctx, dbConfig, dummyLogger)
	require.NoErrorf(t, err, "failed to connect to db,err %v", err)

	defer dbCloseFunc()

	ex := &exchanger.StubExchanger{}
	appLogger := logger.NewLogger(os.Stdout, "APP", logger.L_ERROR)

	var cacheHost string
	if v.GetString("CACHE_HOST") != "" {
		cacheHost = v.GetString("CACHE_HOST")
	} else {
		cacheHost = v.GetString("cache_params.CACHE_HOST")
	}

	redisConnTimeout := v.GetDuration("cache_params.conn_timeout") * time.Second
	ctx, cancel = context.WithTimeout(context.Background(), redisConnTimeout)
	defer cancel()
	redisPool, poolCloseFunc, err := cache.ConnectToRedisWithTimeout(ctx, logger.NewLogger(os.Stdout, "RedisConn\t", logger.L_INFO), &cache.ConnConfig{
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

	cacheConfig := &cache.CacheConfig{
		KeyExpirationTime: v.GetDuration("cache_params.key_expire_time") * time.Second,
		KeyLookupTimeout:  v.GetDuration("cache_params.cache_lookup_timeout") * time.Second,
		KeySetTimeout:     v.GetDuration("cache_params.cache_set_timeout") * time.Second,
	}
	redisCache, err := cache.NewRedisCache(dummyLogger, redisPool, cacheConfig)
	require.NoErrorf(t, err, "must be able to create new redis cache instance, err: %v", err)

	app, err := NewApp(appLogger, db, ex, redisCache, nil)
	require.NoErrorf(t, err, "failed to create BillingApp instance, err %v", err)

	caseTimeout := v.GetDuration("testing_params.test_case_timeout") * time.Second

	userId := int64(20)
	receiverUserId := int64(1)
	userName := "D. Jones"
	purpose := "some operation purpose"
	operationsToPerform := 1000
	amountPerOperation := decimal.NewFromInt(1)

	// Test intended to check single user crediting
	// check for lost updates
	// check for concurrent insertion of new user
	t.Run("check concurrent crediting", func(t *testing.T) {
		t.Log("check concurrent crediting")
		wg := &sync.WaitGroup{}
		maxWorkers := 64
		workersSyncChan := make(chan struct{}, maxWorkers)
		mu := sync.Mutex{}
		ctx, cancel := context.WithTimeout(context.Background(), TimeoutTimeMultiplier*caseTimeout)
		defer cancel()
		err = test_helpers.PrepareDB(ctx, db, test_helpers.Config{
			InitFilePath:    filePathPrefix + v.GetString("testing_params.db_init_file_path"),
			CleanUpFilePath: filePathPrefix + v.GetString("testing_params.db_cleanup_file_path"),
		})
		require.NoError(t, err, "PrepareDB must not return error")

		for i := 0; i < operationsToPerform; i++ {
			wg.Add(1)

			go func() {
				defer func() {
					mu.Unlock()
					<-workersSyncChan
					wg.Done()
				}()
				workersSyncChan <- struct{}{}
				_, err := app.CreditUserAccount(ctx, &CreditAccountRequest{
					UserId:           userId,
					Name:             userName,
					Purpose:          purpose,
					Amount:           amountPerOperation.String(),
					IdempotencyToken: uuid.NewV4().String(),
				})
				mu.Lock()
				require.NoError(t, err, "must be able to Credit user account")
			}()
		}
		wg.Wait()
		userBalance, err := app.GetUserBalance(ctx, &BalanceRequest{
			UserId: userId,
		})
		require.NoError(t, err, "must be able to get user balance")

		expectedBalance := amountPerOperation.Mul(decimal.NewFromInt(int64(operationsToPerform)))
		resultBalance, err := decimal.NewFromString(userBalance.Balance)
		require.NoError(t, err, "must be able to parse balance to decimal")

		assert.Equal(t, expectedBalance.String(), resultBalance.String(), "balance must be equal to expected")

	})

	//test checks concurrent balance withdrawal for single user
	// check for lost updates
	// check for non-negative remaining balance
	t.Run("check concurrent withdrawal", func(t *testing.T) {
		t.Log("check concurrent withdrawal")
		wg := &sync.WaitGroup{}
		maxWorkers := 64
		workersSyncChan := make(chan struct{}, maxWorkers)
		mu := sync.Mutex{}
		ctx, cancel := context.WithTimeout(context.Background(), TimeoutTimeMultiplier*caseTimeout)
		defer cancel()
		err = test_helpers.PrepareDB(ctx, db, test_helpers.Config{
			InitFilePath:    filePathPrefix + v.GetString("testing_params.db_init_file_path"),
			CleanUpFilePath: filePathPrefix + v.GetString("testing_params.db_cleanup_file_path"),
		})
		require.NoError(t, err, "PrepareDB must not return error")

		_, err := app.CreditUserAccount(ctx, &CreditAccountRequest{
			UserId:           userId,
			Name:             userName,
			Purpose:          purpose,
			Amount:           amountPerOperation.Mul(decimal.NewFromInt(int64(operationsToPerform))).String(),
			IdempotencyToken: uuid.NewV4().String(),
		})
		require.NoError(t, err, "must be able to Credit user account")

		//perform  additional withdraw operations after balance became 0
		oddOperations := 30
		for i := 0; i < operationsToPerform+oddOperations; i++ {
			wg.Add(1)

			go func() {
				defer func() {
					mu.Unlock()
					<-workersSyncChan
					wg.Done()
				}()
				workersSyncChan <- struct{}{}
				_, err := app.WithdrawUserAccount(ctx, &WithdrawAccountRequest{
					UserId:           userId,
					Purpose:          purpose,
					Amount:           amountPerOperation.String(),
					IdempotencyToken: uuid.NewV4().String(),
				})
				mu.Lock()
				if !errors.Is(err, ErrUserDoesNotHaveEnoughMoney) {
					require.NoError(t, err, "must be able to Withdraw user account")
				}

			}()
		}
		wg.Wait()
		userBalance, err := app.GetUserBalance(ctx, &BalanceRequest{
			UserId: userId,
		})
		require.NoError(t, err, "must be able to get user balance")

		expectedBalance := decimal.NewFromInt(0)
		resultBalance, err := decimal.NewFromString(userBalance.Balance)
		require.NoError(t, err, "must be able to parse balance to decimal")

		assert.Equal(t, expectedBalance.String(), resultBalance.String(), "balance must be equal to expected")
	})

	// test checks concurrent money transfer from one user to another
	// check for lost updates
	// check for non-negative remaining balance
	t.Run("check concurrent money transfer from one user to another", func(t *testing.T) {
		t.Log("check concurrent money transfer from one user to another")
		wg := &sync.WaitGroup{}
		maxWorkers := 64
		workersSyncChan := make(chan struct{}, maxWorkers)
		mu := sync.Mutex{}
		ctx, cancel := context.WithTimeout(context.Background(), TimeoutTimeMultiplier*caseTimeout)
		defer cancel()

		err = test_helpers.PrepareDB(ctx, db, test_helpers.Config{
			InitFilePath:    filePathPrefix + v.GetString("testing_params.db_init_file_path"),
			CleanUpFilePath: filePathPrefix + v.GetString("testing_params.db_cleanup_file_path"),
		})
		require.NoError(t, err, "PrepareDB must not return error")

		//load certain amount of money to "sender" User
		_, err := app.CreditUserAccount(ctx, &CreditAccountRequest{
			UserId:           userId,
			Name:             userName,
			Purpose:          purpose,
			Amount:           amountPerOperation.Mul(decimal.NewFromInt(int64(operationsToPerform))).String(),
			IdempotencyToken: uuid.NewV4().String(),
		})
		require.NoError(t, err, "must be able to Credit user account")

		//perform  additional transfer operations after balance became 0
		oddOperations := 30
		for i := 0; i < operationsToPerform+oddOperations; i++ {
			wg.Add(1)
			go func() {
				defer func() {
					mu.Unlock()
					<-workersSyncChan
					wg.Done()
				}()
				workersSyncChan <- struct{}{}
				_, err := app.TransferMoneyFromUserToUser(ctx, &MoneyTransferRequest{
					SenderId:         userId,
					ReceiverId:       int64(receiverUserId),
					Amount:           amountPerOperation.String(),
					IdempotencyToken: uuid.NewV4().String(),
				})
				mu.Lock()
				if !errors.Is(err, ErrUserDoesNotHaveEnoughMoney) {
					require.NoError(t, err, "must be able to Withdraw user account")
				}

			}()
		}
		wg.Wait()

		userSenderBalance, err := app.GetUserBalance(ctx, &BalanceRequest{
			UserId: userId,
		})
		require.NoError(t, err, "must be able to get user balance")
		expectedSenderBalance := decimal.NewFromInt(0)

		userReceiverBalance, err := app.GetUserBalance(ctx, &BalanceRequest{
			UserId: receiverUserId,
		})
		require.NoError(t, err, "must be able to get user balance")
		expectedReceiverBalance := amountPerOperation.Mul(decimal.NewFromInt(int64(operationsToPerform)))

		assert.Equal(t, expectedReceiverBalance.String(), userReceiverBalance.Balance, "receiver balance must be equal to expected")
		assert.Equal(t, expectedSenderBalance.String(), userSenderBalance.Balance, "sender balance must be equal to expected")
	})

	//test checks concurrent simultaneous balance withdrawal and crediting for single user
	t.Run("check concurrent withdrawal and crediting", func(t *testing.T) {
		t.Log("check concurrent withdrawal and crediting")
		wg := &sync.WaitGroup{}
		maxWorkers := 64
		workersSyncChan := make(chan struct{}, maxWorkers)
		mu := sync.Mutex{}
		ctx, cancel := context.WithTimeout(context.Background(), TimeoutTimeMultiplier*caseTimeout)
		defer cancel()
		err = test_helpers.PrepareDB(ctx, db, test_helpers.Config{
			InitFilePath:    filePathPrefix + v.GetString("testing_params.db_init_file_path"),
			CleanUpFilePath: filePathPrefix + v.GetString("testing_params.db_cleanup_file_path"),
		})
		require.NoError(t, err, "PrepareDB must not return error")

		_, err := app.CreditUserAccount(ctx, &CreditAccountRequest{
			UserId:           userId,
			Name:             userName,
			Purpose:          purpose,
			Amount:           amountPerOperation.Mul(decimal.NewFromInt(int64(operationsToPerform))).String(),
			IdempotencyToken: uuid.NewV4().String(),
		})

		require.NoError(t, err, "must be able to Credit user account")

		for i := 0; i < operationsToPerform/2; i++ {

			wg.Add(2)

			go func() {
				defer func() {
					mu.Unlock()
					<-workersSyncChan
					wg.Done()
				}()
				workersSyncChan <- struct{}{}
				_, err := app.WithdrawUserAccount(ctx, &WithdrawAccountRequest{
					UserId:           userId,
					Purpose:          purpose,
					Amount:           amountPerOperation.String(),
					IdempotencyToken: uuid.NewV4().String(),
				})
				mu.Lock()
				if !errors.Is(err, ErrUserDoesNotHaveEnoughMoney) {
					require.NoError(t, err, "must be able to Withdraw user account")
				}
			}()

			go func() {
				defer func() {
					mu.Unlock()
					<-workersSyncChan
					wg.Done()
				}()
				workersSyncChan <- struct{}{}
				_, err := app.CreditUserAccount(ctx, &CreditAccountRequest{
					UserId:           userId,
					Purpose:          purpose,
					Amount:           amountPerOperation.String(),
					IdempotencyToken: uuid.NewV4().String(),
				})
				mu.Lock()
				require.NoError(t, err, "must be able to Credit user account")
			}()
		}
		wg.Wait()
		userBalance, err := app.GetUserBalance(ctx, &BalanceRequest{
			UserId: userId,
		})
		require.NoError(t, err, "must be able to get user balance")

		expectedBalance := amountPerOperation.Mul(decimal.NewFromInt(int64(operationsToPerform)))
		resultBalance, err := decimal.NewFromString(userBalance.Balance)
		require.NoError(t, err, "must be able to parse balance to decimal")

		assert.Equal(t, expectedBalance.String(), resultBalance.String(), "balance must be equal to expected")
	})

	//test checks concurrent simultaneous user crediting and transfer to another user
	t.Run("check concurrent user crediting and transfer to another user", func(t *testing.T) {
		t.Log("check concurrent user crediting and transfer to another user")
		wg := &sync.WaitGroup{}
		maxWorkers := 64
		workersSyncChan := make(chan struct{}, maxWorkers)
		mu := sync.Mutex{}

		ctx, cancel := context.WithTimeout(context.Background(), TimeoutTimeMultiplier*caseTimeout)
		defer cancel()
		err = test_helpers.PrepareDB(ctx, db, test_helpers.Config{
			InitFilePath:    filePathPrefix + v.GetString("testing_params.db_init_file_path"),
			CleanUpFilePath: filePathPrefix + v.GetString("testing_params.db_cleanup_file_path"),
		})
		require.NoError(t, err, "PrepareDB must not return error")

		_, err := app.CreditUserAccount(ctx, &CreditAccountRequest{
			UserId:           userId,
			Name:             userName,
			Purpose:          purpose,
			Amount:           amountPerOperation.Mul(decimal.NewFromInt(int64(operationsToPerform))).String(),
			IdempotencyToken: uuid.NewV4().String(),
		})

		require.NoError(t, err, "must be able to Credit user account")

		for i := 0; i < operationsToPerform/2; i++ {
			wg.Add(2)

			go func() {
				defer func() {
					mu.Unlock()
					<-workersSyncChan
					wg.Done()
				}()
				workersSyncChan <- struct{}{}
				_, err := app.CreditUserAccount(ctx, &CreditAccountRequest{
					UserId:           userId,
					Purpose:          purpose,
					Amount:           amountPerOperation.String(),
					IdempotencyToken: uuid.NewV4().String(),
				})
				mu.Lock()
				require.NoError(t, err, "must be able to Credit user account")
			}()

			go func() {
				defer func() {
					mu.Unlock()
					<-workersSyncChan
					wg.Done()
				}()
				workersSyncChan <- struct{}{}
				_, err := app.TransferMoneyFromUserToUser(ctx, &MoneyTransferRequest{
					ReceiverId:       receiverUserId,
					SenderId:         userId,
					Amount:           amountPerOperation.String(),
					IdempotencyToken: uuid.NewV4().String(),
				})
				mu.Lock()
				require.NoError(t, err, "must be able to Transfer user money to another user")
			}()
		}
		wg.Wait()

		senderUserBalance, err := app.GetUserBalance(ctx, &BalanceRequest{
			UserId: userId,
		})
		require.NoError(t, err, "must be able to get sender user balance")

		receiverUserBalance, err := app.GetUserBalance(ctx, &BalanceRequest{
			UserId: receiverUserId,
		})
		require.NoError(t, err, "must be able to get receiver user balance")

		expectedReceiverBalance := amountPerOperation.Mul(decimal.NewFromInt(int64(operationsToPerform / 2)))
		resultReceiverBalance, err := decimal.NewFromString(receiverUserBalance.Balance)
		require.NoError(t, err, "must be able to parse balance to decimal")

		assert.Equal(t, expectedReceiverBalance.String(), resultReceiverBalance.String(), "balance must be equal to expected")

		expectedSenderBalance := amountPerOperation.Mul(decimal.NewFromInt(int64(operationsToPerform)))
		resultSenderBalance, err := decimal.NewFromString(senderUserBalance.Balance)
		require.NoError(t, err, "must be able to parse balance to decimal")

		assert.Equal(t, expectedSenderBalance.String(), resultSenderBalance.String(), "balance must be equal to expected")
	})

	//test checks concurrent simultaneous user to user "ping-pong" transfers
	// check deadlocks
	t.Run("check concurrent simultaneous user to user \"ping-pong\" transfers", func(t *testing.T) {
		t.Log("check concurrent simultaneous user to user \"ping-pong\" transfers")
		wg := &sync.WaitGroup{}
		maxWorkers := 64
		mu := sync.Mutex{}
		workersSyncChan := make(chan struct{}, maxWorkers)

		ctx, cancel := context.WithTimeout(context.Background(), TimeoutTimeMultiplier*caseTimeout)
		defer cancel()
		err = test_helpers.PrepareDB(ctx, db, test_helpers.Config{
			InitFilePath:    filePathPrefix + v.GetString("testing_params.db_init_file_path"),
			CleanUpFilePath: filePathPrefix + v.GetString("testing_params.db_cleanup_file_path"),
		})
		require.NoError(t, err, "PrepareDB must not return error")

		user2Id := receiverUserId
		user1Id := userId
		_, err := app.CreditUserAccount(ctx, &CreditAccountRequest{
			UserId:           user1Id,
			Name:             userName,
			Purpose:          purpose,
			Amount:           amountPerOperation.Mul(decimal.NewFromInt(int64(operationsToPerform))).String(),
			IdempotencyToken: uuid.NewV4().String(),
		})
		require.NoError(t, err, "must be able to Credit user1 account")

		_, err = app.CreditUserAccount(ctx, &CreditAccountRequest{
			UserId:           user2Id,
			Name:             userName,
			Purpose:          purpose,
			Amount:           amountPerOperation.Mul(decimal.NewFromInt(int64(operationsToPerform))).String(),
			IdempotencyToken: uuid.NewV4().String(),
		})
		require.NoError(t, err, "must be able to Credit user2 account")

		for i := 0; i < operationsToPerform/2; i++ {
			wg.Add(2)

			go func() {
				defer func() {
					mu.Unlock()
					<-workersSyncChan
					wg.Done()
				}()
				workersSyncChan <- struct{}{}
				_, err := app.TransferMoneyFromUserToUser(ctx, &MoneyTransferRequest{
					ReceiverId:       user1Id,
					SenderId:         user2Id,
					Amount:           amountPerOperation.String(),
					IdempotencyToken: uuid.NewV4().String(),
				})
				mu.Lock()
				require.NoError(t, err, "must be able to Transfer money from u1 to u2")

			}()

			go func() {
				defer func() {
					mu.Unlock()
					<-workersSyncChan
					wg.Done()
				}()
				workersSyncChan <- struct{}{}
				_, err := app.TransferMoneyFromUserToUser(ctx, &MoneyTransferRequest{
					ReceiverId:       user2Id,
					SenderId:         user1Id,
					Amount:           amountPerOperation.String(),
					IdempotencyToken: uuid.NewV4().String(),
				})
				mu.Lock()
				require.NoError(t, err, "must be able to Transfer money from u2 to u1")
			}()
		}
		wg.Wait()

		user1Balance, err := app.GetUserBalance(ctx, &BalanceRequest{
			UserId: user1Id,
		})
		require.NoError(t, err, "must be able to get sender user1 balance")

		expUser1Balance := amountPerOperation.Mul(decimal.NewFromInt(int64(operationsToPerform)))
		resultUser1Balance, err := decimal.NewFromString(user1Balance.Balance)
		require.NoError(t, err, "must be able to parse balance to decimal")
		assert.Equal(t, expUser1Balance.String(), resultUser1Balance.String(), "balance must be equal to expected")

		user2Balance, err := app.GetUserBalance(ctx, &BalanceRequest{
			UserId: user2Id,
		})
		require.NoError(t, err, "must be able to get receiver user2 balance")

		expUser2Balance := amountPerOperation.Mul(decimal.NewFromInt(int64(operationsToPerform)))
		resultUser2Balance, err := decimal.NewFromString(user2Balance.Balance)
		require.NoError(t, err, "must be able to parse balance to decimal")

		assert.Equal(t, expUser2Balance.String(), resultUser2Balance.String(), "balance must be equal to expected")
	})

	t.Run("check concurrent operations for checking request idempotency", func(t *testing.T) {

		// Try to Credit new user multiple times with single token
		// Check for concurrent access to db
		//  expect insertion of exactly one operation to Operation table
		//	expect crediting user balance only once
		t.Run("check concurrent crediting operations with the same Idempotency Token", func(t *testing.T) {

			wg := &sync.WaitGroup{}
			maxWorkers := 64
			workersSyncChan := make(chan struct{}, maxWorkers)
			mu := sync.Mutex{}
			ctx, cancel := context.WithTimeout(context.Background(), TimeoutTimeMultiplier*caseTimeout)
			defer cancel()
			err = test_helpers.PrepareDB(ctx, db, test_helpers.Config{
				InitFilePath:    filePathPrefix + v.GetString("testing_params.db_init_file_path"),
				CleanUpFilePath: filePathPrefix + v.GetString("testing_params.db_cleanup_file_path"),
			})
			require.NoError(t, err, "PrepareDB must not return error")

			singleUseToken := "TOKEN_1"

			for i := 0; i < operationsToPerform; i++ {
				wg.Add(1)
				go func() {
					defer func() {
						mu.Unlock()
						<-workersSyncChan
						wg.Done()
					}()
					workersSyncChan <- struct{}{}
					_, err := app.CreditUserAccount(ctx, &CreditAccountRequest{
						UserId:           userId,
						Name:             userName,
						Purpose:          purpose,
						Amount:           amountPerOperation.String(),
						IdempotencyToken: singleUseToken,
					})
					mu.Lock()
					require.NoError(t, err, "must be able to Credit user account")
				}()
			}
			wg.Wait()
			userBalanceAfter, err := app.GetUserBalance(ctx, &BalanceRequest{
				UserId: userId,
			})
			require.NoError(t, err, "must be able to get user balance")

			resultBalanceAfter, err := decimal.NewFromString(userBalanceAfter.Balance)
			require.NoError(t, err, "must be able to parse balance to decimal")

			assert.Equal(t, decimal.NewFromInt(1).String(), resultBalanceAfter.String(), "balance must be equal to expected")
		})

		// Try to Withdraw new user multiple times with single token
		// Check for concurrent access to db
		//	expect insertion of exactly one operation to Operation table
		//	expect withdrawing user balance only once
		t.Run("check concurrent withdraw operations with the same Idempotency Token", func(t *testing.T) {
			wg := &sync.WaitGroup{}
			maxWorkers := 64
			workersSyncChan := make(chan struct{}, maxWorkers)
			mu := sync.Mutex{}
			ctx, cancel := context.WithTimeout(context.Background(), TimeoutTimeMultiplier*caseTimeout)
			defer cancel()
			err = test_helpers.PrepareDB(ctx, db, test_helpers.Config{
				InitFilePath:    filePathPrefix + v.GetString("testing_params.db_init_file_path"),
				CleanUpFilePath: filePathPrefix + v.GetString("testing_params.db_cleanup_file_path"),
			})
			require.NoError(t, err, "PrepareDB must not return error")

			_, err := app.CreditUserAccount(ctx, &CreditAccountRequest{
				UserId:           userId,
				Name:             userName,
				Purpose:          purpose,
				Amount:           amountPerOperation.Mul(decimal.NewFromInt(int64(operationsToPerform))).String(),
				IdempotencyToken: uuid.NewV4().String(),
			})

			userBalanceBefore, err := app.GetUserBalance(ctx, &BalanceRequest{
				UserId: userId,
			})
			require.NoError(t, err, "must be able to get user balance")

			singleUseToken := "TOKEN_2"

			for i := 0; i < operationsToPerform; i++ {
				wg.Add(1)
				go func() {
					defer func() {
						mu.Unlock()
						<-workersSyncChan
						wg.Done()
					}()
					workersSyncChan <- struct{}{}
					_, err := app.WithdrawUserAccount(ctx, &WithdrawAccountRequest{
						UserId:           userId,
						Purpose:          purpose,
						Amount:           amountPerOperation.String(),
						IdempotencyToken: singleUseToken,
					})
					mu.Lock()
					require.NoError(t, err, "must be able to Withdraw user account")
				}()
			}
			wg.Wait()
			userBalanceAfter, err := app.GetUserBalance(ctx, &BalanceRequest{
				UserId: userId,
			})
			require.NoError(t, err, "must be able to get user balance")

			resultBalanceAfter, err := decimal.NewFromString(userBalanceAfter.Balance)
			require.NoError(t, err, "must be able to parse balance to decimal")

			resultBalanceBefore, err := decimal.NewFromString(userBalanceBefore.Balance)
			require.NoError(t, err, "must be able to parse balance to decimal")

			assert.Equal(t, resultBalanceBefore.Sub(amountPerOperation).String(), resultBalanceAfter.String(), "balance must be equal to expected")
		})

		t.Run("check concurrent transfer operations with the same Idempotency Token", func(t *testing.T) {
			wg := &sync.WaitGroup{}
			maxWorkers := 64
			workersSyncChan := make(chan struct{}, maxWorkers)
			mu := sync.Mutex{}
			ctx, cancel := context.WithTimeout(context.Background(), TimeoutTimeMultiplier*caseTimeout)
			defer cancel()

			err = test_helpers.PrepareDB(ctx, db, test_helpers.Config{
				InitFilePath:    filePathPrefix + v.GetString("testing_params.db_init_file_path"),
				CleanUpFilePath: filePathPrefix + v.GetString("testing_params.db_cleanup_file_path"),
			})
			require.NoError(t, err, "PrepareDB must not return error")

			//load certain amount of money to "sender" User
			_, err := app.CreditUserAccount(ctx, &CreditAccountRequest{
				UserId:           userId,
				Name:             userName,
				Purpose:          purpose,
				Amount:           amountPerOperation.Mul(decimal.NewFromInt(int64(operationsToPerform))).String(),
				IdempotencyToken: uuid.NewV4().String(),
			})
			require.NoError(t, err, "must be able to Credit user account")

			singleUseToken := "TOKEN_3"

			for i := 0; i < operationsToPerform; i++ {
				wg.Add(1)
				go func() {
					defer func() {
						mu.Unlock()
						<-workersSyncChan
						wg.Done()
					}()
					workersSyncChan <- struct{}{}
					_, err := app.TransferMoneyFromUserToUser(ctx, &MoneyTransferRequest{
						SenderId:         userId,
						ReceiverId:       receiverUserId,
						Amount:           amountPerOperation.String(),
						IdempotencyToken: singleUseToken,
					})
					mu.Lock()
					if !errors.Is(err, ErrUserDoesNotHaveEnoughMoney) {
						require.NoError(t, err, "must be able to Withdraw user account")
					}

				}()
			}
			wg.Wait()

			userSenderBalance, err := app.GetUserBalance(ctx, &BalanceRequest{
				UserId: userId,
			})
			require.NoError(t, err, "must be able to get user balance")
			expectedSenderBalance := amountPerOperation.Mul(decimal.NewFromInt(int64(operationsToPerform))).Sub(amountPerOperation)

			userReceiverBalance, err := app.GetUserBalance(ctx, &BalanceRequest{
				UserId: receiverUserId,
			})
			require.NoError(t, err, "must be able to get user balance")

			assert.Equal(t, amountPerOperation.String(), userReceiverBalance.Balance, "receiver balance must be equal to expected")
			assert.Equal(t, expectedSenderBalance.String(), userSenderBalance.Balance, "sender balance must be equal to expected")
		})
	})
}
