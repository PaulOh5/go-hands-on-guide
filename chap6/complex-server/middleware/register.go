package middleware

import (
	"net/http"

	"github.com/PaulOh5/complex-server/config"
)

func RegisterMiddleware(mux *http.ServeMux, c config.AppConfig, counter *int64) http.Handler {
	return attachIDMiddleware(loggingMiddleware(panicMiddleware(mux, c), c), counter)
}
