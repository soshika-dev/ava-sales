package models

import "github.com/google/uuid"

type AppUser struct {
	ID           uuid.UUID
	Role         ActorRole
	CustomerID   *uuid.UUID
	TechnicianID *uuid.UUID
	IsActive     bool
}
