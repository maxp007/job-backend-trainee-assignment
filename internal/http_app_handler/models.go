package http_app_handler


//Error response wrapper
//swagger:model ErrorResponseBody
type ErrorResponseBody struct {
	// Example: "user with given id does not exist"
	Error string `json:"error"`
}

//Successful results wrapper
//swagger:model SuccessResponseBody
type SuccessResponseBody struct {
	// Example: "{"balance":"100", "currency":"RUB"}"
	Result interface{} `json:"result"`
}


