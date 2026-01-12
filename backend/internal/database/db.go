package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func New(ctx context.Context) (*pgxpool.Pool, error) {
	// Load environment file `env.dev` if present
	// We might need to adjust the path if running from a different working dir,
	// but usually if run from root, just ".env.dev" or similar is fine.
	// The original code used "env.dev".
	// Try loading environment files from various locations
	envFiles := []string{"env.dev", ".env.dev", "../.env.dev", "../../.env.dev"}
	for _, file := range envFiles {
		if err := godotenv.Load(file); err == nil {
			log.Printf("loaded env file: %s", file)
			break
		}
	}

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
			name = "itinera"
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
	if exists {
		log.Println("Schema already exists, skipping migration")
		return nil
	}

	// Try to find and load schema_v1.sql from various paths
	schemaPaths := []string{
		"migrations/schema_v1.sql",
		"backend/migrations/schema_v1.sql",
		"../migrations/schema_v1.sql",
		"../../migrations/schema_v1.sql",
	}

	var schemaSQL []byte
	var schemaPath string
	for _, path := range schemaPaths {
		if data, err := os.ReadFile(path); err == nil {
			schemaSQL = data
			schemaPath = path
			break
		}
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
