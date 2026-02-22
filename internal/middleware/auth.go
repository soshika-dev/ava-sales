package middleware

import (
	"net/http"

	"ava-sales/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func ParseCustomerID(c *gin.Context) (uuid.UUID, *models.APIError) {
	raw := c.GetHeader("X-Customer-Id")
	if raw == "" {
		return uuid.Nil, models.NewAPIError("UNAUTHORIZED", "X-Customer-Id header is required", http.StatusUnauthorized, nil)
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, models.NewAPIError("VALIDATION_ERROR", "X-Customer-Id must be a valid UUID", http.StatusBadRequest, nil)
	}
	return id, nil
}

func ParseTechnicianID(c *gin.Context) (uuid.UUID, *models.APIError) {
	raw := c.GetHeader("X-Technician-Id")
	if raw == "" {
		return uuid.Nil, models.NewAPIError("UNAUTHORIZED", "X-Technician-Id header is required", http.StatusUnauthorized, nil)
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, models.NewAPIError("VALIDATION_ERROR", "X-Technician-Id must be a valid UUID", http.StatusBadRequest, nil)
	}
	return id, nil
}

func ParseActor(c *gin.Context, fallbackRole models.ActorRole, fallbackID uuid.UUID) (models.Actor, *models.APIError) {
	role := fallbackRole
	actorID := fallbackID

	if rawRole := c.GetHeader("X-Actor-Role"); rawRole != "" {
		role = models.ActorRole(rawRole)
		if role != models.RoleCustomer && role != models.RoleTechnician && role != models.RoleAdmin {
			return models.Actor{}, models.NewAPIError("VALIDATION_ERROR", "X-Actor-Role must be CUSTOMER, TECHNICIAN, or ADMIN", http.StatusBadRequest, nil)
		}
	}
	if rawActorID := c.GetHeader("X-Actor-Id"); rawActorID != "" {
		id, err := uuid.Parse(rawActorID)
		if err != nil {
			return models.Actor{}, models.NewAPIError("VALIDATION_ERROR", "X-Actor-Id must be a valid UUID", http.StatusBadRequest, nil)
		}
		actorID = id
	}

	var customerID *uuid.UUID
	if raw := c.GetHeader("X-Customer-Id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return models.Actor{}, models.NewAPIError("VALIDATION_ERROR", "X-Customer-Id must be a valid UUID", http.StatusBadRequest, nil)
		}
		customerID = &id
	}

	var techID *uuid.UUID
	if raw := c.GetHeader("X-Technician-Id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return models.Actor{}, models.NewAPIError("VALIDATION_ERROR", "X-Technician-Id must be a valid UUID", http.StatusBadRequest, nil)
		}
		techID = &id
	}

	return models.Actor{
		Role:         role,
		ActorID:      actorID,
		CustomerID:   customerID,
		TechnicianID: techID,
	}, nil
}
