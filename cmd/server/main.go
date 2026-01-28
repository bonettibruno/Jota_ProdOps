package main

import (
	"log"
	"net/http"

	"github.com/bonettibruno/Jota_ProdOps/internal/api"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", api.HealthHandler)
	mux.HandleFunc("/messages", api.MessagesHandler)

	addr := ":8080"
	log.Printf("Server running on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
