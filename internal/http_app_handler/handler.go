package http_app_handler

import (
	"fmt"
	"job-backend-trainee-assignment/internal/app"
	"job-backend-trainee-assignment/internal/http_handler_router"
	"job-backend-trainee-assignment/internal/logger"
	"net/http"
	"sync"
	"time"
)

type Config struct {
	RequestHandleTimeout time.Duration
}

type AppHttpHandler struct {
	logger logger.ILogger
	app    app.IBillingApp
	router router.IRouter
	cfg    *Config
	mu     sync.Mutex
}

func (h *AppHttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

const (
	contentTypeApplicationJson = "application/json"
)

const (
	pathMethodGetUserBalance    = "/balance"
	pathMethodCreditAccount     = "/credit"
	pathMethodWithdrawAccount   = "/withdraw"
	pathMethodTransferUserMoney = "/transfer"
	pathMethodGetOperationLog   = "/operations"
)

func NewHttpAppHandler(logger logger.ILogger, router router.IRouter, app app.IBillingApp, cfg *Config) (*AppHttpHandler, error) {

	if logger == nil {
		return nil, fmt.Errorf("must provide a non-nil pointer to log.Logger")
	}

	if router == nil {
		return nil, fmt.Errorf("must provide a non-nil pointer to httprouter.Router")
	}

	if app == nil {
		return nil, fmt.Errorf("must provide a non-nil pointer to struct of IChatApp")

	}

	if cfg == nil {
		cfg = &Config{RequestHandleTimeout: 5 * time.Second}
	}

	h := &AppHttpHandler{
		logger: logger,
		app:    app,
		router: router,
		cfg:    cfg,
	}

	h.router.SetMethodNotAllowedHandler(h.MethodNotAllowedHandler)

	HandlerGetUserBalance := h.AccessLogMW(
		h.ContentTypeValidationMW(h.HandlerGetUserBalance, contentTypeApplicationJson))

	HandlerCreditUserAccount := h.AccessLogMW(
		h.ContentTypeValidationMW(h.HandlerCreditUserAccount, contentTypeApplicationJson))

	HandlerWithdrawUserAccount := h.AccessLogMW(
		h.ContentTypeValidationMW(h.HandlerWithdrawUserAccount, contentTypeApplicationJson))

	HandlerTransferUserMoney := h.AccessLogMW(
		h.ContentTypeValidationMW(h.HandlerTransferUserMoney, contentTypeApplicationJson))

	HandlerGetUserOperationsLog := h.AccessLogMW(
		h.ContentTypeValidationMW(h.HandlerGetUserOperationsLog, contentTypeApplicationJson))

	h.router.HandlerFunc(http.MethodPost, pathMethodGetUserBalance, HandlerGetUserBalance)
	h.router.HandlerFunc(http.MethodPost, pathMethodCreditAccount, HandlerCreditUserAccount)
	h.router.HandlerFunc(http.MethodPost, pathMethodWithdrawAccount, HandlerWithdrawUserAccount)
	h.router.HandlerFunc(http.MethodPost, pathMethodTransferUserMoney, HandlerTransferUserMoney)
	h.router.HandlerFunc(http.MethodPost, pathMethodGetOperationLog, HandlerGetUserOperationsLog)
	return h, nil
}
