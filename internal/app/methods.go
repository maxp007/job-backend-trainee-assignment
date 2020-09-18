package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/shopspring/decimal"
	"job-backend-trainee-assignment/internal/exchanger"
	"math"
	"net/http"
	"strings"
	"time"
)

func GetCtxError(ctx context.Context, err error) error {
	if ctx.Err() != nil {
		if err != nil {
			if ctx.Err().Error() == ErrContextCancelled.Error() {
				return fmt.Errorf("operation cancelled err, %w", ErrContextCancelled)
			} else {
				return fmt.Errorf("operation timeout err, %w", ErrContextDeadlineExceeded)
			}
		}
	}
	return nil
}

func (ba *BillingApp) GetUserBalance(ctx context.Context, in *BalanceRequest) (*UserBalance, error) {
	if in == nil {
		ba.logger.Error("GetUserBalance, %s", ErrParamsStructIsNil.Error())
		return nil, &AppError{ErrParamsStructIsNil, http.StatusBadRequest}
	}

	user := &User{}
	tx, err := ba.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted, ReadOnly: false})
	if err != nil {
		if ctxErr := GetCtxError(ctx, err); ctxErr != nil {
			ba.logger.Error("GetUserBalance, %s, err %v", ctxErr.Error(), err)
			return nil, &AppError{ctxErr, http.StatusBadRequest}
		}

		ba.logger.Error("GetUserBalance, %s, err %v", ErrDBTransactionBeginFailed.Error(), err)
		return nil, &AppError{ErrDBTransactionBeginFailed, http.StatusInternalServerError}
	}

	defer func() {
		err := tx.Rollback()
		if err != nil && err != sql.ErrTxDone {
			if ctxErr := GetCtxError(ctx, err); ctxErr != nil {
				ba.logger.Error("GetUserBalance, %s, err %v", ctxErr.Error(), err)
			}

			ba.logger.Error("GetUserBalance, %s, err %v", ErrDBTransactionRollbackFailed.Error(), err)
		}
	}()
	{
		err := tx.GetContext(ctx, user, `SELECT user_id, user_name,
			balance, created_at FROM "User" WHERE user_id = $1`, in.UserId)
		if err != nil {
			if ctxErr := GetCtxError(ctx, err); ctxErr != nil {
				ba.logger.Error("GetUserBalance, %s, err %v", ctxErr.Error(), err)
				return nil, &AppError{ctxErr, http.StatusBadRequest}
			}

			if err == sql.ErrNoRows {
				ba.logger.Error("GetUserBalance, %s, err %v", ErrUserDoesNotExist.Error(), err)
				return nil, &AppError{ErrUserDoesNotExist, http.StatusBadRequest}
			}

			ba.logger.Error("GetUserBalance, %s, err %v", ErrDBFailedToFetchUserRow.Error(), err)
			return nil, &AppError{ErrDBFailedToFetchUserRow, http.StatusInternalServerError}
		}
	}
	err = tx.Commit()
	if err != nil {
		if ctxErr := GetCtxError(ctx, err); ctxErr != nil {
			ba.logger.Error("GetUserBalance, %s, err %v", ctxErr.Error(), err)
			return nil, &AppError{ctxErr, http.StatusBadRequest}
		}

		if err == sql.ErrTxDone {
			ba.logger.Error("GetUserBalance, %s, err %v", ErrDBTransactionCommitFailed.Error(), err)
		}

		ba.logger.Error("GetUserBalance, %s, err %v", ErrDBTransactionCommitFailed.Error(), err)
		return nil, &AppError{ErrDBTransactionCommitFailed, http.StatusInternalServerError}
	}

	var userBalance *UserBalance

	if in.Currency != "" && in.Currency != exchanger.RUBCode {
		amountInCurrency, err := ba.exchanger.GetAmountInCurrency(ctx, user.Balance, in.Currency)
		if err != nil {
			if ctxErr := GetCtxError(ctx, err); ctxErr != nil {
				ba.logger.Error("GetUserBalance, %s, err %v", ctxErr.Error(), err)
				return nil, &AppError{ctxErr, http.StatusBadRequest}
			}

			if errors.Is(err, exchanger.ErrTargetCurrencyNameNotFound) {
				ba.logger.Error("GetUserBalance, %s, err %v", ErrCurrencyDoesNotExist.Error(), err)
				return nil, &AppError{ErrCurrencyDoesNotExist, http.StatusBadRequest}
			}

			ba.logger.Error("GetUserBalance, %s, err %v", ErrCurrencyExchangeFailed.Error(), err)
			return nil, &AppError{ErrCurrencyExchangeFailed, http.StatusInternalServerError}
		}
		userBalance = &UserBalance{
			Balance:  amountInCurrency.String(),
			Currency: in.Currency,
		}

	} else {
		userBalance = &UserBalance{
			Balance:  user.Balance.String(),
			Currency: "RUB",
		}
	}

	return userBalance, nil
}

func (ba *BillingApp) CreditUserAccount(ctx context.Context, in *CreditAccountRequest) (*ResultState, error) {

	if in == nil {
		ba.logger.Error("CreditUserAccount, %s", ErrParamsStructIsNil.Error())
		return nil, &AppError{ErrParamsStructIsNil, http.StatusBadRequest}
	}

	amountToCredit, err := decimal.NewFromString(in.Amount)
	if err != nil {
		ba.logger.Error("CreditUserAccount, %s, err %v", ErrFailedToCastAmountToDecimal.Error(), err)
		return nil, &AppError{ErrFailedToCastAmountToDecimal, http.StatusBadRequest}
	}

	if amountToCredit.IsNegative() {
		ba.logger.Error("CreditUserAccount, %s", ErrAmountValueIsNegative.Error())
		return nil, &AppError{ErrAmountValueIsNegative, http.StatusBadRequest}
	}

	ba.mu.Lock()
	minOpsMonetaryUnit := ba.cfg.MinOpsMonetaryUnit
	maxDecimalWholeDigitsNum := ba.cfg.MaxDecimalWholeDigitsNum
	maxDecimalFracDigitsNum := ba.cfg.MinDecimalFracDigitsNum
	ba.mu.Unlock()

	if amountToCredit.LessThan(minOpsMonetaryUnit) {
		ba.logger.Error("CreditUserAccount, %s", ErrAmountValueIsLessThanMin.Error())
		return nil, &AppError{ErrAmountValueIsLessThanMin, http.StatusBadRequest}
	}

	pointSeparatedDecimalSlice := strings.Split(amountToCredit.Round(0).String(), ".")
	//check number of digits to the right of decimal point
	if len(pointSeparatedDecimalSlice) > 1 {
		gotDecimalFracDigitsNum := len(pointSeparatedDecimalSlice[1])
		if gotDecimalFracDigitsNum > maxDecimalFracDigitsNum {
			ba.logger.Error("CreditUserAccount, %s", ErrAmountHasExcessiveFractionalDigits.Error())
			return nil, &AppError{ErrAmountHasExcessiveFractionalDigits, http.StatusBadRequest}
		}
	}

	//check number of digits to the left of decimal point
	gotDecimalWholeDigitsNum := pointSeparatedDecimalSlice[0]
	if len(gotDecimalWholeDigitsNum) > maxDecimalWholeDigitsNum {
		ba.logger.Error("CreditUserAccount, %s", ErrAmountHasExcessiveWholeDigits.Error())
		return nil, &AppError{ErrAmountHasExcessiveWholeDigits, http.StatusBadRequest}
	}

	tx, err := ba.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted, ReadOnly: false})
	if err != nil {
		if ctxErr := GetCtxError(ctx, err); ctxErr != nil {
			ba.logger.Error("CreditUserAccount, %s, err %v", ctxErr.Error(), err)
			return nil, &AppError{ctxErr, http.StatusBadRequest}
		}

		ba.logger.Error("CreditUserAccount, %s, err %v", ErrDBTransactionBeginFailed.Error(), err)
		return nil, &AppError{ErrDBTransactionBeginFailed, http.StatusInternalServerError}
	}

	defer func() {
		err := tx.Rollback()
		if err != nil && err != sql.ErrTxDone {
			if ctxErr := GetCtxError(ctx, err); ctxErr != nil {
				ba.logger.Error("CreditUserAccount, %s, err %v", ctxErr.Error(), err)
			}

			ba.logger.Error("CreditUserAccount, %s, err %v", ErrDBTransactionRollbackFailed.Error(), err)
		}
	}()
	{
		_, err = tx.ExecContext(ctx, `LOCK TABLE "User" IN ROW SHARE MODE`)
		if err != nil {
			if ctxErr := GetCtxError(ctx, err); ctxErr != nil {
				ba.logger.Error("CreditUserAccount, %s, err %v", ctxErr.Error(), err)
				return nil, &AppError{ctxErr, http.StatusBadRequest}
			}

			ba.logger.Error("CreditUserAccount, %s, err %v", ErrDBFailedToLockUserTableForInsert.Error(), err)
			return nil, &AppError{ErrDBFailedToLockUserTableForInsert, http.StatusInternalServerError}
		}

		user := &User{}
		userAlreadyExist := true
		err := tx.GetContext(ctx, user, `SELECT user_id, user_name,
			balance, created_at FROM "User" WHERE user_id = $1 FOR UPDATE`, in.UserId)
		if err != nil {
			if ctxErr := GetCtxError(ctx, err); ctxErr != nil {
				ba.logger.Error("CreditUserAccount, %s, err %v", ctxErr.Error(), err)
				return nil, &AppError{ctxErr, http.StatusBadRequest}
			}

			if err != sql.ErrNoRows {
				ba.logger.Error("CreditUserAccount, %s, err %v", ErrDBFailedToFetchUserRow.Error(), err)
				return nil, &AppError{ErrDBFailedToFetchUserRow, http.StatusInternalServerError}
			} else {
				_, err = tx.ExecContext(ctx, `INSERT INTO "User" (user_id, user_name, balance, created_at) VALUES ($1,$2,$3,$4)`,
					in.UserId, in.Name, decimal.NewFromInt(0), time.Now())
				if err != nil {
					if ctxErr := GetCtxError(ctx, err); ctxErr != nil {
						ba.logger.Error("CreditUserAccount, %s, err %v", ctxErr.Error(), err)
						return nil, &AppError{ctxErr, http.StatusBadRequest}
					}

					ba.logger.Error("CreditUserAccount, %s, err %v", ErrDBFailedToCreateUserRow.Error(), err)
					return nil, &AppError{ErrDBFailedToCreateUserRow, http.StatusInternalServerError}
				}
				userAlreadyExist = false
			}
		}

		maxPossibleDecimal := decimal.New(1, int32(maxDecimalWholeDigitsNum))
		var expectedReceiverNewBalance decimal.Decimal
		if userAlreadyExist {
			expectedReceiverNewBalance = user.Balance.Add(amountToCredit)
		} else {
				expectedReceiverNewBalance = amountToCredit
		}

		if expectedReceiverNewBalance.GreaterThanOrEqual(maxPossibleDecimal) {
			ba.logger.Error("CreditUserAccount, %s", ErrAmountToStoreExceedsMaximumValue.Error())
			return nil, &AppError{ErrAmountToStoreExceedsMaximumValue, http.StatusBadRequest}
		}

		_, err = tx.ExecContext(ctx, `UPDATE "User" SET balance=balance+$1 WHERE user_id=$2`, amountToCredit, in.UserId)
		if err != nil {
			if ctxErr := GetCtxError(ctx, err); ctxErr != nil {
				ba.logger.Error("CreditUserAccount, %s, err %v", ctxErr.Error(), err)
				return nil, &AppError{ctxErr, http.StatusBadRequest}
			}

			ba.logger.Error("CreditUserAccount, %s, err %v", ErrDBFailedToUpdateUserRow.Error(), err)
			return nil, &AppError{ErrDBFailedToUpdateUserRow, http.StatusInternalServerError}
		}

		_, err = tx.ExecContext(ctx, `INSERT INTO "Operation" (user_id, comment, amount, date)
				VALUES ($1,$2,$3,$4)`, in.UserId, fmt.Sprintf(CommentTransferFromServiceWithComment, in.Purpose), in.Amount, time.Now())
		if err != nil {
			if ctxErr := GetCtxError(ctx, err); ctxErr != nil {
				ba.logger.Error("CreditUserAccount, %s, err %v", ctxErr.Error(), err)
				return nil, &AppError{ctxErr, http.StatusBadRequest}
			}

			ba.logger.Error("CreditUserAccount, %s, err %v", ErrFailedToInsertOperationRow.Error(), err)
			return nil, &AppError{ErrFailedToInsertOperationRow, http.StatusInternalServerError}
		}

	}
	err = tx.Commit()
	if err != nil {
		if ctxErr := GetCtxError(ctx, err); ctxErr != nil {
			ba.logger.Error("CreditUserAccount, %s, err %v", ctxErr.Error(), err)
			return nil, &AppError{ctxErr, http.StatusBadRequest}
		}

		if err == sql.ErrTxDone {
			ba.logger.Error("CreditUserAccount, %s, err %v", ErrDBTransactionCommitFailed.Error(), err)
		}

		ba.logger.Error("CreditUserAccount, %s, err %v", ErrDBTransactionCommitFailed.Error(), err)
		return nil, &AppError{ErrDBTransactionCommitFailed, http.StatusInternalServerError}
	}

	return &ResultState{State: MsgAccountCreditingDone}, nil
}

func (ba *BillingApp) WithdrawUserAccount(ctx context.Context, in *WithdrawAccountRequest) (*ResultState, error) {
	if in == nil {
		ba.logger.Error("WithdrawUserAccount, %s", ErrParamsStructIsNil.Error())
		return nil, &AppError{ErrParamsStructIsNil, http.StatusBadRequest}
	}

	amountToWithdraw, err := decimal.NewFromString(in.Amount)
	if err != nil {
		ba.logger.Error("WithdrawUserAccount, %s, err %v", ErrFailedToCastAmountToDecimal.Error(), err)
		return nil, &AppError{ErrFailedToCastAmountToDecimal, http.StatusBadRequest}
	}
	if amountToWithdraw.IsNegative() {
		ba.logger.Error("WithdrawUserAccount, %s", ErrAmountValueIsNegative.Error())
		return nil, &AppError{ErrAmountValueIsNegative, http.StatusBadRequest}
	}

	ba.mu.Lock()
	minOpsMonetaryUnit := ba.cfg.MinOpsMonetaryUnit
	maxDecimalWholeDigitsNum := ba.cfg.MaxDecimalWholeDigitsNum
	maxDecimalFracDigitsNum := ba.cfg.MinDecimalFracDigitsNum
	ba.mu.Unlock()

	if amountToWithdraw.LessThan(minOpsMonetaryUnit) {
		ba.logger.Error("CreditUserAccount, %s", ErrAmountValueIsLessThanMin.Error())
		return nil, &AppError{ErrAmountValueIsLessThanMin, http.StatusBadRequest}
	}

	pointSeparatedDecimalSlice := strings.Split(amountToWithdraw.Round(0).String(), ".")
	//check number of digits to the right of decimal point
	if len(pointSeparatedDecimalSlice) > 1 {
		gotDecimalFracDigitsNum := len(pointSeparatedDecimalSlice[1])
		if gotDecimalFracDigitsNum > maxDecimalFracDigitsNum {
			ba.logger.Error("CreditUserAccount, %s", ErrAmountHasExcessiveFractionalDigits.Error())
			return nil, &AppError{ErrAmountHasExcessiveFractionalDigits, http.StatusBadRequest}
		}
	}

	//check number of digits to the left of decimal point
	gotDecimalWholeDigitsNum := pointSeparatedDecimalSlice[0]
	if len(gotDecimalWholeDigitsNum) > maxDecimalWholeDigitsNum {
		ba.logger.Error("CreditUserAccount, %s", ErrAmountHasExcessiveWholeDigits.Error())
		return nil, &AppError{ErrAmountHasExcessiveWholeDigits, http.StatusBadRequest}
	}

	tx, err := ba.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted, ReadOnly: false})
	if err != nil {
		if ctxErr := GetCtxError(ctx, err); ctxErr != nil {
			ba.logger.Error("WithdrawUserAccount, %s, err %v", ctxErr.Error(), err)
			return nil, &AppError{ctxErr, http.StatusBadRequest}
		}

		ba.logger.Error("WithdrawUserAccount, %s, err %v", ErrDBTransactionBeginFailed.Error(), err)
		return nil, &AppError{ErrDBTransactionBeginFailed, http.StatusInternalServerError}
	}

	defer func() {
		err := tx.Rollback()
		if err != nil && err != sql.ErrTxDone {
			if ctxErr := GetCtxError(ctx, err); ctxErr != nil {
				ba.logger.Error("WithdrawUserAccount, %s, err %v", ctxErr.Error(), err)
			}

			ba.logger.Error("WithdrawUserAccount, %s, err %v", ErrDBTransactionRollbackFailed.Error(), err)
		}
	}()
	{
		user := &User{}
		err := tx.GetContext(ctx, user, `SELECT user_id, user_name,
			balance, created_at FROM "User" WHERE user_id = $1 FOR NO KEY UPDATE`, in.UserId)
		if err != nil {
			if ctxErr := GetCtxError(ctx, err); ctxErr != nil {
				ba.logger.Error("WithdrawUserAccount, %s, err %v", ctxErr.Error(), err)
				return nil, &AppError{ctxErr, http.StatusBadRequest}
			}

			if err == sql.ErrNoRows {
				ba.logger.Error("WithdrawUserAccount, %s, err %v", ErrUserDoesNotExist.Error(), err)
				return nil, &AppError{ErrUserDoesNotExist, http.StatusBadRequest}
			}

			ba.logger.Error("WithdrawUserAccount, %s, err %v", ErrDBFailedToFetchUserRow.Error(), err)
			return nil, &AppError{ErrDBFailedToFetchUserRow, http.StatusInternalServerError}
		}

		if user.Balance.Sub(amountToWithdraw).IsNegative() {
			ba.logger.Error("WithdrawUserAccount, %s", ErrUserDoesNotHaveEnoughMoney.Error())
			return nil, &AppError{ErrUserDoesNotHaveEnoughMoney, http.StatusBadRequest}
		}

		_, err = tx.ExecContext(ctx, `UPDATE "User" SET balance=balance-$1 WHERE user_id=$2`, in.Amount, in.UserId)
		if err != nil {
			if ctxErr := GetCtxError(ctx, err); ctxErr != nil {
				ba.logger.Error("WithdrawUserAccount, %s, err %v", ctxErr.Error(), err)
				return nil, &AppError{ctxErr, http.StatusBadRequest}
			}

			ba.logger.Error("WithdrawUserAccount, %s, err %v", ErrDBFailedToUpdateUserRow.Error(), err)
			return nil, &AppError{ErrDBFailedToUpdateUserRow, http.StatusInternalServerError}
		}

		_, err = tx.ExecContext(ctx, `INSERT INTO "Operation" (user_id, comment, amount, date)
				VALUES ($1,$2,$3,$4)`, in.UserId, fmt.Sprintf(CommentTransferToServiceWithComment, in.Purpose), "-"+in.Amount, time.Now())
		if err != nil {
			if ctxErr := GetCtxError(ctx, err); ctxErr != nil {
				ba.logger.Error("WithdrawUserAccount, %s, err %v", ctxErr.Error(), err)
				return nil, &AppError{ctxErr, http.StatusBadRequest}
			}

			ba.logger.Error("WithdrawUserAccount, %s, err %v", ErrFailedToInsertOperationRow.Error(), err)
			return nil, &AppError{ErrFailedToInsertOperationRow, http.StatusInternalServerError}
		}

	}
	err = tx.Commit()
	if err != nil {
		if ctxErr := GetCtxError(ctx, err); ctxErr != nil {
			ba.logger.Error("WithdrawUserAccount, %s, err %v", ctxErr.Error(), err)
			return nil, &AppError{ctxErr, http.StatusBadRequest}
		}

		if err == sql.ErrTxDone {
			ba.logger.Error("WithdrawUserAccount, %s, err %v", ErrDBTransactionCommitFailed.Error(), err)
		}

		ba.logger.Error("WithdrawUserAccount, %s, err %v", ErrDBTransactionCommitFailed.Error(), err)
		return nil, &AppError{ErrDBTransactionCommitFailed, http.StatusInternalServerError}
	}

	return &ResultState{State: MsgAccountWithdrawDone}, nil
}

func (ba *BillingApp) TransferMoneyFromUserToUser(ctx context.Context, in *MoneyTransferRequest) (*ResultState, error) {
	if in == nil {
		ba.logger.Error("TransferMoneyFromUserToUser, %s", ErrParamsStructIsNil.Error())
		return nil, &AppError{ErrParamsStructIsNil, http.StatusBadRequest}
	}

	if in.ReceiverId == in.SenderId {
		ba.logger.Error("TransferMoneyFromUserToUser, %s", ErrSenderIdIsEqualToReceiverId.Error())
		return nil, &AppError{ErrSenderIdIsEqualToReceiverId, http.StatusBadRequest}
	}

	amountToTransfer, err := decimal.NewFromString(in.Amount)
	if err != nil {
		ba.logger.Error("TransferMoneyFromUserToUser, %s, err %v", ErrFailedToCastAmountToDecimal.Error(), err)
		return nil, &AppError{ErrFailedToCastAmountToDecimal, http.StatusBadRequest}
	}

	if amountToTransfer.IsNegative() {
		ba.logger.Error("TransferMoneyFromUserToUser, %s", ErrAmountValueIsNegative.Error())
		return nil, &AppError{ErrAmountValueIsNegative, http.StatusBadRequest}
	}

	ba.mu.Lock()
	minOpsMonetaryUnit := ba.cfg.MinOpsMonetaryUnit
	maxDecimalWholeDigitsNum := ba.cfg.MaxDecimalWholeDigitsNum
	maxDecimalFracDigitsNum := ba.cfg.MinDecimalFracDigitsNum
	ba.mu.Unlock()

	if amountToTransfer.LessThan(minOpsMonetaryUnit) {
		ba.logger.Error("CreditUserAccount, %s", ErrAmountValueIsLessThanMin.Error())
		return nil, &AppError{ErrAmountValueIsLessThanMin, http.StatusBadRequest}
	}

	pointSeparatedDecimalSlice := strings.Split(amountToTransfer.Round(0).String(), ".")
	//check number of digits to the right of decimal point
	if len(pointSeparatedDecimalSlice) > 1 {
		gotDecimalFracDigitsNum := len(pointSeparatedDecimalSlice[1])
		if gotDecimalFracDigitsNum > maxDecimalFracDigitsNum {
			ba.logger.Error("CreditUserAccount, %s", ErrAmountHasExcessiveFractionalDigits.Error())
			return nil, &AppError{ErrAmountHasExcessiveFractionalDigits, http.StatusBadRequest}
		}
	}

	//check number of digits to the left of decimal point
	gotDecimalWholeDigitsNum := pointSeparatedDecimalSlice[0]
	if len(gotDecimalWholeDigitsNum) > maxDecimalWholeDigitsNum {
		ba.logger.Error("CreditUserAccount, %s", ErrAmountHasExcessiveWholeDigits.Error())
		return nil, &AppError{ErrAmountHasExcessiveWholeDigits, http.StatusBadRequest}
	}

	tx, err := ba.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted, ReadOnly: false})
	if err != nil {
		if ctxErr := GetCtxError(ctx, err); ctxErr != nil {
			ba.logger.Error("TransferMoneyFromUserToUser, %s, err %v", ctxErr.Error(), err)
			return nil, &AppError{ctxErr, http.StatusBadRequest}
		}

		ba.logger.Error("TransferMoneyFromUserToUser, %s, err %v", ErrDBTransactionBeginFailed.Error(), err)
		return nil, &AppError{ErrDBTransactionBeginFailed, http.StatusInternalServerError}
	}

	defer func() {
		err := tx.Rollback()
		if err != nil && err != sql.ErrTxDone {
			if ctxErr := GetCtxError(ctx, err); ctxErr != nil {
				ba.logger.Error("TransferMoneyFromUserToUser, %s, err %v", ctxErr.Error(), err)
			}

			ba.logger.Error("TransferMoneyFromUserToUser, %s, err %v", ErrDBTransactionRollbackFailed.Error(), err)
		}
	}()
	{
		usersInvolved := make([]User, 0)
		err := tx.SelectContext(ctx, &usersInvolved, `SELECT user_id, user_name,
			balance, created_at FROM "User" WHERE user_id = $1 OR user_id = $2 ORDER BY user_id FOR NO KEY UPDATE`, in.SenderId, in.ReceiverId)
		if err != nil {
			if ctxErr := GetCtxError(ctx, err); ctxErr != nil {
				ba.logger.Error("TransferMoneyFromUserToUser, %s, err %v", ctxErr.Error(), err)
				return nil, &AppError{ctxErr, http.StatusBadRequest}
			}

			if err == sql.ErrNoRows {
				ba.logger.Error("TransferMoneyFromUserToUser, %s, err %v", ErrMoneySenderAndReceiverDoNotExist.Error(), err)
				return nil, &AppError{ErrMoneySenderAndReceiverDoNotExist, http.StatusBadRequest}
			}

			ba.logger.Error("TransferMoneyFromUserToUser, %s, err %v", ErrDBFailedToFetchUsersRows.Error(), err)
			return nil, &AppError{ErrDBFailedToFetchUsersRows, http.StatusInternalServerError}
		}

		senderUser := &User{}
		receiverUser := &User{}
		senderFound := false
		receiverFound := false

		for _, u := range usersInvolved {
			if u.Id == in.SenderId {
				senderFound = true
				senderUser = &u
				continue
			}
			if u.Id == in.ReceiverId {
				receiverFound = true
				receiverUser = &u
				continue
			}
		}

		if !senderFound {
			ba.logger.Error("TransferMoneyFromUserToUser, %s, err %v", ErrMoneySenderDoesNotExist.Error(), err)
			return nil, &AppError{ErrMoneySenderDoesNotExist, http.StatusBadRequest}
		}

		if !receiverFound {
			ba.logger.Error("TransferMoneyFromUserToUser, %s, err %v", ErrMoneyReceiverDoesNotExist.Error(), err)
			return nil, &AppError{ErrMoneyReceiverDoesNotExist, http.StatusBadRequest}
		}

		if senderUser.Balance.Sub(amountToTransfer).IsNegative() {
			ba.logger.Error("TransferMoneyFromUserToUser, %s", ErrUserDoesNotHaveEnoughMoney.Error())
			return nil, &AppError{ErrUserDoesNotHaveEnoughMoney, http.StatusBadRequest}
		}

		maxPossibleDecimal := decimal.New(1, int32(maxDecimalWholeDigitsNum))
		expectedReceiverNewBalance := receiverUser.Balance.Add(amountToTransfer)
		if expectedReceiverNewBalance.GreaterThanOrEqual(maxPossibleDecimal) {
			ba.logger.Error("TransferMoneyFromUserToUser, %s", ErrAmountToStoreExceedsMaximumValue.Error())
			return nil, &AppError{ErrAmountToStoreExceedsMaximumValue, http.StatusBadRequest}

		}

		_, err = tx.ExecContext(ctx, `UPDATE "User" SET balance=balance-$1 WHERE user_id=$2`, amountToTransfer, in.SenderId)
		if err != nil {
			if ctxErr := GetCtxError(ctx, err); ctxErr != nil {
				ba.logger.Error("TransferMoneyFromUserToUser, %s, err %v", ctxErr.Error(), err)
				return nil, &AppError{ctxErr, http.StatusBadRequest}
			}

			ba.logger.Error("TransferMoneyFromUserToUser, %s, err %v", ErrDBFailedToUpdateUserRow.Error(), err)
			return nil, &AppError{ErrDBFailedToUpdateUserRow, http.StatusInternalServerError}
		}

		_, err = tx.ExecContext(ctx, `UPDATE "User" SET balance=balance+$1 WHERE user_id=$2`,
			amountToTransfer, in.ReceiverId)
		if err != nil {
			if ctxErr := GetCtxError(ctx, err); ctxErr != nil {
				ba.logger.Error("TransferMoneyFromUserToUser, %s, err %v", ctxErr.Error(), err)
				return nil, &AppError{ctxErr, http.StatusBadRequest}
			}

			ba.logger.Error("TransferMoneyFromUserToUser, %s, err %v", ErrDBFailedToUpdateUserRow.Error(), err)
			return nil, &AppError{ErrDBFailedToUpdateUserRow, http.StatusInternalServerError}
		}

		_, err = tx.ExecContext(ctx, `INSERT INTO "Operation" ( user_id, comment, amount, date)
						VALUES ($1,$2,$3,$4), ($5,$6,$7,$4)`,
			in.SenderId, fmt.Sprintf(CommentTransferToUserWithName, receiverUser.Name), "-"+in.Amount, time.Now(),
			in.ReceiverId, fmt.Sprintf(CommentTransferFromUserWithName, senderUser.Name), in.Amount)
		if err != nil {
			if ctxErr := GetCtxError(ctx, err); ctxErr != nil {
				ba.logger.Error("TransferMoneyFromUserToUser, %s, err %v", ctxErr.Error(), err)
				return nil, &AppError{ctxErr, http.StatusBadRequest}
			}

			ba.logger.Error("TransferMoneyFromUserToUser, %s, err %v", ErrFailedToInsertOperationRow.Error(), err)
			return nil, &AppError{ErrFailedToInsertOperationRow, http.StatusInternalServerError}
		}
	}
	err = tx.Commit()
	if err != nil {
		if ctxErr := GetCtxError(ctx, err); ctxErr != nil {
			ba.logger.Error("TransferMoneyFromUserToUser, %s, err %v", ctxErr.Error(), err)
			return nil, &AppError{ctxErr, http.StatusBadRequest}
		}

		if err == sql.ErrTxDone {
			ba.logger.Error("TransferMoneyFromUserToUser, %s, err %v", ErrDBTransactionCommitFailed.Error(), err)
		}

		ba.logger.Error("TransferMoneyFromUserToUser, %s, err %v", ErrDBTransactionCommitFailed.Error(), err)
		return nil, &AppError{ErrDBTransactionCommitFailed, http.StatusInternalServerError}
	}
	return &ResultState{State: MsgMoneyTransferDone}, nil

}

func (ba *BillingApp) GetUserOperations(ctx context.Context, in *OperationLogRequest) (*OperationsLog, error) {
	if in == nil {
		ba.logger.Error("GetUserOperations, %s", ErrParamsStructIsNil.Error())
		return nil, &AppError{ErrParamsStructIsNil, http.StatusBadRequest}
	}

	if in.Page == 0 {
		in.Page = 1
	}

	if in.Page < 0 {
		ba.logger.Error("GetUserOperations, %s user %d, page %d", "page must be > 0", in.UserId, in.Page)
		return nil, &AppError{ErrPageParamIsLessThanZero, http.StatusBadRequest}
	}

	if in.Limit < -1 {
		ba.logger.Error("GetUserOperations, %s user %d, limit %d", "limit must be greater or equal to -1", in.UserId, in.Limit)
		return nil, &AppError{ErrLimitParamIsLessThanMin, http.StatusBadRequest}
	}

	if in.OrderField == "" {
		in.OrderField = "date"
	}

	if strings.ToLower(in.OrderField) != "date" && strings.ToLower(in.OrderField) != "amount" {
		ba.logger.Error("GetUserOperations, %s user %d, order field %s", "order field must be either \"date\" or \"amount\"", in.UserId, in.OrderField)
		return nil, &AppError{ErrBadOrderFieldParam, http.StatusBadRequest}
	}

	if in.OrderDirection == "" {
		in.OrderDirection = "desc"
	}

	if strings.ToLower(in.OrderDirection) != "asc" && strings.ToLower(in.OrderDirection) != "desc" {
		ba.logger.Error("GetUserOperations, %s user %d, order direction %s", "order field must be either \"asc\" or \"desc\"", in.UserId, in.OrderDirection)
		return nil, &AppError{ErrBadOrderDirectionParam, http.StatusBadRequest}
	}

	zeroUserOperations := false
	userOperations := make([]Operation, 0)
	var allOperationsNum int64 = 0

	tx, err := ba.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted, ReadOnly: false})
	if err != nil {
		if ctxErr := GetCtxError(ctx, err); ctxErr != nil {
			ba.logger.Error("GetUserOperations, %s, err %v", ctxErr.Error(), err)
			return nil, &AppError{ctxErr, http.StatusBadRequest}
		}

		ba.logger.Error("GetUserOperations, %s, err %v", ErrDBTransactionBeginFailed.Error(), err)
		return nil, &AppError{ErrDBTransactionBeginFailed, http.StatusInternalServerError}
	}
	defer func() {
		err := tx.Rollback()
		if err != nil && err != sql.ErrTxDone {
			if ctxErr := GetCtxError(ctx, err); ctxErr != nil {
				ba.logger.Error("GetUserOperations, %s, err %v", ctxErr.Error(), err)
			}

			ba.logger.Error("GetUserOperations, %s, err %v", ErrDBTransactionRollbackFailed.Error(), err)
		}
	}()
	{
		user := &User{}
		err := tx.GetContext(ctx, user, `SELECT user_id ,user_name,
			balance, created_at FROM "User" WHERE user_id = $1`, in.UserId)
		if err != nil {
			if ctxErr := GetCtxError(ctx, err); ctxErr != nil {
				ba.logger.Error("GetUserOperations, %s, err %v", ctxErr.Error(), err)
				return nil, &AppError{ctxErr, http.StatusBadRequest}
			}

			if err == sql.ErrNoRows {
				ba.logger.Error("GetUserOperations, %s, err %v", ErrUserDoesNotExist.Error(), err)
				return nil, &AppError{ErrUserDoesNotExist, http.StatusBadRequest}
			}

			ba.logger.Error("GetUserOperations, %s, err %v", ErrDBFailedToFetchUserRow.Error(), err)
			return nil, &AppError{ErrDBFailedToFetchUserRow, http.StatusInternalServerError}
		}

		if in.Limit == -1 {
			query := fmt.Sprintf(`SELECT * FROM "Operation" WHERE "Operation".user_id=$1 ORDER BY %s %s`, in.OrderField, in.OrderDirection)
			err = tx.SelectContext(ctx, &userOperations, query, in.UserId)
		} else {
			query := fmt.Sprintf(`SELECT * FROM "Operation" WHERE "Operation".user_id=$1 ORDER BY %s %s LIMIT $2 OFFSET $3`, in.OrderField, in.OrderDirection)
			err = tx.SelectContext(ctx, &userOperations, query, in.UserId, in.Limit, in.Limit*(in.Page-1))
		}

		if err != nil && err != sql.ErrNoRows {
			if ctxErr := GetCtxError(ctx, err); ctxErr != nil {
				ba.logger.Error("GetUserOperations, %s, err %v", ctxErr.Error(), err)
				return nil, &AppError{ctxErr, http.StatusBadRequest}
			}

			ba.logger.Error("GetUserOperations, %s, err %v", ErrDBFailedToFetchOperationRows.Error(), err)
			return nil, &AppError{ErrDBFailedToFetchOperationRows, http.StatusInternalServerError}
		}
		if err == sql.ErrNoRows {
			zeroUserOperations = true
		}

		err = tx.Get(&allOperationsNum, `SELECT count(*) FROM "Operation" WHERE user_id=$1`, in.UserId)
		if err != nil && err != sql.ErrNoRows {
			if ctxErr := GetCtxError(ctx, err); ctxErr != nil {
				ba.logger.Error("GetUserOperations, %s, err %v", ctxErr.Error(), err)
				return nil, &AppError{ctxErr, http.StatusBadRequest}
			}

			ba.logger.Error("GetUserOperations, %s, err %v", ErrDBFailedToFetchOperationCountRow.Error(), err)
			return nil, &AppError{ErrDBFailedToFetchOperationCountRow, http.StatusInternalServerError}
		}

	}
	err = tx.Commit()
	if err != nil {
		if ctxErr := GetCtxError(ctx, err); ctxErr != nil {
			ba.logger.Error("GetUserOperations, %s, err %v", ctxErr.Error(), err)
			return nil, &AppError{ctxErr, http.StatusBadRequest}
		}

		if err == sql.ErrTxDone {
			ba.logger.Error("GetUserOperations, %s, err %v", ErrDBTransactionCommitFailed.Error(), err)
		}

		ba.logger.Error("GetUserOperations, %s, err %v", ErrDBTransactionCommitFailed.Error(), err)
		return nil, &AppError{ErrDBTransactionCommitFailed, http.StatusInternalServerError}
	}

	if zeroUserOperations {
		return &OperationsLog{
			OperationsNum: allOperationsNum,
			Operations:    nil,
			Page:          1,
			PagesTotal:    1,
		}, nil
	}
	var pagesTotal int64

	if in.Limit == -1 || in.Limit == 0 {
		pagesTotal = 1
	} else {
		pagesTotal = int64(math.Ceil(float64(allOperationsNum) / float64(in.Limit)))
	}

	return &OperationsLog{
		OperationsNum: allOperationsNum,
		Operations:    userOperations,
		Page:          in.Page,
		PagesTotal:    pagesTotal,
	}, nil

}
