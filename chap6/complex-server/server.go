package main

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/PaulOh5/complex-server/config"
	"github.com/PaulOh5/complex-server/handlers"
	"github.com/PaulOh5/complex-server/middleware"
)

func setupServer(mux *http.ServeMux, w io.Writer) http.Handler {
	conf := config.InitConfig(w)
	counter := int64(0)
	handlers.Register(mux, conf)
	return middleware.RegisterMiddleware(mux, conf, &counter)
}

func main() {
	mux := http.NewServeMux()
	wrappedMux := setupServer(mux, os.Stdout)

	log.Fatal(http.ListenAndServe(":8080", wrappedMux))
}
