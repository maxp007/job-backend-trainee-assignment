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
	"fmt"
	"job-backend-trainee-assignment/internal/app"
	"job-backend-trainee-assignment/internal/db_connector"
	"job-backend-trainee-assignment/internal/exchanger"
	"job-backend-trainee-assignment/internal/http_app_handler"
	"job-backend-trainee-assignment/internal/http_handler_router"
	"job-backend-trainee-assignment/internal/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"flag"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/spf13/viper"
	_ "job-backend-trainee-assignment/docs"
	"log"
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

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_SYNC, 0755)
	if err != nil {
		log.Printf("ERROR failed to create or open log file at %s, err %v", logFilePath, err)
		return
	}

	mainLogger := logger.NewLogger(logFile, "NewExchanger\t", logLevel)

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
		return
	}
	defer dbCloseFunc()

	exLogger := logger.NewLogger(logFile, "NewExchanger\t", logLevel)
	baseCurrencyCode := v.GetString("app_params.base_currency_code")
	ex, err := exchanger.NewExchanger(exLogger, http.DefaultClient, baseCurrencyCode)
	if err != nil {
		mainLogger.Error("failed to create New NewExchanger,err %v", err)
		return
	}

	appLogger := logger.NewLogger(logFile, "BillingApp\t", logLevel)
	billApp, err := app.NewApp(appLogger, db, ex)
	if err != nil {
		mainLogger.Error("failed to create new App,err %v", err)
		return
	}

	routerLogger := logger.NewLogger(logFile, "Router\t", logLevel)
	r := router.NewRouter(routerLogger)

	httpHandlerLogger := logger.NewLogger(logFile, "HttpHandler\t", logLevel)
	appHandler, err := http_app_handler.NewHttpAppHandler(httpHandlerLogger, r,
		billApp,
		&http_app_handler.Config{RequestHandleTimeout: v.GetDuration("app_params.request_handle_timeout") * time.Second})
	if err != nil {
		mainLogger.Error("failed to create NewHttpAppHandler, err %v", err)
		return
	}
	var appHost string
	if v.GetString("APP_HOST") != "" {
		appHost = v.GetString("APP_HOST")
	} else {
		appHost = v.GetString("http_server_params.APP_HOST")
	}

	hostPort := fmt.Sprintf("%s:%s", appHost, v.GetString("http_server_params.port"))
	readTimeout := v.GetDuration("http_server_params.read_timeout") * time.Second
	writeTimeout := v.GetDuration("http_server_params.write_timeout") * time.Second
	serverLogger := log.New(os.Stdout, "HTTP Server\t", log.LstdFlags|log.Lshortfile|log.Lmicroseconds)
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
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)

		err := server.Shutdown(ctx)
		if err != nil {
			mainLogger.Error("server shutdown timeout, perform force shutdown")
			clientWaitCh <- struct{}{}
			err := server.Close()
			if err != nil {
				mainLogger.Error("on server closing, got err %v", err)
			}
		}
		clientWaitCh <- struct{}{}
		cancel()
	}()

	log.Printf("starting listening on %s", hostPort)
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		mainLogger.Error("listen and serve, got err %s", err)
		return
	}

	log.Printf("waiting for server to serve remaining clients")
	<-clientWaitCh
	log.Printf("remaining clients had been served, exitting")
}
