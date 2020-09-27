package db_connector

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"job-backend-trainee-assignment/internal/logger"
	"os"
	"time"
)

type Config struct {
	DriverName    string
	DBUser        string
	DBPass        string
	DBName        string
	DBPort        string
	DBHost        string
	SSLMode       string
	RetryInterval time.Duration
}

func DBConnectWithTimeout(ctx context.Context, cfg *Config, log logger.ILogger) (db *sqlx.DB, closer func(), err error) {
	if cfg == nil {
		log.Error("DBConnectWithTimeout, Got nil ptr to config struct")
		return nil, nil, fmt.Errorf("got nil pointer to config struct")
	}

	if log == nil {
		log = logger.NewLogger(os.Stdout, "DbConnector ", logger.L_INFO)
	}

	DSN := "user=%s password=%s host=%s port=%s database=%s sslmode=%s"

	connWaitChan := make(chan struct{})
	connString := fmt.Sprintf(DSN, cfg.DBUser, cfg.DBPass, cfg.DBHost,
		cfg.DBPort, cfg.DBName, cfg.SSLMode)
	go func() {
		for {
			db, err = sqlx.Connect(cfg.DriverName, connString)
			if err != nil {
				log.Info("DBConnectWithTimeout, %v,\n trying to connect", err.Error())
				time.Sleep(cfg.RetryInterval)
				continue
			}

			if ctx.Err() != nil {
				break
			}
			connWaitChan <- struct{}{}
			break
		}
	}()

	select {
	case <-ctx.Done():
		{
			log.Error("DBConnectWithTimeout, connection timeout exceeded")
			return nil, nil, fmt.Errorf("db connection context timeout")
		}
	case <-connWaitChan:
		{
			log.Info("DBConnectWithTimeout, connection established")
			return db, func() { db.Close() }, nil
		}
	}
}
