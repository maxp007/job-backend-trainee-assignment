package exchanger

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

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
