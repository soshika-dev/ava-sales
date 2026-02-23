package models

import (
	"time"

	"github.com/google/uuid"
)

type AppUser struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	Username     *string   `json:"username,omitempty"`
	PasswordHash string    `json:"-"`
	Role         ActorRole `json:"role"`
	CustomerID   *uuid.UUID
	TechnicianID *uuid.UUID
	IsActive     bool      `json:"is_active"`
}

type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
