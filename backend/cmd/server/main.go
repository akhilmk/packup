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
	// try to resolve directory
	// try to resolve directory
	staticDir := "frontend/dist"
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		// try relative to cmd/server if running from there (root/backend/cmd/server)
		staticDir = "../../../frontend/dist"
		if _, err := os.Stat(staticDir); os.IsNotExist(err) {
			// Just in case we are in backend root
			staticDir = "../frontend/dist"
			if _, err := os.Stat(staticDir); os.IsNotExist(err) {
				log.Printf("Warning: frontend/dist not found at %s or ../../../frontend/dist or ../frontend/dist", "frontend/dist")
			}
		}
	}

	// We only serve if it exists, otherwise it might panic or 404.
	if _, err := os.Stat(staticDir); err == nil {
		fs := http.FileServer(http.Dir(staticDir))
		mux.Handle("/", fs)
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
