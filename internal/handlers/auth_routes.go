package handlers

import (
	"ava-sales/internal/middleware"
	"ava-sales/internal/repository"
	"github.com/gin-gonic/gin"
)

func (h *Handler) registerAuthRoutes(api *gin.RouterGroup, tx repository.TxManager, jwtSecret string) {
	authGroup := api.Group("/auth")
	authGroup.POST("/login", h.Login)
	authGroup.POST("/refresh", h.Refresh)
	authGroup.GET("/me", middleware.JWTAuth(tx, jwtSecret), h.Me)
}
