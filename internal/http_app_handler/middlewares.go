package http_app_handler

import (
	"net/http"
	"time"
)

func (h *AppHttpHandler) ContentTypeValidationMW(handlerFunc http.HandlerFunc, contentType string) http.HandlerFunc {
	logger := h.logger
	return func(w http.ResponseWriter, r *http.Request) {
		if h := r.Header.Get("Content-Type"); h != contentType {
			logger.Error("ContentTypeValidationMW, got wrong content type on Path %s, host %s, method:%s, content-type:%s", r.URL, r.Host, r.Method, r.Header.Get("Content-Type"))
			err := WriteResponse(w, nil, ErrUnsupportedContentType,
				http.StatusUnsupportedMediaType)
			if err != nil {
				logger.Error("ContentTypeValidationMW, failed to write response on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
			}

			return
		}

		handlerFunc(w, r)
	}
}

func (h *AppHttpHandler) MethodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	h.logger.Error("MethodNotAllowedHandler, got wrong method on Path %s, host %s, method:%s", r.URL, r.Host, r.Method)
	err := WriteResponse(w, nil, ErrUnsupportedMethod, http.StatusMethodNotAllowed)
	if err != nil {
		h.logger.Error("MethodValidMiddleware, failed to write response on Path %s, host %s, method:%s, err:%s", r.URL, r.Host, r.Method, err.Error())
	}

	return
}

func (h *AppHttpHandler) AccessLogMW(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		handler.ServeHTTP(w, r)
		h.logger.Info("AccessLogMW, remote_host: %s, method: %s, url: %s, elapsed_time: %s", r.Host, r.Method, r.URL, time.Since(start))
	}
}
