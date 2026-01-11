package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	proto "todo-app/gen/proto"
	"todo-app/gen/proto/todov1connect"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment file `env.dev` if present
	if err := godotenv.Load("env.dev"); err == nil {
		log.Println("loaded env.dev")
	} else {
		log.Printf("env.dev not found or failed to load: %v", err)
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
			name = "todos"
		}
		if ssl == "" {
			ssl = "disable"
		}

		dbURL = "postgres://" + user + ":" + pass + "@" + host + ":" + port + "/" + name + "?sslmode=" + ssl
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer pool.Close()

	if err := ensureSchema(ctx, pool); err != nil {
		log.Fatalf("failed to ensure schema: %v", err)
	}

	// Register ConnectRPC handler
	path, handler := todov1connect.NewTodoServiceHandler(&TodoServer{db: pool})

	mux := http.NewServeMux()
	mux.Handle(path+"/", handler)
	mux.Handle(path, handler)

	// Serve static frontend files
	fs := http.FileServer(http.Dir("../frontend/dist"))
	mux.Handle("/", fs)

	// Provide a simple GET endpoint for listing todos (JSON)
	mux.HandleFunc("/api/todos", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		rows, err := pool.Query(context.Background(), `SELECT id, text, completed FROM todos ORDER BY created DESC LIMIT 100`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		type todoResp struct {
			Id        string `json:"id"`
			Text      string `json:"text"`
			Completed bool   `json:"completed"`
		}

		var todos []todoResp
		for rows.Next() {
			var id, text string
			var completed bool
			if err := rows.Scan(&id, &text, &completed); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			todos = append(todos, todoResp{Id: id, Text: text, Completed: completed})
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"todos": todos})
	})

	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}

	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

type TodoServer struct {
	db *pgxpool.Pool
}

func (s *TodoServer) AddTodo(ctx context.Context, req *connect.Request[proto.AddTodoRequest]) (*connect.Response[proto.Todo], error) {
	text := req.Msg.Text
	if len(text) > 200 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("text limit of 200 characters exceeded"))
	}
	if text == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("text cannot be empty"))
	}

	id := uuid.NewString()
	_, err := s.db.Exec(ctx, `INSERT INTO todos(id, text, completed, created) VALUES($1,$2,$3,$4)`, id, text, false, time.Now())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&proto.Todo{Id: id, Text: text, Completed: false}), nil
}

func (s *TodoServer) ListTodos(ctx context.Context, req *connect.Request[proto.ListTodosRequest]) (*connect.Response[proto.ListTodosResponse], error) {
	rows, err := s.db.Query(ctx, `SELECT id, text, completed FROM todos ORDER BY created DESC LIMIT 100`)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	defer rows.Close()

	var todos []*proto.Todo
	for rows.Next() {
		var id, text string
		var completed bool
		if err := rows.Scan(&id, &text, &completed); err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		todos = append(todos, &proto.Todo{Id: id, Text: text, Completed: completed})
	}

	return connect.NewResponse(&proto.ListTodosResponse{Todos: todos}), nil
}

func (s *TodoServer) UpdateTodo(ctx context.Context, req *connect.Request[proto.UpdateTodoRequest]) (*connect.Response[proto.Todo], error) {
	// Check existence first
	var exists bool
	err := s.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM todos WHERE id=$1)`, req.Msg.Id).Scan(&exists)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if !exists {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("todo not found"))
	}

	if req.Msg.Text != "" && len(req.Msg.Text) > 200 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("text limit of 200 characters exceeded"))
	}

	// Perform update
	_, err = s.db.Exec(ctx, `UPDATE todos SET text = COALESCE(NULLIF($2,''), text), completed = $3 WHERE id=$1`, req.Msg.Id, req.Msg.Text, req.Msg.Completed)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Return updated row
	var text string
	var completed bool
	if err := s.db.QueryRow(ctx, `SELECT text, completed FROM todos WHERE id=$1`, req.Msg.Id).Scan(&text, &completed); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&proto.Todo{Id: req.Msg.Id, Text: text, Completed: completed}), nil
}

func (s *TodoServer) DeleteTodo(ctx context.Context, req *connect.Request[proto.DeleteTodoRequest]) (*connect.Response[proto.DeleteTodoResponse], error) {
	cmd, err := s.db.Exec(ctx, `DELETE FROM todos WHERE id=$1`, req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if cmd.RowsAffected() == 0 {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("todo not found"))
	}
	return connect.NewResponse(&proto.DeleteTodoResponse{Id: req.Msg.Id}), nil
}

func ensureSchema(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `
	CREATE TABLE IF NOT EXISTS todos (
		id TEXT PRIMARY KEY,
		text TEXT NOT NULL,
		completed BOOLEAN NOT NULL DEFAULT false,
		created TIMESTAMPTZ NOT NULL DEFAULT now()
	);`)
	return err
}
