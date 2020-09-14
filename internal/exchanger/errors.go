package exchanger

import (
	"errors"
	"fmt"
)

var (
	ErrContextCancelled        = fmt.Errorf("context canceled")
	ErrContextDeadlineExceeded = fmt.Errorf("context deadline exceeded")

	ErrTargetCurrencyNameNotFound  = errors.New("target currency with given name was not found")
	ErrExchangeServiceRequestError = errors.New("failed to get exchange rates from remote service")
	ErrBaseCurrencyNameNotFound    = errors.New("base currency with given name was not found")
	ErrAmountParamPtrIsNil         = errors.New("got nil ptr in amount param")
	ErrResponseJSONUnmarshalFailed = errors.New("failed to unmarshal response body")
	ErrResponseBodyReadFailed      = errors.New("failed to read response body")
	ErrNewRequestCreateFailed      = errors.New("failed to unmarshal response body")
	ErrErrorResponseUnknownError   = errors.New("got response with unknown error")
	ErrRequestDoerError            = errors.New("error occured in request doer")
)
