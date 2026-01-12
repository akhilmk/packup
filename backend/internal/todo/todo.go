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
	IsCustomerTask  bool      `json:"is_customer_task"`
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

	userRole, _ := r.Context().Value("user_role").(string)

	// Check if we should exclude admin todos (for admin's personal todo view)
	// Admins should only see their personal todos in this view ("My Todos") unless configured otherwise.
	excludeAdminTodos := r.URL.Query().Get("exclude_admin_todos") == "true" || userRole == "admin"

	var query string
	if excludeAdminTodos {
		// Only return user's personal todos (exclude customer tasks)
		query = `
			SELECT 
				t.id, 
				t.text, 
				t.status,
				t.created, 
				t.position,
				t.created_by_user_id, 
				t.is_customer_task
			FROM todos t
			WHERE t.user_id = $1 AND t.is_customer_task = false
			ORDER BY position ASC, created DESC 
			LIMIT 100
		`
	} else {
		// Return user's own todos + all customer tasks
		// For customer tasks, use user-specific status/position from user_todo_state if exists
		query = `
			SELECT 
				t.id, 
				t.text, 
				CASE 
					WHEN t.is_customer_task THEN COALESCE(uts.status, t.status)
					ELSE t.status
				END as status,
				t.created, 
				CASE 
					WHEN t.is_customer_task THEN COALESCE(uts.position, t.position)
					ELSE t.position
				END as position,
				t.created_by_user_id, 
				t.is_customer_task
			FROM todos t
			LEFT JOIN user_todo_state uts ON t.id = uts.todo_id AND uts.user_id = $1 AND t.is_customer_task = true
			WHERE t.user_id = $1 OR t.is_customer_task = true
			ORDER BY position ASC, created DESC 
			LIMIT 100
		`
	}

	rows, err := h.db.Query(r.Context(), query, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.Text, &t.Status, &t.Created, &t.Position, &t.CreatedByUserID, &t.IsCustomerTask); err != nil {
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

	// userRole is not needed for personal todo creation anymore
	// userRole, _ := r.Context().Value("user_role").(string)

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

	// This endpoint is for PERSONAL todos only.
	// Admin (Global, Customer) todos are created via the Admin API.
	isCustomerTask := false

	// Get min position for this user to put at top
	var minPos float64
	_ = h.db.QueryRow(r.Context(), `SELECT COALESCE(MIN(position), 0) FROM todos WHERE user_id=$1`, userID).Scan(&minPos)
	position := minPos - 1024.0

	// Insert personal todo
	_, err := h.db.Exec(r.Context(), `
		INSERT INTO todos(id, text, status, created, position, user_id, created_by_user_id, is_customer_task) 
		VALUES($1,$2,$3,$4,$5,$6,$7,$8)
	`, id, req.Text, status, created, position, userID, userID, isCustomerTask)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
		IsCustomerTask:  isCustomerTask,
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
	var isCustomerTask bool
	var todoUserID *string
	err := h.db.QueryRow(r.Context(), `
		SELECT is_customer_task, user_id 
		FROM todos 
		WHERE id=$1
	`, id).Scan(&isCustomerTask, &todoUserID)

	if err != nil {
		http.Error(w, "todo not found", http.StatusNotFound)
		return
	}

	// Users can update:
	// 1. Their own todos (user_id matches)
	// 2. Customer tasks (is_customer_task=true)
	canUpdate := isCustomerTask || (todoUserID != nil && *todoUserID == userID)
	if !canUpdate {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	// Handle update based on todo type
	if isCustomerTask {
		// For customer tasks, update user_todo_state (per-user status)
		// Only update status if provided (text updates not allowed for customer tasks)
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
				WHEN t.is_customer_task THEN COALESCE(uts.status, t.status)
				ELSE t.status
			END as status,
			t.created, 
			CASE 
				WHEN t.is_customer_task THEN COALESCE(uts.position, t.position)
				ELSE t.position
			END as position,
			t.created_by_user_id, t.is_customer_task
		FROM todos t
		LEFT JOIN user_todo_state uts ON t.id = uts.todo_id AND uts.user_id = $2 AND t.is_customer_task = true
		WHERE t.id=$1
	`, id, userID).Scan(&t.ID, &t.Text, &t.Status, &t.Created, &t.Position, &t.CreatedByUserID, &t.IsCustomerTask); err != nil {
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

		// Check if this is a customer task
		var isCustomerTask bool
		err := tx.QueryRow(r.Context(), `SELECT is_customer_task FROM todos WHERE id=$1`, id).Scan(&isCustomerTask)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if isCustomerTask {
			// For customer tasks, UPSERT into user_todo_state
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

	// Check if todo is a customer task
	var isCustomerTask bool
	var todoUserID *string
	err := h.db.QueryRow(r.Context(), `
		SELECT is_customer_task, user_id 
		FROM todos 
		WHERE id=$1
	`, id).Scan(&isCustomerTask, &todoUserID)

	if err != nil {
		http.Error(w, "todo not found", http.StatusNotFound)
		return
	}

	// Only admins can delete customer tasks
	if isCustomerTask && userRole != "admin" {
		http.Error(w, "forbidden: only admins can delete customer tasks", http.StatusForbidden)
		return
	}

	// Regular users can only delete their own todos
	if !isCustomerTask && (todoUserID == nil || *todoUserID != userID) {
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
