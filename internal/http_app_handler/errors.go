package http_app_handler

import "errors"

var (
	ErrJsonMarshalFailed      = errors.New("failed to marshal json response")
	ErrJsonUnmarshalFailed    = errors.New("failed to unmarshal json request")
	ErrResponseWriteFailed    = errors.New("failed to write response")
	ErrRequestBodyReadFailed  = errors.New("failed to read request body")
	ErrUnsupportedContentType = errors.New("unsupported request content type")
	ErrUnsupportedMethod      = errors.New("unsupported request method")
	ErrUnknownError           = errors.New("got unknown internal error")
	ErrAmbiguousResponseBody  = errors.New("got ambiguous response data to send")
	ErrRequestTimeout         = errors.New("request processing timeout exceeded")
	ErrBadHttpCodeToResponse  = errors.New("got invalid http code value to respond")

	)

