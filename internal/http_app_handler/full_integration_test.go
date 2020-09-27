// +build integration

package http_app_handler

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"job-backend-trainee-assignment/internal/app"
	"job-backend-trainee-assignment/internal/cache"
	"job-backend-trainee-assignment/internal/db_connector"
	"job-backend-trainee-assignment/internal/exchanger"
	"job-backend-trainee-assignment/internal/http_handler_router"
	"job-backend-trainee-assignment/internal/logger"
	"job-backend-trainee-assignment/internal/test_helpers"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAppHttpHandler_WithAppIntegration_WithStubExchanger(t *testing.T) {
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

	reqDoer := &http.Client{
		Timeout: v.GetDuration("app_params.exchange_timeout") * time.Second,
	}
	ex, err := exchanger.NewExchanger(dummyLogger, reqDoer, v.GetString("app_params.base_currency_code"))
	require.NoErrorf(t, err, "failed to create NewExchanger instance %v", err)

	var cacheHost string
	if v.GetString("CACHE_HOST") != "" {
		cacheHost = v.GetString("CACHE_HOST")
	} else {
		cacheHost = v.GetString("cache_params.CACHE_HOST")
	}

	redisConnTimeout := v.GetDuration("cache_params.conn_timeout") * time.Second
	ctx, cancel = context.WithTimeout(context.Background(), redisConnTimeout)
	defer cancel()

	redisPool, poolCloseFunc, err := cache.ConnectToRedisWithTimeout(ctx, dummyLogger, &cache.ConnConfig{
		Host:          cacheHost,
		Port:          v.GetString("cache_params.port"),
		DBName:        v.GetInt("cache_params.db_name"),
		Pass:          v.GetString("cache_params.pass"),
		RetryInterval: v.GetDuration("cache_params.conn_retry_interval") * time.Second,
		MaxConn:       v.GetInt("cache_params.max_conn"),
		MaxIdleConn:   v.GetInt("cache_params.max_idle_conn"),
		IdleTimeout:   v.GetDuration("cache_params.idle_timeout") * time.Second,
	})
	require.NoError(t, err, "Must be able to connect to redis with timeout")
	defer poolCloseFunc()

	cacheConfig := &cache.CacheConfig{KeyExpirationTime: v.GetDuration("cache_params.key_expire_time") * time.Second}
	redisCache, err := cache.NewRedisCache(dummyLogger, redisPool, cacheConfig)
	require.NoError(t, err, "Must be able to create redis cache instance")

	commonApp, err := app.NewApp(dummyLogger, db, ex, redisCache, nil)
	require.NoErrorf(t, err, "failed to create NewApp instance %v", err)

	r, err := router.NewRouter(dummyLogger)
	require.NoErrorf(t, err, "NewRouter must not return error, err %v", err)

	appHandler, err := NewHttpAppHandler(dummyLogger, r, commonApp, &Config{RequestHandleTimeout: v.GetDuration("app_params.request_handle_timeout") * time.Second})
	require.NoErrorf(t, err, "failed to create NewHttpAppHandlers instance %v", err)

	//Testing involves check "HttpHandler" + "app" + "database" + "STUB exchanger"
	//
	t.Run("test existent user credit-withdraw operations", func(t *testing.T) {
		alreadyUsedIdempotencyToken := "1"
		testCases := []TestCaseWithPath{
			{
				CaseName:       "positive path, existent user, GetUserBalance",
				Path:           pathMethodGetUserBalance,
				ReqMethod:      http.MethodPost,
				ReqContentType: contentTypeApplicationJson,
				ReqBody: &app.BalanceRequest{
					UserId: 1,
				},
				RespStatus: http.StatusOK,
				RespBody: &SuccessResponseBody{
					Result: &app.UserBalance{
						Balance:  "0",
						Currency: "RUB",
					},
				},
			},
			{
				CaseName:       "positive path, existent user, CreditUserAccount",
				Path:           pathMethodCreditAccount,
				ReqMethod:      http.MethodPost,
				ReqContentType: contentTypeApplicationJson,
				ReqBody: &app.CreditAccountRequest{
					UserId:           1,
					Purpose:          "payment from service",
					Amount:           "15",
					IdempotencyToken: uuid.NewV4().String(),
				},
				RespStatus: http.StatusOK,
				RespBody:   &SuccessResponseBody{Result: &app.ResultState{State: app.MsgAccountCreditingDone}},
			},
			{
				CaseName:       "positive path, existent user, GetUserBalance after CreditUserAccount",
				Path:           pathMethodGetUserBalance,
				ReqMethod:      http.MethodPost,
				ReqContentType: contentTypeApplicationJson,
				ReqBody: &app.BalanceRequest{
					UserId: 1,
				},
				RespStatus: http.StatusOK,
				RespBody: &SuccessResponseBody{
					Result: &app.UserBalance{
						Balance:  "15",
						Currency: "RUB",
					},
				},
			},
			{
				CaseName:       "positive path, existent user, WithdrawUserAccount after credit user account",
				Path:           pathMethodWithdrawAccount,
				ReqMethod:      http.MethodPost,
				ReqContentType: contentTypeApplicationJson,
				ReqBody: &app.WithdrawAccountRequest{
					UserId:           1,
					Purpose:          "payment to service",
					Amount:           "10",
					IdempotencyToken: uuid.NewV4().String(),
				},
				RespStatus: http.StatusOK,
				RespBody:   &SuccessResponseBody{Result: &app.ResultState{State: app.MsgAccountWithdrawDone}},
			},
			{
				CaseName:       "positive path, existent user, GetUserBalance after Credit - Withdraw",
				Path:           pathMethodGetUserBalance,
				ReqMethod:      http.MethodPost,
				ReqContentType: contentTypeApplicationJson,
				ReqBody: &app.BalanceRequest{
					UserId: 1,
				},
				RespStatus: http.StatusOK,
				RespBody: &SuccessResponseBody{
					Result: &app.UserBalance{
						Balance:  "5",
						Currency: "RUB",
					},
				},
			},
			{
				CaseName:       "positive path, existent user, try to CreditUserAccount using already used Idempotancy Token",
				Path:           pathMethodCreditAccount,
				ReqMethod:      http.MethodPost,
				ReqContentType: contentTypeApplicationJson,
				ReqBody: &app.CreditAccountRequest{
					UserId:           1,
					Purpose:          "payment from service",
					Amount:           "15",
					IdempotencyToken: alreadyUsedIdempotencyToken,
				},
				RespStatus: http.StatusOK,
				RespBody:   &SuccessResponseBody{Result: &app.ResultState{State: app.OperationTokenIsAlreadyUsed}},
			},
			{
				CaseName:       "positive path, existent user, GetUserBalance after Credit attempt with used token",
				Path:           pathMethodGetUserBalance,
				ReqMethod:      http.MethodPost,
				ReqContentType: contentTypeApplicationJson,
				ReqBody: &app.BalanceRequest{
					UserId: 1,
				},
				RespStatus: http.StatusOK,
				RespBody: &SuccessResponseBody{
					Result: &app.UserBalance{
						Balance:  "5",
						Currency: "RUB",
					},
				},
			},
			{
				CaseName:       "positive path, existent user,try to WithdrawUserAccount with aready Used token",
				Path:           pathMethodWithdrawAccount,
				ReqMethod:      http.MethodPost,
				ReqContentType: contentTypeApplicationJson,
				ReqBody: &app.WithdrawAccountRequest{
					UserId:           1,
					Purpose:          "payment to service",
					Amount:           "10",
					IdempotencyToken: alreadyUsedIdempotencyToken,
				},
				RespStatus: http.StatusOK,
				RespBody:   &SuccessResponseBody{Result: &app.ResultState{State: app.OperationTokenIsAlreadyUsed}},
			},
			{
				CaseName:       "positive path, existent user, GetUserBalance after Withdraw attempt with used token",
				Path:           pathMethodGetUserBalance,
				ReqMethod:      http.MethodPost,
				ReqContentType: contentTypeApplicationJson,
				ReqBody: &app.BalanceRequest{
					UserId: 1,
				},
				RespStatus: http.StatusOK,
				RespBody: &SuccessResponseBody{
					Result: &app.UserBalance{
						Balance:  "5",
						Currency: "RUB",
					},
				},
			},
		}

		testTimeout := v.GetDuration("testing_params.test_case_timeout") * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()

		err := test_helpers.PrepareDB(ctx, db, test_helpers.Config{
			InitFilePath:    v.GetString("testing_params.db_init_file_path"),
			CleanUpFilePath: v.GetString("testing_params.db_cleanup_file_path"),
		})
		require.NoError(t, err, "PrepareDB must not return error")

		for caseIdx, tc := range testCases {
			t.Logf("\ttesting case:%d \"%s\"", caseIdx, tc.CaseName)
			{

				b, err := json.Marshal(tc.ReqBody)
				require.NoError(t, err)

				buf := bytes.NewBuffer(b)

				req, err := http.NewRequestWithContext(ctx, http.MethodPost, tc.Path, buf)
				require.NoError(t, err)

				req.Header.Add("Content-Type", contentTypeApplicationJson)
				rr := httptest.NewRecorder()

				appHandler.ServeHTTP(rr, req)

				responseBody, err := ioutil.ReadAll(rr.Body)
				require.NoError(t, err)

				expectedBody, err := json.Marshal(tc.RespBody)
				require.NoError(t, err)

				assert.JSONEq(t, string(expectedBody), string(responseBody), "\t\tresponse body must match")
				assert.Equal(t, tc.RespStatus, rr.Code, "\t\tresponse status mush match")

			}
		}
	})

	t.Run("test user creating, and credit-withdraw operations", func(t *testing.T) {
		testCases := []TestCaseWithPath{
			{
				CaseName:       "negative path, yet nonexistent user, GetUserBalance",
				Path:           pathMethodGetUserBalance,
				ReqMethod:      http.MethodPost,
				ReqContentType: contentTypeApplicationJson,
				ReqBody: &app.BalanceRequest{
					UserId: 100,
				},
				RespStatus: http.StatusBadRequest,
				RespBody: &ErrorResponseBody{
					Error: app.ErrUserDoesNotExist.Error(),
				},
			},
			{
				CaseName:       "negative path, yet nonexistent user, Withdraw account",
				Path:           pathMethodWithdrawAccount,
				ReqMethod:      http.MethodPost,
				ReqContentType: contentTypeApplicationJson,
				ReqBody: &app.WithdrawAccountRequest{
					UserId:           100,
					Purpose:          "payment to service",
					Amount:           "15",
					IdempotencyToken: uuid.NewV4().String(),
				},
				RespStatus: http.StatusBadRequest,
				RespBody: &ErrorResponseBody{
					Error: app.ErrUserDoesNotExist.Error(),
				},
			},
			{
				CaseName:       "positive path, nonexistent user, CreditUserAccount, new user creating",
				Path:           pathMethodCreditAccount,
				ReqMethod:      http.MethodPost,
				ReqContentType: contentTypeApplicationJson,
				ReqBody: &app.CreditAccountRequest{
					UserId:           100,
					Purpose:          "payment from service",
					Amount:           "25",
					IdempotencyToken: uuid.NewV4().String(),
				},
				RespStatus: http.StatusOK,
				RespBody:   &SuccessResponseBody{Result: &app.ResultState{State: app.MsgAccountCreditingDone}},
			},
			{
				CaseName:       "positive path, nonexistent user, GetUserBalance after CreditUserAccount",
				Path:           pathMethodGetUserBalance,
				ReqMethod:      http.MethodPost,
				ReqContentType: contentTypeApplicationJson,
				ReqBody: &app.BalanceRequest{
					UserId: 100,
				},
				RespStatus: http.StatusOK,
				RespBody: &SuccessResponseBody{
					Result: &app.UserBalance{
						Balance:  "25",
						Currency: "RUB",
					},
				},
			},
			{
				CaseName:       "positive path, nonexistent user, WithdrawUserAccount after credit account",
				Path:           pathMethodWithdrawAccount,
				ReqMethod:      http.MethodPost,
				ReqContentType: contentTypeApplicationJson,
				ReqBody: &app.WithdrawAccountRequest{
					UserId:           100,
					Amount:           "15",
					Purpose:          "payment to service",
					IdempotencyToken: uuid.NewV4().String(),
				},
				RespStatus: http.StatusOK,
				RespBody:   &SuccessResponseBody{Result: &app.ResultState{State: app.MsgAccountWithdrawDone}},
			},
			{
				CaseName:       "positive path, nonexistent user, GetUserBalance after Credit - Withdraw",
				Path:           pathMethodGetUserBalance,
				ReqMethod:      http.MethodPost,
				ReqContentType: contentTypeApplicationJson,
				ReqBody: &app.BalanceRequest{
					UserId: 100,
				},
				RespStatus: http.StatusOK,
				RespBody: &SuccessResponseBody{
					Result: &app.UserBalance{
						Balance:  "10",
						Currency: "RUB",
					},
				},
			},
		}

		testTimeout := v.GetDuration("testing_params.test_case_timeout") * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()

		err := test_helpers.PrepareDB(ctx, db, test_helpers.Config{
			InitFilePath:    v.GetString("testing_params.db_init_file_path"),
			CleanUpFilePath: v.GetString("testing_params.db_cleanup_file_path"),
		})
		require.NoError(t, err, "PrepareDB must not return error")

		for caseIdx, tc := range testCases {
			t.Logf("\ttesting case:%d \"%s\"", caseIdx, tc.CaseName)
			{

				b, err := json.Marshal(tc.ReqBody)
				require.NoError(t, err)

				buf := bytes.NewBuffer(b)

				req, err := http.NewRequestWithContext(ctx, http.MethodPost, tc.Path, buf)
				require.NoError(t, err)

				req.Header.Add("Content-Type", contentTypeApplicationJson)
				rr := httptest.NewRecorder()

				appHandler.ServeHTTP(rr, req)

				responseBody, err := ioutil.ReadAll(rr.Body)
				require.NoError(t, err)

				expectedBody, err := json.Marshal(tc.RespBody)
				require.NoError(t, err)

				assert.JSONEq(t, string(expectedBody), string(responseBody), "\t\tresponse body must match")
				assert.Equal(t, tc.RespStatus, rr.Code, "\t\tresponse status mush match")

			}
		}
	})

	t.Run("test money transfer from one user to another", func(t *testing.T) {
		alreadyUsedIdempotencyToken := "1"
		testCases := []TestCaseWithPath{
			{
				CaseName:       "positive path, existent sender user, GetUserBalance",
				Path:           pathMethodGetUserBalance,
				ReqMethod:      http.MethodPost,
				ReqContentType: contentTypeApplicationJson,
				ReqBody: &app.BalanceRequest{
					UserId: 2,
				},
				RespStatus: http.StatusOK,
				RespBody: &SuccessResponseBody{
					Result: &app.UserBalance{
						Balance:  "10",
						Currency: "RUB",
					},
				},
			},
			{
				CaseName:       "positive path, existent receiver user, GetUserBalance",
				Path:           pathMethodGetUserBalance,
				ReqMethod:      http.MethodPost,
				ReqContentType: contentTypeApplicationJson,
				ReqBody: &app.BalanceRequest{
					UserId: 1,
				},
				RespStatus: http.StatusOK,
				RespBody: &SuccessResponseBody{
					Result: &app.UserBalance{
						Balance:  "0",
						Currency: "RUB",
					},
				},
			},
			{
				CaseName:       "positive path, TransferUserMoney",
				Path:           pathMethodTransferUserMoney,
				ReqMethod:      http.MethodPost,
				ReqContentType: contentTypeApplicationJson,
				ReqBody: &app.MoneyTransferRequest{
					SenderId:         2,
					ReceiverId:       1,
					Amount:           "10",
					IdempotencyToken: uuid.NewV4().String(),
				},
				RespStatus: http.StatusOK,
				RespBody:   &SuccessResponseBody{Result: &app.ResultState{State: app.MsgMoneyTransferDone}},
			},
			{
				CaseName:       "positive path, existent sender user, GetUserBalance, after outcoming transfer",
				Path:           pathMethodGetUserBalance,
				ReqMethod:      http.MethodPost,
				ReqContentType: contentTypeApplicationJson,
				ReqBody: &app.BalanceRequest{
					UserId: 2,
				},
				RespStatus: http.StatusOK,
				RespBody: &SuccessResponseBody{
					Result: &app.UserBalance{
						Balance:  "0",
						Currency: "RUB",
					},
				},
			},
			{
				CaseName:       "positive path, existent receiver user, GetUserBalance, after incoming transfer",
				Path:           pathMethodGetUserBalance,
				ReqMethod:      http.MethodPost,
				ReqContentType: contentTypeApplicationJson,
				ReqBody: &app.BalanceRequest{
					UserId: 1,
				},
				RespStatus: http.StatusOK,
				RespBody: &SuccessResponseBody{
					Result: &app.UserBalance{
						Balance:  "10",
						Currency: "RUB",
					},
				},
			},
			{
				CaseName:       "positive path,try to TransferUserMoney,from u2 to u1 with used idempotency token",
				Path:           pathMethodTransferUserMoney,
				ReqMethod:      http.MethodPost,
				ReqContentType: contentTypeApplicationJson,
				ReqBody: &app.MoneyTransferRequest{
					SenderId:         2,
					ReceiverId:       1,
					Amount:           "10",
					IdempotencyToken: alreadyUsedIdempotencyToken,
				},
				RespStatus: http.StatusOK,
				RespBody:   &SuccessResponseBody{Result: &app.ResultState{State: app.OperationTokenIsAlreadyUsed}},
			},
			{
				CaseName:       "positive path, existent sender user, GetUserBalance,u2, after transfer attempt from u2 to u1 with used token",
				Path:           pathMethodGetUserBalance,
				ReqMethod:      http.MethodPost,
				ReqContentType: contentTypeApplicationJson,
				ReqBody: &app.BalanceRequest{
					UserId: 2,
				},
				RespStatus: http.StatusOK,
				RespBody: &SuccessResponseBody{
					Result: &app.UserBalance{
						Balance:  "0",
						Currency: "RUB",
					},
				},
			},
			{
				CaseName:       "positive path, existent receiver user, GetUserBalance,u1, after transfer attempt from u2 to u1 with used token",
				Path:           pathMethodGetUserBalance,
				ReqMethod:      http.MethodPost,
				ReqContentType: contentTypeApplicationJson,
				ReqBody: &app.BalanceRequest{
					UserId: 1,
				},
				RespStatus: http.StatusOK,
				RespBody: &SuccessResponseBody{
					Result: &app.UserBalance{
						Balance:  "10",
						Currency: "RUB",
					},
				},
			},
			{
				CaseName:       "positive path,try to TransferUserMoney,from u1 to u2 with used idempotency token",
				Path:           pathMethodTransferUserMoney,
				ReqMethod:      http.MethodPost,
				ReqContentType: contentTypeApplicationJson,
				ReqBody: &app.MoneyTransferRequest{
					SenderId:         1,
					ReceiverId:       2,
					Amount:           "10",
					IdempotencyToken: alreadyUsedIdempotencyToken,
				},
				RespStatus: http.StatusOK,
				RespBody:   &SuccessResponseBody{Result: &app.ResultState{State: app.OperationTokenIsAlreadyUsed}},
			},
			{
				CaseName:       "positive path, existent sender user, GetUserBalance,u2, after transfer attempt from u1 to u2 with used token",
				Path:           pathMethodGetUserBalance,
				ReqMethod:      http.MethodPost,
				ReqContentType: contentTypeApplicationJson,
				ReqBody: &app.BalanceRequest{
					UserId: 2,
				},
				RespStatus: http.StatusOK,
				RespBody: &SuccessResponseBody{
					Result: &app.UserBalance{
						Balance:  "0",
						Currency: "RUB",
					},
				},
			},
			{
				CaseName:       "positive path, existent receiver user, GetUserBalance,u1, after transfer attempt from u1 to u2 with used token",
				Path:           pathMethodGetUserBalance,
				ReqMethod:      http.MethodPost,
				ReqContentType: contentTypeApplicationJson,
				ReqBody: &app.BalanceRequest{
					UserId: 1,
				},
				RespStatus: http.StatusOK,
				RespBody: &SuccessResponseBody{
					Result: &app.UserBalance{
						Balance:  "10",
						Currency: "RUB",
					},
				},
			},
		}

		testTimeout := v.GetDuration("testing_params.test_case_timeout") * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()

		err := test_helpers.PrepareDB(ctx, db, test_helpers.Config{
			InitFilePath:    v.GetString("testing_params.db_init_file_path"),
			CleanUpFilePath: v.GetString("testing_params.db_cleanup_file_path"),
		})
		require.NoError(t, err, "PrepareDB must not return error")

		for caseIdx, tc := range testCases {
			t.Logf("\ttesting case:%d \"%s\"", caseIdx, tc.CaseName)
			{

				b, err := json.Marshal(tc.ReqBody)
				require.NoError(t, err)

				buf := bytes.NewBuffer(b)

				req, err := http.NewRequestWithContext(ctx, http.MethodPost, tc.Path, buf)
				require.NoError(t, err)

				req.Header.Add("Content-Type", contentTypeApplicationJson)
				rr := httptest.NewRecorder()

				appHandler.ServeHTTP(rr, req)

				responseBody, err := ioutil.ReadAll(rr.Body)
				require.NoError(t, err)

				expectedBody, err := json.Marshal(tc.RespBody)
				require.NoError(t, err)

				assert.JSONEq(t, string(expectedBody), string(responseBody), "\t\tresponse body must match")
				assert.Equal(t, tc.RespStatus, rr.Code, "\t\tresponse status mush match")

			}
		}
	})

	t.Run("test get user operation log", func(t *testing.T) {
		testCases := []TestCaseWithPath{
			{
				CaseName:       "positive path, non-existent user, CreditUserAccount",
				Path:           pathMethodCreditAccount,
				ReqMethod:      http.MethodPost,
				ReqContentType: contentTypeApplicationJson,
				ReqBody: &app.CreditAccountRequest{
					UserId:           100,
					Purpose:          "from user card",
					Amount:           "10",
					IdempotencyToken: "TOKEN1",
				},
				RespStatus: http.StatusOK,
				RespBody:   &SuccessResponseBody{Result: &app.ResultState{State: app.MsgAccountCreditingDone}},
			},

			{
				CaseName:       "positive path, non-existent user, WithdrawUserAccount after credit user account",
				Path:           pathMethodWithdrawAccount,
				ReqMethod:      http.MethodPost,
				ReqContentType: contentTypeApplicationJson,
				ReqBody: &app.WithdrawAccountRequest{
					UserId:           100,
					Purpose:          "ad service",
					Amount:           "10",
					IdempotencyToken: "TOKEN2",
				},
				RespStatus: http.StatusOK,
				RespBody:   &SuccessResponseBody{Result: &app.ResultState{State: app.MsgAccountWithdrawDone}},
			},
			{
				CaseName:       "positive path, non-existent user, get operations log after credit-withdraw",
				Path:           pathMethodGetOperationLog,
				ReqMethod:      http.MethodPost,
				ReqContentType: contentTypeApplicationJson,
				ReqBody: &app.OperationLogRequest{
					UserId:         100,
					Limit:          -1,
					OrderField:     "amount",
					OrderDirection: "desc",
				},
				RespStatus: http.StatusOK,
				RespBody: &SuccessResponseBody{
					Result: &app.OperationsLog{
						OperationsNum: 2,
						Operations: []app.Operation{{
							Id:               6,
							UserId:           100,
							Comment:          "payment from service, from user card",
							Amount:           decimal.NewFromInt(10),
							Date:             time.Time{},
							IdempotencyToken: "TOKEN1",
						}, {
							Id:               7,
							UserId:           100,
							Comment:          "payment to service, ad service",
							Amount:           decimal.NewFromInt(-10),
							Date:             time.Time{},
							IdempotencyToken: "TOKEN2",
						}},
						Page:       1,
						PagesTotal: 1,
					}},
			},
		}

		testTimeout := v.GetDuration("testing_params.test_case_timeout") * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()

		err := test_helpers.PrepareDB(ctx, db, test_helpers.Config{
			InitFilePath:    v.GetString("testing_params.db_init_file_path"),
			CleanUpFilePath: v.GetString("testing_params.db_cleanup_file_path"),
		})
		require.NoError(t, err, "PrepareDB must not return error")

		for caseIdx, tc := range testCases {
			t.Logf("\ttesting case:%d \"%s\"", caseIdx, tc.CaseName)
			{

				b, err := json.Marshal(tc.ReqBody)
				require.NoError(t, err)

				buf := bytes.NewBuffer(b)

				req, err := http.NewRequestWithContext(ctx, http.MethodPost, tc.Path, buf)
				require.NoError(t, err)

				req.Header.Add("Content-Type", contentTypeApplicationJson)
				rr := httptest.NewRecorder()

				appHandler.ServeHTTP(rr, req)

				responseBody, err := ioutil.ReadAll(rr.Body)
				require.NoError(t, err)

				//Check if response body is OperationsLog, edit time in each operation to 0
				res := SuccessResponseBody{}
				d := json.NewDecoder(bytes.NewReader(responseBody))
				d.DisallowUnknownFields()
				if err := d.Decode(&res); err == nil {
					b, err := json.Marshal(res.Result)
					require.NoError(t, err)

					opLog := &app.OperationsLog{}
					d := json.NewDecoder(bytes.NewReader(b))
					d.DisallowUnknownFields()

					if err := d.Decode(opLog); err == nil {
						for i := 0; i < len(opLog.Operations); i++ {
							opLog.Operations[i].Date = time.Time{}
						}
						responseBody, err = json.Marshal(&SuccessResponseBody{Result: opLog})
						require.NoError(t, err)
					}
				}

				expectedBody, err := json.Marshal(tc.RespBody)
				require.NoError(t, err)

				assert.JSONEq(t, string(expectedBody), string(responseBody), "\t\tresponse body must match")
				assert.Equal(t, tc.RespStatus, rr.Code, "\t\tresponse status mush match")

			}
		}
	})
}
