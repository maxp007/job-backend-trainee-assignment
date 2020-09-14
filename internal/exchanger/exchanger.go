package exchanger

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"job-backend-trainee-assignment/internal/logger"
	"net/http"
)

type RequestDoer interface {
	Do(*http.Request) (*http.Response, error)
}

type ICurrencyExchanger interface {
	GetAmountInCurrency(ctx context.Context, amount decimal.Decimal, targetCurrencyName string) (
		amountInCurrency *decimal.Decimal, err error)
}

const StackExchangeApiURL = "https://api.exchangeratesapi.io/latest?base=%s"

//CurrencyExchanger implementation using remote service
type CurrencyExchanger struct {
	client       RequestDoer
	logger       logger.ILogger
	baseCurrency string
	exchangeURL  string
}

type ExchangeRates struct {
	Rates map[string]float64 `json:"rates"`
	Base  string             `json:"base"`
	Date  string             `json:"date"`
}

type ErrorResponseBody struct {
	Err string `json:"error"`
}

func NewExhanger(logger logger.ILogger, requestDoer RequestDoer, baseCurrency string) (*CurrencyExchanger, error) {

	exURL := fmt.Sprintf(StackExchangeApiURL, baseCurrency)
	return &CurrencyExchanger{logger: logger, client: requestDoer, baseCurrency: baseCurrency, exchangeURL: exURL}, nil
}

func (ce *CurrencyExchanger) GetAmountInCurrency(ctx context.Context, amount decimal.Decimal,
	targetCurrencyName string) (*decimal.Decimal, error) {



	if targetCurrencyName == ce.baseCurrency {
		return &amount, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ce.exchangeURL, nil)
	if err != nil {
		ce.logger.Error("Failed to create currency rates request  at:%s, err:%v", ce.exchangeURL, err)
		return nil, fmt.Errorf("NewRequest err: %w", err)
	}

	resp, err := ce.client.Do(req)
	if err != nil {
		ce.logger.Error("Failed to get currency rates at:%s, err:%v, %v", ce.exchangeURL, err)
		return nil, fmt.Errorf("client.Get err: %v, err: %w", err, ErrRequestDoerError)
	}

	resBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ce.logger.Error("Failed to read currency rates response body ,at:%s, err:%v", ce.exchangeURL, err)
		return nil, fmt.Errorf("ReadAll(resp.Body) err: %v, err:%w", err, ErrResponseBodyReadFailed)
	}

	if resp.StatusCode != http.StatusOK {
		errBody := &ErrorResponseBody{}
		err = json.Unmarshal(resBytes, errBody)
		if err != nil {
			ce.logger.Error("Failed to unmarshal error response body ,at:%s, err:%v", ce.exchangeURL, err)
			return nil, fmt.Errorf("reponse error json Unmarshal err: %w", ErrResponseJSONUnmarshalFailed)
		}

		if errBody.Err == fmt.Sprintf("Base '%s' is not supported.", ce.baseCurrency) {
			ce.logger.Error("base currency %s is not supported by remote service at:%s", ce.baseCurrency, ce.exchangeURL)
			return nil, fmt.Errorf("base currency %s is not supported by remote service, err: %w", ce.baseCurrency, ErrBaseCurrencyNameNotFound)
		}

		ce.logger.Error("got non-ok status code from exchange rates service at:%s, err:%s", ce.exchangeURL, errBody.Err)
		return nil, fmt.Errorf("got non-ok status code from exchange rates service err: %s, %w", errBody.Err, ErrErrorResponseUnknownError)
	}

	exchangeRatesResult := &ExchangeRates{}
	err = json.Unmarshal(resBytes, exchangeRatesResult)
	if err != nil {
		ce.logger.Error("Failed to unmarshal currency rates response body ,at:%s, err:%v", ce.exchangeURL, err)
		return nil, fmt.Errorf("exchangeRates json Unmarshal err: %w", ErrResponseJSONUnmarshalFailed)
	}

	found := false
	var rate float64
	for currName, currRate := range exchangeRatesResult.Rates {
		if currName == targetCurrencyName {
			found = true
			rate = currRate
			break
		}
	}

	if !found {
		ce.logger.Error("failed to convert base to target currency, currency with name %s was not found", targetCurrencyName)
		return nil, fmt.Errorf("unable to find specified currency name:%s in response, err:%w", exchangeRatesResult.Base, ErrTargetCurrencyNameNotFound)
	}

	if amount.IsZero() {
		return &amount, nil
	}

	amountInCurrency := amount.Mul(decimal.NewFromFloat(rate)).RoundBank(2)
	return &amountInCurrency, nil
}

