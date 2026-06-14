package main

import (
	"log"
	"net/http"

	"github.com/giovaniif/sharding-test/db"
	"github.com/giovaniif/sharding-test/handlers"
)

func main() {
	database, err := db.Connect()
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer database.Close()

	orders := handlers.NewOrderHandler(database)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /orders", orders.GetAll)
	mux.HandleFunc("GET /orders/{id}", orders.GetByID)
	mux.HandleFunc("POST /orders", orders.Create)
	mux.HandleFunc("DELETE /orders/{id}", orders.Delete)

	log.Println("server listening on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
