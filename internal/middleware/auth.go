package middleware

import (
	"net/http"
	"strings"

	"ava-sales/internal/models"
	"ava-sales/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const AuthActorKey = "auth_actor"

type JWTClaims struct {
	jwt.RegisteredClaims
}

func JWTAuth(tx repository.TxManager, secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": models.NewAPIError("UNAUTHORIZED", "Authorization header is required", http.StatusUnauthorized, nil)})
			c.Abort()
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": models.NewAPIError("UNAUTHORIZED", "Authorization must be Bearer token", http.StatusUnauthorized, nil)})
			c.Abort()
			return
		}

		claims := &JWTClaims{}
		token, err := jwt.ParseWithClaims(parts[1], claims, func(token *jwt.Token) (any, error) {
			return []byte(secret), nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": models.NewAPIError("UNAUTHORIZED", "invalid token", http.StatusUnauthorized, nil)})
			c.Abort()
			return
		}
		if strings.TrimSpace(claims.Subject) == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": models.NewAPIError("UNAUTHORIZED", "token subject (sub) is required", http.StatusUnauthorized, nil)})
			c.Abort()
			return
		}

		userID, parseErr := uuid.Parse(claims.Subject)
		if parseErr != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": models.NewAPIError("UNAUTHORIZED", "token sub must be UUID", http.StatusUnauthorized, nil)})
			c.Abort()
			return
		}

		user, repoErr := tx.Repo().GetAppUserByID(c.Request.Context(), userID)
		if repoErr != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": models.NewAPIError("UNAUTHORIZED", "user not found", http.StatusUnauthorized, nil)})
			c.Abort()
			return
		}
		if !user.IsActive {
			c.JSON(http.StatusUnauthorized, gin.H{"error": models.NewAPIError("UNAUTHORIZED", "user is inactive", http.StatusUnauthorized, nil)})
			c.Abort()
			return
		}

		actor := models.Actor{
			Role:         user.Role,
			ActorID:      user.ID,
			CustomerID:   user.CustomerID,
			TechnicianID: user.TechnicianID,
		}
		c.Set(AuthActorKey, actor)
		c.Next()
	}
}

func GetActor(c *gin.Context) (models.Actor, *models.APIError) {
	v, ok := c.Get(AuthActorKey)
	if !ok {
		return models.Actor{}, models.NewAPIError("UNAUTHORIZED", "authentication context missing", http.StatusUnauthorized, nil)
	}
	actor, ok := v.(models.Actor)
	if !ok {
		return models.Actor{}, models.NewAPIError("UNAUTHORIZED", "invalid authentication context", http.StatusUnauthorized, nil)
	}
	return actor, nil
}

func RequireCustomer() gin.HandlerFunc {
	return func(c *gin.Context) {
		actor, err := GetActor(c)
		if err != nil {
			c.JSON(err.HTTPStatus, gin.H{"error": err})
			c.Abort()
			return
		}
		if actor.Role != models.RoleCustomer || actor.CustomerID == nil {
			c.JSON(http.StatusForbidden, gin.H{"error": models.ErrForbidden})
			c.Abort()
			return
		}
		c.Next()
	}
}

func RequireTechOrAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		actor, err := GetActor(c)
		if err != nil {
			c.JSON(err.HTTPStatus, gin.H{"error": err})
			c.Abort()
			return
		}
		if actor.Role != models.RoleTechnician && actor.Role != models.RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": models.ErrForbidden})
			c.Abort()
			return
		}
		if actor.Role == models.RoleTechnician && actor.TechnicianID == nil {
			c.JSON(http.StatusForbidden, gin.H{"error": models.ErrForbidden})
			c.Abort()
			return
		}
		c.Next()
	}
}
