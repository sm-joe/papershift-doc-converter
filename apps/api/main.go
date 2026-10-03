package main

import (
	"log"
	"net/http"
	"os"

	"github.com/sm-joe/papershift-doc-converter/internal/api"
)

func main() {
	addr := getEnv("PAPERSHIFT_ADDR", ":8080")
	workspaceRoot := getEnv("PAPERSHIFT_WORK_DIR", os.TempDir())

	app := api.NewApplication(workspaceRoot)
	handler := api.NewHandler(app)

	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	log.Printf("PaperShift API listening on %s", addr)

	if err := server.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
