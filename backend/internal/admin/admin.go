package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/akhilmk/packup/internal/auth"
	"github.com/akhilmk/packup/internal/httputil"
	"github.com/akhilmk/packup/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	db *pgxpool.Pool
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{db: db}
}

// RegisterRoutes registers the admin routes to a mux using Go 1.22 enhanced routing
func (h *Handler) RegisterRoutes(mux *http.ServeMux, adminMiddleware func(http.HandlerFunc) http.HandlerFunc) {
	mux.HandleFunc("GET /api/admin/users", adminMiddleware(h.ListUsers))
	mux.HandleFunc("GET /api/admin/todos", adminMiddleware(h.ListAdminTodos))
	mux.HandleFunc("POST /api/admin/todos", adminMiddleware(h.CreateAdminTodo))
	mux.HandleFunc("PUT /api/admin/todos/{id}", adminMiddleware(h.UpdateAdminTodo))
	mux.HandleFunc("DELETE /api/admin/todos/{id}", adminMiddleware(h.DeleteAdminTodo))
	mux.HandleFunc("GET /api/admin/users/{userId}/todos", adminMiddleware(h.ListUserTodos))
	mux.HandleFunc("POST /api/admin/users/{userId}/todos", adminMiddleware(h.CreateUserTodo))
	mux.HandleFunc("PUT /api/admin/users/{userId}/todos/{todoId}", adminMiddleware(h.UpdateUserTodo))
	mux.HandleFunc("DELETE /api/admin/users/{userId}/todos/{todoId}", adminMiddleware(h.DeleteUserTodo))
}

// Middleware to check if user is admin
func (h *Handler) RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !auth.IsAdmin(r.Context()) {
			httputil.Forbidden(w, "forbidden: admin access required")
			return
		}
		next(w, r)
	}
}

// ListUsers returns all users (admin only)
// ListUsers returns all users excluding admins.
// @Summary List users
// @Description Get a list of all users excluding admins.
// @Tags admin
// @Produce json
// @Success 200 {object} map[string][]models.User
// @Failure 401 {object} httputil.APIError
// @Failure 403 {object} httputil.APIError
// @Failure 500 {object} httputil.APIError
// @Router /api/admin/users [get]
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// ListUsers returns all users excluding admins (admin only)
	rows, err := h.db.Query(r.Context(), `
		SELECT id, email, name, avatar_url, role, created_at 
		FROM users 
		WHERE role != 'admin'
		ORDER BY created_at DESC
	`)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.AvatarURL, &u.Role, &u.CreatedAt); err != nil {
			httputil.InternalError(w, err.Error())
			return
		}
		users = append(users, u)
	}

	if users == nil {
		users = []models.User{}
	}

	httputil.WriteJSON(w, map[string]any{"users": users}, http.StatusOK)
}

// ListAdminTodos returns all admin todos (admin only)
// ListAdminTodos returns all global default tasks.
// @Summary List global default tasks
// @Description Get a list of all global default tasks.
// @Tags admin
// @Produce json
// @Success 200 {object} map[string][]models.Todo
// @Failure 401 {object} httputil.APIError
// @Failure 403 {object} httputil.APIError
// @Failure 500 {object} httputil.APIError
// @Router /api/admin/todos [get]
func (h *Handler) ListAdminTodos(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(r.Context(), `
		SELECT id, text, status, created, position, created_by_user_id, is_default_task, shared_with_admin
		FROM todos 
		WHERE is_default_task = true 
		ORDER BY position ASC, created DESC
	`)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	defer rows.Close()

	var todos []models.Todo
	for rows.Next() {
		var t models.Todo
		if err := rows.Scan(&t.ID, &t.Text, &t.Status, &t.Created, &t.Position, &t.CreatedByUserID, &t.IsDefaultTask, &t.SharedWithAdmin); err != nil {
			httputil.InternalError(w, err.Error())
			return
		}
		todos = append(todos, t)
	}

	if todos == nil {
		todos = []models.Todo{}
	}

	httputil.WriteJSON(w, map[string]any{"todos": todos}, http.StatusOK)
}

// CreateAdminTodo creates a new admin todo (admin only)
// CreateAdminTodo creates a new global default task.
// @Summary Create global default task
// @Description Create a new global default task.
// @Tags admin
// @Accept json
// @Produce json
// @Param todo body object true "Todo text"
// @Success 201 {object} models.Todo
// @Failure 400 {object} httputil.APIError
// @Failure 401 {object} httputil.APIError
// @Failure 403 {object} httputil.APIError
// @Failure 500 {object} httputil.APIError
// @Router /api/admin/todos [post]
func (h *Handler) CreateAdminTodo(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w)
		return
	}

	var req struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid json")
		return
	}

	if !models.ValidateText(req.Text) {
		httputil.BadRequest(w, fmt.Sprintf("text cannot be empty or exceed %d characters", models.MaxTextLength))
		return
	}

	id := uuid.NewString()
	status := string(models.StatusPending)
	created := time.Now()

	// Get min position to put at top
	var minPos float64
	_ = h.db.QueryRow(r.Context(), `SELECT COALESCE(MIN(position), 0) FROM todos WHERE is_default_task=true`).Scan(&minPos)
	position := minPos - models.PositionIncrement

	// Insert admin todo (default task)
	_, err := h.db.Exec(r.Context(), `
		INSERT INTO todos(id, text, status, created, position, created_by_user_id, is_default_task, user_id, shared_with_admin) 
		VALUES($1,$2,$3,$4,$5,$6,$7,NULL,$8)
	`, id, req.Text, status, created, position, userID, true, false) // SharedWithAdmin is irrelevant for default tasks but defaulting to false
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}

	createdByUserID := userID
	httputil.WriteJSON(w, models.Todo{
		ID:              id,
		Text:            req.Text,
		Status:          status,
		Created:         created,
		Position:        position,
		CreatedByUserID: &createdByUserID,
		IsDefaultTask:   true,
		SharedWithAdmin: false,
	}, http.StatusCreated)
}

// UpdateAdminTodo updates an admin todo's text (admin only)
// UpdateAdminTodo updates a global default task's text.
// @Summary Update global default task
// @Description Update a global default task's text.
// @Tags admin
// @Accept json
// @Produce json
// @Param id path string true "Todo ID"
// @Param todo body object true "New text"
// @Success 200 {object} models.Todo
// @Failure 400 {object} httputil.APIError
// @Failure 401 {object} httputil.APIError
// @Failure 403 {object} httputil.APIError
// @Failure 404 {object} httputil.APIError
// @Failure 500 {object} httputil.APIError
// @Router /api/admin/todos/{id} [put]
func (h *Handler) UpdateAdminTodo(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		httputil.BadRequest(w, "id required")
		return
	}

	var req struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid json")
		return
	}

	if req.Text != "" && !models.ValidateText(req.Text) {
		httputil.BadRequest(w, fmt.Sprintf("text limit of %d characters exceeded", models.MaxTextLength))
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
		httputil.InternalError(w, err.Error())
		return
	}

	// Fetch and return updated todo
	var t models.Todo
	if err := h.db.QueryRow(r.Context(), `
		SELECT id, text, status, created, position, created_by_user_id, is_default_task, shared_with_admin
		FROM todos 
		WHERE id=$1
	`, id).Scan(&t.ID, &t.Text, &t.Status, &t.Created, &t.Position, &t.CreatedByUserID, &t.IsDefaultTask, &t.SharedWithAdmin); err != nil {
		httputil.InternalError(w, err.Error())
		return
	}

	httputil.WriteJSON(w, t, http.StatusOK)
}

// DeleteAdminTodo deletes an admin todo (admin only)
// DeleteAdminTodo deletes a global default task.
// @Summary Delete global default task
// @Description Delete a global default task.
// @Tags admin
// @Produce json
// @Param id path string true "Todo ID"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} httputil.APIError
// @Failure 401 {object} httputil.APIError
// @Failure 403 {object} httputil.APIError
// @Failure 404 {object} httputil.APIError
// @Failure 500 {object} httputil.APIError
// @Router /api/admin/todos/{id} [delete]
func (h *Handler) DeleteAdminTodo(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		httputil.BadRequest(w, "id required")
		return
	}

	// Verify it's a default task
	var isDefaultTask bool
	err := h.db.QueryRow(r.Context(), `SELECT is_default_task FROM todos WHERE id=$1`, id).Scan(&isDefaultTask)
	if err != nil {
		httputil.NotFound(w, "todo not found")
		return
	}
	if !isDefaultTask {
		httputil.BadRequest(w, "not a default task")
		return
	}

	// Delete the todo
	cmd, err := h.db.Exec(r.Context(), `DELETE FROM todos WHERE id=$1`, id)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	if cmd.RowsAffected() == 0 {
		httputil.NotFound(w, "todo not found")
		return
	}

	httputil.WriteSuccess(w)
}

// ListUserTodos returns all todos for a specific user (admin only)
// ListUserTodos returns all todos for a specific user.
// @Summary List user's todos
// @Description Get all personal (if shared) and default tasks for a specific user.
// @Tags admin
// @Produce json
// @Param userId path string true "User ID"
// @Success 200 {object} map[string][]models.Todo
// @Failure 401 {object} httputil.APIError
// @Failure 403 {object} httputil.APIError
// @Failure 404 {object} httputil.APIError
// @Failure 500 {object} httputil.APIError
// @Router /api/admin/users/{userId}/todos [get]
func (h *Handler) ListUserTodos(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userId")
	if userID == "" {
		httputil.BadRequest(w, "userId required")
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
			t.user_id,
			t.hidden_from_user
		FROM todos t
		LEFT JOIN user_todo_state uts ON t.id = uts.todo_id AND uts.user_id = $1 AND t.is_default_task = true
		WHERE 
			(t.user_id = $1 AND t.shared_with_admin = true) -- Show personal only if shared
			OR 
			(t.is_default_task = true) -- Always show default tasks
		ORDER BY position ASC, created DESC
	`, userID)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	defer rows.Close()

	var todos []models.Todo
	for rows.Next() {
		var t models.Todo
		if err := rows.Scan(&t.ID, &t.Text, &t.Status, &t.Created, &t.Position, &t.CreatedByUserID, &t.IsDefaultTask, &t.SharedWithAdmin, &t.UserID, &t.HiddenFromUser); err != nil {
			httputil.InternalError(w, err.Error())
			return
		}
		todos = append(todos, t)
	}

	if todos == nil {
		todos = []models.Todo{}
	}

	httputil.WriteJSON(w, map[string]any{"todos": todos}, http.StatusOK)
}

// CreateUserTodo creates a new todo for a specific user (admin only)
// CreateUserTodo creates a new todo for a specific user.
// @Summary Create todo for user
// @Description Create a new personal todo for a specific user (admin created).
// @Tags admin
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Param todo body object true "Todo content"
// @Success 201 {object} models.Todo
// @Failure 400 {object} httputil.APIError
// @Failure 401 {object} httputil.APIError
// @Failure 403 {object} httputil.APIError
// @Failure 500 {object} httputil.APIError
// @Router /api/admin/users/{userId}/todos [post]
func (h *Handler) CreateUserTodo(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("userId")
	if userId == "" {
		httputil.BadRequest(w, "userId required")
		return
	}

	adminID, ok := auth.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w)
		return
	}

	var req struct {
		Text           string `json:"text"`
		HiddenFromUser bool   `json:"hidden_from_user"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid json")
		return
	}

	if !models.ValidateText(req.Text) {
		httputil.BadRequest(w, fmt.Sprintf("text cannot be empty or exceed %d characters", models.MaxTextLength))
		return
	}

	id := uuid.NewString()
	status := string(models.StatusPending)
	created := time.Now()

	// Get min position for this user to put at top
	var minPos float64
	_ = h.db.QueryRow(r.Context(), `SELECT COALESCE(MIN(position), 0) FROM todos WHERE user_id=$1`, userId).Scan(&minPos)
	position := minPos - models.PositionIncrement

	// Insert todo for user, created by admin
	_, err := h.db.Exec(r.Context(), `
		INSERT INTO todos(id, text, status, created, position, user_id, created_by_user_id, is_default_task, shared_with_admin, hidden_from_user) 
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
	`, id, req.Text, status, created, position, userId, adminID, false, true, req.HiddenFromUser)

	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}

	createdByUserID := adminID
	httputil.WriteJSON(w, models.Todo{
		ID:              id,
		Text:            req.Text,
		Status:          status,
		Created:         created,
		Position:        position,
		CreatedByUserID: &createdByUserID,
		IsDefaultTask:   false,
		SharedWithAdmin: true,
		HiddenFromUser:  req.HiddenFromUser,
		UserID:          &userId,
	}, http.StatusCreated)
}

// UpdateUserTodo updates a specific user's todo status (admin only)
// UpdateUserTodo updates a specific user's todo status or text.
// @Summary Update user's todo
// @Description Update a specific user's personal or default todo status/text.
// @Tags admin
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Param todoId path string true "Todo ID"
// @Param todo body object true "Update content"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} httputil.APIError
// @Failure 401 {object} httputil.APIError
// @Failure 403 {object} httputil.APIError
// @Failure 404 {object} httputil.APIError
// @Failure 500 {object} httputil.APIError
// @Router /api/admin/users/{userId}/todos/{todoId} [put]
func (h *Handler) UpdateUserTodo(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userId")
	todoID := r.PathValue("todoId")
	if userID == "" || todoID == "" {
		httputil.BadRequest(w, "userId and todoId required")
		return
	}

	var req struct {
		Text           *string `json:"text,omitempty"`
		Status         *string `json:"status,omitempty"`
		HiddenFromUser *bool   `json:"hidden_from_user,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid json")
		return
	}

	// Verify user exists
	var exists bool
	err := h.db.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM users WHERE id=$1)`, userID).Scan(&exists)
	if err != nil || !exists {
		httputil.NotFound(w, "user not found")
		return
	}

	// Check if todo is a default task
	var isDefaultTask bool
	var createdByUserID *string
	err = h.db.QueryRow(r.Context(), `SELECT is_default_task, created_by_user_id FROM todos WHERE id=$1`, todoID).Scan(&isDefaultTask, &createdByUserID)
	if err != nil {
		httputil.NotFound(w, "todo not found")
		return
	}

	if isDefaultTask {
		if req.HiddenFromUser != nil {
			// Admins shouldn't be making default tasks hidden locally for a user (not requested, complicates logic)
			// But if they want to change status, they can.
		}

		// For default tasks, UPSERT into user_todo_state
		// We only update status, position remains checked/default
		if req.Status != nil {
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
			`, userID, todoID, *req.Status)

			if err != nil {
				httputil.InternalError(w, err.Error())
				return
			}
		}
	} else {
		// Personal / Admin-Assigned User Task
		// Allow updating hidden_from_user status and text (for admin-created tasks only)
		if req.HiddenFromUser != nil {
			_, err = h.db.Exec(r.Context(), `UPDATE todos SET hidden_from_user = $1 WHERE id = $2`, *req.HiddenFromUser, todoID)
			if err != nil {
				httputil.InternalError(w, err.Error())
				return
			}
		}

		// Allow text update only for admin-created tasks (where created_by != user_id)
		if req.Text != nil {
			// Check if admin created this task (created_by_user_id != user_id means admin created it)
			if createdByUserID != nil && *createdByUserID != userID {
				if !models.ValidateText(*req.Text) {
					httputil.BadRequest(w, fmt.Sprintf("text cannot be empty or exceed %d characters", models.MaxTextLength))
					return
				}
				_, err = h.db.Exec(r.Context(), `UPDATE todos SET text = $1 WHERE id = $2`, *req.Text, todoID)
				if err != nil {
					httputil.InternalError(w, err.Error())
					return
				}
			} else {
				httputil.Forbidden(w, "cannot edit text of user-created tasks")
				return
			}
		}

		// Allow status update for any user task (Shared Responsibility)
		// Both Admin and User can update status of shared tasks.
		if req.Status != nil {
			// Update status in todos table
			_, err = h.db.Exec(r.Context(), `UPDATE todos SET status = $1 WHERE id = $2`, *req.Status, todoID)
			if err != nil {
				httputil.InternalError(w, err.Error())
				return
			}
		}
	}

	httputil.WriteSuccess(w)
}

// DeleteUserTodo deletes a user-specific todo that was created by an admin (admin only)
// DeleteUserTodo deletes a user-specific todo created by an admin.
// @Summary Delete user's admin-created todo
// @Description Delete a personal todo that was created for the user by an admin.
// @Tags admin
// @Produce json
// @Param userId path string true "User ID"
// @Param todoId path string true "Todo ID"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} httputil.APIError
// @Failure 401 {object} httputil.APIError
// @Failure 403 {object} httputil.APIError
// @Failure 404 {object} httputil.APIError
// @Failure 500 {object} httputil.APIError
// @Router /api/admin/users/{userId}/todos/{todoId} [delete]
func (h *Handler) DeleteUserTodo(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userId")
	todoID := r.PathValue("todoId")
	if userID == "" || todoID == "" {
		httputil.BadRequest(w, "userId and todoId required")
		return
	}

	// Check if todo exists and verify it's an admin-created task for this user
	var isDefaultTask bool
	var todoUserID *string
	var createdByUserID *string
	err := h.db.QueryRow(r.Context(), `
		SELECT is_default_task, user_id, created_by_user_id 
		FROM todos 
		WHERE id=$1
	`, todoID).Scan(&isDefaultTask, &todoUserID, &createdByUserID)

	if err != nil {
		httputil.NotFound(w, "todo not found")
		return
	}

	// Cannot delete default tasks through this endpoint
	if isDefaultTask {
		httputil.BadRequest(w, "use the default task endpoint to delete default tasks")
		return
	}

	// Verify the todo belongs to the specified user
	if todoUserID == nil || *todoUserID != userID {
		httputil.Forbidden(w, "todo does not belong to this user")
		return
	}

	// Only allow deleting admin-created tasks (created_by != user_id)
	if createdByUserID == nil || *createdByUserID == userID {
		httputil.Forbidden(w, "can only delete admin-created tasks")
		return
	}

	// Delete the todo
	cmd, err := h.db.Exec(r.Context(), `DELETE FROM todos WHERE id=$1`, todoID)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	if cmd.RowsAffected() == 0 {
		httputil.NotFound(w, "todo not found")
		return
	}

	httputil.WriteSuccess(w)
}
