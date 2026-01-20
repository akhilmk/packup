package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/akhilmk/packup/internal/admin"
	"github.com/akhilmk/packup/internal/auth"
	"github.com/akhilmk/packup/internal/database"
	"github.com/akhilmk/packup/internal/todo"

	_ "github.com/akhilmk/packup/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title PackUp API
// @version 1.0
// @description This is the API server for the PackUp application.
// @host localhost:8080
// @BasePath /
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
	// we pass the auth middleware to the handler, so it can wrap its routes
	mw := authHandler.Middleware
	todoHandler.RegisterRoutes(mux, mw)

	// Register Admin routes (require admin role)
	// We wrap the standard auth middleware AND the admin check
	adminMw := func(next http.HandlerFunc) http.HandlerFunc {
		return mw(adminHandler.RequireAdmin(next))
	}
	adminHandler.RegisterRoutes(mux, adminMw)

	// Swagger documentation protected by admin-only auth
	mux.HandleFunc("GET /swagger/", authHandler.AdminMiddlewareWithRedirect(httpSwagger.WrapHandler))

	// Serve static frontend files
	staticDir := "frontend/dist"
	if _, err := os.Stat(staticDir); err == nil {
		fs := http.FileServer(http.Dir(staticDir))
		mux.Handle("/", fs)
		log.Printf("Serving static files from: %s", staticDir)
	} else {
		log.Printf("Warning: %s not found. Frontend will not be served.", staticDir)
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
