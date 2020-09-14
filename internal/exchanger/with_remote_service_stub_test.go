package exchanger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"job-backend-trainee-assignment/internal/logger"
	"net/http"
	"testing"
	"time"
)

type TestCase struct {
	caseName       string
	inAmount       decimal.Decimal
	targetCurrency string
	expectedResult *decimal.Decimal
	expectedError  error
}

type StubRequestDoerCommon struct{}

func (srd *StubRequestDoerCommon) Do(*http.Request) (*http.Response, error) {
	body, _ := json.Marshal(&ExchangeRates{
		Rates: map[string]float64{
			USDCode: 0.05,
			RUBCode: 1,
		},
		Base: RUBCode,
		Date: "2020-08-15",
	})
	return &http.Response{
		Status:     "OK",
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewReader(body)),
	}, nil
}

func TestExchanger_WithStubRequestDoer_Common(t *testing.T) {
	var nilPtrToDecimal *decimal.Decimal = nil
	decimalBaseValue := decimal.NewFromFloat(10.0)
	decimalCurrencyValue := decimal.NewFromFloat(0.5)
	decimalZeroValue := decimal.NewFromFloat(0.0)

	testCases := []TestCase{
		{
			caseName:       "positive path, amount > 0",
			inAmount:       decimalBaseValue,
			targetCurrency: USDCode,
			expectedResult: &decimalCurrencyValue,
			expectedError:  nil,
		},
		{
			caseName:       "positive path, amount = 0",
			inAmount:       decimalZeroValue,
			targetCurrency: USDCode,
			expectedResult: &decimalZeroValue,
			expectedError:  nil,
		},
		{
			caseName:       "positive path, target Currency name = base currency name",
			inAmount:       decimalBaseValue,
			targetCurrency: RUBCode,
			expectedResult: &decimalBaseValue,
			expectedError:  nil,
		},
		{
			caseName:       "negative path, target currency is unknown",
			inAmount:       decimalBaseValue,
			targetCurrency: "UNKNOWN_CURRENCY",
			expectedResult: nilPtrToDecimal,
			expectedError:  ErrTargetCurrencyNameNotFound,
		},
	}

	ex, err := NewExhanger(&logger.DummyLogger{}, &StubRequestDoerCommon{}, RUBCode)
	assert.NoError(t, err, "NewExchanger Must return no errors")

	for caseIdx, testCase := range testCases {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)

		t.Logf("testing case [%d] %s", caseIdx, testCase.caseName)

		amount, err := ex.GetAmountInCurrency(ctx, testCase.inAmount, testCase.targetCurrency)
		cancel()

		assert.ErrorIsf(t, err, testCase.expectedError, "method returned unexpected error")

		if testCase.expectedResult != nil && amount != nil {
			if !amount.Equal(*testCase.expectedResult) {
				t.Errorf("expected Result: %v, got %v", testCase.expectedResult, amount)
			}
		} else if testCase.expectedResult == nil && amount == nil {

		} else {
			t.Errorf("expected Result: %v, got %v", testCase.expectedResult, amount)
		}
	}
}

type StubRequestDoerBadJSONResponseBody struct{}

func (srd *StubRequestDoerBadJSONResponseBody) Do(*http.Request) (*http.Response, error) {

	reader := bytes.NewReader([]byte(""))
	response := ioutil.NopCloser(reader)

	response.Close()
	return &http.Response{
		Status:     "OK",
		StatusCode: http.StatusOK,
		Body:       response,
	}, nil
}
func TestCurrencyExchanger_WithStubRequestDoer_BadJSONResponseBody(t *testing.T) {
	t.Logf("Given the need to test exchanger with bad response body")
	{
		var nilPtrToDecimal *decimal.Decimal = nil
		decimalBaseValue := decimal.NewFromFloat(10.0)

		testCase := TestCase{
			caseName:       "testing bad response body processing",
			inAmount:       decimalBaseValue,
			targetCurrency: "USD",
			expectedResult: nilPtrToDecimal,
			expectedError:  ErrResponseJSONUnmarshalFailed,
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()

		ex, err := NewExhanger(&logger.DummyLogger{}, &StubRequestDoerBadJSONResponseBody{}, RUBCode)
		if err != nil {
			t.Fatalf("\t\tShould not receive err, got \"%v\"", err)
		}

		t.Logf("\tWhen checking exchanger for \"%v\" error", testCase.expectedError)
		{
			amount, err := ex.GetAmountInCurrency(ctx, testCase.inAmount, testCase.targetCurrency)
			if amount == testCase.expectedResult {
				t.Logf("\t\tShould get matching results")
			} else {
				t.Errorf("\t\tShould receive result \"%v\", got \"%v\"", testCase.expectedResult, amount)

			}

			if err != nil && errors.Is(err, testCase.expectedError) {
				t.Logf("\t\tShould receive err: %v", testCase.expectedError)
			} else {
				t.Errorf("\t\tShould receive \"%v\", got \"%v\"", testCase.expectedError, err)
			}
		}
	}
}

type StubRequestDoerBaseCurrencyNotSupported struct{}

func (srd *StubRequestDoerBaseCurrencyNotSupported) Do(*http.Request) (*http.Response, error) {

	reader := bytes.NewReader([]byte(`{"error":"Base 'RUB' is not supported."}`))
	response := ioutil.NopCloser(reader)

	response.Close()
	return &http.Response{
		Status:     "Bad Request",
		StatusCode: http.StatusBadRequest,
		Body:       response,
	}, nil
}
func TestCurrencyExchanger_WithStubRequestDoer_BadRequestBaseCurrency(t *testing.T) {
	t.Logf("Given the need to test exchanger with bad response body")
	{
		var nilPtrToDecimal *decimal.Decimal = nil
		decimalBaseValue := decimal.NewFromFloat(10.0)

		testCase := TestCase{
			caseName:       "testing bad request base currency",
			inAmount:       decimalBaseValue,
			targetCurrency: "USD",
			expectedResult: nilPtrToDecimal,
			expectedError:  ErrBaseCurrencyNameNotFound,
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()

		ex, err := NewExhanger(&logger.DummyLogger{}, &StubRequestDoerBaseCurrencyNotSupported{}, RUBCode)
		if err != nil {
			t.Fatalf("\t\tShould not receive err, got \"%v\"", err)
		}

		t.Logf("\tWhen checking exchanger for \"%v\" error", testCase.expectedError)
		{
			amount, err := ex.GetAmountInCurrency(ctx, testCase.inAmount, testCase.targetCurrency)
			if amount == testCase.expectedResult {
				t.Logf("\t\tShould get matching results")
			} else {
				t.Errorf("\t\tShould receive result \"%v\", got \"%v\"", testCase.expectedResult, amount)

			}

			if err != nil && errors.Is(err, testCase.expectedError) {
				t.Logf("\t\tShould receive err: %v", testCase.expectedError)
			} else {
				t.Errorf("\t\tShould receive \"%v\", got \"%v\"", testCase.expectedError, err)
			}
		}
	}
}

type StubRequestDoerBadJSONErrorResponseBody struct{}

func (srd *StubRequestDoerBadJSONErrorResponseBody) Do(*http.Request) (*http.Response, error) {

	reader := bytes.NewReader([]byte(`{"error":"BAD JSON STRING`))
	response := ioutil.NopCloser(reader)

	response.Close()
	return &http.Response{
		Status:     "Bad Request",
		StatusCode: http.StatusBadRequest,
		Body:       response,
	}, nil
}
func TestCurrencyExchanger_WithStubRequestDoer_BadJSONErrorResponseBody(t *testing.T) {
	t.Logf("Given the need to test exchanger with bad error response body")
	{
		var nilPtrToDecimal *decimal.Decimal = nil
		decimalBaseValue := decimal.NewFromFloat(10.0)

		testCase := TestCase{
			caseName:       "testing bad response body processing",
			inAmount:       decimalBaseValue,
			targetCurrency: "USD",
			expectedResult: nilPtrToDecimal,
			expectedError:  ErrResponseJSONUnmarshalFailed,
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()

		ex, err := NewExhanger(&logger.DummyLogger{}, &StubRequestDoerBadJSONErrorResponseBody{}, RUBCode)
		if err != nil {
			t.Fatalf("\t\tShould not receive err, got \"%v\"", err)
		}

		t.Logf("\tWhen checking exchanger for \"%v\" error", testCase.expectedError)
		{
			amount, err := ex.GetAmountInCurrency(ctx, testCase.inAmount, testCase.targetCurrency)
			if amount == testCase.expectedResult {
				t.Logf("\t\tShould get matching results")
			} else {
				t.Errorf("\t\tShould receive result \"%v\", got \"%v\"", testCase.expectedResult, amount)

			}

			if err != nil && errors.Is(err, testCase.expectedError) {
				t.Logf("\t\tShould receive err: %v", testCase.expectedError)
			} else {
				t.Errorf("\t\tShould receive \"%v\", got \"%v\"", testCase.expectedError, err)
			}
		}
	}
}

type StubRequestDoerWithErrorResponseUnknownError struct{}

func (srd *StubRequestDoerWithErrorResponseUnknownError) Do(*http.Request) (*http.Response, error) {

	reader := bytes.NewReader([]byte(`{"error":"some unknown error"}`))
	response := ioutil.NopCloser(reader)

	response.Close()
	return &http.Response{
		Status:     "Bad Request",
		StatusCode: http.StatusBadRequest,
		Body:       response,
	}, nil
}

func TestCurrencyExchanger_WithStubRequestDoer_ErrorResponseUnknownError(t *testing.T) {
	t.Logf("Given the need to test exchanger with unknown error in response")
	{
		var nilPtrToDecimal *decimal.Decimal = nil
		decimalBaseValue := decimal.NewFromFloat(10.0)

		testCase := TestCase{
			caseName:       "testing unknown error in error response",
			inAmount:       decimalBaseValue,
			targetCurrency: "USD",
			expectedResult: nilPtrToDecimal,
			expectedError:  ErrErrorResponseUnknownError,
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()

		ex, err := NewExhanger(&logger.DummyLogger{}, &StubRequestDoerWithErrorResponseUnknownError{}, RUBCode)
		if err != nil {
			t.Fatalf("\t\tShould not receive err, got \"%v\"", err)
		}

		t.Logf("\tWhen checking exchanger for \"%v\" error", testCase.expectedError)
		{
			amount, err := ex.GetAmountInCurrency(ctx, testCase.inAmount, testCase.targetCurrency)
			if amount == testCase.expectedResult {
				t.Logf("\t\tShould get matching results")
			} else {
				t.Errorf("\t\tShould receive result \"%v\", got \"%v\"", testCase.expectedResult, amount)

			}

			if err != nil && errors.Is(err, testCase.expectedError) {
				t.Logf("\t\tShould receive err: %v", testCase.expectedError)
			} else {
				t.Errorf("\t\tShould receive \"%v\", got \"%v\"", testCase.expectedError, err)
			}
		}
	}
}

type StubRequestDoerWithError struct{}

func (srd *StubRequestDoerWithError) Do(*http.Request) (*http.Response, error) {

	return nil, fmt.Errorf("some error occured while request")
}

func TestCurrencyExchanger_WithStubRequestDoer_WithDoerError(t *testing.T) {
	t.Logf("Given the need to test exchanger with request doer error")
	{
		var nilPtrToDecimal *decimal.Decimal = nil
		decimalBaseValue := decimal.NewFromFloat(10.0)

		testCase := TestCase{
			caseName:       "testing unknown error in error response",
			inAmount:       decimalBaseValue,
			targetCurrency: "USD",
			expectedResult: nilPtrToDecimal,
			expectedError:  ErrRequestDoerError,
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()

		ex, err := NewExhanger(&logger.DummyLogger{}, &StubRequestDoerWithError{}, RUBCode)
		if err != nil {
			t.Fatalf("\t\tShould not receive err, got \"%v\"", err)
		}

		t.Logf("\tWhen checking exchanger for \"%v\" error", testCase.expectedError)
		{
			amount, err := ex.GetAmountInCurrency(ctx, testCase.inAmount, testCase.targetCurrency)
			if amount == testCase.expectedResult {
				t.Logf("\t\tShould get matching results")
			} else {
				t.Errorf("\t\tShould receive result \"%v\", got \"%v\"", testCase.expectedResult, amount)

			}

			if err != nil && errors.Is(err, testCase.expectedError) {
				t.Logf("\t\tShould receive err: %v", testCase.expectedError)
			} else {
				t.Errorf("\t\tShould receive \"%v\", got \"%v\"", testCase.expectedError, err)
			}
		}
	}
}

type StubRequestDoerWithBadResponseBody struct{}

type StubReaderWithError struct {
	bytes []byte
}

func (StubReaderWithError) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("some body reading error")
}

func (srd *StubRequestDoerWithBadResponseBody) Do(*http.Request) (*http.Response, error) {
	someBadBody := ioutil.NopCloser(StubReaderWithError{[]byte("body")})

	return &http.Response{
		Status:     "OK",
		StatusCode: http.StatusOK,
		Body:       someBadBody,
	}, nil
}

func TestCurrencyExchanger_WithStubRequestDoer_WithResponseBodyReadError(t *testing.T) {
	t.Logf("Given the need to test exchanger with body reading error")
	{
		var nilPtrToDecimal *decimal.Decimal = nil
		decimalBaseValue := decimal.NewFromFloat(10.0)

		testCase := TestCase{
			caseName:       "testing response body reading error",
			inAmount:       decimalBaseValue,
			targetCurrency: "USD",
			expectedResult: nilPtrToDecimal,
			expectedError:  ErrResponseBodyReadFailed,
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()

		ex, err := NewExhanger(&logger.DummyLogger{}, &StubRequestDoerWithBadResponseBody{}, RUBCode)
		if err != nil {
			t.Fatalf("\t\tShould not receive err, got \"%v\"", err)
		}

		t.Logf("\tWhen checking exchanger for \"%v\" error", testCase.expectedError)
		{
			amount, err := ex.GetAmountInCurrency(ctx, testCase.inAmount, testCase.targetCurrency)
			if amount == testCase.expectedResult {
				t.Logf("\t\tShould get matching results")
			} else {
				t.Errorf("\t\tShould receive result \"%v\", got \"%v\"", testCase.expectedResult, amount)

			}

			if err != nil && errors.Is(err, testCase.expectedError) {
				t.Logf("\t\tShould receive err: %v", testCase.expectedError)
			} else {
				t.Errorf("\t\tShould receive \"%v\", got \"%v\"", testCase.expectedError, err)
			}
		}
	}
}
