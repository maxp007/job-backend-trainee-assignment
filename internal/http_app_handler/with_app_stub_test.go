package http_app_handler

import (
	"bytes"
	"encoding/json"
	"github.com/shopspring/decimal"
	assert "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"job-backend-trainee-assignment/internal/app"
	"job-backend-trainee-assignment/internal/http_handler_router"
	"job-backend-trainee-assignment/internal/logger"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAppHttpHandler_WithStubApp_Common(t *testing.T) {
	operationCreateDatetime, _ := time.Parse(time.RFC3339, "2020-08-11T10:23:58+03:00")

	testCases := []TestCaseWithPath{
		//Get User Balance Cases
		//
		{
			CaseName:       "positive path, handler GetUserBalance, Common",
			Path:           pathMethodGetUserBalance,
			ReqMethod:      http.MethodPost,
			ReqContentType: contentTypeApplicationJson,
			ReqBody: &app.BalanceRequest{
				UserId:   2,
				Currency: "RUB",
			},
			RespStatus: http.StatusOK,
			RespBody: &SuccessResponseBody{Result: app.UserBalance{
				Balance:  "10",
				Currency: "RUB",
			}},
		},
		{
			CaseName:       "negative path, handler GetUserBalance, unsupported request method",
			Path:           pathMethodGetUserBalance,
			ReqMethod:      "UNSUPPORTED_METHOD",
			ReqContentType: contentTypeApplicationJson,
			ReqBody:        &app.BalanceRequest{},
			RespStatus:     http.StatusMethodNotAllowed,
			RespBody:       &ErrorResponseBody{Error: ErrUnsupportedMethod.Error()},
		},
		{
			CaseName:       "negative path, handler GetUserBalance, unsupported content-type",
			Path:           pathMethodGetUserBalance,
			ReqMethod:      http.MethodPost,
			ReqContentType: "UNKNOWN_CONTENT_TYPE",
			ReqBody:        &app.BalanceRequest{},
			RespStatus:     http.StatusUnsupportedMediaType,
			RespBody:       &ErrorResponseBody{Error: ErrUnsupportedContentType.Error()},
		},
		{
			CaseName:       "negative path, handler GetUserBalance, app method returned error (user not exist)",
			Path:           pathMethodGetUserBalance,
			ReqMethod:      http.MethodPost,
			ReqContentType: contentTypeApplicationJson,
			ReqBody: &app.BalanceRequest{
				UserId: 100500,
			},
			RespStatus: http.StatusBadRequest,
			RespBody:   &ErrorResponseBody{Error: app.ErrUserDoesNotExist.Error()},
		},
		{
			CaseName:       "negative path, handler GetUserBalance, corrupted request json",
			Path:           pathMethodGetUserBalance,
			ReqMethod:      http.MethodPost,
			ReqContentType: contentTypeApplicationJson,
			ReqBody:        `{"some":"corrupted json}`,
			RespStatus:     http.StatusBadRequest,
			RespBody:       &ErrorResponseBody{Error: ErrJsonUnmarshalFailed.Error()},
		},

		//Credit User Account Cases
		//
		{
			CaseName:       "positive path, handler CreditUserAccount, Common",
			Path:           pathMethodCreditAccount,
			ReqMethod:      http.MethodPost,
			ReqContentType: contentTypeApplicationJson,
			ReqBody: &app.CreditAccountRequest{
				UserId: 1,
				Amount: "10",
			},
			RespStatus: http.StatusOK,
			RespBody:   &SuccessResponseBody{Result: app.ResultState{State: app.MsgAccountCreditingDone}},
		},
		{
			CaseName:       "negative path, handler CreditUserAccount, unsupported request method",
			Path:           pathMethodCreditAccount,
			ReqMethod:      "UNSUPPORTED_METHOD",
			ReqContentType: contentTypeApplicationJson,
			ReqBody:        &app.CreditAccountRequest{},
			RespStatus:     http.StatusMethodNotAllowed,
			RespBody:       &ErrorResponseBody{Error: ErrUnsupportedMethod.Error()},
		},
		{
			CaseName:       "negative path, handler CreditUserAccount, unsupported content-type",
			Path:           pathMethodCreditAccount,
			ReqMethod:      http.MethodPost,
			ReqContentType: "UNKNOWN_CONTENT_TYPE",
			ReqBody:        &app.CreditAccountRequest{},
			RespStatus:     http.StatusUnsupportedMediaType,
			RespBody:       &ErrorResponseBody{Error: ErrUnsupportedContentType.Error()},
		},
		{
			CaseName:       "negative path, handler CreditUserAccount, corrupted request json",
			Path:           pathMethodCreditAccount,
			ReqMethod:      http.MethodPost,
			ReqContentType: contentTypeApplicationJson,
			ReqBody:        `{"some":"corrupted json}`,
			RespStatus:     http.StatusBadRequest,
			RespBody:       &ErrorResponseBody{Error: ErrJsonUnmarshalFailed.Error()},
		},

		//Withdraw User Account Cases
		//
		{
			CaseName:  "positive path, handler WithdrawUserAccount, Common",
			Path:      pathMethodWithdrawAccount,
			ReqMethod: http.MethodPost,
			ReqBody: &app.WithdrawAccountRequest{
				UserId:  2,
				Purpose: "some purpose",
				Amount:  "10",
			},
			ReqContentType: contentTypeApplicationJson,
			RespStatus:     http.StatusOK,
			RespBody:       &SuccessResponseBody{Result: app.ResultState{State: app.MsgAccountWithdrawDone}},
		},
		{
			CaseName:       "negative path, handler WithdrawUserAccount, unsupported request method",
			Path:           pathMethodWithdrawAccount,
			ReqMethod:      "UNSUPPORTED_METHOD",
			ReqContentType: contentTypeApplicationJson,
			ReqBody:        &app.WithdrawAccountRequest{},
			RespStatus:     http.StatusMethodNotAllowed,
			RespBody:       &ErrorResponseBody{Error: ErrUnsupportedMethod.Error()},
		},
		{
			CaseName:       "negative path, handler WithdrawUserAccount, unsupported content-type",
			Path:           pathMethodWithdrawAccount,
			ReqMethod:      http.MethodPost,
			ReqContentType: "UNKNOWN_CONTENT_TYPE",
			ReqBody:        &app.WithdrawAccountRequest{},
			RespStatus:     http.StatusUnsupportedMediaType,
			RespBody:       &ErrorResponseBody{Error: ErrUnsupportedContentType.Error()},
		},
		{
			CaseName:       "negative path, handler WithdrawUserAccount, app method returned error (user does not exist)",
			Path:           pathMethodWithdrawAccount,
			ReqMethod:      http.MethodPost,
			ReqContentType: contentTypeApplicationJson,
			ReqBody: &app.WithdrawAccountRequest{
				UserId: 100500,
			},
			RespStatus: http.StatusBadRequest,
			RespBody:   &ErrorResponseBody{Error: app.ErrUserDoesNotExist.Error()},
		},
		{
			CaseName:       "negative path, handler WithdrawUserAccount, corrupted request json",
			Path:           pathMethodWithdrawAccount,
			ReqMethod:      http.MethodPost,
			ReqContentType: contentTypeApplicationJson,
			ReqBody:        `{"some":"corrupted json}`,
			RespStatus:     http.StatusBadRequest,
			RespBody:       &ErrorResponseBody{Error: ErrJsonUnmarshalFailed.Error()},
		},

		//Transfer User Money Cases
		//
		{
			CaseName:       "positive path, handler TransferUserMoney, Common",
			Path:           pathMethodTransferUserMoney,
			ReqMethod:      http.MethodPost,
			ReqContentType: contentTypeApplicationJson,
			ReqBody: &app.MoneyTransferRequest{
				SenderId:   2,
				ReceiverId: 1,
				Amount:     "10",
			},
			RespStatus: http.StatusOK,
			RespBody:   &SuccessResponseBody{Result: app.ResultState{State: app.MsgMoneyTransferDone}},
		},
		{
			CaseName:       "negative path, handler TransferUserMoney, unsupported request method",
			Path:           pathMethodTransferUserMoney,
			ReqMethod:      "UNSUPPORTED_METHOD",
			ReqContentType: contentTypeApplicationJson,
			ReqBody:        &app.MoneyTransferRequest{},
			RespStatus:     http.StatusMethodNotAllowed,
			RespBody:       &ErrorResponseBody{Error: ErrUnsupportedMethod.Error()},
		},
		{
			CaseName:       "negative path, handler TransferUserMoney, unsupported content-type",
			Path:           pathMethodTransferUserMoney,
			ReqMethod:      http.MethodPost,
			ReqContentType: "UNKNOWN_CONTENT_TYPE",
			ReqBody:        &app.MoneyTransferRequest{},
			RespStatus:     http.StatusUnsupportedMediaType,
			RespBody:       &ErrorResponseBody{Error: ErrUnsupportedContentType.Error()},
		},
		{
			CaseName:       "negative path, handler TransferUserMoney, app method returned error (user does not exist)",
			Path:           pathMethodTransferUserMoney,
			ReqMethod:      http.MethodPost,
			ReqContentType: contentTypeApplicationJson,
			ReqBody: &app.MoneyTransferRequest{
				SenderId:   100500,
				ReceiverId: 1,
				Amount:     "10",
			},
			RespStatus: http.StatusBadRequest,
			RespBody:   &ErrorResponseBody{Error: app.ErrMoneySenderDoesNotExist.Error()},
		},
		{
			CaseName:       "negative path, handler TransferUserMoney, corrupted request json",
			Path:           pathMethodTransferUserMoney,
			ReqContentType: contentTypeApplicationJson,
			ReqMethod:      http.MethodPost,
			ReqBody:        `{"some":"corrupted json}`,
			RespStatus:     http.StatusBadRequest,
			RespBody:       &ErrorResponseBody{Error: ErrJsonUnmarshalFailed.Error()},
		},

		//User operation log testing
		{
			CaseName:       "positive path, handler GetUserOperations",
			Path:           pathMethodGetOperationLog,
			ReqContentType: contentTypeApplicationJson,
			ReqMethod:      http.MethodPost,
			ReqBody: app.OperationLogRequest{
				UserId: 1,
				Limit:  -1,
			},
			RespStatus: http.StatusOK,
			RespBody: &SuccessResponseBody{Result: &app.OperationsLog{
				OperationsNum: 2,
				Operations: []app.Operation{{
					Id:      1,
					UserId:  1,
					Comment: "incoming payment",
					Amount:  decimal.NewFromInt(10),
					Date:    operationCreateDatetime,
				}, {
					Id:      3,
					UserId:  1,
					Comment: "transfer to Mr. Jones",
					Amount:  decimal.NewFromInt(-10),
					Date:    operationCreateDatetime,
				}},
				Page:       1,
				PagesTotal: 1,
			}},
		},
		{
			CaseName:       "negative path, handler GetUserOperations, corrupted request json",
			Path:           pathMethodGetOperationLog,
			ReqContentType: contentTypeApplicationJson,
			ReqMethod:      http.MethodPost,
			ReqBody:        `{"some":"corrupted json}`,
			RespStatus:     http.StatusBadRequest,
			RespBody:       &ErrorResponseBody{Error: ErrJsonUnmarshalFailed.Error()},
		},
	}

	dummyLogger := &logger.DummyLogger{}

	commonApp := &app.StubBillingAppCommon{}

	r, err := router.NewRouter(dummyLogger)
	require.NoError(t, err, "NewRouter must not return error")

	appHandler, err := NewHttpAppHandler(dummyLogger, r, commonApp, &Config{RequestHandleTimeout: 5 * time.Second})
	require.NoError(t, err, "NewHttpAppHandler must not return error")

	for caseIdx, tc := range testCases {
		t.Logf("\ttesting case:%d \"%s\"", caseIdx, tc.CaseName)
		{
			var req *http.Request
			if str, ok := tc.ReqBody.(string); ok {

				require.True(t, ok, "tc.ReqBody must be a string")
				buf := bytes.NewBufferString(str)
				req, err = http.NewRequest(tc.ReqMethod, tc.Path, buf)
				require.NoError(t, err, "must be able to create request obj")
			} else {

				b, err := json.Marshal(tc.ReqBody)
				require.NoError(t, err, "must be able to marshal request body")
				buf := bytes.NewBuffer(b)
				req, err = http.NewRequest(tc.ReqMethod, tc.Path, buf)
				require.NoError(t, err, "must be able to create request obj")

			}

			req.Header.Add("Content-Type", tc.ReqContentType)
			rr := httptest.NewRecorder()

			appHandler.ServeHTTP(rr, req)

			responseBody, err := ioutil.ReadAll(rr.Body)
			require.NoError(t, err, "Must be able to read response body")

			expectedBody, err := json.Marshal(tc.RespBody)
			require.NoError(t, err, "must be able to unmarshal response body")
			assert.JSONEq(t, string(expectedBody), string(responseBody), "\t\tresponse body must match")
			assert.Equal(t, tc.RespStatus, rr.Code, "\t\tresponse status mush match")

		}
	}
}
