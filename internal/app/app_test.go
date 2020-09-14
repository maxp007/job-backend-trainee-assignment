// +build integration

package app

import (
	"context"
	"flag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"job-backend-trainee-assignment/internal/db_connector"
	"job-backend-trainee-assignment/internal/exchanger"
	logger2 "job-backend-trainee-assignment/internal/logger"
	"testing"
	"time"
)

type TestCaseWithTimeout struct {
	TestCase
	timeout time.Duration
}

type TestCase struct {
	caseName       string
	inParams       interface{}
	expectedResult interface{}
	expectedError  error
}

var configPath = flag.String("config", "config", "speicify the path to app's config.json file")

func TestNewApp_Function(t *testing.T) {
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

	db, dbCloseFunc, err := db_connector.DBConnectWithTimeout(ctx, dbConfig, nil)
	if err != nil {
		t.Errorf("failed to connect to db,err %v", err)
		return
	}
	defer dbCloseFunc()

	t.Run("positive path, common", func(t *testing.T) {

		logger := &logger2.DummyLogger{}
		ex := &exchanger.StubExchanger{}
		app, err := NewApp(logger, db, ex)
		assert.NoError(t, err, "must get no errors on NewApp Creating")
		assert.NotNil(t, app, "ptr to app instance must be not nil")
	})

	t.Run("negative path, logger is nil", func(t *testing.T) {
		ex := &exchanger.StubExchanger{}
		app, err := NewApp(nil, db, ex)
		assert.Error(t, err, "must get error on NewApp Creating")
		assert.Nil(t, app, "ptr to app instance must be nil")
	})

	t.Run("negative path, db is nil", func(t *testing.T) {
		logger := &logger2.DummyLogger{}
		ex := &exchanger.StubExchanger{}
		app, err := NewApp(logger, nil, ex)
		assert.Error(t, err, "must get error on NewApp Creating")
		assert.Nil(t, app, "ptr to app instance must be nil")
	})

	t.Run("negative path, ex is nil", func(t *testing.T) {
		logger := &logger2.DummyLogger{}
		app, err := NewApp(logger, db, nil)
		assert.Error(t, err, "must get error on NewApp Creating")
		assert.Nil(t, app, "ptr to app instance must be nil")
	})
}
