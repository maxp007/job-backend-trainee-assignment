package http_handler

import (
	"fmt"
	"log"
	"net/http"
)

func ContentTypeValidatorMiddleware(handlerFunc http.HandlerFunc, contentType string) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h := r.Header.Get("Content-Type"); h != contentTypeApplicationJson {
			log.Printf("got wrong content type on Path %s, host %s, method:%s, content-type:%s", r.URL, r.Host, r.Method, r.Header.Get("Content-Type"))
			err := WriteResponse(w, nil, fmt.Errorf(handlerErrUnsupportedContentType),
				http.StatusUnsupportedMediaType)
			if err != nil {
				log.Printf("failed to write response on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
			}
			return
		}
		handlerFunc(w, r)
	})
}

func MethodValidationMiddleware(handlerFunc http.HandlerFunc, contentType string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			log.Printf("got wrong method on Path %s, host %s, method:%s", r.URL, r.Host, r.Method)

			err := WriteResponse(w, nil, fmt.Errorf(handlerErrUnsupportedMethod), http.StatusMethodNotAllowed)
			if err != nil {
				log.Printf("HandlerCreateNewUser, failed to write response on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
			}
			return
		}

	})

}
