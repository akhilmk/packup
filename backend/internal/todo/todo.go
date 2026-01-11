package todo

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Todo struct {
	ID       string    `json:"id"`
	Text     string    `json:"text"`
	Status   string    `json:"status"`
	Created  time.Time `json:"created"`
	Position float64   `json:"position"`
}

type Handler struct {
	db *pgxpool.Pool
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{db: db}
}

// RegisterRoutes registers the specific routes to a mux using Go 1.22 enhanced routing
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/todos", h.List)
	mux.HandleFunc("POST /api/todos", h.Create)
	mux.HandleFunc("PUT /api/todos/{id}", h.Update)
	mux.HandleFunc("PUT /api/todos/reorder", h.Reorder)
	mux.HandleFunc("DELETE /api/todos/{id}", h.Delete)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(r.Context(), `SELECT id, text, status, created, position FROM todos ORDER BY position ASC, created DESC LIMIT 100`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.Text, &t.Status, &t.Created, &t.Position); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		todos = append(todos, t)
	}

	// Return empty array instead of null if nil
	if todos == nil {
		todos = []Todo{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"todos": todos})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if len(req.Text) > 200 {
		http.Error(w, "text limit of 200 characters exceeded", http.StatusBadRequest)
		return
	}
	if req.Text == "" {
		http.Error(w, "text cannot be empty", http.StatusBadRequest)
		return
	}

	id := uuid.NewString()
	// Default status is 'pending'
	status := "pending"
	created := time.Now()

	// Get min position to put at top
	var minPos float64
	_ = h.db.QueryRow(r.Context(), `SELECT COALESCE(MIN(position), 0) FROM todos`).Scan(&minPos)
	position := minPos - 1024.0

	_, err := h.db.Exec(r.Context(), `INSERT INTO todos(id, text, status, created, position) VALUES($1,$2,$3,$4,$5)`, id, req.Text, status, created, position)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(Todo{ID: id, Text: req.Text, Status: status, Created: created, Position: position})
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}

	var req struct {
		Text   string `json:"text"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if req.Text != "" && len(req.Text) > 200 {
		http.Error(w, "text limit of 200 characters exceeded", http.StatusBadRequest)
		return
	}

	// Validate status if provided
	if req.Status != "" {
		switch req.Status {
		case "pending", "in-progress", "done":
			// valid
		default:
			http.Error(w, "invalid status", http.StatusBadRequest)
			return
		}
	}

	// Check existence
	var exists bool
	err := h.db.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM todos WHERE id=$1)`, id).Scan(&exists)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "todo not found", http.StatusNotFound)
		return
	}

	// Dynamic update: only update fields that are provided
	// Note: We use COALESCE(NULLIF($2,''), text) for text to allow empty updates (if client sends empty string to mean "no change").
	// But actually our client sends partial JSON.
	// A better way for SQL updates with partial data is often detailed.
	// Here simplified: If req.Status is empty, we keep old status.

	_, err = h.db.Exec(r.Context(), `
		UPDATE todos 
		SET text = CASE WHEN $2 = '' THEN text ELSE $2 END, 
		    status = CASE WHEN $3 = '' THEN status ELSE $3 END 
		WHERE id=$1`, id, req.Text, req.Status)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var t Todo
	if err := h.db.QueryRow(r.Context(), `SELECT id, text, status, created FROM todos WHERE id=$1`, id).Scan(&t.ID, &t.Text, &t.Status, &t.Created); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

func (h *Handler) Reorder(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if len(req.IDs) == 0 {
		return
	}

	// Update positions based on index
	// We'll reset normalization here: 0, 1024, 2048...
	// Using a transaction would be better but keeping it simple

	// Prepare batch update is cleaner but one-by-one inside tx is okay for small lists
	tx, err := h.db.Begin(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	for i, id := range req.IDs {
		pos := float64(i) * 1024.0
		_, err := tx.Exec(r.Context(), `UPDATE todos SET position = $1 WHERE id = $2`, pos, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success":true}`))
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}

	cmd, err := h.db.Exec(r.Context(), `DELETE FROM todos WHERE id=$1`, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if cmd.RowsAffected() == 0 {
		http.Error(w, "todo not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success":true}`))
}
