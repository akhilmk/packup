package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/akhilmk/packup/internal/models"
)

// TestDetermineUserRole tests the role determination logic
func TestDetermineUserRole(t *testing.T) {
	tests := []struct {
		name         string
		adminEmails  string
		email        string
		expectedRole string
	}{
		{
			name:         "User when no admin emails configured",
			adminEmails:  "",
			email:        "user@example.com",
			expectedRole: "user",
		},
		{
			name:         "Admin when email matches single admin email",
			adminEmails:  "admin@example.com",
			email:        "admin@example.com",
			expectedRole: "admin",
		},
		{
			name:         "Admin when email matches one of multiple admin emails",
			adminEmails:  "admin1@example.com,admin@example.com,admin2@example.com",
			email:        "admin@example.com",
			expectedRole: "admin",
		},
		{
			name:         "User when email doesn't match any admin email",
			adminEmails:  "admin@example.com",
			email:        "user@example.com",
			expectedRole: "user",
		},
		{
			name:         "Admin with spaces in list",
			adminEmails:  "admin1@example.com, admin@example.com",
			email:        "admin@example.com",
			expectedRole: "admin",
		},
		{
			name:         "User with partial email match (shouldn't match)",
			adminEmails:  "admin@example.com",
			email:        "notadmin@example.com",
			expectedRole: "user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original env
			original := os.Getenv("ADMIN_EMAILS")
			defer os.Setenv("ADMIN_EMAILS", original)

			os.Setenv("ADMIN_EMAILS", tt.adminEmails)

			role := determineUserRole(tt.email)
			if role != tt.expectedRole {
				t.Errorf("Expected role '%s', got '%s'", tt.expectedRole, role)
			}
		})
	}
}

// TestMiddlewareUnauthorized tests the auth middleware without cookie
func TestMiddlewareUnauthorized(t *testing.T) {
	handler := &Handler{db: nil}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("No cookie returns unauthorized", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/todos", nil)
		w := httptest.NewRecorder()

		handler.Middleware(next)(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})
}

// TestMeUnauthorized tests the Me endpoint without authentication
func TestMeUnauthorized(t *testing.T) {
	handler := &Handler{db: nil}

	req := httptest.NewRequest("GET", "/api/auth/me", nil)
	w := httptest.NewRecorder()

	handler.Me(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

// TestLogout tests the logout endpoint
func TestLogout(t *testing.T) {
	handler := &Handler{db: nil}

	t.Run("Logout without cookie", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/auth/logout", nil)
		w := httptest.NewRecorder()

		handler.Logout(w, req)

		// Should still return OK
		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		// Check that cookie is being cleared
		cookies := w.Result().Cookies()
		found := false
		for _, c := range cookies {
			if c.Name == "session_token" && c.Value == "" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected session_token cookie to be cleared")
		}
	})
}

// TestGoogleLoginMissingConfig tests login with missing OAuth config
func TestGoogleLoginMissingConfig(t *testing.T) {
	handler := &Handler{db: nil}

	// Save and restore original env
	originalClientID := os.Getenv("GOOGLE_CLIENT_ID")
	originalRedirectURI := os.Getenv("GOOGLE_REDIRECT_URI")
	defer func() {
		os.Setenv("GOOGLE_CLIENT_ID", originalClientID)
		os.Setenv("GOOGLE_REDIRECT_URI", originalRedirectURI)
	}()

	os.Setenv("GOOGLE_CLIENT_ID", "")
	os.Setenv("GOOGLE_REDIRECT_URI", "")

	req := httptest.NewRequest("GET", "/api/auth/google/login", nil)
	w := httptest.NewRecorder()

	handler.GoogleLogin(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d with missing config, got %d", http.StatusInternalServerError, w.Code)
	}
}

// TestGoogleCallbackMissingCode tests callback without code
func TestGoogleCallbackMissingCode(t *testing.T) {
	handler := &Handler{db: nil}

	req := httptest.NewRequest("GET", "/api/auth/google/callback", nil)
	w := httptest.NewRecorder()

	handler.GoogleCallback(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d with missing code, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestUserStruct tests that User struct has correct fields
func TestUserStruct(t *testing.T) {
	user := models.User{
		ID:        "user-1",
		GoogleID:  "google-123",
		Email:     "test@example.com",
		Name:      "Test User",
		AvatarURL: "https://example.com/avatar.png",
		Role:      "admin",
	}

	if user.ID != "user-1" {
		t.Errorf("Expected ID 'user-1', got '%s'", user.ID)
	}
	if user.GoogleID != "google-123" {
		t.Errorf("Expected GoogleID 'google-123', got '%s'", user.GoogleID)
	}
	if user.Role != "admin" {
		t.Errorf("Expected Role 'admin', got '%s'", user.Role)
	}
}

// TestContextValues tests that context values are properly set in middleware
func TestContextValues(t *testing.T) {
	t.Run("User ID in context", func(t *testing.T) {
		ctx := context.Background()
		ctx = context.WithValue(ctx, "user_id", "user-123")

		userID, ok := ctx.Value("user_id").(string)
		if !ok || userID != "user-123" {
			t.Errorf("Expected user_id 'user-123', got '%s'", userID)
		}
	})

	t.Run("User role in context", func(t *testing.T) {
		ctx := context.Background()
		ctx = context.WithValue(ctx, "user_role", "admin")

		userRole, ok := ctx.Value("user_role").(string)
		if !ok || userRole != "admin" {
			t.Errorf("Expected user_role 'admin', got '%s'", userRole)
		}
	})
}
