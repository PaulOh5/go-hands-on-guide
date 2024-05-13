package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func apiHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, world!")
}

func setupHandlersAndMiddleWare(mux *http.ServeMux, l *log.Logger) http.Handler {
	mux.HandleFunc("/api", apiHandler)
	return loggingMiddleware(mux, l)
}

func loggingMiddleware(h http.Handler, l *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		h.ServeHTTP(w, r)
		l.Printf(
			"protocoal=%s path=%s method=%s duration=%f",
			r.Proto, r.URL.Path, r.Method, time.Since(startTime).Seconds(),
		)
	})
}

func main() {
	tlsCertFile := "./server.crt"
	tlsKeyFile := "./server.key"

	mux := http.NewServeMux()

	l := log.New(os.Stdout, "tls-server", log.Lshortfile|log.LstdFlags)
	m := setupHandlersAndMiddleWare(mux, l)

	log.Fatal(
		http.ListenAndServeTLS(":8443", tlsCertFile, tlsKeyFile, m),
	)
}
