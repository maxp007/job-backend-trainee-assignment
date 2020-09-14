package mock

import (
	"context"
	"github.com/shopspring/decimal"
	"job-backend-trainee-assignment/internal/app"
	"job-backend-trainee-assignment/internal/app/domain"
	"net/http"
	"time"
)

type StubBillingAppCommon struct {
}

func (dba *StubBillingAppCommon) GetUserBalance(ctx context.Context, in *app.BalanceRequest) (*app.UserBalance, error) {
	if in.UserId != 1 && in.UserId != 2 {
		return nil, &app.AppError{app.ErrUserDoesNotExist, http.StatusBadRequest}
	}

	if in.UserId == 2 {
		return &app.UserBalance{
			Balance:  "10",
			Currency: in.Currency,
		}, nil
	} else {
		return &app.UserBalance{
			Balance:  "0",
			Currency: in.Currency,
		}, nil
	}

}

func (dba *StubBillingAppCommon) CreditUserAccount(ctx context.Context, in *app.CreditAccountRequest) (*app.ResultState, error) {
	if in.UserId != 1 && in.UserId != 2 {
		return nil, &app.AppError{app.ErrUserDoesNotExist, http.StatusBadRequest}
	}

	return &app.ResultState{State: app.MsgAccountCreditingDone}, nil
}

func (dba *StubBillingAppCommon) WithdrawUserAccount(ctx context.Context, in *app.WithdrawAccountRequest) (*app.ResultState, error) {
	if in.UserId != 1 && in.UserId != 2 {
		return nil, &app.AppError{app.ErrUserDoesNotExist, http.StatusBadRequest}
	}

	if in.UserId == 1 {
		return nil, &app.AppError{app.ErrUserDoesNotHaveEnoughMoney, http.StatusBadRequest}
	}

	return &app.ResultState{State: app.MsgAccountWithdrawDone}, nil
}

func (dba *StubBillingAppCommon) TransferMoneyFromUserToUser(ctx context.Context, in *app.MoneyTransferRequest) (*app.ResultState, error) {
	if in.SenderId != 1 && in.SenderId != 2 {
		return nil, &app.AppError{app.ErrMoneySenderDoesNotExist, http.StatusBadRequest}
	}

	if in.ReceiverId != 1 && in.ReceiverId != 2 {
		return nil, &app.AppError{app.ErrMoneyReceiverDoesNotExist, http.StatusBadRequest}
	}

	if in.SenderId == 1 {
		return nil, &app.AppError{app.ErrUserDoesNotHaveEnoughMoney, http.StatusBadRequest}
	}

	return &app.ResultState{State: app.MsgMoneyTransferDone}, nil
}

func (dba *StubBillingAppCommon) GetUserOperations(ctx context.Context, in *app.OperationLogRequest) (*app.OperationsLog, error) {
	if in.UserId != 1 && in.UserId != 2 {
		return nil, &app.AppError{app.ErrMoneySenderDoesNotExist, http.StatusBadRequest}
	}

	if in.UserId != 1 && in.UserId != 2 {
		return nil, &app.AppError{app.ErrMoneyReceiverDoesNotExist, http.StatusBadRequest}
	}
	datetime, _ := time.Parse(time.RFC3339, "2020-08-11T10:23:58+03:00")

	return &app.OperationsLog{
		OperationsNum: 2,
		Operations: []domain.Operation{{
			Id:      1,
			UserId:  1,
			Comment: "incoming payment",
			Amount:  decimal.NewFromInt(10),
			Date:    datetime,
		}, {
			Id:      3,
			UserId:  1,
			Comment: "transfer to Mr. Jones",
			Amount:  decimal.NewFromInt(-10),
			Date:    datetime,
		}},
		Page:       0,
		PagesTotal: 0,
	}, nil
}
