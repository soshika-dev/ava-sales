package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"ava-sales/internal/models"
	"github.com/gin-gonic/gin"
)

func writeError(c *gin.Context, err error) {
	var apiErr *models.APIError
	if errors.As(err, &apiErr) {
		c.JSON(apiErr.HTTPStatus, gin.H{"error": apiErr})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": models.ErrInternal})
}

func parsePagination(c *gin.Context) (page int, pageSize int, err *models.APIError) {
	page = 1
	pageSize = 20

	if raw := c.Query("page"); raw != "" {
		v, e := strconv.Atoi(raw)
		if e != nil || v <= 0 {
			return 0, 0, models.NewAPIError("VALIDATION_ERROR", "page must be a positive integer", http.StatusBadRequest, nil)
		}
		page = v
	}
	if raw := c.Query("page_size"); raw != "" {
		v, e := strconv.Atoi(raw)
		if e != nil || v <= 0 {
			return 0, 0, models.NewAPIError("VALIDATION_ERROR", "page_size must be a positive integer", http.StatusBadRequest, nil)
		}
		if v > 100 {
			v = 100
		}
		pageSize = v
	}
	return page, pageSize, nil
}
