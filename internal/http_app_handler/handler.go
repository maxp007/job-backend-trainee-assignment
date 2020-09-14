package http_handler

import (
	"fmt"
	"job-backend-trainee-assignment/internal/app"
	"job-backend-trainee-assignment/internal/logger"
	"job-backend-trainee-assignment/internal/router"
	"net/http"
)

type AppHttpHandler struct {
	logger logger.ILogger
	app    app.IBillingApp
	router router.IRouter
}

func (h *AppHttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

type ErrorResponseBody struct {
	Error string `json:"error"`
}

type SuccessResponseBody struct {
	Result interface{} `json:"result"`
}

const (
	contentTypeApplicationJson = "application/json"
)

func NewHttpAppHandler(logger logger.ILogger, router router.IRouter, app app.IBillingApp) (*AppHttpHandler, error) {

	if logger == nil {
		return nil, fmt.Errorf("must provide a non-nil pointer to log.Logger")
	}

	if router == nil {
		return nil, fmt.Errorf("must provide a non-nil pointer to httprouter.Router")
	}

	if app == nil {
		return nil, fmt.Errorf("must provide a non-nil pointer to struct of IChatApp")

	}

	csa := &AppHttpHandler{
		logger: logger,
		app:    app,
		router: router,
	}

	csa.router.HandlerFunc(http.MethodPost, pathMethodGetUserBalance, nil)
	csa.router.HandlerFunc(http.MethodPost, pathMethodCreditAccount, nil)
	csa.router.HandlerFunc(http.MethodPost, pathMethodWithdrawAccount, nil)
	csa.router.HandlerFunc(http.MethodPost, pathMethodTransferUserMoney, nil)
	csa.router.HandlerFunc(http.MethodPost, pathMethodGetTransactionLog, nil)

	return csa, nil
}
