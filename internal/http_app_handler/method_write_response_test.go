package http_app_handler

import (
	"encoding/json"
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

			var respResult *DummyStruct = &DummyStruct{SomeStr: "some str"}
			var respErr error = nil
			httpCode := 100500
			err := WriteResponse(rr, respResult, respErr, httpCode)
			assert.ErrorIsf(t, err, ErrBadHttpCodeToResponse, "must get ErrBadHttpCodeToResponse error, got %v", err)
		})

		t.Run("testing result arg is nil ", func(t *testing.T) {
			rr := httptest.NewRecorder()

			var respResult *DummyStruct  = nil
			var respErr error = ErrUnknownError
			httpCode := http.StatusInternalServerError
			expectedResp :=ErrorResponseBody{Error: ErrUnknownError.Error()}

			err := WriteResponse(rr, respResult, respErr, httpCode)
			assert.NoError(t, err, "must get no errors")

			responseBody, err := ioutil.ReadAll(rr.Body)
			require.NoError(t, err)

			expectedBody, err := json.Marshal(expectedResp)
			require.NoError(t, err)
			assert.JSONEq(t,string(expectedBody),string(responseBody), "must get expected error response")
			assert.EqualValuesf(t, httpCode, rr.Code,"must get expected HttpCode")
		})

		t.Run("testing err arg is nil", func(t *testing.T) {
			rr := httptest.NewRecorder()

			var respResult *DummyStruct  = &DummyStruct{SomeStr: "Some string"}
			var respErr error = nil
			httpCode := http.StatusOK
			expectedResp :=SuccessResponseBody{Result: respResult}

			err := WriteResponse(rr, respResult, respErr, httpCode)
			assert.NoError(t, err, "must get no errors")

			responseBody, err := ioutil.ReadAll(rr.Body)
			require.NoError(t, err)

			expectedBody, err := json.Marshal(expectedResp)
			require.NoError(t, err)
			assert.JSONEq(t,string(expectedBody),string(responseBody), "must get expected error response")
			assert.EqualValuesf(t, httpCode, rr.Code,"must get expected HttpCode")

		})

		t.Run("testing both result and err args are nil", func(t *testing.T) {
			rr := httptest.NewRecorder()

			var respResult *DummyStruct  = nil
			var respErr error = nil
			httpCode := http.StatusOK

			err := WriteResponse(rr, respResult, respErr, httpCode)
			assert.NoError(t, err, "must get no errors")

			responseBody, err := ioutil.ReadAll(rr.Body)
			require.NoError(t, err)

			assert.JSONEq(t,"{\"result\":null}",string(responseBody), "must get expected error response")
			assert.EqualValuesf(t, httpCode, rr.Code,"must get expected HttpCode")
		})

		t.Run("testing both result and err args are non-nil", func(t *testing.T) {
			rr := httptest.NewRecorder()

			var respResult *DummyStruct = &DummyStruct{SomeStr: "some str"}
			var respErr error = ErrAmbiguousResponseBody
			httpCode := http.StatusInternalServerError
			expectedResp := ErrorResponseBody{Error: ErrAmbiguousResponseBody.Error()}

			err := WriteResponse(rr, respResult, respErr, httpCode)
			assert.NoError(t, err, "must get no errors")

			responseBody, err := ioutil.ReadAll(rr.Body)
			require.NoError(t, err)

			expectedBody, err := json.Marshal(expectedResp)
			require.NoError(t, err)
			assert.JSONEq(t,string(expectedBody),string(responseBody), "must get expected error response")
			assert.EqualValuesf(t, httpCode, rr.Code,"must get expected HttpCode")

		})

	}
}
