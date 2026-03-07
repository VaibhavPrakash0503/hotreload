package main

import (
	"log"
	"net/http"

	routes "github.com/VaibhavPrakash0503/hotreload/testserver/router"
)

func main() {
	router := routes.SetupRoutes()

	log.Println("Server starting on :8080")

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatal(err)
	}
}
