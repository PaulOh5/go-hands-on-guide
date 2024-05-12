package handlers

import (
	"fmt"
	"net/http"

	"github.com/PaulOh5/complex-server/config"
)

type app struct {
	conf    config.AppConfig
	handler func(w http.ResponseWriter, r *http.Request, conf config.AppConfig)
}

func (a app) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.handler(w, r, a.conf)
}

func apiHandler(w http.ResponseWriter, r *http.Request, conf config.AppConfig) {
	fmt.Fprint(w, "Hello, world!\n")
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request, conf config.AppConfig) {
	if r.Method != http.MethodGet {
		conf.Logger.Printf("error=\"Invalid request\" path=%s method=%s", r.URL.Path, r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fmt.Fprintf(w, "ok\n")
}

func panicHandler(w http.ResponseWriter, r *http.Request, conf config.AppConfig) {
	panic("I panicked")
}
