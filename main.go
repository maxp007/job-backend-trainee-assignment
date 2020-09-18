// Job-trainee-assignment.
//
// Project description.
//
//     Schemes: http
//     Version: 0.1
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
//
// swagger:meta
package main

import (
	"context"
	"flag"
	"fmt"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/shopspring/decimal"
	"github.com/spf13/viper"
	_ "job-backend-trainee-assignment/docs"
	"job-backend-trainee-assignment/internal/app"
	"job-backend-trainee-assignment/internal/db_connector"
	"job-backend-trainee-assignment/internal/exchanger"
	"job-backend-trainee-assignment/internal/http_app_handler"
	"job-backend-trainee-assignment/internal/http_handler_router"
	"job-backend-trainee-assignment/internal/logger"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var configPath = flag.String("config", "config", "specify the path to app's config.json file")

func main() {
	flag.Parse()
	v := viper.New()
	v.SetConfigName(*configPath)
	v.AddConfigPath(".")
	v.AutomaticEnv()
	err := v.ReadInConfig()
	if err != nil {
		log.Printf("ERROR failed to read config file at: %s, err %v", *configPath, err)
		return
	}

	logFilePath := v.GetString("log_params.log_path")
	logLevel := v.GetInt64("log_params.log_level")
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_SYNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		log.Printf("ERROR failed to create or open log file at %s, err %v", logFilePath, err)
		return
	}

	mainLogger := logger.NewLogger(logFile, "Main\t", logLevel)
	mainLoggerToStdout := logger.NewLogger(os.Stdout, "Main\t", logLevel)

	mainLogger.Info("starting application")
	mainLoggerToStdout.Info("starting application")

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

	db, dbCloseFunc, err := db_connector.DBConnectWithTimeout(ctx, dbConfig, mainLogger)
	if err != nil {
		mainLogger.Error("failed to connect to db,err %v", err)
		mainLoggerToStdout.Error("failed to connect to db,err %v", err)
		return
	}
	defer dbCloseFunc()

	exLogger := logger.NewLogger(logFile, "NewExchanger\t", logLevel)
	baseCurrencyCode := v.GetString("app_params.base_currency_code")
	ex, err := exchanger.NewExchanger(exLogger, http.DefaultClient, baseCurrencyCode)
	if err != nil {
		mainLogger.Error("failed to create New NewExchanger,err %v", err)
		mainLoggerToStdout.Error("failed to create New NewExchanger,err %v", err)
		return
	}

	appLogger := logger.NewLogger(logFile, "BillingApp\t", logLevel)
	minAmountUnit := v.GetString("app_params.min_monetary_unit")
	decimalMinAmount, err := decimal.NewFromString(minAmountUnit)
	if err != nil {
		mainLogger.Error("failed to create NewApp Config,err: %v", err)
		mainLoggerToStdout.Error("failed to create NewApp Config,err: %v", err)
		return
	}

	decimalWholeDigitNum := v.GetInt("app_params.money_value_params.decimal_whole_digits_num")
	decimalFracDigitNum := v.GetInt("app_params.money_value_params.decimal_frac_digits_num")
	billApp, err := app.NewApp(appLogger, db, ex, &app.Config{
		MinOpsMonetaryUnit:       decimalMinAmount,
		MaxDecimalWholeDigitsNum: decimalWholeDigitNum,
		MinDecimalFracDigitsNum:  decimalFracDigitNum,
	})
	if err != nil {
		mainLogger.Error("failed to create new App,err %v", err)
		mainLoggerToStdout.Error("failed to create new App,err %v", err)
		return
	}

	routerLogger := logger.NewLogger(logFile, "Router\t", logLevel)
	r, err := router.NewRouter(routerLogger)
	if err != nil {
		mainLogger.Error("failed to create NewRouter, err %v", err)
		mainLoggerToStdout.Error("failed to create NewRouter, err %v", err)
		return
	}

	httpHandlerLogger := logger.NewLogger(logFile, "HttpHandler\t", logLevel)
	requestHandleTimeout := v.GetDuration("app_params.request_handle_timeout") * time.Second
	cfg := &http_app_handler.Config{
		RequestHandleTimeout: requestHandleTimeout,
	}

	appHandler, err := http_app_handler.NewHttpAppHandler(httpHandlerLogger, r, billApp, cfg)
	if err != nil {
		mainLogger.Error("failed to create NewHttpAppHandler, err %v", err)
		mainLoggerToStdout.Error("failed to create NewHttpAppHandler, err %v", err)
		return
	}

	readTimeout := v.GetDuration("http_server_params.read_timeout") * time.Second
	writeTimeout := v.GetDuration("http_server_params.write_timeout") * time.Second
	serverLogger := log.New(os.Stdout, "HTTP Server\t", log.LstdFlags|log.Lshortfile|log.Lmicroseconds)

	appPort := v.GetString("http_server_params.port")
	var appHost string
	if v.GetString("APP_HOST") != "" {
		appHost = v.GetString("APP_HOST")
	} else {
		appHost = v.GetString("http_server_params.APP_HOST")
	}

	hostPort := fmt.Sprintf("%s:%s", appHost, appPort)

	server := http.Server{
		Addr:         hostPort,
		Handler:      appHandler,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		ErrorLog:     serverLogger,
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	clientWaitCh := make(chan struct{})

	shutdownTimeout := v.GetDuration("http_server_params.shutdown_timeout") * time.Second

	go func() {
		<-c

		mainLogger.Info("got sigterm signal, shutting down the server")
		mainLoggerToStdout.Info("got sigterm signal, shutting down the server")

		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)

		err := server.Shutdown(ctx)
		if err != nil {
			mainLogger.Error("server shutdown timeout, perform force shutdown")
			mainLoggerToStdout.Error("server shutdown timeout, perform force shutdown")

			clientWaitCh <- struct{}{}
			err := server.Close()
			if err != nil {
				mainLogger.Error("on server closing, got err %v", err)
				mainLoggerToStdout.Error("on server closing, got err %v", err)
			}
		}
		clientWaitCh <- struct{}{}
		cancel()
	}()

	mainLogger.Info("starting listening on %v", hostPort)
	mainLoggerToStdout.Info("starting listening on %v", hostPort)
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		mainLogger.Error("listen and serve, got err %v", err)
		mainLoggerToStdout.Error("listen and serve, got err %v", err)
		return
	}

	mainLogger.Info("waiting for server to serve remaining clients")
	mainLoggerToStdout.Info("waiting for server to serve remaining clients")

	<-clientWaitCh

	mainLogger.Info("remaining clients had been served, exitting")
	mainLoggerToStdout.Info("remaining clients had been served, exitting")
}
