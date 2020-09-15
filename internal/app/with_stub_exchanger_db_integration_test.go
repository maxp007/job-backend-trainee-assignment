// +build integration

package app

import (
	"context"
	"flag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"job-backend-trainee-assignment/internal/db_connector"
	"job-backend-trainee-assignment/internal/exchanger"
	"job-backend-trainee-assignment/internal/logger"
	"job-backend-trainee-assignment/internal/test_helpers"

	"testing"
	"time"
)

//special case for testing context cancellation
const (
	testContextTimeoutInstant = 0 * time.Nanosecond
)

func TestBillingApp_WithStubExchanger_Common(t *testing.T) {
	flag.Parse()
	v := viper.New()

	v.AddConfigPath(".")
	v.AddConfigPath("../../")
	v.SetConfigName(*configPath)
	v.AutomaticEnv()

	err := v.ReadInConfig()
	require.NoErrorf(t, err, "failed to read config file at: %s, err %v", *configPath, err)

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

	app, err := NewApp(dummyLogger, db, ex)
	require.NoErrorf(t, err, "failed to create BillingApp instance, err %v", err)

	caseTimeout := v.GetDuration("testing_params.test_case_timeout") * time.Second
	t.Run("GetUserBalance method of BillingApp", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), caseTimeout)
		defer cancel()

		err := test_helpers.PrepareDB(ctx, db, test_helpers.Config{
			InitFilePath:    v.GetString("testing_params.db_init_file_path"),
			CleanUpFilePath: v.GetString("testing_params.db_cleanup_file_path"),
		})
		require.NoError(t, err, "PrepareDB must not return error")

		var nilUserBalanceResult *UserBalance = nil
		var nilBalanceRequest *BalanceRequest = nil
		testCases := []TestCase{
			{
				caseName: "positive path, when balance > 0",
				inParams: &BalanceRequest{
					UserId:   2,
					Currency: exchanger.RUBCode,
				},
				expectedResult: &UserBalance{
					Balance:  "10",
					Currency: exchanger.RUBCode,
				},
				expectedError: nil,
			},
			{
				caseName: "positive path,when balance = 0 ",
				inParams: &BalanceRequest{
					UserId:   1,
					Currency: exchanger.RUBCode,
				},
				expectedResult: &UserBalance{
					Balance:  "0",
					Currency: exchanger.RUBCode,
				},
				expectedError: nil,
			},
			{
				caseName: "positive path, when currency field is empty",
				inParams: &BalanceRequest{
					UserId: 2,
				},
				expectedResult: &UserBalance{
					Balance:  "10",
					Currency: exchanger.RUBCode,
				},
				expectedError: nil,
			},
			{
				caseName: "positive path, currency exchange",
				inParams: &BalanceRequest{
					UserId:   2,
					Currency: exchanger.USDCode,
				},
				expectedResult: &UserBalance{
					Balance:  "750",
					Currency: exchanger.USDCode,
				},
				expectedError: nil,
			},
			{
				caseName: "negative path, got nonexistent user_id",
				inParams: &BalanceRequest{
					UserId:   100500,
					Currency: exchanger.RUBCode,
				},
				expectedResult: nilUserBalanceResult,
				expectedError:  ErrUserDoesNotExist,
			},

			{
				caseName: "negative path, got unknown currency",
				inParams: &BalanceRequest{
					UserId:   2,
					Currency: "UNKNOWN_CURRENCY",
				},
				expectedResult: nilUserBalanceResult,
				expectedError:  ErrCurrencyDoesNotExist,
			},
			{
				caseName:       "negative path, given params structure is nil",
				inParams:       nilBalanceRequest,
				expectedResult: nilUserBalanceResult,
				expectedError:  ErrParamsStructIsNil,
			},
		}

		for caseIdx, testCase := range testCases {
			ctx, cancel := context.WithTimeout(context.Background(), caseTimeout)

			t.Logf("testing case [%d] %s", caseIdx, testCase.caseName)

			inParams, ok := testCase.inParams.(*BalanceRequest)
			assert.Equal(t, true, ok, "expected inParam to be of type *BalanceRequest")

			userBalance, err := app.GetUserBalance(ctx, inParams)
			assert.ErrorIsf(t, err, testCase.expectedError, "method returned unexpected error: %v", err)
			assert.EqualValues(t, testCase.expectedResult, userBalance, "method returned unexpected result")
			cancel()
		}
	})

	t.Run("CreditUserAccount method of BillingApp", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), caseTimeout)
		defer cancel()

		err := test_helpers.PrepareDB(ctx, db, test_helpers.Config{
			InitFilePath:    v.GetString("testing_params.db_init_file_path"),
			CleanUpFilePath: v.GetString("testing_params.db_cleanup_file_path"),
		})
		require.NoError(t, err, "PrepareDB must not return error")

		var nilCreditAccountRequest *CreditAccountRequest = nil
		var nilResultState *ResultState = nil
		testCases := []TestCase{
			{
				caseName: "positive path, when amount > 0",
				inParams: &CreditAccountRequest{
					UserId:  1,
					Purpose: "credits from user payment",
					Amount:  "10",
				},
				expectedResult: &ResultState{State: MsgAccountCreditingDone},
				expectedError:  nil,
			},
			{
				caseName: "positive path amount = 0",
				inParams: &CreditAccountRequest{
					UserId:  1,
					Purpose: "credits from user payment",
					Amount:  "0",
				},
				expectedResult: &ResultState{State: MsgAccountCreditingDone},
				expectedError:  nil,
			},
			{
				caseName: "negative path, amount value is negative",
				inParams: &CreditAccountRequest{
					UserId:  1,
					Purpose: "credits from user payment",
					Amount:  "-10",
				},
				expectedResult: nilResultState,
				expectedError:  ErrAmountValueIsNegative,
			},
			{
				caseName: "negative path, Amount value is not a number",
				inParams: &CreditAccountRequest{
					UserId:  2,
					Purpose: "credits from user payment",
					Amount:  "not-a-number",
				},
				expectedResult: nilResultState,
				expectedError:  ErrFailedToCastAmountToDecimal,
			},
			{
				caseName: "positive path, when crediting yet nonexistent user",
				inParams: &CreditAccountRequest{
					UserId:  100,
					Purpose: "credits from user payment",
					Amount:  "10",
				},
				expectedResult: &ResultState{State: MsgAccountCreditingDone},
				expectedError:  nil,
			},
			{
				caseName:       "negative path, given nil in params",
				inParams:       nilCreditAccountRequest,
				expectedResult: nilResultState,
				expectedError:  ErrParamsStructIsNil,
			},
		}

		for caseIdx, testCase := range testCases {
			ctx, cancel := context.WithTimeout(context.Background(), caseTimeout)

			t.Logf("testing case [%d] %s", caseIdx, testCase.caseName)

			inParams, ok := testCase.inParams.(*CreditAccountRequest)
			assert.Equal(t, true, ok, "expected inParam to be of type *BalanceRequest")

			userBalance, err := app.CreditUserAccount(ctx, inParams)
			assert.ErrorIsf(t, err, testCase.expectedError, "method returned unexpected error: %v", err)
			assert.EqualValues(t, testCase.expectedResult, userBalance, "method returned unexpected result")
			cancel()
		}
	})

	t.Run("WithdrawUserAccount method of BillingApp", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), caseTimeout)
		defer cancel()

		err := test_helpers.PrepareDB(ctx, db, test_helpers.Config{
			InitFilePath:    v.GetString("testing_params.db_init_file_path"),
			CleanUpFilePath: v.GetString("testing_params.db_cleanup_file_path"),
		})
		require.NoError(t, err, "PrepareDB must not return error")

		var nilWithdrawAccountRequest *WithdrawAccountRequest = nil
		var nilResultState *ResultState = nil
		testCases := []TestCase{
			{
				caseName: "positive path, when amount > 0",
				inParams: &WithdrawAccountRequest{
					UserId:  2,
					Purpose: "payment to advertisement service",
					Amount:  "10",
				},
				expectedResult: &ResultState{State: MsgAccountWithdrawDone},
				expectedError:  nil,
			},
			{
				caseName: "positive path amount = 0",
				inParams: &WithdrawAccountRequest{
					UserId:  1,
					Purpose: "payment to advertisement service",
					Amount:  "0",
				},
				expectedResult: &ResultState{State: MsgAccountWithdrawDone},
				expectedError:  nil,
			},
			{
				caseName: "negative path, user with id not exist",
				inParams: &WithdrawAccountRequest{
					UserId:  100500,
					Purpose: "payment to advertisement service",
					Amount:  "10",
				},
				expectedResult: nilResultState,
				expectedError:  ErrUserDoesNotExist,
			},
			{
				caseName:       "negative path, given nil in params",
				inParams:       nilWithdrawAccountRequest,
				expectedResult: nilResultState,
				expectedError:  ErrParamsStructIsNil,
			},
			{
				caseName: "negative path, amount value is negative",
				inParams: &WithdrawAccountRequest{
					UserId:  1,
					Purpose: "payment to advertisement service",
					Amount:  "-10",
				},
				expectedResult: nilResultState,
				expectedError:  ErrAmountValueIsNegative,
			},
			{
				caseName: "negative path, amount value is not a number",
				inParams: &WithdrawAccountRequest{
					UserId:  1,
					Purpose: "payment to advertisement service",
					Amount:  "not-a-number",
				},
				expectedResult: nilResultState,
				expectedError:  ErrFailedToCastAmountToDecimal,
			},
			{
				caseName: "negative path, user has not enough money",
				inParams: &WithdrawAccountRequest{
					UserId:  1,
					Purpose: "payment to advertisement service",
					Amount:  "100500",
				},
				expectedResult: nilResultState,
				expectedError:  ErrUserDoesNotHaveEnoughMoney,
			},
		}

		for caseIdx, testCase := range testCases {
			ctx, cancel := context.WithTimeout(context.Background(), caseTimeout)

			t.Logf("testing case [%d] %s", caseIdx, testCase.caseName)

			inParams, ok := testCase.inParams.(*WithdrawAccountRequest)
			assert.Equal(t, true, ok, "expected inParam to be of type *BalanceRequest")

			userBalance, err := app.WithdrawUserAccount(ctx, inParams)
			assert.ErrorIsf(t, err, testCase.expectedError, "method returned unexpected error: %v", err)
			assert.EqualValues(t, testCase.expectedResult, userBalance, "method returned unexpected result")
			cancel()
		}
	})

	t.Run("TransferMoneyFromUserToUser method of BillingApp", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), caseTimeout)
		defer cancel()

		err := test_helpers.PrepareDB(ctx, db, test_helpers.Config{
			InitFilePath:    v.GetString("testing_params.db_init_file_path"),
			CleanUpFilePath: v.GetString("testing_params.db_cleanup_file_path"),
		})
		require.NoError(t, err, "PrepareDB must not return error")

		var nilMoneyTransferRequest *MoneyTransferRequest = nil
		var nilResultState *ResultState = nil

		testCases := []TestCase{
			{
				caseName: "positive path, transfer amount > 0",
				inParams: &MoneyTransferRequest{
					SenderId:   2,
					ReceiverId: 1,
					Amount:     "10",
				},
				expectedResult: &ResultState{MsgMoneyTransferDone},
				expectedError:  nil,
			},
			{
				caseName: "positive path, transfer amount = 0",
				inParams: &MoneyTransferRequest{
					SenderId:   2,
					ReceiverId: 1,
					Amount:     "0",
				},
				expectedResult: &ResultState{MsgMoneyTransferDone},
				expectedError:  nil,
			},
			{
				caseName: "negative path, sender with id does not exist",
				inParams: &MoneyTransferRequest{
					SenderId:   100500,
					ReceiverId: 1,
					Amount:     "10",
				},
				expectedResult: nilResultState,
				expectedError:  ErrMoneySenderDoesNotExist,
			},
			{
				caseName: "negative path, receiver with id does not exist",
				inParams: &MoneyTransferRequest{
					SenderId:   1,
					ReceiverId: 100500,
					Amount:     "10",
				},
				expectedResult: nilResultState,
				expectedError:  ErrMoneyReceiverDoesNotExist,
			},
			{
				caseName: "negative path, receiver id is equal to sender id",
				inParams: &MoneyTransferRequest{
					SenderId:   1,
					ReceiverId: 1,
					Amount:     "10",
				},
				expectedResult: nilResultState,
				expectedError:  ErrSenderIdIsEqualToReceiverId,
			},
			{
				caseName:       "negative path, params is nil",
				inParams:       nilMoneyTransferRequest,
				expectedResult: nilResultState,
				expectedError:  ErrParamsStructIsNil,
			},
			{
				caseName: "negative path, Amount value is negative",
				inParams: &MoneyTransferRequest{
					SenderId:   2,
					ReceiverId: 1,
					Amount:     "-10",
				},
				expectedResult: nilResultState,
				expectedError:  ErrAmountValueIsNegative,
			},
			{
				caseName: "negative path, Amount value is not a number",
				inParams: &MoneyTransferRequest{
					SenderId:   2,
					ReceiverId: 1,
					Amount:     "not-a-number",
				},
				expectedResult: nilResultState,
				expectedError:  ErrFailedToCastAmountToDecimal,
			},
			{
				caseName: "negative path, sender does not have enough money",
				inParams: &MoneyTransferRequest{
					SenderId:   2,
					ReceiverId: 1,
					Amount:     "100500",
				},
				expectedResult: nilResultState,
				expectedError:  ErrUserDoesNotHaveEnoughMoney,
			},
		}

		for caseIdx, testCase := range testCases {
			ctx, cancel := context.WithTimeout(context.Background(), caseTimeout)

			t.Logf("testing case [%d] %s", caseIdx, testCase.caseName)

			inParams, ok := testCase.inParams.(*MoneyTransferRequest)
			assert.Equal(t, true, ok, "expected inParam to be of type *BalanceRequest")

			userBalance, err := app.TransferMoneyFromUserToUser(ctx, inParams)
			assert.ErrorIsf(t, err, testCase.expectedError, "method returned unexpected error: %v", err)
			assert.EqualValues(t, testCase.expectedResult, userBalance, "method returned unexpected result")
			cancel()
		}
	})

	t.Run("GetUserOperations method of BillingApp", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), caseTimeout)
		defer cancel()

		err := test_helpers.PrepareDB(ctx, db, test_helpers.Config{
			InitFilePath:    v.GetString("testing_params.db_init_file_path"),
			CleanUpFilePath: v.GetString("testing_params.db_cleanup_file_path"),
		})
		require.NoError(t, err, "PrepareDB must not return error")

		var nilOperationLogRequest *OperationLogRequest = nil
		var nilOperationsLog *OperationsLog = nil
		testCases := []TestCase{
			{
				caseName:       "negative path, in params struct is nil",
				inParams:       nilOperationLogRequest,
				expectedResult: nilOperationsLog,
				expectedError:  ErrParamsStructIsNil,
			},
			{
				caseName: "negative path, user does not exist",
				inParams: &OperationLogRequest{
					UserId: 100500,
				},
				expectedResult: nilOperationsLog,
				expectedError:  ErrUserDoesNotExist,
			},
			{
				caseName: "negative path, page < 0",
				inParams: &OperationLogRequest{
					Page: -1,
				},
				expectedResult: nilOperationsLog,
				expectedError:  ErrPageParamIsLessThanZero,
			},
			{
				caseName: "negative path, limit < -1",
				inParams: &OperationLogRequest{
					Limit: -2,
				},
				expectedResult: nilOperationsLog,
				expectedError:  ErrLimitParamIsLessThanMin,
			},
			{
				caseName: "negative path, bad order field",
				inParams: &OperationLogRequest{
					OrderField: "SOME BAD ORDER FIELD",
				},
				expectedResult: nilOperationsLog,
				expectedError:  ErrBadOrderFieldParam,
			},
			{
				caseName: "negative path, bad order direction",
				inParams: &OperationLogRequest{
					OrderDirection: "SOME BAD ORDER DIRECTION",
				},
				expectedResult: nilOperationsLog,
				expectedError:  ErrBadOrderDirectionParam,
			},
		}

		for caseIdx, testCase := range testCases {
			ctx, cancel := context.WithTimeout(context.Background(), caseTimeout)

			t.Logf("testing case [%d] %s", caseIdx, testCase.caseName)

			inParams, ok := testCase.inParams.(*OperationLogRequest)
			assert.Equal(t, true, ok, "expected inParam to be of type *BalanceRequest")

			userBalance, err := app.GetUserOperations(ctx, inParams)

			assert.ErrorIsf(t, err, testCase.expectedError, "method returned unexpected error: %v", err)
			assert.EqualValues(t, testCase.expectedResult, userBalance, "method returned unexpected result")
			cancel()
		}
	})

}

func TestBillingApp_WithStubExchanger_WithContextTimeout(t *testing.T) {
	flag.Parse()
	v := viper.New()

	v.AddConfigPath(".")
	v.AddConfigPath("../../")
	v.SetConfigName(*configPath)
	v.AutomaticEnv()
	err := v.ReadInConfig()
	require.Errorf(t, err, "failed to read config file at: %s, err %v", *configPath, err)

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

	db, dbCloseFunc, err := db_connector.DBConnectWithTimeout(ctx, dbConfig, &logger.DummyLogger{})
	require.NoErrorf(t, err, "failed to connect to db,err %v", err)
	defer dbCloseFunc()

	ex := &exchanger.StubExchanger{}

	logger := &logger.DummyLogger{}
	app, err := NewApp(logger, db, ex)
	require.NoErrorf(t, err, "failed to create BillingApp instance, err %v", err)

	testTimeout := v.GetDuration("testing_params.test_case_timeout") * time.Second

	t.Run("GetUserBalance method of BillingApp, test context Timeout handling", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()

		err := test_helpers.PrepareDB(ctx, db, test_helpers.Config{
			InitFilePath:    v.GetString("testing_params.db_init_file_path"),
			CleanUpFilePath: v.GetString("testing_params.db_cleanup_file_path"),
		})
		require.NoError(t, err, "PrepareDB must not return error")

		var nilUserBalanceResult *UserBalance = nil
		caseData := TestCase{
			caseName: "negative path, context timed out",
			inParams: &BalanceRequest{
				UserId:   1,
				Currency: exchanger.USDCode,
			},
			expectedResult: nilUserBalanceResult,
			expectedError:  ErrContextDeadlineExceeded,
		}
		testCases := []TestCaseWithTimeout{
			{
				timeout: testContextTimeoutInstant, TestCase: caseData,
			},
		}
		for caseIdx, testCase := range testCases {
			ctx, cancel := context.WithTimeout(context.Background(), testCase.timeout)

			t.Logf("testing case [%d] %s, timeout:%d ns", caseIdx, testCase.caseName, testCase.timeout.Nanoseconds())

			inParams, ok := testCase.inParams.(*BalanceRequest)
			assert.Equal(t, true, ok, "expected inParam to be of type *BalanceRequest")

			userBalance, err := app.GetUserBalance(ctx, inParams)
			assert.ErrorIsf(t, err, testCase.expectedError, "method returned unexpected error: %v", err)
			assert.EqualValues(t, testCase.expectedResult, userBalance, "method returned unexpected result")
			cancel()
		}
	})

	t.Run("GetUserBalance method of BillingApp, test context Cancel handling", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()

		err := test_helpers.PrepareDB(ctx, db, test_helpers.Config{
			InitFilePath:    v.GetString("testing_params.db_init_file_path"),
			CleanUpFilePath: v.GetString("testing_params.db_cleanup_file_path"),
		})
		require.NoError(t, err, "PrepareDB must not return error")

		var nilUserBalanceResult *UserBalance = nil
		caseData := TestCase{
			caseName: "negative path, context cancelled",
			inParams: &BalanceRequest{
				UserId:   1,
				Currency: exchanger.USDCode,
			},
			expectedResult: nilUserBalanceResult,
			expectedError:  ErrContextCancelled,
		}
		testCases := []TestCaseWithTimeout{
			{
				timeout: testContextTimeoutInstant, TestCase: caseData,
			},
		}
		for caseIdx, testCase := range testCases {
			ctx, cancel := context.WithTimeout(context.Background(), testCase.timeout+time.Second)

			t.Logf("testing case [%d] %s, timeout:%d ns", caseIdx, testCase.caseName, testCase.timeout.Nanoseconds())

			inParams, ok := testCase.inParams.(*BalanceRequest)
			assert.Equal(t, true, ok, "expected inParam to be of type *BalanceRequest")

			var userBalance *UserBalance
			var err error
			ch := make(chan struct{})
			go func() {
				userBalance, err = app.GetUserBalance(ctx, inParams)

				ch <- struct{}{}
			}()

			time.Sleep(testCase.timeout)
			cancel()
			<-ch

			assert.ErrorIsf(t, err, testCase.expectedError, "method returned unexpected error: %v", err)
			assert.EqualValues(t, testCase.expectedResult, userBalance, "method returned unexpected result")
		}
	})

	t.Run("CreditUserAccount method of BillingApp, test context Timeout handling", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()

		err := test_helpers.PrepareDB(ctx, db, test_helpers.Config{
			InitFilePath:    v.GetString("testing_params.db_init_file_path"),
			CleanUpFilePath: v.GetString("testing_params.db_cleanup_file_path"),
		})
		require.NoError(t, err, "PrepareDB must not return error")

		var nilResultState *ResultState = nil
		caseData := TestCase{
			caseName: "negative path, context timed out",
			inParams: &CreditAccountRequest{
				UserId:  1,
				Purpose: "credits from user payment",
				Amount:  "1",
			},
			expectedResult: nilResultState,
			expectedError:  ErrContextDeadlineExceeded,
		}

		testCases := []TestCaseWithTimeout{
			{
				timeout: testContextTimeoutInstant, TestCase: caseData,
			},
		}

		for caseIdx, testCase := range testCases {
			ctx, cancel := context.WithTimeout(context.Background(), testCase.timeout)

			t.Logf("testing case [%d] %s, timeout:%d ns", caseIdx, testCase.caseName, testCase.timeout.Nanoseconds())

			inParams, ok := testCase.inParams.(*CreditAccountRequest)
			assert.Equal(t, true, ok, "expected inParam to be of type *BalanceRequest")

			res, err := app.CreditUserAccount(ctx, inParams)
			assert.ErrorIsf(t, err, testCase.expectedError, "method returned unexpected error: %v", err)
			assert.EqualValues(t, testCase.expectedResult, res, "method returned unexpected result")
			cancel()
		}
	})

	t.Run("CreditUserAccount method of BillingApp, test context Cancel handling", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()

		err := test_helpers.PrepareDB(ctx, db, test_helpers.Config{
			InitFilePath:    v.GetString("testing_params.db_init_file_path"),
			CleanUpFilePath: v.GetString("testing_params.db_cleanup_file_path"),
		})
		require.NoError(t, err, "PrepareDB must not return error")

		var nilResultState *ResultState = nil
		caseData := TestCase{
			caseName: "negative path, context cancelled",
			inParams: &CreditAccountRequest{
				UserId:  1,
				Purpose: "credits from user payment",
				Amount:  "1",
			},
			expectedResult: nilResultState,
			expectedError:  ErrContextCancelled,
		}

		testCases := []TestCaseWithTimeout{
			{
				timeout: testContextTimeoutInstant, TestCase: caseData,
			},
		}

		for caseIdx, testCase := range testCases {
			ctx, cancel := context.WithTimeout(context.Background(), testCase.timeout+time.Second)

			t.Logf("testing case [%d] %s, timeout:%d ns", caseIdx, testCase.caseName, testCase.timeout.Nanoseconds())

			inParams, ok := testCase.inParams.(*CreditAccountRequest)
			assert.Equal(t, true, ok, "expected inParam to be of type *BalanceRequest")

			var res *ResultState
			var err error
			ch := make(chan struct{})
			go func() {
				res, err = app.CreditUserAccount(ctx, inParams)

				ch <- struct{}{}
			}()

			time.Sleep(testCase.timeout)
			cancel()
			<-ch

			assert.ErrorIsf(t, err, testCase.expectedError, "method returned unexpected error: %v", err)
			assert.EqualValues(t, testCase.expectedResult, res, "method returned unexpected result")
		}
	})

	t.Run("WithdrawUserAccount method of BillingApp, test context Timeout handling", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()

		err := test_helpers.PrepareDB(ctx, db, test_helpers.Config{
			InitFilePath:    v.GetString("testing_params.db_init_file_path"),
			CleanUpFilePath: v.GetString("testing_params.db_cleanup_file_path"),
		})
		require.NoError(t, err, "PrepareDB must not return error")

		var nilResultState *ResultState = nil
		caseData := TestCase{
			caseName: "negative path, context timed out",
			inParams: &WithdrawAccountRequest{
				UserId:  2,
				Purpose: "payment to advertisement service",
				Amount:  "1",
			},
			expectedResult: nilResultState,
			expectedError:  ErrContextDeadlineExceeded,
		}
		testCases := []TestCaseWithTimeout{
			{
				timeout: testContextTimeoutInstant, TestCase: caseData,
			},
		}

		for caseIdx, testCase := range testCases {
			ctx, cancel := context.WithTimeout(context.Background(), testCase.timeout)

			t.Logf("testing case [%d] %s, timeout:%d ns", caseIdx, testCase.caseName, testCase.timeout.Nanoseconds())

			inParams, ok := testCase.inParams.(*WithdrawAccountRequest)
			assert.Equal(t, true, ok, "expected inParam to be of type *BalanceRequest")

			res, err := app.WithdrawUserAccount(ctx, inParams)
			assert.ErrorIsf(t, err, testCase.expectedError, "method returned unexpected error: %v", err)
			assert.EqualValues(t, testCase.expectedResult, res, "method returned unexpected result")
			cancel()
		}
	})

	t.Run("WithdrawUserAccount method of BillingApp, test context Cancel handling", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()

		err := test_helpers.PrepareDB(ctx, db, test_helpers.Config{
			InitFilePath:    v.GetString("testing_params.db_init_file_path"),
			CleanUpFilePath: v.GetString("testing_params.db_cleanup_file_path"),
		})
		require.NoError(t, err, "PrepareDB must not return error")

		var nilResultState *ResultState = nil
		caseData := TestCase{
			caseName: "negative path, context cancelled",
			inParams: &WithdrawAccountRequest{
				UserId:  2,
				Purpose: "payment to advertisement service",
				Amount:  "1",
			},
			expectedResult: nilResultState,
			expectedError:  ErrContextCancelled,
		}
		testCases := []TestCaseWithTimeout{
			{
				timeout: testContextTimeoutInstant, TestCase: caseData,
			},
		}

		for caseIdx, testCase := range testCases {
			ctx, cancel := context.WithTimeout(context.Background(), testCase.timeout+time.Second)

			t.Logf("testing case [%d] %s, timeout:%d ns", caseIdx, testCase.caseName, testCase.timeout.Nanoseconds())

			inParams, ok := testCase.inParams.(*WithdrawAccountRequest)
			assert.Equal(t, true, ok, "expected inParam to be of type *BalanceRequest")

			var res *ResultState
			var err error
			ch := make(chan struct{})
			go func() {
				res, err = app.WithdrawUserAccount(ctx, inParams)

				ch <- struct{}{}
			}()

			time.Sleep(testCase.timeout)
			cancel()
			<-ch

			assert.ErrorIsf(t, err, testCase.expectedError, "method returned unexpected error: %v", err)
			assert.EqualValues(t, testCase.expectedResult, res, "method returned unexpected result")

		}
	})

	t.Run("TransferMoneyFromUserToUser method of BillingApp, test context Timeout handling", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()

		err := test_helpers.PrepareDB(ctx, db, test_helpers.Config{
			InitFilePath:    v.GetString("testing_params.db_init_file_path"),
			CleanUpFilePath: v.GetString("testing_params.db_cleanup_file_path"),
		})
		require.NoError(t, err, "PrepareDB must not return error")

		var nilResultState *ResultState = nil
		caseData := TestCase{
			caseName: "negative path, context timed out",
			inParams: &MoneyTransferRequest{
				SenderId:   2,
				ReceiverId: 1,
				Amount:     "1",
			},
			expectedResult: nilResultState,
			expectedError:  ErrContextDeadlineExceeded,
		}

		testCases := []TestCaseWithTimeout{
			{
				timeout: testContextTimeoutInstant, TestCase: caseData,
			},
		}
		for caseIdx, testCase := range testCases {
			ctx, cancel := context.WithTimeout(context.Background(), testCase.timeout)

			t.Logf("testing case [%d] %s, timeout:%d ns", caseIdx, testCase.caseName, testCase.timeout.Nanoseconds())

			inParams, ok := testCase.inParams.(*MoneyTransferRequest)
			assert.Equal(t, true, ok, "expected inParam to be of type *BalanceRequest")

			res, err := app.TransferMoneyFromUserToUser(ctx, inParams)
			assert.ErrorIsf(t, err, testCase.expectedError, "method returned unexpected error: %v", err)
			assert.EqualValues(t, testCase.expectedResult, res, "method returned unexpected result")
			cancel()
		}
	})

	t.Run("TransferMoneyFromUserToUser method of BillingApp, test context Cancel handling", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()

		err := test_helpers.PrepareDB(ctx, db, test_helpers.Config{
			InitFilePath:    v.GetString("testing_params.db_init_file_path"),
			CleanUpFilePath: v.GetString("testing_params.db_cleanup_file_path"),
		})
		require.NoError(t, err, "PrepareDB must not return error")

		var nilResultState *ResultState = nil
		caseData := TestCase{
			caseName: "negative path, context cancelled",
			inParams: &MoneyTransferRequest{
				SenderId:   2,
				ReceiverId: 1,
				Amount:     "1",
			},
			expectedResult: nilResultState,
			expectedError:  ErrContextCancelled,
		}

		testCases := []TestCaseWithTimeout{
			{
				timeout: testContextTimeoutInstant, TestCase: caseData,
			},
		}
		for caseIdx, testCase := range testCases {
			ctx, cancel := context.WithTimeout(context.Background(), testCase.timeout+time.Second)

			t.Logf("testing case [%d] %s, timeout:%d ns", caseIdx, testCase.caseName, testCase.timeout.Nanoseconds())

			inParams, ok := testCase.inParams.(*MoneyTransferRequest)
			assert.Equal(t, true, ok, "expected inParam to be of type *BalanceRequest")

			var res *ResultState
			var err error
			ch := make(chan struct{})
			go func() {
				res, err = app.TransferMoneyFromUserToUser(ctx, inParams)
				ch <- struct{}{}
			}()

			time.Sleep(testCase.timeout)
			cancel()
			<-ch

			assert.ErrorIsf(t, err, testCase.expectedError, "method returned unexpected error: %v", err)
			assert.EqualValues(t, testCase.expectedResult, res, "method returned unexpected result")
		}
	})

}
