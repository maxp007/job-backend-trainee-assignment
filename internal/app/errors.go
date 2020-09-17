package app

import (
	"errors"
	"fmt"
)

type AppError struct {
	Err  error
	Code int
}

func (ae *AppError) Error() string {
	return ae.Err.Error()
}

func (ae *AppError) Unwrap() error {
	return ae.Err
}

var (
	ErrParamsStructIsNil = fmt.Errorf("params struct is nil")

	ErrAmountValueIsNegative      = errors.New("amount value must be non-negative number")
	ErrUserDoesNotExist           = errors.New("user with specified id does not exist")
	ErrUserDoesNotHaveEnoughMoney = errors.New("user with specified id does not have enough money on balance")
	ErrMoneySenderDoesNotExist    = errors.New("money sender user with specified id was not found")
	ErrMoneyReceiverDoesNotExist  = errors.New("money receiver user with specified id was not found")
	ErrMoneySenderAndReceiverDoNotExist    = errors.New("both user and receiver users with id were not found")


	ErrCurrencyExchangeFailed      = errors.New("failed to get user balance in specified currency")
	ErrSenderIdIsEqualToReceiverId = errors.New("sender user and receiver user must have different identifiers")
	ErrCurrencyDoesNotExist        = errors.New("currency with provided name was not found")

	ErrDBTransactionBeginFailed    = fmt.Errorf("failed to begin transaction")
	ErrDBTransactionRollbackFailed = fmt.Errorf("failed to rollback transaction")
	ErrDBTransactionCommitFailed   = fmt.Errorf("failed to commit transaction")

	ErrDBFailedToUpdateUserRow = fmt.Errorf("failed to update user row to database")
	ErrDBFailedToCreateUserRow = fmt.Errorf("failed to create user row to database")
	ErrDBFailedToLockUserTableForInsert = fmt.Errorf("failed to lock user table to insert new user")


	ErrFailedToInsertOperationRow       = fmt.Errorf("failed to insert user operation row to database")
	ErrDBFailedToFetchOperationRows     = fmt.Errorf("failed to fetch operation rows from database")
	ErrDBFailedToFetchOperationCountRow = fmt.Errorf("failed to fetch operation count row from database")

	ErrFailedToCastAmountToDecimal = fmt.Errorf("failed to cast incoming string amount to decimal amount")
	ErrDBFailedToFetchUserRow      = fmt.Errorf("failed to fetch user row from database")

	ErrDBFailedToFetchUsersRows      = fmt.Errorf("failed to fetch users rows from database")

	ErrPageParamIsLessThanZero = errors.New("given param page is negative")
	ErrLimitParamIsLessThanMin = errors.New("given param limit is less than min of -1")
	ErrBadOrderFieldParam      = errors.New("given param order field has bad value")
	ErrBadOrderDirectionParam  = errors.New("given param order direction has bad value")
	ErrContextCancelled        = fmt.Errorf("context canceled")
	ErrContextDeadlineExceeded = fmt.Errorf("context deadline exceeded")
)
