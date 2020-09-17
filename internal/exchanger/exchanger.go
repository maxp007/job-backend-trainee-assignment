package exchanger

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"job-backend-trainee-assignment/internal/logger"
	"net/http"
	"sync"
	"time"
)

type RequestDoer interface {
	Do(*http.Request) (*http.Response, error)
}

type ICurrencyExchanger interface {
	GetAmountInCurrency(ctx context.Context, amount decimal.Decimal, targetCurrencyName string) (
		amountInCurrency *decimal.Decimal, err error)
}

const (
	USDCode   = "USD"
	RUBCode   = "RUB"
	EURCode   = "EUR"
	layoutISO = "2006-01-02"
)

const StackExchangeApiURL = "https://api.exchangeratesapi.io/latest?base=%s"

//CurrencyExchanger implementation using remote service
type CurrencyExchanger struct {
	client       RequestDoer
	logger       logger.ILogger
	baseCurrency string
	exchangeURL  string
	cachedResult *ExchangeRates
	cachedTime   time.Time
	mu           sync.Mutex
}

func NewExchanger(logger logger.ILogger, requestDoer RequestDoer, baseCurrency string) (*CurrencyExchanger, error) {
	exURL := fmt.Sprintf(StackExchangeApiURL, baseCurrency)
	return &CurrencyExchanger{logger: logger, client: requestDoer, baseCurrency: baseCurrency, exchangeURL: exURL}, nil
}

func (ce *CurrencyExchanger) GetAmountInCurrency(ctx context.Context, amount decimal.Decimal,
	targetCurrencyName string) (*decimal.Decimal, error) {

	baseCurrency := ""
	exchangeURL := ""
	var cachedTime time.Time
	cachedResultIsNil := true

	ce.mu.Lock()
	{
		cachedTime = ce.cachedTime
		baseCurrency = ce.baseCurrency
		exchangeURL = ce.exchangeURL
		if ce.cachedResult != nil {
			cachedResultIsNil = false
		}
	}
	ce.mu.Unlock()

	if time.Since(cachedTime).Minutes() >= 24*60 || cachedResultIsNil {
		ce.logger.Info("updating cached ExchangeRates value")
		if targetCurrencyName == baseCurrency {
			return &amount, nil
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, exchangeURL, nil)
		if err != nil {
			ce.logger.Error("Failed to create currency rates request  at:%s, err:%v", exchangeURL, err)
			return nil, fmt.Errorf("NewRequest err: %w", err)
		}

		resp, err := ce.client.Do(req)
		if err != nil {
			ce.logger.Error("Failed to get currency rates at:%s, err:%v, %v", exchangeURL, err)
			return nil, fmt.Errorf("client.Get err: %v, err: %w", err, ErrRequestDoerError)
		}

		resBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			ce.logger.Error("Failed to read currency rates response body ,at:%s, err:%v", exchangeURL, err)
			return nil, fmt.Errorf("ReadAll(resp.Body) err: %v, err:%w", err, ErrResponseBodyReadFailed)
		}

		if resp.StatusCode != http.StatusOK {
			errBody := &ErrorResponseBody{}
			err = json.Unmarshal(resBytes, errBody)
			if err != nil {
				ce.logger.Error("Failed to unmarshal error response body ,at:%s, err:%v", exchangeURL, err)
				return nil, fmt.Errorf("reponse error json Unmarshal err: %w", ErrResponseJSONUnmarshalFailed)
			}

			if errBody.Err == fmt.Sprintf("Base '%s' is not supported.", baseCurrency) {
				ce.logger.Error("base currency %s is not supported by remote service at:%s", baseCurrency, exchangeURL)
				return nil, fmt.Errorf("base currency %s is not supported by remote service, err: %w", baseCurrency, ErrBaseCurrencyNameNotFound)
			}

			ce.logger.Error("got non-ok status code from exchange rates service at:%s, err:%s", exchangeURL, errBody.Err)
			return nil, fmt.Errorf("got non-ok status code from exchange rates service err: %s, %w", errBody.Err, ErrErrorResponseUnknownError)
		}
		exchangeRatesResult := &ExchangeRates{}
		err = json.Unmarshal(resBytes, exchangeRatesResult)
		if err != nil {
			ce.logger.Error("Failed to unmarshal currency rates response body ,at:%s, err:%v", exchangeURL, err)
			return nil, fmt.Errorf("exchangeRates json Unmarshal err: %w", ErrResponseJSONUnmarshalFailed)
		}

		ratesLastUpdatedDate, err := time.Parse(layoutISO, exchangeRatesResult.Date)
		if err != nil {
			ce.logger.Error("Failed to parse current rates date ,at:%s, err:%v", exchangeURL, err)
		}

		ce.mu.Lock()
		ce.cachedTime = ratesLastUpdatedDate
		ce.cachedResult = exchangeRatesResult
		ce.mu.Unlock()
	} else {
		ce.logger.Info("using cached ExchangeRates value")
	}

	ce.mu.Lock()
	result := ce.cachedResult
	ce.mu.Unlock()

	found := false
	var rate float64
	for currName, currRate := range result.Rates {
		if currName == targetCurrencyName {
			found = true
			rate = currRate
			break
		}
	}

	if !found {
		ce.logger.Error("failed to convert base to target currency, currency with name %s was not found", targetCurrencyName)
		return nil, fmt.Errorf("unable to find specified currency name:%s in rates, err:%w", result.Base, ErrTargetCurrencyNameNotFound)
	}

	if amount.IsZero() {
		return &amount, nil
	}

	amountInCurrency := amount.Mul(decimal.NewFromFloat(rate)).RoundBank(2)
	return &amountInCurrency, nil
}
