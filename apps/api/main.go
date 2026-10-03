package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/sm-joe/papershift-doc-converter/internal/formats"
)

type healthResponse struct {
	Status string `json:"status"`
}

func main() {
	addr := getEnv("PAPERSHIFT_ADDR", ":8080")

	mux := http.NewServeMux()

	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/ready", readyHandler)
	mux.HandleFunc("/api/v1/formats", formatsHandler)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	log.Printf("PaperShift API listening on %s", addr)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	writeJSON(w, http.StatusOK, healthResponse{
		Status: "ok",
	})
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	writeJSON(w, http.StatusOK, healthResponse{
		Status: "ready",
	})
}

func formatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"formats": formats.All(),
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
