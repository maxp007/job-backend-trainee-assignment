package router

import "net/http"

type DummyRouter struct{}

func (mr *DummyRouter) ServeHTTP(w http.ResponseWriter, r *http.Request)                    {}
func (mr *DummyRouter) HandlerFunc(httpMethod string, path string, handle http.HandlerFunc) {}
func (mr *DummyRouter) SetMethodNotAllowedHandler(handlerFunc http.HandlerFunc)             {}
