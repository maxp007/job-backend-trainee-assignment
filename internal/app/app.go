package app

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
	"job-backend-trainee-assignment/internal/exchanger"
	"job-backend-trainee-assignment/internal/logger"
	"sync"
)

type IBillingApp interface {
	CreditUserAccount(ctx context.Context, in *CreditAccountRequest) (*ResultState, error)
	WithdrawUserAccount(ctx context.Context, in *WithdrawAccountRequest) (*ResultState, error)
	TransferMoneyFromUserToUser(ctx context.Context, in *MoneyTransferRequest) (*ResultState, error)
	GetUserBalance(ctx context.Context, in *BalanceRequest) (*UserBalance, error)
	GetUserOperations(ctx context.Context, in *OperationLogRequest) (*OperationsLog, error)
}

type BillingApp struct {
	db        *sqlx.DB
	logger    logger.ILogger
	exchanger exchanger.ICurrencyExchanger
	cfg       *Config
	mu        sync.Mutex
}

type Config struct {
	MinOpsMonetaryUnit       decimal.Decimal
	MaxDecimalWholeDigitsNum int
	MinDecimalFracDigitsNum  int
}

var (
	defaultMinOpsMonetaryUnit    = "0.01"
	defaultDecimalWholeDigitsNum = 12
	defaultDecimalFracDigitsNum  = 4
)

func NewApp(logger logger.ILogger, db *sqlx.DB, exchanger exchanger.ICurrencyExchanger, cfg *Config) (
	*BillingApp, error) {
	if logger == nil {
		return nil, fmt.Errorf("must provide non-nil logger instance")
	}

	if cfg == nil {
		defaultMinAmount, err := decimal.NewFromString(defaultMinOpsMonetaryUnit)
		if err != nil {
			return nil, fmt.Errorf("fail to create default config, %v", err)
		}
		cfg = &Config{MinOpsMonetaryUnit: defaultMinAmount, MaxDecimalWholeDigitsNum: defaultDecimalWholeDigitsNum, MinDecimalFracDigitsNum: defaultDecimalFracDigitsNum}
	}

	if db == nil {
		return nil, fmt.Errorf("must provide non-nil sqlx.DB pointer")
	}

	if exchanger == nil {
		return nil, fmt.Errorf("must provide non-nil exchanger instance")
	}

	return &BillingApp{
		logger:    logger,
		db:        db,
		exchanger: exchanger,
		cfg:       cfg,
		mu:        sync.Mutex{},
	}, nil
}
