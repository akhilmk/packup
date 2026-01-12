package auth

import (
	"context"

	"github.com/akhilmk/itinera/internal/models"
)

// contextKey is a type for context keys to avoid collisions.
type contextKey string

// Context keys for user information.
const (
	userIDKey   contextKey = "user_id"
	userRoleKey contextKey = "user_role"
)

// SetUserContext adds user ID and role to the context.
func SetUserContext(ctx context.Context, userID, userRole string) context.Context {
	ctx = context.WithValue(ctx, userIDKey, userID)
	ctx = context.WithValue(ctx, userRoleKey, userRole)
	return ctx
}

// GetUserID retrieves the user ID from the context.
// Returns empty string and false if not found.
func GetUserID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDKey).(string)
	return id, ok
}

// GetUserRole retrieves the user role from the context.
// Returns empty string and false if not found.
func GetUserRole(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(userRoleKey).(string)
	return role, ok
}

// IsAdmin checks if the user in the context has admin role.
func IsAdmin(ctx context.Context) bool {
	role, ok := GetUserRole(ctx)
	return ok && role == string(models.RoleAdmin)
}
