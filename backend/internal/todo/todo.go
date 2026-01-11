package todo

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Todo struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	Completed bool   `json:"completed"`
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
	mux.HandleFunc("DELETE /api/todos/{id}", h.Delete)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(r.Context(), `SELECT id, text, completed FROM todos ORDER BY created DESC LIMIT 100`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.Text, &t.Completed); err != nil {
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
	_, err := h.db.Exec(r.Context(), `INSERT INTO todos(id, text, completed, created) VALUES($1,$2,$3,$4)`, id, req.Text, false, time.Now())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(Todo{ID: id, Text: req.Text, Completed: false})
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}

	var req struct {
		Text      string `json:"text"`
		Completed bool   `json:"completed"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if req.Text != "" && len(req.Text) > 200 {
		http.Error(w, "text limit of 200 characters exceeded", http.StatusBadRequest)
		return
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

	_, err = h.db.Exec(r.Context(), `UPDATE todos SET text = COALESCE(NULLIF($2,''), text), completed = $3 WHERE id=$1`, id, req.Text, req.Completed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var t Todo
	if err := h.db.QueryRow(r.Context(), `SELECT id, text, completed FROM todos WHERE id=$1`, id).Scan(&t.ID, &t.Text, &t.Completed); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
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
