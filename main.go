package main

import (
	"log"
	"net/http"
	"os"

	"github.com/giovaniif/sharding-test/db"
	"github.com/giovaniif/sharding-test/handlers"
)

func main() {
	sm, err := db.ConnectShards([db.NumShards]string{
		os.Getenv("DB_HOST_0"),
		os.Getenv("DB_HOST_1"),
		os.Getenv("DB_HOST_2"),
	})
	if err != nil {
		log.Fatalf("failed to connect to shards: %v", err)
	}
	defer sm.Close()

	orders := handlers.NewOrderHandler(sm)

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
