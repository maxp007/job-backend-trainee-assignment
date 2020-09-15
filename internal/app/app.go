package app

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"job-backend-trainee-assignment/internal/exchanger"
	"job-backend-trainee-assignment/internal/logger"
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
}


func NewApp(logger logger.ILogger, db *sqlx.DB, exchanger exchanger.ICurrencyExchanger) (
	*BillingApp, error) {
	if logger == nil {
		return nil, fmt.Errorf("must provide non-nil logger instance")
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
	}, nil
}
