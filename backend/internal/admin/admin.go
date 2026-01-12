package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	AvatarURL string    `json:"avatar_url"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type Todo struct {
	ID              string    `json:"id"`
	Text            string    `json:"text"`
	Status          string    `json:"status"`
	Created         time.Time `json:"created"`
	Position        float64   `json:"position"`
	CreatedByUserID *string   `json:"created_by_user_id,omitempty"`
	IsDefaultTask   bool      `json:"is_default_task"`
	SharedWithAdmin bool      `json:"shared_with_admin"`
	UserID          *string   `json:"user_id,omitempty"`
}

type Handler struct {
	db *pgxpool.Pool
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{db: db}
}

// Middleware to check if user is admin
func (h *Handler) RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userRole, ok := r.Context().Value("user_role").(string)
		if !ok || userRole != "admin" {
			http.Error(w, "forbidden: admin access required", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}

// ListUsers returns all users (admin only)
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// ListUsers returns all users excluding admins (admin only)
	rows, err := h.db.Query(r.Context(), `
		SELECT id, email, name, avatar_url, role, created_at 
		FROM users 
		WHERE role != 'admin'
		ORDER BY created_at DESC
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.AvatarURL, &u.Role, &u.CreatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	if users == nil {
		users = []User{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"users": users})
}

// ListAdminTodos returns all admin todos (admin only)
func (h *Handler) ListAdminTodos(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(r.Context(), `
		SELECT id, text, status, created, position, created_by_user_id, is_default_task, shared_with_admin
		FROM todos 
		WHERE is_default_task = true 
		ORDER BY position ASC, created DESC
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.Text, &t.Status, &t.Created, &t.Position, &t.CreatedByUserID, &t.IsDefaultTask, &t.SharedWithAdmin); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		todos = append(todos, t)
	}

	if todos == nil {
		todos = []Todo{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"todos": todos})
}

// CreateAdminTodo creates a new admin todo (admin only)
func (h *Handler) CreateAdminTodo(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

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

	// Get min position to put at top
	var minPos float64
	_ = h.db.QueryRow(r.Context(), `SELECT COALESCE(MIN(position), 0) FROM todos WHERE is_default_task=true`).Scan(&minPos)
	position := minPos - 1024.0

	// Insert admin todo (default task)
	_, err := h.db.Exec(r.Context(), `
		INSERT INTO todos(id, text, status, created, position, created_by_user_id, is_default_task, user_id, shared_with_admin) 
		VALUES($1,$2,$3,$4,$5,$6,$7,NULL,$8)
	`, id, req.Text, status, created, position, userID, true, false) // SharedWithAdmin is irrelevant for default tasks but defaulting to false
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
		IsDefaultTask:   true,
		SharedWithAdmin: false,
	})
}

// UpdateAdminTodo updates an admin todo's text (admin only)
func (h *Handler) UpdateAdminTodo(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}

	var req struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if req.Text != "" && len(req.Text) > 200 {
		http.Error(w, "text limit of 200 characters exceeded", http.StatusBadRequest)
		return
	}

	// Verify it's a default task
	var isDefaultTask bool
	err := h.db.QueryRow(r.Context(), `SELECT is_default_task FROM todos WHERE id=$1`, id).Scan(&isDefaultTask)
	if err != nil {
		http.Error(w, "todo not found", http.StatusNotFound)
		return
	}
	if !isDefaultTask {
		http.Error(w, "not a default task", http.StatusBadRequest)
		return
	}

	// Update text only
	_, err = h.db.Exec(r.Context(), `UPDATE todos SET text = $1 WHERE id = $2`, req.Text, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Fetch and return updated todo
	var t Todo
	if err := h.db.QueryRow(r.Context(), `
		SELECT id, text, status, created, position, created_by_user_id, is_default_task, shared_with_admin
		FROM todos 
		WHERE id=$1
	`, id).Scan(&t.ID, &t.Text, &t.Status, &t.Created, &t.Position, &t.CreatedByUserID, &t.IsDefaultTask, &t.SharedWithAdmin); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

// DeleteAdminTodo deletes an admin todo (admin only)
func (h *Handler) DeleteAdminTodo(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}

	// Verify it's a default task
	var isDefaultTask bool
	err := h.db.QueryRow(r.Context(), `SELECT is_default_task FROM todos WHERE id=$1`, id).Scan(&isDefaultTask)
	if err != nil {
		http.Error(w, "todo not found", http.StatusNotFound)
		return
	}
	if !isDefaultTask {
		http.Error(w, "not a default task", http.StatusBadRequest)
		return
	}

	// Delete the todo
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

// ListUserTodos returns all todos for a specific user (admin only)
func (h *Handler) ListUserTodos(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userId")
	if userID == "" {
		http.Error(w, "userId required", http.StatusBadRequest)
		return
	}

	// Verify user exists
	var exists bool
	err := h.db.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM users WHERE id=$1)`, userID).Scan(&exists)
	if err != nil || !exists {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	// Get user's todos (personal + default with their specific status)
	// IMPORTANT: For personal todos, ONLY show if shared_with_admin = true
	rows, err := h.db.Query(r.Context(), `
		SELECT 
			t.id, 
			t.text, 
			CASE 
				WHEN t.is_default_task THEN COALESCE(uts.status, t.status)
				ELSE t.status
			END as status,
			t.created, 
			CASE 
				WHEN t.is_default_task THEN COALESCE(uts.position, t.position)
				ELSE t.position
			END as position,
			t.created_by_user_id, 
			t.is_default_task,
			t.shared_with_admin,
			t.user_id
		FROM todos t
		LEFT JOIN user_todo_state uts ON t.id = uts.todo_id AND uts.user_id = $1 AND t.is_default_task = true
		WHERE 
			(t.user_id = $1 AND t.shared_with_admin = true) -- Show personal only if shared
			OR 
			(t.is_default_task = true) -- Always show default tasks
		ORDER BY position ASC, created DESC
	`, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.Text, &t.Status, &t.Created, &t.Position, &t.CreatedByUserID, &t.IsDefaultTask, &t.SharedWithAdmin, &t.UserID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		todos = append(todos, t)
	}

	if todos == nil {
		todos = []Todo{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"todos": todos})
}

// UpdateUserTodo updates a specific user's todo status (admin only)
func (h *Handler) UpdateUserTodo(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userId")
	todoID := r.PathValue("todoId")
	if userID == "" || todoID == "" {
		http.Error(w, "userId and todoId required", http.StatusBadRequest)
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	// Verify user exists
	var exists bool
	err := h.db.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM users WHERE id=$1)`, userID).Scan(&exists)
	if err != nil || !exists {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	// Check if todo is a default task
	var isDefaultTask bool
	err = h.db.QueryRow(r.Context(), `SELECT is_default_task FROM todos WHERE id=$1`, todoID).Scan(&isDefaultTask)
	if err != nil {
		http.Error(w, "todo not found", http.StatusNotFound)
		return
	}

	if !isDefaultTask {
		// As per requirements, admins cannot change status of personal shared tasks
		http.Error(w, "forbidden: admins can only update status of default tasks", http.StatusForbidden)
		return
	}

	// For default tasks, UPSERT into user_todo_state
	// We only update status, position remains checked/default
	_, err = h.db.Exec(r.Context(), `
		INSERT INTO user_todo_state (user_id, todo_id, status, position, updated_at)
		VALUES ($1, $2, $3, 
			(SELECT COALESCE(
				(SELECT position FROM user_todo_state WHERE user_id=$1 AND todo_id=$2),
				(SELECT position FROM todos WHERE id=$2)
			)),
			NOW()
		)
		ON CONFLICT (user_id, todo_id) 
		DO UPDATE SET status = EXCLUDED.status, updated_at = NOW()
	`, userID, todoID, req.Status)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success":true}`))
}
