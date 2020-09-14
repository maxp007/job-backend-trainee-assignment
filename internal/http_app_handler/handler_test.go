package http_handler

import (
	"github.com/stretchr/testify/assert"
	"job-backend-trainee-assignment/internal/app"
	"job-backend-trainee-assignment/internal/logger"
	"job-backend-trainee-assignment/internal/router"
	"testing"
)

func TestNewApp_Function(t *testing.T) {

	t.Run("positive path, common", func(t *testing.T) {
		dummyLogger := &logger.DummyLogger{}
		dummyRouter := &router.DummyRouter{}
		dummyApp := &app.DummyBillingApp{}
		handler, err := NewHttpAppHandler(dummyLogger, dummyRouter, dummyApp)
		assert.NoError(t, err, "must get no errors on NewHttpAppHandler Creating")
		assert.NotNil(t, handler, "ptr to app instance must be not nil")
	})

	t.Run("negative path, logger is nil", func(t *testing.T) {
		dummyRouter := &router.DummyRouter{}
		dummyApp := &app.DummyBillingApp{}
		handler, err := NewHttpAppHandler(nil, dummyRouter, dummyApp)
		assert.Error(t, err, "must get error on NewHttpAppHandler Creating")
		assert.Nil(t, handler, "ptr to app instance must be nil")
	})

	t.Run("negative path, router is nil", func(t *testing.T) {
		dummyLogger := &logger.DummyLogger{}
		dummyApp := &app.DummyBillingApp{}
		handler, err := NewHttpAppHandler(dummyLogger, nil, dummyApp)
		assert.Error(t, err, "must get error on NewHttpAppHandler Creating")
		assert.Nil(t, handler, "ptr to app instance must be nil")
	})

	t.Run("negative path, app is nil", func(t *testing.T) {
		dummyLogger := &logger.DummyLogger{}
		dummyRouter := &router.DummyRouter{}
		handler, err := NewHttpAppHandler(dummyLogger, dummyRouter, nil)
		assert.Error(t, err, "must get error on NewHttpAppHandler Creating")
		assert.Nil(t, handler, "ptr to app instance must be nil")
	})
}
