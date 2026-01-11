package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/akhilmk/focus-flow/internal/auth"
	"github.com/akhilmk/focus-flow/internal/database"
	"github.com/akhilmk/focus-flow/internal/todo"
)

func main() {
	ctx := context.Background() // Wait, need to import context

	// Initialize DB
	pool, err := database.New(ctx)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer pool.Close()

	// Initialize Handlers
	authHandler := auth.NewHandler(pool)
	todoHandler := todo.NewHandler(pool)

	mux := http.NewServeMux()

	// Register Auth routes
	authHandler.RegisterRoutes(mux)

	// Register Protected Todo API routes
	// We wrap these with the auth middleware
	mw := authHandler.Middleware
	mux.HandleFunc("GET /api/todos", mw(todoHandler.List))
	mux.HandleFunc("POST /api/todos", mw(todoHandler.Create))
	mux.HandleFunc("PUT /api/todos/{id}", mw(todoHandler.Update))
	mux.HandleFunc("PUT /api/todos/reorder", mw(todoHandler.Reorder))
	mux.HandleFunc("DELETE /api/todos/{id}", mw(todoHandler.Delete))

	// Serve static frontend files
	// Try multiple paths for different deployment scenarios
	staticPaths := []string{
		"frontend/dist",          // Running from bin/ folder (production)
		"../frontend/dist",       // Running from backend/ folder (development)
		"../../frontend/dist",    // Running from backend/cmd/server (development)
		"../../../frontend/dist", // Alternative development path
	}

	var staticDir string
	for _, path := range staticPaths {
		if _, err := os.Stat(path); err == nil {
			staticDir = path
			break
		}
	}

	if staticDir != "" {
		fs := http.FileServer(http.Dir(staticDir))
		mux.Handle("/", fs)
		log.Printf("Serving static files from: %s", staticDir)
	} else {
		log.Printf("Warning: frontend/dist not found in any expected location")
	}

	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}

	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
