package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
)

type appConfig struct {
	logger *log.Logger
}

type app struct {
	config  appConfig
	handler func(w http.ResponseWriter, r *http.Request, config appConfig) (int, error)
}

func (a *app) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	code, err := a.handler(w, r, a.config)
	if err != nil {
		log.Printf("%s", err.Error())
		http.Error(w, err.Error(), code)
	}
}

func apiHandler(w http.ResponseWriter, r *http.Request, config appConfig) (int, error) {
	config.logger.Println("Handling API request")
	fmt.Fprintf(w, "Hello, world\n")
	return 200, nil
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request, config appConfig) (int, error) {
	if r.Method != http.MethodGet {
		return http.StatusMethodNotAllowed, errors.New("method not allowed")
	}

	config.logger.Println("Handling healthcheck request")
	fmt.Fprint(w, "ok\n")
	return 200, nil
}

func setupHandler(mux *http.ServeMux, config appConfig) {
	mux.Handle("/healthz", &app{config: config, handler: healthCheckHandler})
	mux.Handle("/api", &app{config: config, handler: apiHandler})
}

func main() {
	config := appConfig{
		logger: log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile),
	}

	mux := http.NewServeMux()
	setupHandler(mux, config)

	log.Fatal(http.ListenAndServe(":8080", mux))
}
