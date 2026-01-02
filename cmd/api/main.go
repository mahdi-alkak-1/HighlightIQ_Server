package main

import (
	"log"
	"net/http"

	"highlightiq-server/internal/config"
	"highlightiq-server/internal/db"

	authhandlers "highlightiq-server/internal/http/handlers/auth"
	recordinghandlers "highlightiq-server/internal/http/handlers/recordings"
	"highlightiq-server/internal/http/middleware"
	"highlightiq-server/internal/http/router"

	recordingrepo "highlightiq-server/internal/repos/recordings"
	"highlightiq-server/internal/repos/users"

	authsvc "highlightiq-server/internal/services/auth"
	recordingsvc "highlightiq-server/internal/services/recordings"
)

func main() {
	cfg := config.Load()

	conn, err := db.NewMySQL(cfg.MySQL)
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	defer conn.Close()

	// repos
	usersRepo := users.New(conn)
	recRepo := recordingrepo.New(conn)

	// services
	authService := authsvc.New(usersRepo, cfg.JWTSecret)
	recService := recordingsvc.New(recRepo, cfg.RecordingsDir)

	// handlers
	authHandler := authhandlers.New(authService)
	recHandler := recordinghandlers.New(recService)

	// middleware
	jwtAuth := middleware.NewJWTAuth(usersRepo, cfg.JWTSecret)

	// router (3rd arg = clipCandidatesHandler for now)
	r := router.New(authHandler, recHandler, nil, jwtAuth.Middleware)

	log.Println("API listening on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
