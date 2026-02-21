package handlers

import (
	"net/http"
	"strconv"

	"ava-sales/internal/repositories"
	"ava-sales/internal/response"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type AgencyHandler struct {
	repo *repositories.AgencyRepository
}

func NewAgencyHandler(repo *repositories.AgencyRepository) *AgencyHandler {
	return &AgencyHandler{repo: repo}
}

func (h *AgencyHandler) List(c *gin.Context) {
	limit := 20
	offset := 0
	if q := c.Query("limit"); q != "" {
		if n, err := strconv.Atoi(q); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	if q := c.Query("offset"); q != "" {
		if n, err := strconv.Atoi(q); err == nil && n >= 0 {
			offset = n
		}
	}

	agencies, err := h.repo.List(c.Request.Context(), repositories.AgencyFilter{
		City: c.Query("city"), Province: c.Query("province"), Query: c.Query("q"), Limit: limit, Offset: offset,
	})
	if err != nil {
		response.JSONError(c, http.StatusInternalServerError, "internal_error", "failed to fetch agencies")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": agencies, "limit": limit, "offset": offset})
}

func (h *AgencyHandler) GetByID(c *gin.Context) {
	agency, err := h.repo.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		if err == pgx.ErrNoRows {
			response.JSONError(c, http.StatusNotFound, "not_found", "agency not found")
			return
		}
		response.JSONError(c, http.StatusInternalServerError, "internal_error", "failed to fetch agency")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": agency})
}
