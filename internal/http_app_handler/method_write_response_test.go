package http_app_handler

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

type DummyStruct struct {
	SomeStr string `json:"some_str"`
}

func TestWriteResponseMethod(t *testing.T) {
	t.Logf("Given the need to test writeResponse Helper")
	{
		t.Run("testing httpCode bad value", func(t *testing.T) {
			rr := httptest.NewRecorder()
			respResult := &DummyStruct{SomeStr: "some str"}
			httpCode := 100500
			err := WriteResponse(rr, respResult, httpCode)
			assert.ErrorIsf(t, err, ErrBadHttpCodeToResponse, "must get ErrBadHttpCodeToResponse error, got %v", err)
		})

		t.Run("testing response body is nil", func(t *testing.T) {
			rr := httptest.NewRecorder()
			httpCode := http.StatusOK

			err := WriteResponse(rr, nil, httpCode)
			assert.NoError(t, err, "must get no errors")

			responseBody, err := ioutil.ReadAll(rr.Body)
			require.NoError(t, err)

			assert.JSONEq(t, "null", string(responseBody), "must get expected error response")
			assert.EqualValuesf(t, httpCode, rr.Code, "must get expected HttpCode")
		})

	}
}
