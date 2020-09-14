package http_handler

import "errors"

var (
	handlerErrJsonMarshalFailed      = errors.New("failed to marshal json request")
	handlerErrJsonUnmarshalFailed    = errors.New("failed to unmarshal json response")
	handlerErrResponseWriteFailed    = errors.New("failed to write response")
	handlerErrRequestBodyReadFailed  = errors.New("failed to read request body")
	handlerErrUnsupportedContentType = errors.New("unsupported request content type")
	handlerErrUnsupportedMethod      = errors.New("unsupported request method")
	handlerErrUnknownError           = errors.New("got unknown internal error")
	handlerErrAmbiguousResponseBody  = errors.New("got ambiguous response data to send")
	handlerErrRequestTimeout         = errors.New("request processin timeout exceeded")
)

