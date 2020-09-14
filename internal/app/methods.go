package app

import "context"

func GiveUserMoney(ctx context.Context, in *GiveMoneyRequest) (err error) {
	return nil
}

func TakeUserMoney(ctx context.Context, in *TakeMoneyRequest) (err error) {
	return nil
}

func SendMoneyFromUserToUser(ctx context.Context, in *OrderToExecute) (err error) {
	return nil
}

func GetUserBalance(ctx context.Context, in *BalanceRequest) (result *UserBalance, err error) {
	return nil, nil
}

func GetUserTransactionLog(ctx context.Context, in *TransactionLogRequest) (result *TransactionLog, err error) {
	return nil, nil
}
