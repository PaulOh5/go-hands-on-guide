package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type httpLog struct {
	Url      string `json:"url"`
	Method   string `json:"method"`
	BodySize int    `json:"body-size"`
	Protocol string `json:"protocol"`
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var bodyBytes []byte
		if r.Body != nil {
			bodyBytes, _ = io.ReadAll(r.Body)
		}
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		bodySize := len(bodyBytes)
		logger := httpLog{
			Url:      r.URL.Path,
			Method:   r.Method,
			BodySize: bodySize,
			Protocol: r.Proto,
		}
		jsonData, _ := json.Marshal(logger)
		log.Print(string(jsonData))
		next.ServeHTTP(w, r)
	})
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, world!")
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "ok")
}

func setupHandlers(mux *http.ServeMux) {
	mux.Handle("/healthz", loggingMiddleware(http.HandlerFunc(healthCheckHandler)))
	mux.Handle("/api", loggingMiddleware(http.HandlerFunc(apiHandler)))
}

func main() {
	listenAddr := os.Getenv("LISTEN_ADDR")
	if len(listenAddr) == 0 {
		listenAddr = ":8080"
	}

	mux := http.NewServeMux()
	setupHandlers(mux)

	log.Fatal(http.ListenAndServe(listenAddr, mux))
}
