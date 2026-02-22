package models

import (
	"time"

	"github.com/google/uuid"
)

type Agency struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Province  string    `json:"province"`
	City      string    `json:"city"`
	Address   string    `json:"address"`
	Phone     string    `json:"phone"`
	Latitude  *float64  `json:"latitude,omitempty"`
	Longitude *float64  `json:"longitude,omitempty"`
	Version   float32   `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
