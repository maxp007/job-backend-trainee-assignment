package http_app_handler

import (
	"github.com/stretchr/testify/assert"
	"job-backend-trainee-assignment/internal/app"
	"job-backend-trainee-assignment/internal/http_handler_router"
	"job-backend-trainee-assignment/internal/logger"
	"testing"
	"time"
)


type TestCaseWithPath struct {
	CaseName       string
	Path           string
	ReqMethod      string
	ReqContentType string
	ReqBody        interface{}
	RespStatus     int
	RespBody       interface{}
}

func TestNewAppHandler_Function(t *testing.T) {

	t.Run("positive path, common", func(t *testing.T) {
		dummyLogger := &logger.DummyLogger{}
		dummyRouter := &router.DummyRouter{}
		dummyApp := &app.StubBillingAppCommon{}
		handler, err := NewHttpAppHandler(dummyLogger, dummyRouter, dummyApp, &Config{RequestHandleTimeout: time.Second})
		assert.NoError(t, err, "must get no errors on NewHttpAppHandler Creating")
		assert.NotNil(t, handler, "ptr to app instance must be not nil")
	})

	t.Run("negative path, logger is nil", func(t *testing.T) {
		dummyRouter := &router.DummyRouter{}
		dummyApp := &app.StubBillingAppCommon{}
		handler, err := NewHttpAppHandler(nil, dummyRouter, dummyApp, &Config{RequestHandleTimeout: time.Second})
		assert.Error(t, err, "must get error on NewHttpAppHandler Creating")
		assert.Nil(t, handler, "ptr to app instance must be nil")
	})

	t.Run("negative path, router is nil", func(t *testing.T) {
		dummyLogger := &logger.DummyLogger{}
		dummyApp := &app.StubBillingAppCommon{}
		handler, err := NewHttpAppHandler(dummyLogger, nil, dummyApp, &Config{RequestHandleTimeout: time.Second})
		assert.Error(t, err, "must get error on NewHttpAppHandler Creating")
		assert.Nil(t, handler, "ptr to app instance must be nil")
	})

	t.Run("negative path, app is nil", func(t *testing.T) {
		dummyLogger := &logger.DummyLogger{}
		dummyRouter := &router.DummyRouter{}
		handler, err := NewHttpAppHandler(dummyLogger, dummyRouter, nil, &Config{RequestHandleTimeout: time.Second})
		assert.Error(t, err, "must get error on NewHttpAppHandler Creating")
		assert.Nil(t, handler, "ptr to app instance must be nil")
	})
}
