package todo

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

// RegisterRoutes registers the specific routes to a mux using Go 1.22 enhanced routing
func (h *Handler) RegisterRoutes(mux *http.ServeMux, middleware func(http.HandlerFunc) http.HandlerFunc) {
	mux.HandleFunc("GET /api/todos", middleware(h.List))
	mux.HandleFunc("POST /api/todos", middleware(h.Create))
	mux.HandleFunc("PUT /api/todos/{id}", middleware(h.Update))
	mux.HandleFunc("PUT /api/todos/reorder", middleware(h.Reorder))
	mux.HandleFunc("DELETE /api/todos/{id}", middleware(h.Delete))
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w)
		return
	}

	// Check if this is explicitly requested (e.g. for some future use case), but default to false.
	// We want Admins to see global default tasks in their personal view as well, acting as "normal users".
	userRole, _ := auth.GetUserRole(r.Context())
	excludeAdminTodos := r.URL.Query().Get("exclude_admin_todos") == "true" || userRole == "admin"

	var query string
	if excludeAdminTodos {
		// Only return user's personal todos (exclude default tasks)
		query = `
			SELECT 
				t.id, 
				t.text, 
				t.status,
				t.created, 
				t.position,
				t.created_by_user_id, 
				t.is_default_task,
				t.shared_with_admin,
				t.hidden_from_user
			FROM todos t
			WHERE t.user_id = $1 AND t.is_default_task = false
			ORDER BY position ASC, created DESC 
			LIMIT 100
		`
	} else {
		// Return user's own todos + all default tasks
		// For default tasks, use user-specific status/position from user_todo_state if exists
		query = `
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
				t.hidden_from_user
			FROM todos t
			LEFT JOIN user_todo_state uts ON t.id = uts.todo_id AND uts.user_id = $1 AND t.is_default_task = true
			WHERE (t.user_id = $1 OR t.is_default_task = true) AND t.hidden_from_user = false
			ORDER BY position ASC, created DESC 
			LIMIT 100
		`
	}

	rows, err := h.db.Query(r.Context(), query, userID)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	defer rows.Close()

	var todos []models.Todo
	for rows.Next() {
		var t models.Todo
		if err := rows.Scan(&t.ID, &t.Text, &t.Status, &t.Created, &t.Position, &t.CreatedByUserID, &t.IsDefaultTask, &t.SharedWithAdmin, &t.HiddenFromUser); err != nil {
			httputil.InternalError(w, err.Error())
			return
		}
		todos = append(todos, t)
	}

	// Return empty array instead of null if nil
	if todos == nil {
		todos = []models.Todo{}
	}

	httputil.WriteJSON(w, map[string]any{"todos": todos}, http.StatusOK)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w)
		return
	}

	// userRole is not needed for personal todo creation anymore
	// userRole, _ := r.Context().Value("user_role").(string)

	var req struct {
		Text            string `json:"text"`
		SharedWithAdmin *bool  `json:"shared_with_admin"`
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

	// This endpoint is for PERSONAL todos only.
	// Admin (Global, Default) todos are created via the Admin API.
	isDefaultTask := false

	// Default to shared (true) unless explicitly set to false
	sharedWithAdmin := true
	if req.SharedWithAdmin != nil {
		sharedWithAdmin = *req.SharedWithAdmin
	}

	// Get min position for this user to put at top
	var minPos float64
	_ = h.db.QueryRow(r.Context(), `SELECT COALESCE(MIN(position), 0) FROM todos WHERE user_id=$1`, userID).Scan(&minPos)
	position := minPos - models.PositionIncrement

	// Insert personal todo
	_, err := h.db.Exec(r.Context(), `
		INSERT INTO todos(id, text, status, created, position, user_id, created_by_user_id, is_default_task, shared_with_admin) 
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)
	`, id, req.Text, status, created, position, userID, userID, isDefaultTask, sharedWithAdmin)

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
		IsDefaultTask:   isDefaultTask,
		SharedWithAdmin: sharedWithAdmin,
	}, http.StatusCreated)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		httputil.BadRequest(w, "id required")
		return
	}

	var req struct {
		Text            string `json:"text"`
		Status          string `json:"status"`
		SharedWithAdmin *bool  `json:"shared_with_admin,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid json")
		return
	}

	if req.Text != "" && !models.ValidateText(req.Text) {
		httputil.BadRequest(w, fmt.Sprintf("text limit of %d characters exceeded", models.MaxTextLength))
		return
	}

	// Validate status if provided
	if req.Status != "" {
		if !models.TodoStatus(req.Status).IsValid() {
			httputil.BadRequest(w, "invalid status")
			return
		}
	}

	// Check if todo exists and if user can update it
	var isDefaultTask bool
	var todoUserID *string
	err := h.db.QueryRow(r.Context(), `
		SELECT is_default_task, user_id 
		FROM todos 
		WHERE id=$1
	`, id).Scan(&isDefaultTask, &todoUserID)

	if err != nil {
		httputil.NotFound(w, "todo not found")
		return
	}

	// Users can update:
	// 1. Their own todos (user_id matches)
	// 2. Default tasks (is_default_task=true)
	canUpdate := isDefaultTask || (todoUserID != nil && *todoUserID == userID)
	if !canUpdate {
		httputil.Forbidden(w, "forbidden")
		return
	}

	// Handle update based on todo type
	if isDefaultTask {
		// For default tasks, update user_todo_state (per-user status)
		// Only update status if provided (text updates not allowed for default tasks)
		// Sharing cannot be toggled for default tasks
		if req.Status != "" {
			_, err = h.db.Exec(r.Context(), `
				INSERT INTO user_todo_state (user_id, todo_id, status, position, updated_at)
				VALUES ($1, $2, $3, 
					(SELECT COALESCE(
						(SELECT position FROM user_todo_state WHERE user_id=$1 AND todo_id=$2),
						(SELECT position FROM todos WHERE id=$2)
					)),
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
		// For personal todos, check permissions based on who created it
		var createdBy *string
		err := h.db.QueryRow(r.Context(), `SELECT created_by_user_id FROM todos WHERE id=$1`, id).Scan(&createdBy)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Check if this is an admin-created task (created_by != user_id)
		isAdminCreatedTask := createdBy != nil && *createdBy != userID

		if isAdminCreatedTask {
			// User can ONLY update status on admin-created tasks
			// Text and shared_with_admin updates are forbidden
			if req.Text != "" {
				httputil.Forbidden(w, "forbidden: cannot edit text of admin-assigned task")
				return
			}
			if req.SharedWithAdmin != nil {
				httputil.Forbidden(w, "forbidden: cannot change sharing status of admin-assigned task")
				return
			}
			// Allow status update only
			if req.Status != "" {
				query := "UPDATE todos SET status = $1 WHERE id = $2"
				_, err = h.db.Exec(r.Context(), query, req.Status, id)
			}
		} else {
			// User's own task - allow all updates
			// Build dynamic query
			query := "UPDATE todos SET "
			var args []interface{}
			argID := 1

			if req.Text != "" {
				query += fmt.Sprintf("text = $%d, ", argID)
				args = append(args, req.Text)
				argID++
			}
			if req.Status != "" {
				query += fmt.Sprintf("status = $%d, ", argID)
				args = append(args, req.Status)
				argID++
			}
			if req.SharedWithAdmin != nil {
				query += fmt.Sprintf("shared_with_admin = $%d, ", argID)
				args = append(args, *req.SharedWithAdmin)
				argID++
			}

			if len(args) > 0 {
				// Remove trailing comma and space
				query = query[:len(query)-2]
				query += fmt.Sprintf(" WHERE id = $%d", argID)
				args = append(args, id)
				_, err = h.db.Exec(r.Context(), query, args...)
			}
		}
	}
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}

	// Fetch and return updated todo with user-specific state
	var t models.Todo
	if err := h.db.QueryRow(r.Context(), `
		SELECT 
			t.id, t.text, 
			CASE 
				WHEN t.is_default_task THEN COALESCE(uts.status, t.status)
				ELSE t.status
			END as status,
			t.created, 
			CASE 
				WHEN t.is_default_task THEN COALESCE(uts.position, t.position)
				ELSE t.position
			END as position,
			t.created_by_user_id, t.is_default_task, t.shared_with_admin
		FROM todos t
		LEFT JOIN user_todo_state uts ON t.id = uts.todo_id AND uts.user_id = $2 AND t.is_default_task = true
		WHERE t.id=$1
	`, id, userID).Scan(&t.ID, &t.Text, &t.Status, &t.Created, &t.Position, &t.CreatedByUserID, &t.IsDefaultTask, &t.SharedWithAdmin); err != nil {
		httputil.InternalError(w, err.Error())
		return
	}

	httputil.WriteJSON(w, t, http.StatusOK)
}

func (h *Handler) Reorder(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w)
		return
	}

	var req struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid json")
		return
	}

	if len(req.IDs) == 0 {
		return
	}

	tx, err := h.db.Begin(r.Context())
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	defer tx.Rollback(r.Context())

	for i, id := range req.IDs {
		pos := float64(i) * models.PositionIncrement

		// Check if this is a default task
		var isDefaultTask bool
		err := tx.QueryRow(r.Context(), `SELECT is_default_task FROM todos WHERE id=$1`, id).Scan(&isDefaultTask)
		if err != nil {
			httputil.InternalError(w, err.Error())
			return
		}

		if isDefaultTask {
			// For default tasks, UPSERT into user_todo_state
			_, err := tx.Exec(r.Context(), `
				INSERT INTO user_todo_state (user_id, todo_id, status, position, updated_at)
				VALUES ($1, $2, 
					(SELECT COALESCE(status, 'pending') FROM user_todo_state WHERE user_id=$1 AND todo_id=$2),
					$3, now())
				ON CONFLICT (user_id, todo_id) 
				DO UPDATE SET position = $3, updated_at = now()
			`, userID, id, pos)
			if err != nil {
				httputil.InternalError(w, err.Error())
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
				httputil.InternalError(w, err.Error())
				return
			}
		}
	}

	if err := tx.Commit(r.Context()); err != nil {
		httputil.InternalError(w, err.Error())
		return
	}

	httputil.WriteSuccess(w)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		httputil.Unauthorized(w)
		return
	}

	userRole, _ := auth.GetUserRole(r.Context())

	id := r.PathValue("id")
	if id == "" {
		httputil.BadRequest(w, "id required")
		return
	}

	// Check if todo is a default task and who created it
	var isDefaultTask bool
	var todoUserID *string
	var createdByUserID *string
	err := h.db.QueryRow(r.Context(), `
		SELECT is_default_task, user_id, created_by_user_id 
		FROM todos 
		WHERE id=$1
	`, id).Scan(&isDefaultTask, &todoUserID, &createdByUserID)

	if err != nil {
		httputil.NotFound(w, "todo not found")
		return
	}

	// Only admins can delete default tasks
	if isDefaultTask && userRole != "admin" {
		httputil.Forbidden(w, "forbidden: only admins can delete default tasks")
		return
	}

	// Regular users can only delete their own todos (where they are the owner)
	if !isDefaultTask && (todoUserID == nil || *todoUserID != userID) {
		httputil.Forbidden(w, "forbidden")
		return
	}

	// Users can only delete todos they created themselves (not admin-created tasks)
	if !isDefaultTask && userRole != "admin" {
		if createdByUserID != nil && *createdByUserID != userID {
			httputil.Forbidden(w, "forbidden: cannot delete admin-assigned task")
			return
		}
	}

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
