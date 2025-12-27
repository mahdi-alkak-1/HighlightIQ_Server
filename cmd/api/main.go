package main

import (
	"log"
	"net/http"

	"highlightiq-server/internal/config"
	"highlightiq-server/internal/db"
	"highlightiq-server/internal/http/router"
)

func main() {
	cfg := config.Load()

	// Connect DB (fail fast if DB is not reachable)
	conn, err := db.NewMySQL(cfg.MySQL)
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	defer conn.Close()

	// For now router doesn't use DB yet — next step we will inject it.
	r := router.New()

	log.Println("API listening on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
