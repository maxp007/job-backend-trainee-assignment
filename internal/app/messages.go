package app

var (
	MsgAccountCreditingDone = "Account crediting Done"
	MsgAccountWithdrawDone  = "Account withdraw Done"
	MsgMoneyTransferDone    = "Money transfer Done"

	OperationTokenIsAlreadyUsed = "Operation with specified token had already been done"
)

//message structure to form user operation comments
var (
	CommentTransferFromServiceWithComment = "payment from service, %s"
	CommentTransferToServiceWithComment   = "payment to service, %s"
	CommentTransferToUserWithName         = "transfer to user %s"
	CommentTransferFromUserWithName       = "transfer from user %s"
)
