package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/akhilmk/itinera/internal/admin"
	"github.com/akhilmk/itinera/internal/auth"
	"github.com/akhilmk/itinera/internal/database"
	"github.com/akhilmk/itinera/internal/todo"
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
	adminHandler := admin.NewHandler(pool)

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

	// Register Admin routes (require admin role)
	adminMw := func(next http.HandlerFunc) http.HandlerFunc {
		return mw(adminHandler.RequireAdmin(next))
	}
	mux.HandleFunc("GET /api/admin/users", adminMw(adminHandler.ListUsers))
	mux.HandleFunc("GET /api/admin/todos", adminMw(adminHandler.ListAdminTodos))
	mux.HandleFunc("POST /api/admin/todos", adminMw(adminHandler.CreateAdminTodo))
	mux.HandleFunc("PUT /api/admin/todos/{id}", adminMw(adminHandler.UpdateAdminTodo))
	mux.HandleFunc("DELETE /api/admin/todos/{id}", adminMw(adminHandler.DeleteAdminTodo))
	mux.HandleFunc("GET /api/admin/users/{userId}/todos", adminMw(adminHandler.ListUserTodos))

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
