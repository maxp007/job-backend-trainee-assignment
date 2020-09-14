package exchanger


type ExchangeRates struct {
	Rates map[string]float64 `json:"rates"`
	Base  string             `json:"base"`
	Date  string             `json:"date"`
}

type ErrorResponseBody struct {
	Err string `json:"error"`
}