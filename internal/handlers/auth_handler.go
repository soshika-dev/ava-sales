package handlers

import (
	"net/http"
	"strings"

	"ava-sales/internal/middleware"
	"ava-sales/internal/models"
	"github.com/gin-gonic/gin"
)

type loginRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password" binding:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, models.NewAPIError("VALIDATION_ERROR", "invalid request body", http.StatusBadRequest, err.Error()))
		return
	}

	identity := strings.TrimSpace(req.Email)
	if identity == "" {
		identity = strings.TrimSpace(req.Username)
	}
	if identity == "" {
		writeError(c, models.NewAPIError("VALIDATION_ERROR", "email or username is required", http.StatusBadRequest, nil))
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), identity, req.Password)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, models.NewAPIError("VALIDATION_ERROR", "invalid request body", http.StatusBadRequest, err.Error()))
		return
	}

	resp, err := h.authService.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) Me(c *gin.Context) {
	actor, err := middleware.GetActor(c)
	if err != nil {
		writeError(c, err)
		return
	}

	resp, svcErr := h.authService.Me(c.Request.Context(), actor.ActorID)
	if svcErr != nil {
		writeError(c, svcErr)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}
