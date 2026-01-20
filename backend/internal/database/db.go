package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func New(ctx context.Context) (*pgxpool.Pool, error) {

	// The application relies on environment variables provided by the OS/Docker runtime.
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		user := os.Getenv("DB_USER")
		pass := os.Getenv("DB_PASS")
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		name := os.Getenv("DB_NAME")
		ssl := os.Getenv("SSLMODE")

		if user == "" {
			user = "postgres"
		}
		if pass == "" {
			pass = "postgres"
		}
		if host == "" {
			host = "localhost"
		}
		if port == "" {
			port = "5432"
		}
		if name == "" {
			name = "packup"
		}
		if ssl == "" {
			ssl = "disable"
		}

		dbURL = "postgres://" + user + ":" + pass + "@" + host + ":" + port + "/" + name + "?sslmode=" + ssl
	}

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, err
	}

	// Auto-load schema if tables don't exist
	if err := ensureSchema(ctx, pool); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}

func ensureSchema(ctx context.Context, db *pgxpool.Pool) error {
	// Check if tables exist
	var exists bool
	err := db.QueryRow(ctx, "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'users')").Scan(&exists)
	if err != nil {
		return err
	}

	// If tables already exist, skip schema loading
	// If tables already exist, we still want to run the schema to ensure new columns/tables are added.
	// The schema file should be idempotent (using IF NOT EXISTS).
	if exists {
		log.Println("Schema loaded previously, but running again to ensure updates...")
	}

	schemaPath := "migrations/schema_v1.sql"
	schemaSQL, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("schema file not found at %s: %w", schemaPath, err)
	}

	if schemaSQL == nil {
		return fmt.Errorf("schema_v1.sql not found in any expected location")
	}

	log.Printf("Loading schema from: %s", schemaPath)
	_, err = db.Exec(ctx, string(schemaSQL))
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	log.Println("Schema loaded successfully")
	return nil
}
