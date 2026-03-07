package routes

import (
	"net/http"

	"github.com/VaibhavPrakash0503/hotreload/testserver/handlers"
)

func SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handlers.HelloHandler)
	mux.HandleFunc("/health", handlers.HealthHandler)

	return mux
}
