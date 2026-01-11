package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"todo-app/internal/database"
	"todo-app/internal/todo"
)

func main() {
	ctx := context.Background() // Wait, need to import context

	// Initialize DB
	pool, err := database.New(ctx)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer pool.Close()

	// Initialize Handler
	todoHandler := todo.NewHandler(pool)

	mux := http.NewServeMux()

	// Register Todo API routes
	todoHandler.RegisterRoutes(mux)

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
