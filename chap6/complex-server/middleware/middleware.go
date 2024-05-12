package middleware

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/PaulOh5/complex-server/config"
)

type requestIDContextKey struct{}
type requestIDContextValue struct {
	requestID string
}

func loggingMiddleware(h http.Handler, c config.AppConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		h.ServeHTTP(w, r)
		requestDuration := time.Since(startTime)
		c.Logger.Printf(
			"protocol=%s path=%s method=%s duration=%f",
			r.Proto, r.URL.Path,
			r.Method, requestDuration.Seconds(),
		)
	})
}

func panicMiddleware(h http.Handler, c config.AppConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rValue := recover(); rValue != nil {
				c.Logger.Println("panic detected", rValue)
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Unexpected server error")
			}
		}()
		h.ServeHTTP(w, r)
	})
}

func attachIDMiddleware(h http.Handler, counter *int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt64(counter, 1)
		requestID := fmt.Sprintf("Request(%d)", count)
		requestIDContext := requestIDContextValue{requestID: requestID}
		currentCtx := r.Context()
		newCtx := context.WithValue(currentCtx, requestIDContextKey{}, requestIDContext)
		r = r.WithContext(newCtx)
		h.ServeHTTP(w, r)
	})
}
