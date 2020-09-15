package router

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"job-backend-trainee-assignment/internal/logger"
	"net/http"
)

type IRouter interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	HandlerFunc(httpMethod string, path string, handle http.HandlerFunc)
	SetMethodNotAllowedHandler(h http.HandlerFunc)
}

type HttpRouter struct {
	logger logger.ILogger
	router *httprouter.Router
}

func (hr *HttpRouter) HandlerFunc(httpMethod string, path string, handle http.HandlerFunc) {
	hr.router.HandlerFunc(httpMethod, path, handle)
}

func (hr *HttpRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hr.router.ServeHTTP(w, r)
}

func (hr *HttpRouter) SetMethodNotAllowedHandler(h http.HandlerFunc) {
	hr.router.MethodNotAllowed = h
}

func NewRouter(logger logger.ILogger) (*HttpRouter, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger must be not nil")
	}

	r := httprouter.New()
	return &HttpRouter{router: r, logger: logger}, nil
}
