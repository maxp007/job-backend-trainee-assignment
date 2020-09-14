package http_handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"job-backend-trainee-assignment/internal/app"
	"log"
	"net/http"
	"time"
)

const requestHandleTimeout = time.Second * 3

func WriteResponse(w http.ResponseWriter, result interface{}, err error, httpCode int) error {

	if httpCode == 0 {
		httpCode = http.StatusOK
	}

	var respBody interface{}
	var b []byte

	if err != nil {
		respBody = ErrorResponseBody{Error: err.Error()}
	} else if result != nil {
		respBody = SuccessResponseBody{Result: result}
	} else {
		log.Printf("WriteResponse, AmbiguousResponseBody err, got result %v, err %v", result, err)

		respBody = ErrorResponseBody{Error: handlerErrAmbiguousResponseBody.Error()}
	}

	b, err = json.Marshal(respBody)
	if err != nil {
		log.Printf("WriteResponse, failed to marshal response err, got result %v, err %v", result, err)
		return fmt.Errorf(handlerErrJsonMarshalFailed.Error())
	}

	w.WriteHeader(httpCode)
	_, err = w.Write(b)
	if err != nil {
		log.Printf("WriteResponse, failed to write response err, got result %v, err %v", result, err)

		return fmt.Errorf(handlerErrResponseWriteFailed.Error())
	}
	return nil
}

func (h *AppHttpHandler) HandlerGetUserBalance(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestHandleTimeout)
	defer cancel()

	params := &app.BalanceRequest{}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		h.logger.Printf("failed to request read body on Path %s, host %s, method:%s", r.URL, r.Host, r.Method)
		err := WriteResponse(w, nil, handlerErrRequestBodyReadFailed, http.StatusBadRequest)
		if err != nil {
			h.logger.Printf("HandlerGetUserBalance,failed to write response on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
		}
		return
	}

	err = json.Unmarshal(b, params)
	if err != nil {
		h.logger.Printf("failed to unmarshal request body on Path %s, host %s, method:%s", r.URL, r.Host, r.Method)
		err := WriteResponse(w, nil, handlerErrJsonUnmarshalFailed, http.StatusBadRequest)
		if err != nil {
			h.logger.Printf("HandlerGetUserBalance, failed to write response on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
		}
		return
	}

	var httpCode int

	result, err := h.app.GetUserBalance(ctx, params)
	if err != nil {
		h.logger.Printf("HandlerGetUserBalance err, on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
		err = WriteResponse(w, nil, err, httpCode)
		if err != nil {
			h.logger.Printf("HandlerGetUserBalance,  failed to write response on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
		}
		return
	}
	err = WriteResponse(w, result, nil, httpCode)
	if err != nil {
		h.logger.Printf("HandlerGetUserBalance, failed to write response on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
	}

}

func (h *AppHttpHandler) HandlerCreditUserAccount(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestHandleTimeout)
	defer cancel()

	params := &app.CreditAccountRequest{}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		h.logger.Printf("failed to request read body on Path %s, host %s, method:%s", r.URL, r.Host, r.Method)
		err := WriteResponse(w, nil, handlerErrRequestBodyReadFailed, http.StatusBadRequest)
		if err != nil {
			h.logger.Printf("HandlerCreditUserAccount,failed to write response on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
		}
		return
	}

	err = json.Unmarshal(b, params)
	if err != nil {
		h.logger.Printf("failed to unmarshal request body on Path %s, host %s, method:%s", r.URL, r.Host, r.Method)
		err := WriteResponse(w, nil, handlerErrJsonUnmarshalFailed, http.StatusBadRequest)
		if err != nil {
			h.logger.Printf("HandlerCreditUserAccount, failed to write response on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
		}
		return
	}

	var httpCode int

	result, err := h.app.CreditUserAccount(ctx, params)
	if err != nil {
		h.logger.Printf("HandlerCreditUserAccount err, on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
		err = WriteResponse(w, nil, err, httpCode)
		if err != nil {
			h.logger.Printf("HandlerCreditUserAccount,  failed to write response on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
		}
		return
	}

	err = WriteResponse(w, result, nil, httpCode)
	if err != nil {
		h.logger.Printf("HandlerCreditUserAccount, failed to write response on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
	}

}

func (h *AppHttpHandler) HandlerWithdrawUserAccount(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestHandleTimeout)
	defer cancel()

	params := &app.WithdrawAccountRequest{}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		h.logger.Printf("failed to request read body on Path %s, host %s, method:%s", r.URL, r.Host, r.Method)
		err := WriteResponse(w, nil, handlerErrRequestBodyReadFailed, http.StatusBadRequest)
		if err != nil {
			h.logger.Printf("HandlerWithdrawUserAccount,failed to write response on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
		}
		return
	}

	err = json.Unmarshal(b, params)
	if err != nil {
		h.logger.Printf("failed to unmarshal request body on Path %s, host %s, method:%s", r.URL, r.Host, r.Method)
		err := WriteResponse(w, nil, handlerErrJsonUnmarshalFailed, http.StatusBadRequest)
		if err != nil {
			h.logger.Printf("HandlerWithdrawUserAccount, failed to write response on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
		}
		return
	}

	var httpCode int

	result, err := h.app.WithdrawUserAccount(ctx, params)
	if err != nil {
		h.logger.Printf("HandlerWithdrawUserAccount err, on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
		err = WriteResponse(w, nil, err, httpCode)
		if err != nil {
			h.logger.Printf("HandlerWithdrawUserAccount,  failed to write response on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
		}
		return
	}

	err = WriteResponse(w, result, nil, httpCode)
	if err != nil {
		h.logger.Printf("HandlerWithdrawUserAccount, failed to write response on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
	}

}

func (h *AppHttpHandler) HandlerTransferUserMoney(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestHandleTimeout)
	defer cancel()

	params := &app.MoneyTransferRequest{}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		h.logger.Printf("failed to request read body on Path %s, host %s, method:%s", r.URL, r.Host, r.Method)
		err := WriteResponse(w, nil, handlerErrRequestBodyReadFailed, http.StatusBadRequest)
		if err != nil {
			h.logger.Printf("HandlerTransferUserMoney,failed to write response on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
		}
		return
	}

	err = json.Unmarshal(b, params)
	if err != nil {
		h.logger.Printf("failed to unmarshal request body on Path %s, host %s, method:%s", r.URL, r.Host, r.Method)
		err := WriteResponse(w, nil, handlerErrJsonUnmarshalFailed, http.StatusBadRequest)
		if err != nil {
			h.logger.Printf("HandlerTransferUserMoney, failed to write response on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
		}
		return
	}

	var httpCode int

	result, err := h.app.TransferMoneyFromUserToUser(ctx, params)
	if err != nil {
		h.logger.Printf("HandlerTransferUserMoney err, on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
		err = WriteResponse(w, nil, err, httpCode)
		if err != nil {
			h.logger.Printf("HandlerTransferUserMoney,  failed to write response on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
		}
		return
	}

	err = WriteResponse(w, result, nil, httpCode)
	if err != nil {
		h.logger.Printf("HandlerTransferUserMoney, failed to write response on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
	}
}
