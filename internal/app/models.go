package app

import (
	"github.com/shopspring/decimal"
	"time"
)



//swagger:model BalanceRequest
//BalanceRequest represents a request body for getting balance of user with specified Id
type BalanceRequest struct {
	//identifier of user who's balance is required to get
	//required: true
	//example: 1
	UserId int64 `json:"user_id"`
	//name of currency in which balance value is required
	//required: false
	//enum: RUB,USD,EUR
	//default: RUB
	Currency string `json:"currency"`
}

//swagger:model MoneyTransferRequest
//MoneyTransferRequest represents a request to perform a money transfer operation
//
type MoneyTransferRequest struct {
	//identifier of party sending money
	//required: true
	//example: 1
	SenderId int64 `json:"sender_id"`
	//identifier of party receiving money
	//required: true
	//example: 2
	ReceiverId int64 `json:"receiver_id"`
	//amount of money to be sent to another user
	//required: true
	//minimum: 1.00
	//example: 100
	Amount string `json:"amount"`
}

//swagger:model WithdrawAccountRequest
//Represents request to withdraw certain amount of money from user account
type WithdrawAccountRequest struct {
	//identifier of user who's balance is required to withdraw
	//required: true
	//example: 1
	UserId int64 `json:"user_id"`
	//money withdraw purpose
	//example: advertisement service payment
	Purpose string `json:"purpose"`
	//amount of money to be Withdrawn
	//required: true
	//minimum: 1.00
	//example: 100
	Amount string `json:"amount"`
}

//swagger:model CreditAccountRequest
//Represents request to credit certain amount of money to user account
type CreditAccountRequest struct {
	//identifier of user who's balance is required to add
	//required: true
	//example: 1
	UserId int64 `json:"user_id"`
	//user name to be shown to other users
	//example: Mr. Jones
	Name string `json:"name"`
	//money adding purpose
	//example: payment from debit card
	Purpose string `json:"purpose"`
	//amount of money to be Added to account
	//required: true
	//minimum: 1.00
	//example: 100
	Amount string `json:"amount"`
}

//swagger:model
//Operation represents a operation model used to store user operations
type Operation struct {
	//operation identifier
	//example: 1
	Id int64 `json:"operation_id" db:"operation_id"`
	//identifier of user, involved in operation
	//example: 2
	UserId int64 `json:"user_id" db:"user_id"`
	//comment of operation
	//example: transfer from user to user
	Comment string `json:"purpose" db:"comment"`
	//amount of money in "RUB" sent or received in operation
	//example: -100 or 100
	Amount decimal.Decimal `json:"amount" db:"amount"`
	//operation creating date
	//example: 2020-08-10
	Date time.Time `json:"date" db:"date"`
}

// swagger:model
//User represents a user model stored in a database
type User struct {
	//User identifier
	//example: 1
	Id int64 `json:"user_id" db:"user_id"`
	//user name to be shown to other users
	//example: Mr. Jones
	Name string `json:"name" db:"user_name"`
	//current user balance in russian roubles
	//minimum: 0.00
	//example: 100
	Balance decimal.Decimal `db:"balance"`
	//date, the user record was created
	//example: 2020-08-10
	CreatedAt time.Time `db:"created_at"`
}

//swagger:model OperationLogRequest
//OperationLogRequest represents a request body for getting user operation log
//
type OperationLogRequest struct {
	//identifier of user to get his operation log
	//required: true
	//example: 1
	UserId int64 `json:"user_id"`
	//field name to order operations by
	//enum: ["date", "amount"]
	//default: date
	//required: false
	//example: date
	OrderField string `json:"order_field"`
	//operations order direction
	//enum: ["desc", "asc"]
	//default: desc
	//required: false
	OrderDirection string `json:"order_direction"`
	//desired page of operation log
	//default: 1
	//required: false
	Page int64 `json:"page"`
	//limit the number of operations per page
	//default: -1
	//required: false
	Limit int64 `json:"limit"`
}

// swagger:model UserBalance
//UserBalance represent a response body for user balance request
//
type UserBalance struct {
	//User balance in requested currency
	//example: 100
	Balance string `json:"balance"`
	//currency name of given balance value
	//example: RUB
	Currency string `json:"currency"`
}

// swagger:model OperationsLog
//OperationsLog represents a response body page of user operations log
//
type OperationsLog struct {
	//number of user operation
	OperationsNum int64 `json:"operations_num"`
	//List of user operations
	Operations []Operation `json:"operations"`
	//Current page number
	Page int64 `json:"page"`
	//total amount of operation log pages
	PagesTotal int64 `json:"pages_total"`
}

// swagger:model ResultState
// represents a message about successful operation
type ResultState struct {
	// Example: account crediting done
	//example: Money transfer operation done
	State string `json:"state"`
}
