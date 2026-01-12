package models

import "time"

// UserRole represents the role of a user in the system.
type UserRole string

// Valid user roles.
const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
)

// User represents a user in the system.
type User struct {
	ID        string    `json:"id"`
	GoogleID  string    `json:"google_id,omitempty"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	AvatarURL string    `json:"avatar_url"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}
