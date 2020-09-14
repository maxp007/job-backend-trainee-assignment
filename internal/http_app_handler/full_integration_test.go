package http_app_handler

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"job-backend-trainee-assignment/internal/app"
	"job-backend-trainee-assignment/internal/db_connector"
	"job-backend-trainee-assignment/internal/logger"
	"job-backend-trainee-assignment/internal/router"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var DSN = "user=%s password=%s host=%s port=%s database=%s sslmode=%s"

var configPath = flag.String("config", "config", "speicify the path to app's config.json file")

type TestCaseWithPath struct {
	CaseName       string
	Path           string
	ReqMethod      string
	ReqContentType string
	ReqBody        interface{}
	RespStatus     int
	RespBody       interface{}
}

func TestAppHttpHandler_WithAppIntegration_WithStubExchanger(t *testing.T) {
	flag.Parse()
	v := viper.New()

	v.AddConfigPath(".")
	v.AddConfigPath("../../")
	v.SetConfigName(*configPath)
	v.AutomaticEnv()
	err := v.ReadInConfig()
	if err != nil {
		t.Fatalf("failed to read config file at: %s, err %v", *configPath, err)
	}

	t.Log("connecting to db")
	dbConnWaitTimeout := v.GetDuration("db_params.db_conn_timeout") * time.Second
	db, err := db_connector.DBConnectWithTimeout(context.Background(), v, dbConnWaitTimeout)
	if err != nil {
		t.Fatalf("failed to connect to database, err %v", err)
	}

	defer db.Close()

	httpHandlerLogger := &logger.DummyLogger{}
	commonApp := &app.StubBillingAppCommon{}
	routerLogger := &logger.DummyLogger{}

	r := router.NewRouter(routerLogger)

	appHandler, err := NewHttpAppHandler(httpHandlerLogger, r, commonApp)
	if err != nil {
		t.Fatal("failed to create NewHttpAppHandlers instance %v", err)
	}

	//Testing involves check "HttpHandler" + "app" + "database" + "STUB exchanger"
	//
	testCases := []TestCaseWithPath{
		//Test balance withdraw and crediting of existent user
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
				UserId:  1,
				Purpose: "payment from service",
				Amount:  "15",
			},
			RespStatus: http.StatusOK,
			RespBody:   &SuccessResponseBody{Result: &app.ResultState{State: app.MsgAccountCreditingDone}},
		},
		{
			CaseName:       "positive path, existent user, GetUserBalance after CreditUserAccount",
			Path:           pathMethodCreditAccount,
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
			ReqBody:	&app.WithdrawAccountRequest{
				UserId:  1,
				Purpose: "payment to service",
				Amount:  "10",
			},
			RespStatus:     http.StatusOK,
			RespBody:       &SuccessResponseBody{Result: &app.ResultState{State: app.MsgAccountWithdrawDone}},
		},
		{
			CaseName:       "positive path, existent user, GetUserBalance after Credit - Withdraw",
			Path:           pathMethodCreditAccount,
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
		//Test balance withdraw and crediting of previously non-existent user
		//Also test new user adding, through crediting account
		{
			CaseName:       "negative path, nonexistent user, GetUserBalance",
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
			CaseName:       "positive path, nonexistent user, CreditUserAccount",
			Path:           pathMethodCreditAccount,
			ReqMethod:      http.MethodPost,
			ReqContentType: contentTypeApplicationJson,
			ReqBody: &app.CreditAccountRequest{
				UserId:  100,
				Purpose: "payment from service",
				Amount:  "25",
			},
			RespStatus: http.StatusOK,
			RespBody:   &SuccessResponseBody{Result: &app.ResultState{State: app.MsgAccountCreditingDone}},
		},
		{
			CaseName:       "positive path, nonexistent user, GetUserBalance after CreditUserAccount",
			Path:           pathMethodCreditAccount,
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
			CaseName:       "positive path, nonexistent user, WithdrawUserAccount after credit user account",
			Path:           pathMethodWithdrawAccount,
			ReqMethod:      http.MethodPost,
			ReqContentType: contentTypeApplicationJson,
			ReqBody:	&app.WithdrawAccountRequest{
				UserId:  100,
				Purpose: "payment to service",
				Amount:  "15",
			},
			RespStatus:     http.StatusOK,
			RespBody:       &SuccessResponseBody{Result: &app.ResultState{State: app.MsgAccountWithdrawDone}},
		},
		{
			CaseName:       "positive path, nonexistent user, GetUserBalance after Credit - Withdraw",
			Path:           pathMethodCreditAccount,
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
		




		{
			CaseName:       "positive path, TransferUserMoney",
			Path:           pathMethodTransferUserMoney,
			ReqMethod:      http.MethodPost,
			ReqContentType: contentTypeApplicationJson,
			ReqBody:,
			RespStatus:     int,
			RespBody:       &SuccessResponseBody{},
		},
		{
			CaseName:       "negative path, GetUserBalance, user does not exist",
			Path:           pathMethodGetUserBalance,
			ReqMethod:      http.MethodPost,
			ReqContentType: contentTypeApplicationJson,
			ReqBody:,
			RespStatus:     int,
			RespBody:       &SuccessResponseBody{},
		},
		{
			CaseName:       "negative path, CreditUserAccount, user does not exist",
			Path:           pathMethodCreditAccount,
			ReqMethod:      http.MethodPost,
			ReqContentType: contentTypeApplicationJson,
			ReqBody:,
			RespStatus:     int,
			RespBody:       &SuccessResponseBody{},
		},
		{
			CaseName:       "negative path, WithdrawUserAccount, user does not exist",
			Path:           pathMethodWithdrawAccount,
			ReqMethod:      http.MethodPost,
			ReqContentType: contentTypeApplicationJson,
			ReqBody:,
			RespStatus:     int,
			RespBody:       &SuccessResponseBody{},
		},
		{
			CaseName:       "negative path, TransferUserMoney, receiver does not exist",
			Path:           pathMethodTransferUserMoney,
			ReqMethod:      http.MethodPost,
			ReqContentType: contentTypeApplicationJson,
			ReqBody:,
			RespStatus:     int,
			RespBody:       &SuccessResponseBody{},
		},
		{
			CaseName:       "negative path, TransferUserMoney, sender does not exist",
			Path:           pathMethodTransferUserMoney,
			ReqMethod:      http.MethodPost,
			ReqContentType: contentTypeApplicationJson,
			ReqBody:,
			RespStatus:     int,
			RespBody:       &SuccessResponseBody{},
		},
		{
			CaseName:       "negative path, WithdrawUserAccount, user does not have enough money",
			Path:           pathMethodWithdrawAccount,
			ReqMethod:      http.MethodPost,
			ReqContentType: contentTypeApplicationJson,
			ReqBody:,
			RespStatus:     int,
			RespBody:       &SuccessResponseBody{},
		},
		{
			CaseName:       "negative path, TransferUserMoney, sender does not have enough money",
			Path:           pathMethodTransferUserMoney,
			ReqMethod:      http.MethodPost,
			ReqContentType: contentTypeApplicationJson,
			ReqBody:,
			RespStatus:     http.StatusBadRequest,
			RespBody:       &SuccessResponseBody{},
		},
	}

	for caseIdx, tc := range testCases {
		t.Logf("\ttesting case:%d \"%s\"", caseIdx, tc.CaseName)
		{

			b, err := json.Marshal(tc.ReqBody)
			require.NoError(t, err)

			buf := bytes.NewBuffer(b)

			req, err := http.NewRequest(http.MethodPost, tc.Path, buf)
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
}
