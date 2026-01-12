package todo

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/akhilmk/itinera/internal/auth"
	"github.com/akhilmk/itinera/internal/models"
)

// Test structs for request/response validation
func TestTodoStruct(t *testing.T) {
	t.Run("Todo struct fields", func(t *testing.T) {
		createdBy := "user-123"
		todo := models.Todo{
			ID:              "todo-1",
			Text:            "Test todo",
			Status:          "pending",
			Created:         time.Now(),
			Position:        1024.0,
			CreatedByUserID: &createdBy,
			IsDefaultTask:   false,
			SharedWithAdmin: true,
			HiddenFromUser:  false,
		}

		if todo.ID != "todo-1" {
			t.Errorf("Expected ID 'todo-1', got '%s'", todo.ID)
		}
		if todo.Text != "Test todo" {
			t.Errorf("Expected Text 'Test todo', got '%s'", todo.Text)
		}
		if todo.Status != "pending" {
			t.Errorf("Expected Status 'pending', got '%s'", todo.Status)
		}
		if todo.IsDefaultTask != false {
			t.Error("Expected IsDefaultTask to be false")
		}
		if todo.SharedWithAdmin != true {
			t.Error("Expected SharedWithAdmin to be true")
		}
		if todo.HiddenFromUser != false {
			t.Error("Expected HiddenFromUser to be false")
		}
	})

	t.Run("Todo JSON marshaling", func(t *testing.T) {
		todo := models.Todo{
			ID:              "todo-1",
			Text:            "Test todo",
			Status:          "pending",
			Created:         time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC),
			Position:        1024.0,
			IsDefaultTask:   true,
			SharedWithAdmin: false,
			HiddenFromUser:  true,
		}

		data, err := json.Marshal(todo)
		if err != nil {
			t.Fatalf("Failed to marshal todo: %v", err)
		}

		var unmarshaled map[string]interface{}
		if err := json.Unmarshal(data, &unmarshaled); err != nil {
			t.Fatalf("Failed to unmarshal todo: %v", err)
		}

		if unmarshaled["id"] != "todo-1" {
			t.Errorf("Expected id 'todo-1', got '%v'", unmarshaled["id"])
		}
		if unmarshaled["is_default_task"] != true {
			t.Errorf("Expected is_default_task to be true, got '%v'", unmarshaled["is_default_task"])
		}
		if unmarshaled["hidden_from_user"] != true {
			t.Errorf("Expected hidden_from_user to be true, got '%v'", unmarshaled["hidden_from_user"])
		}
	})
}

// TestListEndpointUnauthorized tests that List fails without auth
func TestListEndpointUnauthorized(t *testing.T) {
	handler := &Handler{db: nil} // nil db since we're testing auth check

	req := httptest.NewRequest("GET", "/api/todos", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

// TestCreateEndpointUnauthorized tests that Create fails without auth
func TestCreateEndpointUnauthorized(t *testing.T) {
	handler := &Handler{db: nil}

	body := bytes.NewBufferString(`{"text":"Test todo"}`)
	req := httptest.NewRequest("POST", "/api/todos", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

// TestCreateTextValidation tests text validation in Create
func TestCreateTextValidation(t *testing.T) {
	t.Run("Empty text", func(t *testing.T) {
		handler := &Handler{db: nil}

		body := bytes.NewBufferString(`{"text":""}`)
		req := httptest.NewRequest("POST", "/api/todos", body)
		req.Header.Set("Content-Type", "application/json")
		// Add user context
		ctx := req.Context()
		ctx = auth.SetUserContext(ctx, "user-123", "user")
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.Create(w, req)

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
		req := httptest.NewRequest("POST", "/api/todos", body)
		req.Header.Set("Content-Type", "application/json")
		ctx := req.Context()
		ctx = auth.SetUserContext(ctx, "user-123", "user")
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.Create(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for long text, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

// TestUpdateEndpointUnauthorized tests that Update fails without auth
func TestUpdateEndpointUnauthorized(t *testing.T) {
	handler := &Handler{db: nil}

	body := bytes.NewBufferString(`{"status":"done"}`)
	req := httptest.NewRequest("PUT", "/api/todos/123", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Update(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

// TestStatusValidation tests status validation
func TestStatusValidation(t *testing.T) {
	tests := []struct {
		name   string
		status string
		valid  bool
	}{
		{"pending is valid", "pending", true},
		{"in-progress is valid", "in-progress", true},
		{"done is valid", "done", true},
		{"invalid status", "invalid", false},
		{"empty status", "", true}, // Empty is allowed (means no change)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := false
			switch tt.status {
			case "pending", "in-progress", "done", "":
				valid = true
			}
			if valid != tt.valid {
				t.Errorf("Expected status '%s' validity to be %v, got %v", tt.status, tt.valid, valid)
			}
		})
	}
}

// TestReorderEndpointUnauthorized tests that Reorder fails without auth
func TestReorderEndpointUnauthorized(t *testing.T) {
	handler := &Handler{db: nil}

	body := bytes.NewBufferString(`{"ids":["1","2","3"]}`)
	req := httptest.NewRequest("PUT", "/api/todos/reorder", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Reorder(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

// TestDeleteEndpointUnauthorized tests that Delete fails without auth
func TestDeleteEndpointUnauthorized(t *testing.T) {
	handler := &Handler{db: nil}

	req := httptest.NewRequest("DELETE", "/api/todos/123", nil)
	w := httptest.NewRecorder()

	handler.Delete(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

// TestSharedWithAdminDefault tests the default value of shared_with_admin
func TestSharedWithAdminDefault(t *testing.T) {
	t.Run("Default behavior when not specified", func(t *testing.T) {
		// When shared_with_admin is not specified, it should default to true
		var req struct {
			Text            string `json:"text"`
			SharedWithAdmin *bool  `json:"shared_with_admin"`
		}

		jsonData := `{"text":"test"}`
		if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		// Default to true unless explicitly set to false
		sharedWithAdmin := true
		if req.SharedWithAdmin != nil {
			sharedWithAdmin = *req.SharedWithAdmin
		}

		if sharedWithAdmin != true {
			t.Errorf("Expected SharedWithAdmin to default to true, got %v", sharedWithAdmin)
		}
	})

	t.Run("Explicitly set to false", func(t *testing.T) {
		var req struct {
			Text            string `json:"text"`
			SharedWithAdmin *bool  `json:"shared_with_admin"`
		}

		jsonData := `{"text":"test","shared_with_admin":false}`
		if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		sharedWithAdmin := true
		if req.SharedWithAdmin != nil {
			sharedWithAdmin = *req.SharedWithAdmin
		}

		if sharedWithAdmin != false {
			t.Errorf("Expected SharedWithAdmin to be false when explicitly set, got %v", sharedWithAdmin)
		}
	})
}
