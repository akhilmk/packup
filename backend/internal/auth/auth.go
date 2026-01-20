package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

// GoogleLogin redirects to Google OAuth2 login page.
// @Summary Login with Google
// @Description Redirects to Google OAuth2 login page.
// @Tags auth
// @Success 307
// @Router /api/auth/google/login [get]
func (h *Handler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	redirectURI := os.Getenv("GOOGLE_REDIRECT_URI")

	if clientID == "" || redirectURI == "" {
		http.Error(w, "OAuth not configured", http.StatusInternalServerError)
		return
	}

	// Get return_to path from query params, default to "/"
	returnTo := r.URL.Query().Get("return_to")
	if returnTo == "" {
		returnTo = "/"
	}

	// Use state to pass the return path through the OAuth flow
	state := base64.URLEncoding.EncodeToString([]byte(returnTo))

	scope := "https://www.googleapis.com/auth/userinfo.email https://www.googleapis.com/auth/userinfo.profile"
	authURL := fmt.Sprintf("https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s",
		url.QueryEscape(clientID), url.QueryEscape(redirectURI), url.QueryEscape(scope), url.QueryEscape(state))

	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
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
		Secure:   os.Getenv("SESSION_SECURE") == "true",
	})

	// Decode return path from state
	state := r.URL.Query().Get("state")
	targetPath := "/"
	if state != "" {
		if decoded, err := base64.URLEncoding.DecodeString(state); err == nil {
			targetPath = string(decoded)
		}
	}

	http.Redirect(w, r, targetPath, http.StatusSeeOther)
}

// Me returns the current authenticated user.
// @Summary Get current user
// @Description Get current authenticated user details from session cookie.
// @Tags auth
// @Produce json
// @Success 200 {object} models.User
// @Failure 401 {string} string "unauthorized"
// @Router /api/auth/me [get]
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

// Logout clears the session cookie.
// @Summary Logout
// @Description Clear session cookie and delete session from DB.
// @Tags auth
// @Success 200
// @Router /api/auth/logout [post]
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
		SameSite: http.SameSiteLaxMode,
		Secure:   os.Getenv("SESSION_SECURE") == "true",
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

// AdminMiddlewareWithRedirect allows only admins and redirects to login if unauthenticated.
// For logged-in non-admins, it returns 403 Forbidden.
func (h *Handler) AdminMiddlewareWithRedirect(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Prevent caching of protected pages (like Swagger UI)
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		cookie, err := r.Cookie("session_token")
		if err != nil {
			loginURL := "/api/auth/google/login?return_to=" + r.URL.Path
			http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
			return
		}

		user, err := h.getUserBySession(r.Context(), cookie.Value)
		if err != nil {
			loginURL := "/api/auth/google/login?return_to=" + r.URL.Path
			http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
			return
		}

		if user.Role != string(models.RoleAdmin) {
			http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
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

	tokenURL := fmt.Sprintf("https://oauth2.googleapis.com/token?client_id=%s&client_secret=%s&code=%s&grant_type=authorization_code&redirect_uri=%s",
		url.QueryEscape(clientID), url.QueryEscape(clientSecret), url.QueryEscape(code), url.QueryEscape(redirectURI))

	req, _ := http.NewRequest("POST", tokenURL, nil)
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
