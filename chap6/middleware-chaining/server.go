package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"
)

type appConfig struct {
	logger *log.Logger
}

type app struct {
	config  appConfig
	handler func(w http.ResponseWriter, r *http.Request, config appConfig)
}

func (a app) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.handler(w, r, a.config)
}

type requestContextKey struct{}
type requestContextValue struct {
	requestID string
}

func addRequestID(r *http.Request, requestID string) *http.Request {
	c := requestContextValue{requestID: requestID}
	currentCtx := r.Context()
	newCtx := context.WithValue(currentCtx, requestContextKey{}, c)
	return r.WithContext(newCtx)
}

func apiHandler(w http.ResponseWriter, r *http.Request, config appConfig) {
	config.logger.Println("Handling API request")
	fmt.Fprintf(w, "Hello, world!\n")
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request, config appConfig) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	config.logger.Println("Handling healthcheck request")
	fmt.Fprint(w, "ok\n")
}

func panicHandler(w http.ResponseWriter, r *http.Request, config appConfig) {
	panic("I panicked")
}

func setupHandler(mux *http.ServeMux, config appConfig) {
	mux.Handle("/healthz", &app{config: config, handler: healthCheckHandler})
	mux.Handle("/api", &app{config: config, handler: apiHandler})
	mux.Handle("/panic", &app{config: config, handler: panicHandler})
}

func loggingMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var requestID string
		startTime := time.Now()
		h.ServeHTTP(w, r)
		v := r.Context().Value(requestContextKey{})
		if m, ok := v.(requestContextValue); ok {
			requestID = m.requestID
		}
		log.Printf(
			"ID=%s path=%s method=%s, duration=%f",
			requestID,
			r.URL.Path, r.Method,
			time.Since(startTime).Seconds(),
		)
	})
}

func panicMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rValue := recover(); rValue != nil {
				log.Println("panic detected", rValue)
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
		r = addRequestID(r, fmt.Sprintf("Request(%d)", count))
		h.ServeHTTP(w, r)
	})
}

func main() {
	config := appConfig{
		logger: log.New(
			os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile,
		),
	}

	counter := int64(0)

	mux := http.NewServeMux()
	setupHandler(mux, config)
	m := attachIDMiddleware(loggingMiddleware(panicMiddleware(mux)), &counter)

	log.Fatal(http.ListenAndServe(":8080", m))
}
