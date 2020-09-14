package app

import (
	"errors"
	"fmt"
)

//TODO custom value in errors

//postgres error codes
const (
	NotNullConstraintViolation = "23502"
	ForeignKeyViolation        = "23503"
	UniqueConstraintViolation  = "23505"
)

//ApiError struct conforming error, returned by app methods
type ApiError struct {
	Err      error
	HttpCode int
}

func (ae *ApiError) Error() string {
	return fmt.Sprintf("err: %s, code:%d", ae.Err.Error(), ae.HttpCode)
}

var (
	ErrUserDoesNotExist           = errors.New("user with specified id doe not exist")
	ErrUserDoesNotHaveEnoughMoney = errors.New("user with specified id does not have enough money on balance")
	ErrMoneySenderDoesNotExist    = errors.New("money sender user with specified id was not found")
	ErrMoneyReceiverDoesNotExist  = errors.New("money receiver user with specified id was not found")
	ErrAmountIsLessThanMin        = errors.New("provided amount must be greater than minimin of 10 rub")

	ErrSenderIdIsEqualToReceiverId   = errors.New("sender user and receiver user must have different identifiers")
	ErrExchangeServiceIsNotAvailable = errors.New("error occured during balance currency exchange")
	ErrCurrencyDoesNotExist          = errors.New("provided currency code was not found")

	ErrOrderFieldDoesNotExist        = errors.New("cannot order by specified field, field does not exist")
	ErrOrderDirectionHasInvalidValue = errors.New("specified order direction has invalid value")
)
