package http_app_handler

import (
	"context"
	"encoding/json"
	"job-backend-trainee-assignment/internal/app"
	"log"
	"net/http"
	"time"
)


func WriteResponse(w http.ResponseWriter, result interface{}, err error, httpCode int) error {

	if httpCode == 0 {
		httpCode = http.StatusOK
	}

	if code := http.StatusText(httpCode); code == "" {
		return ErrBadHttpCodeToResponse
	}

	var respBody interface{}
	var b []byte

	if err != nil && result != nil {
		log.Printf("WriteResponse, AmbiguousResponseBody err, got not-nil result %v and not nil err %v", result, err)
		respBody = ErrorResponseBody{Error: ErrAmbiguousResponseBody.Error()}
		httpCode = http.StatusInternalServerError
	}

	if err != nil {
		respBody = ErrorResponseBody{Error: err.Error()}
	} else if result != nil {
		respBody = SuccessResponseBody{Result: result}
	}

	b, err = json.Marshal(respBody)
	if err != nil {
		log.Printf("WriteResponse, failed to marshal response err, got result %v, err %v", result, err)
		return ErrJsonMarshalFailed
	}

	w.WriteHeader(httpCode)
	_, err = w.Write(b)
	if err != nil {
		log.Printf("WriteResponse, failed to write response err, got result %v, err %v", result, err)
		return ErrResponseWriteFailed
	}
	return nil

}


// swagger:route POST /balance methods GetUserBalance
// Returns balance of user with given id.
// responses:
//   200: BalanceResponseBody (UserBalance model, wrapped in SuccessResponseBody)
// 	 400: ErrorResponseBody
// 	 500: ErrorResponseBody
func (h *AppHttpHandler) HandlerGetUserBalance(w http.ResponseWriter, r *http.Request) {
	var requestHandleTimeout time.Duration

	h.mu.Lock()
	requestHandleTimeout = h.cfg.RequestHandleTimeout
	h.mu.Unlock()

	ctx, cancel := context.WithTimeout(r.Context(), requestHandleTimeout)
	defer cancel()

	params := &app.BalanceRequest{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	defer r.Body.Close()

	err := d.Decode(params)
	if err != nil {
		h.logger.Error("failed to unmarshal request body on Path %s, host %s, method:%s", r.URL, r.Host, r.Method)
		err := WriteResponse(w, nil, ErrJsonUnmarshalFailed, http.StatusBadRequest)
		if err != nil {
			h.logger.Error("HandlerGetUserBalance, failed to write response on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
		}
		return
	}

	var httpCode int

	result, err := h.app.GetUserBalance(ctx, params)
	if err != nil {
		if appErr, ok := err.(*app.AppError); ok {
			httpCode = appErr.Code
		}

		h.logger.Error("HandlerGetUserBalance err, on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
		err = WriteResponse(w, nil, err, httpCode)
		if err != nil {
			h.logger.Error("HandlerGetUserBalance,  failed to write response on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
		}
		return
	}

	err = WriteResponse(w, result, nil, httpCode)
	if err != nil {
		h.logger.Error("HandlerGetUserBalance, failed to write response on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
	}

}

// 	swagger:route POST /credit methods CreditUserAccount
// 	Adds given amount of money to given users account.
// 	Responses:
//		200: CreditAccountResponseBody (ResultState model, wrapped in SuccessResponseBody)
// 		400: ErrorResponseBody
// 		500: ErrorResponseBody
func (h *AppHttpHandler) HandlerCreditUserAccount(w http.ResponseWriter, r *http.Request) {
	var requestHandleTimeout time.Duration

	h.mu.Lock()
	requestHandleTimeout = h.cfg.RequestHandleTimeout
	h.mu.Unlock()

	ctx, cancel := context.WithTimeout(r.Context(), requestHandleTimeout)
	defer cancel()

	params := &app.CreditAccountRequest{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	defer r.Body.Close()

	err := d.Decode(params)
	if err != nil {
		h.logger.Error("failed to unmarshal request body on Path %s, host %s, method:%s", r.URL, r.Host, r.Method)
		err := WriteResponse(w, nil, ErrJsonUnmarshalFailed, http.StatusBadRequest)
		if err != nil {
			h.logger.Error("HandlerCreditUserAccount, failed to write response on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
		}
		return
	}

	var httpCode int

	result, err := h.app.CreditUserAccount(ctx, params)
	if err != nil {
		if appErr, ok := err.(*app.AppError); ok {
			httpCode = appErr.Code
		}
		h.logger.Error("HandlerCreditUserAccount err, on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
		err = WriteResponse(w, nil, err, httpCode)
		if err != nil {
			h.logger.Error("HandlerCreditUserAccount,  failed to write response on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
		}
		return
	}

	err = WriteResponse(w, result, nil, httpCode)
	if err != nil {
		h.logger.Error("HandlerCreditUserAccount, failed to write response on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
	}

}

// swagger:route POST /withdraw methods WithdrawUserAccount
// Withdraws given amount of money from given users account.
// 	Responses:
//		200: WithdrawAccountResponseBody (ResultState model, wrapped in SuccessResponseBody)
// 		400: ErrorResponseBody
// 		500: ErrorResponseBody
func (h *AppHttpHandler) HandlerWithdrawUserAccount(w http.ResponseWriter, r *http.Request) {
	var requestHandleTimeout time.Duration

	h.mu.Lock()
	requestHandleTimeout = h.cfg.RequestHandleTimeout
	h.mu.Unlock()

	ctx, cancel := context.WithTimeout(r.Context(), requestHandleTimeout)
	defer cancel()

	params := &app.WithdrawAccountRequest{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	defer r.Body.Close()

	err := d.Decode(params)
	if err != nil {
		h.logger.Error("failed to unmarshal request body on Path %s, host %s, method:%s", r.URL, r.Host, r.Method)
		err := WriteResponse(w, nil, ErrJsonUnmarshalFailed, http.StatusBadRequest)
		if err != nil {
			h.logger.Error("HandlerWithdrawUserAccount, failed to write response on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
		}
		return
	}

	var httpCode int

	result, err := h.app.WithdrawUserAccount(ctx, params)
	if err != nil {
		if appErr, ok := err.(*app.AppError); ok {
			httpCode = appErr.Code
		}
		h.logger.Error("HandlerWithdrawUserAccount err, on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
		err = WriteResponse(w, nil, err, httpCode)
		if err != nil {
			h.logger.Error("HandlerWithdrawUserAccount,  failed to write response on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
		}
		return
	}

	err = WriteResponse(w, result, nil, httpCode)
	if err != nil {
		h.logger.Error("HandlerWithdrawUserAccount, failed to write response on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
	}

}

// swagger:route POST /transfer methods TransferUserMoney
// Transfer user money to another user.
// 	Responses:
//		200: MoneyTransferResponseBody (ResultState model, wrapped in SuccessResponseBody)
// 		400: ErrorResponseBody
// 		500: ErrorResponseBody
func (h *AppHttpHandler) HandlerTransferUserMoney(w http.ResponseWriter, r *http.Request) {
	var requestHandleTimeout time.Duration

	h.mu.Lock()
	requestHandleTimeout = h.cfg.RequestHandleTimeout
	h.mu.Unlock()

	ctx, cancel := context.WithTimeout(r.Context(), requestHandleTimeout)
	defer cancel()

	params := &app.MoneyTransferRequest{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	defer r.Body.Close()

	err := d.Decode(params)
	if err != nil {
		h.logger.Error("failed to decode request body on Path %s, host %s, method:%s", r.URL, r.Host, r.Method)
		err := WriteResponse(w, nil, ErrJsonUnmarshalFailed, http.StatusBadRequest)
		if err != nil {
			h.logger.Error("HandlerTransferUserMoney, failed to write response on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
		}
		return
	}

	var httpCode int
	result, err := h.app.TransferMoneyFromUserToUser(ctx, params)
	if err != nil {
		httpCode = http.StatusInternalServerError
		if appErr, ok := err.(*app.AppError); ok {
			httpCode = appErr.Code
		}
		h.logger.Error("HandlerTransferUserMoney err, on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
		err = WriteResponse(w, nil, err, httpCode)
		if err != nil {
			h.logger.Error("HandlerTransferUserMoney,  failed to write response on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
		}
		return
	}

	err = WriteResponse(w, result, nil, httpCode)
	if err != nil {
		h.logger.Error("HandlerTransferUserMoney, failed to write response on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
	}
}

// swagger:route POST /operations methods GetUserOperationsLog
// Get user operations log.
// 	Responses:
//		200: OperationsLogResponseBody (OperationsLog model, wrapped in SuccessResponseBody)
// 		400: ErrorResponseBody
// 		500: ErrorResponseBody
func (h *AppHttpHandler) HandlerGetUserOperationsLog(w http.ResponseWriter, r *http.Request) {
	var requestHandleTimeout time.Duration

	h.mu.Lock()
	requestHandleTimeout = h.cfg.RequestHandleTimeout
	h.mu.Unlock()

	ctx, cancel := context.WithTimeout(r.Context(), requestHandleTimeout)
	defer cancel()

	params := &app.OperationLogRequest{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	defer r.Body.Close()

	err := d.Decode(params)
	if err != nil {
		h.logger.Error("failed to decode request body on Path %s, host %s, method:%s", r.URL, r.Host, r.Method)
		err := WriteResponse(w, nil, ErrJsonUnmarshalFailed, http.StatusBadRequest)
		if err != nil {
			h.logger.Error("HandlerGetUserOperationsLog, failed to write response on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
		}
		return
	}

	var httpCode int
	result, err := h.app.GetUserOperations(ctx, params)
	if err != nil {
		httpCode = http.StatusInternalServerError
		if appErr, ok := err.(*app.AppError); ok {
			httpCode = appErr.Code
		}

		h.logger.Error("HandlerGetUserOperationsLog err, on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
		err = WriteResponse(w, nil, err, httpCode)
		if err != nil {
			h.logger.Error("HandlerGetUserOperationsLog,  failed to write response on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
		}
		return
	}

	err = WriteResponse(w, result, nil, httpCode)
	if err != nil {
		h.logger.Error("HandlerGetUserOperationsLog, failed to write response on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
	}
}
