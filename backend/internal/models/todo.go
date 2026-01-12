// Package models contains shared data types used across the application.
package models

import "time"

// TodoStatus represents the status of a todo item.
type TodoStatus string

// Valid todo statuses.
const (
	StatusPending    TodoStatus = "pending"
	StatusInProgress TodoStatus = "in-progress"
	StatusDone       TodoStatus = "done"
)

// IsValid checks if the status is a valid todo status.
func (s TodoStatus) IsValid() bool {
	switch s {
	case StatusPending, StatusInProgress, StatusDone:
		return true
	}
	return false
}

// String returns the string representation of the status.
func (s TodoStatus) String() string {
	return string(s)
}

// Todo constants.
const (
	// MaxTextLength is the maximum length of todo text.
	MaxTextLength = 200

	// PositionIncrement is the increment used for positioning todos.
	PositionIncrement = 1024.0
)

// Todo represents a todo item.
type Todo struct {
	ID              string    `json:"id"`
	Text            string    `json:"text"`
	Status          string    `json:"status"`
	Created         time.Time `json:"created"`
	Position        float64   `json:"position"`
	CreatedByUserID *string   `json:"created_by_user_id,omitempty"`
	IsDefaultTask   bool      `json:"is_default_task"`
	SharedWithAdmin bool      `json:"shared_with_admin"`
	HiddenFromUser  bool      `json:"hidden_from_user"`
	UserID          *string   `json:"user_id,omitempty"`
}

// ValidateText validates the todo text length.
func ValidateText(text string) bool {
	return len(text) > 0 && len(text) <= MaxTextLength
}
