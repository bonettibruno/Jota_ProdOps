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
	// Load environment variables from .env file
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	// Initialize Gemini LLM client
	g, err := gemini.New()
	if err != nil {
		log.Fatal(err)
	}

	// Set global LLM client for handlers
	api.SetLLMClient(g)

	// Route definitions
	mux := http.NewServeMux()
	mux.HandleFunc("/health", api.HealthHandler)     // Service health check
	mux.HandleFunc("/messages", api.MessagesHandler) // Main chat and orchestration endpoint
	mux.HandleFunc("/metrics", api.MetricsHandler)   // Telemetry and ProdOps KPIs

	log.Printf("Server running on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
