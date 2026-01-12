package todo

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Todo struct {
	ID              string    `json:"id"`
	Text            string    `json:"text"`
	Status          string    `json:"status"`
	Created         time.Time `json:"created"`
	Position        float64   `json:"position"`
	CreatedByUserID *string   `json:"created_by_user_id,omitempty"`
	IsAdminTodo     bool      `json:"is_admin_todo"`
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
	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Return user's own todos + all admin todos
	// For admin todos, use user-specific status/position from user_todo_state if exists
	rows, err := h.db.Query(r.Context(), `
		SELECT 
			t.id, 
			t.text, 
			CASE 
				WHEN t.is_admin_todo THEN COALESCE(uts.status, t.status)
				ELSE t.status
			END as status,
			t.created, 
			CASE 
				WHEN t.is_admin_todo THEN COALESCE(uts.position, t.position)
				ELSE t.position
			END as position,
			t.created_by_user_id, 
			t.is_admin_todo 
		FROM todos t
		LEFT JOIN user_todo_state uts ON t.id = uts.todo_id AND uts.user_id = $1 AND t.is_admin_todo = true
		WHERE t.user_id = $1 OR t.is_admin_todo = true 
		ORDER BY position ASC, created DESC 
		LIMIT 100
	`, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.Text, &t.Status, &t.Created, &t.Position, &t.CreatedByUserID, &t.IsAdminTodo); err != nil {
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
	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userRole, _ := r.Context().Value("user_role").(string)

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
	status := "pending"
	created := time.Now()

	// Determine if this is an admin todo
	isAdminTodo := userRole == "admin"

	// Get min position to put at top
	var minPos float64
	if isAdminTodo {
		// For admin todos, get global min position
		_ = h.db.QueryRow(r.Context(), `SELECT COALESCE(MIN(position), 0) FROM todos`).Scan(&minPos)
	} else {
		// For user todos, get min position for this user
		_ = h.db.QueryRow(r.Context(), `SELECT COALESCE(MIN(position), 0) FROM todos WHERE user_id=$1`, userID).Scan(&minPos)
	}
	position := minPos - 1024.0

	// Insert todo
	if isAdminTodo {
		// Admin todo: set is_admin_todo=true, created_by_user_id=userID, user_id=NULL
		_, err := h.db.Exec(r.Context(), `
			INSERT INTO todos(id, text, status, created, position, created_by_user_id, is_admin_todo, user_id) 
			VALUES($1,$2,$3,$4,$5,$6,$7,NULL)
		`, id, req.Text, status, created, position, userID, true)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// Regular user todo
		_, err := h.db.Exec(r.Context(), `
			INSERT INTO todos(id, text, status, created, position, user_id, created_by_user_id, is_admin_todo) 
			VALUES($1,$2,$3,$4,$5,$6,$7,$8)
		`, id, req.Text, status, created, position, userID, userID, false)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	createdByUserID := userID
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(Todo{
		ID:              id,
		Text:            req.Text,
		Status:          status,
		Created:         created,
		Position:        position,
		CreatedByUserID: &createdByUserID,
		IsAdminTodo:     isAdminTodo,
	})
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

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

	// Check if todo exists and if user can update it
	var isAdminTodo bool
	var todoUserID *string
	err := h.db.QueryRow(r.Context(), `
		SELECT is_admin_todo, user_id 
		FROM todos 
		WHERE id=$1
	`, id).Scan(&isAdminTodo, &todoUserID)

	if err != nil {
		http.Error(w, "todo not found", http.StatusNotFound)
		return
	}

	// Users can update:
	// 1. Their own todos (user_id matches)
	// 2. Admin todos (is_admin_todo=true)
	canUpdate := isAdminTodo || (todoUserID != nil && *todoUserID == userID)
	if !canUpdate {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	// Handle update based on todo type
	if isAdminTodo {
		// For admin todos, update user_todo_state (per-user status)
		// Only update status if provided (text updates not allowed for admin todos)
		if req.Status != "" {
			_, err = h.db.Exec(r.Context(), `
				INSERT INTO user_todo_state (user_id, todo_id, status, position, updated_at)
				VALUES ($1, $2, $3, 
					(SELECT COALESCE(position, 0) FROM user_todo_state WHERE user_id=$1 AND todo_id=$2),
					now())
				ON CONFLICT (user_id, todo_id) 
				DO UPDATE SET status = $3, updated_at = now()
			`, userID, id, req.Status)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	} else {
		// For personal todos, update the todos table
		_, err = h.db.Exec(r.Context(), `
			UPDATE todos 
			SET text = CASE WHEN $2 = '' THEN text ELSE $2 END, 
			    status = CASE WHEN $3 = '' THEN status ELSE $3 END 
			WHERE id=$1
		`, id, req.Text, req.Status)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Fetch and return updated todo with user-specific state
	var t Todo
	if err := h.db.QueryRow(r.Context(), `
		SELECT 
			t.id, t.text, 
			CASE 
				WHEN t.is_admin_todo THEN COALESCE(uts.status, t.status)
				ELSE t.status
			END as status,
			t.created, 
			CASE 
				WHEN t.is_admin_todo THEN COALESCE(uts.position, t.position)
				ELSE t.position
			END as position,
			t.created_by_user_id, t.is_admin_todo 
		FROM todos t
		LEFT JOIN user_todo_state uts ON t.id = uts.todo_id AND uts.user_id = $2 AND t.is_admin_todo = true
		WHERE t.id=$1
	`, id, userID).Scan(&t.ID, &t.Text, &t.Status, &t.Created, &t.Position, &t.CreatedByUserID, &t.IsAdminTodo); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

func (h *Handler) Reorder(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

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

	tx, err := h.db.Begin(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	for i, id := range req.IDs {
		pos := float64(i) * 1024.0

		// Check if this is an admin todo
		var isAdminTodo bool
		err := tx.QueryRow(r.Context(), `SELECT is_admin_todo FROM todos WHERE id=$1`, id).Scan(&isAdminTodo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if isAdminTodo {
			// For admin todos, UPSERT into user_todo_state
			_, err := tx.Exec(r.Context(), `
				INSERT INTO user_todo_state (user_id, todo_id, status, position, updated_at)
				VALUES ($1, $2, 
					(SELECT COALESCE(status, 'pending') FROM user_todo_state WHERE user_id=$1 AND todo_id=$2),
					$3, now())
				ON CONFLICT (user_id, todo_id) 
				DO UPDATE SET position = $3, updated_at = now()
			`, userID, id, pos)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			// For personal todos, update todos table
			_, err := tx.Exec(r.Context(), `
				UPDATE todos 
				SET position = $1 
				WHERE id = $2 AND user_id = $3
			`, pos, id, userID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
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
	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userRole, _ := r.Context().Value("user_role").(string)

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}

	// Check if todo is an admin todo
	var isAdminTodo bool
	var todoUserID *string
	err := h.db.QueryRow(r.Context(), `
		SELECT is_admin_todo, user_id 
		FROM todos 
		WHERE id=$1
	`, id).Scan(&isAdminTodo, &todoUserID)

	if err != nil {
		http.Error(w, "todo not found", http.StatusNotFound)
		return
	}

	// Only admins can delete admin todos
	if isAdminTodo && userRole != "admin" {
		http.Error(w, "forbidden: only admins can delete admin todos", http.StatusForbidden)
		return
	}

	// Regular users can only delete their own todos
	if !isAdminTodo && (todoUserID == nil || *todoUserID != userID) {
		http.Error(w, "forbidden", http.StatusForbidden)
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
