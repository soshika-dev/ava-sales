package models

import "github.com/google/uuid"

type Actor struct {
	Role         ActorRole
	ActorID      uuid.UUID
	CustomerID   *uuid.UUID
	TechnicianID *uuid.UUID
}
