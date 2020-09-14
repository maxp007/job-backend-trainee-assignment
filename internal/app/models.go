package app

import "time"

//BalanceRequest represents a request body for getting balance of user with specified Id
//
//swagger:model
type BalanceRequest struct {
	//identifier of user whom balance is required to get
	//required: true
	UserId int64 `json:"user_id"`
	//name of currency in which balance value is required
	//required: false
	//enum: ["RUB","USD","EUR"]
	//default:RUB
	Currency string `json:"currency"`
}

//UserBalance represent a response body for user balance request
//
//swagger:model
type UserBalance struct {
	//User balance in requested currency
	Balance float64 `json:"balance"`
	//currency name of given balance value
	Currency string `json:"currency"`
}

//TransactionLogRequest represents a request body for getting user transaction log
//
//swagger:model
type TransactionLogRequest struct {
	//identifier of user to get his transaction log
	//required: true
	UserId int64 `json:"user_id"`
	//field name to order transaction by
	//enum: ["date", "amount"]
	//default: "date"
	//required: false
	OrderField string `json:"order_field"`
	//transactions order direction
	//enum: ["desc", "asc"]
	//default: "desc"
	//required: false
	OrderDirection string `json:"order_direction"`
}

//TransactionLog represents a response body page of user transactions log
//
//swagger:model
type TransactionLog struct {
	//List of user transactions
	Transactions []Transaction `json:"transactions"`
	//Current page number
	Page int64 `json:"page"`
	//total amount of transaction log pages
	PagesTotal int64 `json:"pages_total"`
}

//Transaction represents a transaction model used to store user transactions
//swagger:model
type Transaction struct {
	//transaction identifier
	Id int64 `json:"id" db:"transaction_id"`
	//identifier of party sending money
	SenderId int64 `json:"sender_id" db:"sender_id"`
	//identifier of party receiving money
	ReceiverID int64 `json:"receiver_id" db:"receiver_id"`
	//purpose of transaction
	Purpose string `json:"purpose" db:"purpose"`
	//amount of money to be sent or received after transaction
	Amount float64 `json:"amount" db:"amount"`
	//transaction creating date
	CreatedAt time.Duration `json:"created_at" db:"created_at"`
}

//OrderToExecute represents a request to perform a money transaction operation
//
//swagger: model
type TransactionRequest struct {
	//identifier of party sending money
	//required: true
	SenderId int64 `json:"sender_id"`
	//identifier of party receiving money
	//required: true
	ReceiverId int64 `json:"receiver_id"`
	//amount of money to be sent as a result of transaction
	//required: true
	//minimum: 1.00
	Amount float64 `json:"amount"`
}

//swagger:model
type TakeMoneyRequest struct {
	UserId  int64   `json:"user_id"`
	Purpose string  `json:"purpose"`
	Amount  float64 `json:"amount"`
}

//swagger:model
type GiveMoneyRequest struct {
	UserId  int64  `json:"user_id"`
	Purpose string `json:"purpose"`
	Amount  int64  `json:"amount"`
}

//Party is a name of transaction subjects
// party can either be a user or internal avito service
type Party struct {
	//party identifier
	Id int64 `json:"id" db:"party_id"`
	//party creation date
	CreatedAt time.Duration `json:"created_at" db:"created_at"`
}

//User represents a user model stored in a database
//swagger:model
type User struct {
	//User identifier
	Id int64 `json:"id" db:"user_id"`
	//current user balance in russian roubles
	//minimum: 0.00
	Balance float64 `json:"balance" db:"user_balance"`
	//user name to be shown to other users
	Name string `json:"name" db:"name"`
}

//AvitoService represents avito service model stored in database
//the purpose of using this model is to distinguish transactions between real users
// and the transaction between user and the internal avito services
//swagger:model
type AvitoService struct {
	//avito service name to be shown to user
	Name string `json:"service_name" db:"service_name"`
}

type ResultState struct {
	State string `json:"state"`
}
