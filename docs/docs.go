package docs

import (
	"job-backend-trainee-assignment/internal/app"
)

//
// Response body wrappers for swagger docs
//

//swagger:model BalanceResponseBody
//BalanceResponseBody represents a response body of user balance with specified Id
type BalanceResponseBody struct {
	//in: body
	BalanceResponseBody app.UserBalance `json:"result"`
}

//swagger:model CreditAccountResponseBody
//CreditAccountResponseBody represents a message about successful user account crediting operation
type CreditAccountResponseBody struct {
	//in: body
	CreditAccountResponseBody app.ResultState `json:"result"`
}

//swagger:model WithdrawAccountResponseBody
//WithdrawAccountResponseBody represents a message about successful user account withdraw operation
type WithdrawAccountResponseBody struct {
	//in: body
	WithdrawAccountResponseBody app.ResultState `json:"result"`
}

//swagger:model MoneyTransferResponseBody
//MoneyTransferResponseBody represents a message about successful money transfer from one to another user operation
type MoneyTransferResponseBody struct {
	//in: body
	MoneyTransferResponseBody app.ResultState `json:"result"`
}

//swagger:model OperationsLogResponseBody
//OperationsLog represents a response body page of user operations log
type OperationsLogResponseBody struct {
	//in: body
	OperationsLogResponseBody app.OperationsLog `json:"result"`
}

//
// Request body wrappers for swagger docs
//

//swagger:parameters GetUserBalance
type BalanceRequestBody struct {
	//BalanceRequest represents a request body for getting balance of user with specified Id
	//in: body
	BalanceRequestBody app.BalanceRequest
}

//swagger:parameters TransferUserMoney
type MoneyTransferRequestBody struct {
	//MoneyTransferRequest represents a request to perform a money transfer operation
	//in: body
	MoneyTransferRequestBody app.MoneyTransferRequest
}

//swagger:parameters WithdrawUserAccount
type WithdrawAccountRequestBody struct {
	//Represents request to withdraw certain amount of money from user account
	//in: body
	WithdrawAccountRequestBody app.WithdrawAccountRequest
}

//swagger:parameters CreditUserAccount
type CreditAccountRequestBody struct {
	//Represents request to credit certain amount of money to user account
	//in: body
	CreditAccountRequestBody app.CreditAccountRequest
}

//swagger:parameters GetUserOperationsLog
type OperationLogRequestBody struct {
	//OperationLogRequest represents a request body for getting user operation log
	//in: body
	OperationLogRequestBody app.OperationLogRequest
}
