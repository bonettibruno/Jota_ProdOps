package main

import (
	"log"
	"net/http"
	"os"

	"github.com/bonettibruno/Jota_ProdOps/internal/api"
	"github.com/bonettibruno/Jota_ProdOps/internal/llm/gemini"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	g, err := gemini.New()
	if err != nil {
		log.Fatal(err)
	}

	api.SetLLMClient(g)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", api.HealthHandler)
	mux.HandleFunc("/messages", api.MessagesHandler)
	mux.HandleFunc("/metrics", api.MetricsHandler)

	log.Printf("Server running on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
