package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/akhilmk/itinera/internal/auth"
	"github.com/akhilmk/itinera/internal/models"
)

// Test structs
func TestUserStruct(t *testing.T) {
	user := models.User{
		ID:        "user-1",
		Email:     "test@example.com",
		Name:      "Test User",
		AvatarURL: "https://example.com/avatar.png",
		Role:      "user",
		CreatedAt: time.Now(),
	}

	if user.ID != "user-1" {
		t.Errorf("Expected ID 'user-1', got '%s'", user.ID)
	}
	if user.Email != "test@example.com" {
		t.Errorf("Expected Email 'test@example.com', got '%s'", user.Email)
	}
	if user.Role != "user" {
		t.Errorf("Expected Role 'user', got '%s'", user.Role)
	}
}

func TestTodoStruct(t *testing.T) {
	t.Run("Todo with HiddenFromUser", func(t *testing.T) {
		todo := models.Todo{
			ID:              "todo-1",
			Text:            "Admin assigned task",
			Status:          "pending",
			Created:         time.Now(),
			Position:        1024.0,
			IsDefaultTask:   false,
			SharedWithAdmin: true,
			HiddenFromUser:  true,
		}

		if todo.HiddenFromUser != true {
			t.Error("Expected HiddenFromUser to be true")
		}
	})
}

// TestRequireAdminMiddleware tests the admin middleware
func TestRequireAdminMiddleware(t *testing.T) {
	handler := &Handler{db: nil}

	t.Run("Non-admin user is forbidden", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/api/admin/users", nil)
		// Add non-admin user context
		ctx := req.Context()
		ctx = auth.SetUserContext(ctx, "user-123", "user")
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.RequireAdmin(next)(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status %d for non-admin, got %d", http.StatusForbidden, w.Code)
		}
	})

	t.Run("Admin user is allowed", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/api/admin/users", nil)
		ctx := req.Context()
		ctx = auth.SetUserContext(ctx, "admin-123", "admin")
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.RequireAdmin(next)(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d for admin, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("No role is forbidden", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/api/admin/users", nil)
		w := httptest.NewRecorder()

		handler.RequireAdmin(next)(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status %d when no role, got %d", http.StatusForbidden, w.Code)
		}
	})
}

// TestCreateAdminTodoUnauthorized tests that CreateAdminTodo fails without auth
func TestCreateAdminTodoUnauthorized(t *testing.T) {
	handler := &Handler{db: nil}

	body := bytes.NewBufferString(`{"text":"Default task"}`)
	req := httptest.NewRequest("POST", "/api/admin/todos", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateAdminTodo(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

// TestCreateAdminTodoTextValidation tests text validation
func TestCreateAdminTodoTextValidation(t *testing.T) {
	t.Run("Empty text", func(t *testing.T) {
		handler := &Handler{db: nil}

		body := bytes.NewBufferString(`{"text":""}`)
		req := httptest.NewRequest("POST", "/api/admin/todos", body)
		req.Header.Set("Content-Type", "application/json")
		ctx := req.Context()
		ctx = auth.SetUserContext(ctx, "admin-123", "admin")
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.CreateAdminTodo(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for empty text, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Text over 200 characters", func(t *testing.T) {
		handler := &Handler{db: nil}

		longText := make([]byte, 201)
		for i := range longText {
			longText[i] = 'a'
		}

		bodyStr := `{"text":"` + string(longText) + `"}`
		body := bytes.NewBufferString(bodyStr)
		req := httptest.NewRequest("POST", "/api/admin/todos", body)
		req.Header.Set("Content-Type", "application/json")
		ctx := req.Context()
		ctx = auth.SetUserContext(ctx, "admin-123", "admin")
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.CreateAdminTodo(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for long text, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

// TestCreateUserTodoValidation tests CreateUserTodo validation
func TestCreateUserTodoValidation(t *testing.T) {
	t.Run("Missing userId", func(t *testing.T) {
		handler := &Handler{db: nil}

		body := bytes.NewBufferString(`{"text":"Task for user"}`)
		req := httptest.NewRequest("POST", "/api/admin/users//todos", body)
		req.Header.Set("Content-Type", "application/json")
		ctx := context.WithValue(req.Context(), "user_id", "admin-123")
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.CreateUserTodo(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for missing userId, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Unauthorized", func(t *testing.T) {
		handler := &Handler{db: nil}

		body := bytes.NewBufferString(`{"text":"Task for user"}`)
		req := httptest.NewRequest("POST", "/api/admin/users/user-123/todos", body)
		req.Header.Set("Content-Type", "application/json")
		// Note: not setting pathvalue properly but userId will be empty
		w := httptest.NewRecorder()

		handler.CreateUserTodo(w, req)

		// userId will be empty from PathValue, so we expect bad request first
		if w.Code != http.StatusBadRequest && w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d or %d, got %d", http.StatusBadRequest, http.StatusUnauthorized, w.Code)
		}
	})
}

// TestHiddenFromUserRequest tests the request struct for hidden_from_user
func TestHiddenFromUserRequest(t *testing.T) {
	t.Run("Parse with hidden_from_user true", func(t *testing.T) {
		var req struct {
			Text           string `json:"text"`
			HiddenFromUser bool   `json:"hidden_from_user"`
		}

		jsonData := `{"text":"test","hidden_from_user":true}`
		if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if req.HiddenFromUser != true {
			t.Errorf("Expected HiddenFromUser to be true, got %v", req.HiddenFromUser)
		}
	})

	t.Run("Parse with hidden_from_user false (default)", func(t *testing.T) {
		var req struct {
			Text           string `json:"text"`
			HiddenFromUser bool   `json:"hidden_from_user"`
		}

		jsonData := `{"text":"test"}`
		if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		// Default value for bool is false
		if req.HiddenFromUser != false {
			t.Errorf("Expected HiddenFromUser to default to false, got %v", req.HiddenFromUser)
		}
	})
}

// TestUpdateUserTodoValidation tests UpdateUserTodo validation
func TestUpdateUserTodoValidation(t *testing.T) {
	t.Run("Missing userId and todoId", func(t *testing.T) {
		handler := &Handler{db: nil}

		body := bytes.NewBufferString(`{"status":"done"}`)
		req := httptest.NewRequest("PUT", "/api/admin/users//todos/", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.UpdateUserTodo(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for missing ids, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

// TestUpdateAdminTodoValidation tests UpdateAdminTodo validation
func TestUpdateAdminTodoValidation(t *testing.T) {
	t.Run("Missing id", func(t *testing.T) {
		handler := &Handler{db: nil}

		body := bytes.NewBufferString(`{"text":"Updated text"}`)
		req := httptest.NewRequest("PUT", "/api/admin/todos/", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.UpdateAdminTodo(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for missing id, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

// TestDeleteAdminTodoValidation tests DeleteAdminTodo validation
func TestDeleteAdminTodoValidation(t *testing.T) {
	t.Run("Missing id", func(t *testing.T) {
		handler := &Handler{db: nil}

		req := httptest.NewRequest("DELETE", "/api/admin/todos/", nil)
		w := httptest.NewRecorder()

		handler.DeleteAdminTodo(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for missing id, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

// TestListUserTodosValidation tests ListUserTodos validation
func TestListUserTodosValidation(t *testing.T) {
	t.Run("Missing userId", func(t *testing.T) {
		handler := &Handler{db: nil}

		req := httptest.NewRequest("GET", "/api/admin/users//todos", nil)
		w := httptest.NewRecorder()

		handler.ListUserTodos(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for missing userId, got %d", http.StatusBadRequest, w.Code)
		}
	})
}
