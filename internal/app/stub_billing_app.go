package app

import (
	"context"
	"github.com/shopspring/decimal"
	"net/http"
	"time"
)

type StubBillingAppCommon struct {
}

func (dba *StubBillingAppCommon) GetUserBalance(ctx context.Context, in *BalanceRequest) (*UserBalance, error) {
	if in.UserId != 1 && in.UserId != 2 {
		return nil, &AppError{ErrUserDoesNotExist, http.StatusBadRequest}
	}

	if in.UserId == 2 {
		return &UserBalance{
			Balance:  "10",
			Currency: in.Currency,
		}, nil
	} else {
		return &UserBalance{
			Balance:  "0",
			Currency: in.Currency,
		}, nil
	}

}

func (dba *StubBillingAppCommon) CreditUserAccount(ctx context.Context, in *CreditAccountRequest) (*ResultState, error) {

	return &ResultState{State: MsgAccountCreditingDone}, nil
}

func (dba *StubBillingAppCommon) WithdrawUserAccount(ctx context.Context, in *WithdrawAccountRequest) (*ResultState, error) {
	if in.UserId != 1 && in.UserId != 2 {
		return nil, &AppError{ErrUserDoesNotExist, http.StatusBadRequest}
	}

	if in.UserId == 1 {
		return nil, &AppError{ErrUserDoesNotHaveEnoughMoney, http.StatusBadRequest}
	}

	return &ResultState{State: MsgAccountWithdrawDone}, nil
}

func (dba *StubBillingAppCommon) TransferMoneyFromUserToUser(ctx context.Context, in *MoneyTransferRequest) (*ResultState, error) {
	if in.SenderId != 1 && in.SenderId != 2 {
		return nil, &AppError{ErrMoneySenderDoesNotExist, http.StatusBadRequest}
	}

	if in.ReceiverId != 1 && in.ReceiverId != 2 {
		return nil, &AppError{ErrMoneyReceiverDoesNotExist, http.StatusBadRequest}
	}

	if in.SenderId == 1 {
		return nil, &AppError{ErrUserDoesNotHaveEnoughMoney, http.StatusBadRequest}
	}

	return &ResultState{State: MsgMoneyTransferDone}, nil
}

func (dba *StubBillingAppCommon) GetUserOperations(ctx context.Context, in *OperationLogRequest) (*OperationsLog, error) {
	if in.UserId != 1 && in.UserId != 2 {
		return nil, &AppError{ErrMoneySenderDoesNotExist, http.StatusBadRequest}
	}

	if in.UserId != 1 && in.UserId != 2 {
		return nil, &AppError{ErrMoneyReceiverDoesNotExist, http.StatusBadRequest}
	}
	datetime, _ := time.Parse(time.RFC3339, "2020-08-11T10:23:58+03:00")

	return &OperationsLog{
		OperationsNum: 2,
		Operations: []Operation{{
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
		Page:       1,
		PagesTotal: 1,
	}, nil
}
