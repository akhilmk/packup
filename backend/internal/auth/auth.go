package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/akhilmk/packup/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	db *pgxpool.Pool
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{db: db}
}

// RegisterRoutes registers auth routes interactively
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/auth/google/login", h.GoogleLogin)
	mux.HandleFunc("GET /api/auth/google/callback", h.GoogleCallback)
	mux.HandleFunc("GET /api/auth/me", h.Me)
	mux.HandleFunc("POST /api/auth/logout", h.Logout)
}

func (h *Handler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	redirectURI := os.Getenv("GOOGLE_REDIRECT_URI")

	if clientID == "" || redirectURI == "" {
		http.Error(w, "OAuth not configured", http.StatusInternalServerError)
		return
	}

	// In real world, use 'state' parameter to prevent CSRF
	state := "random-state"

	scope := "https://www.googleapis.com/auth/userinfo.email https://www.googleapis.com/auth/userinfo.profile"
	url := fmt.Sprintf("https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s",
		clientID, redirectURI, scope, state)

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *Handler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "code not found", http.StatusBadRequest)
		return
	}

	token, err := h.exchangeCode(code)
	if err != nil {
		http.Error(w, "failed to exchange code: "+err.Error(), http.StatusInternalServerError)
		return
	}

	googleUser, err := h.getGoogleUser(token)
	if err != nil {
		http.Error(w, "failed to get user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	user, err := h.getOrCreateUser(r.Context(), googleUser)
	if err != nil {
		http.Error(w, "failed to save user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sessionToken, err := h.createSession(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "failed to create session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		// Secure:   true, // Uncomment in production with HTTPS
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.getUserBySession(r.Context(), cookie.Value)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err == nil {
		h.db.Exec(r.Context(), "DELETE FROM sessions WHERE token=$1", cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
	})

	w.WriteHeader(http.StatusOK)
}

// Middleware
func (h *Handler) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		user, err := h.getUserBySession(r.Context(), cookie.Value)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := SetUserContext(r.Context(), user.ID, user.Role)
		next(w, r.WithContext(ctx))
	}
}

// Helper function to determine user role based on admin emails
func determineUserRole(email string) string {
	adminEmails := os.Getenv("ADMIN_EMAILS")
	if adminEmails == "" {
		return string(models.RoleUser)
	}

	// Parse comma-separated admin emails
	emails := []string{}
	for _, e := range []byte(adminEmails) {
		if e == ',' {
			emails = append(emails, "")
		} else if len(emails) == 0 {
			emails = append(emails, string(e))
		} else {
			emails[len(emails)-1] += string(e)
		}
	}

	// Trim spaces and check if email matches
	for _, adminEmail := range emails {
		trimmed := ""
		for _, c := range adminEmail {
			if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
				trimmed += string(c)
			}
		}
		if trimmed == email {
			return string(models.RoleAdmin)
		}
	}

	return string(models.RoleUser)
}

// Helpers

func (h *Handler) exchangeCode(code string) (string, error) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURI := os.Getenv("GOOGLE_REDIRECT_URI")

	if clientID == "" || clientSecret == "" {
		return "", fmt.Errorf("client credentials missing")
	}

	url := fmt.Sprintf("https://oauth2.googleapis.com/token?client_id=%s&client_secret=%s&code=%s&grant_type=authorization_code&redirect_uri=%s",
		clientID, clientSecret, code, redirectURI)

	req, _ := http.NewRequest("POST", url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token exchange failed: %s", string(body))
	}

	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.AccessToken, nil
}

type googleUser struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func (h *Handler) getGoogleUser(token string) (googleUser, error) {
	req, _ := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return googleUser{}, err
	}
	defer resp.Body.Close()

	var gu googleUser
	if err := json.NewDecoder(resp.Body).Decode(&gu); err != nil {
		return googleUser{}, err
	}
	return gu, nil
}

func (h *Handler) getOrCreateUser(ctx context.Context, gu googleUser) (models.User, error) {
	var user models.User
	// Check if exists
	err := h.db.QueryRow(ctx, "SELECT id, google_id, email, name, avatar_url, role, created_at FROM users WHERE google_id=$1", gu.ID).
		Scan(&user.ID, &user.GoogleID, &user.Email, &user.Name, &user.AvatarURL, &user.Role, &user.CreatedAt)

	if err == pgx.ErrNoRows {
		// Determine role based on admin emails
		role := determineUserRole(gu.Email)

		// Create
		user = models.User{
			ID:        uuid.NewString(),
			GoogleID:  gu.ID,
			Email:     gu.Email,
			Name:      gu.Name,
			AvatarURL: gu.Picture,
			Role:      role,
			CreatedAt: time.Now(),
		}
		_, err = h.db.Exec(ctx, "INSERT INTO users(id, google_id, email, name, avatar_url, role, created_at) VALUES($1,$2,$3,$4,$5,$6,$7)",
			user.ID, user.GoogleID, user.Email, user.Name, user.AvatarURL, user.Role, user.CreatedAt)
	} else if err == nil {
		// Update role if it changed (in case admin emails were updated)
		newRole := determineUserRole(gu.Email)
		if newRole != user.Role {
			user.Role = newRole
			_, _ = h.db.Exec(ctx, "UPDATE users SET role=$1 WHERE id=$2", user.Role, user.ID)
		}
	}

	return user, err
}

func (h *Handler) createSession(ctx context.Context, userID string) (string, error) {
	b := make([]byte, 32)
	rand.Read(b)
	token := base64.URLEncoding.EncodeToString(b)
	expiresAt := time.Now().Add(24 * time.Hour)

	_, err := h.db.Exec(ctx, "INSERT INTO sessions(token, user_id, expires_at) VALUES($1,$2,$3)", token, userID, expiresAt)
	return token, err
}

func (h *Handler) getUserBySession(ctx context.Context, token string) (models.User, error) {
	var user models.User
	err := h.db.QueryRow(ctx, `
		SELECT u.id, u.google_id, u.email, u.name, u.avatar_url, u.role, u.created_at 
		FROM sessions s
		JOIN users u ON s.user_id = u.id
		WHERE s.token = $1 AND s.expires_at > now()
	`, token).Scan(&user.ID, &user.GoogleID, &user.Email, &user.Name, &user.AvatarURL, &user.Role, &user.CreatedAt)

	return user, err
}
