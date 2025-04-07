package middleware

import (
	"net/http"
	"strings"

	"github.com/KoLili12/bulb-server/internal/services"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware представляет middleware для проверки аутентификации
type AuthMiddleware struct {
	authService services.AuthService
}

// NewAuthMiddleware создает новый экземпляр AuthMiddleware
func NewAuthMiddleware(authService services.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

// RequireAuth проверяет JWT-токен и устанавливает ID пользователя в контекст
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := m.authService.ValidateAccessToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Устанавливаем ID пользователя в контекст
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)

		c.Next()
	}
}

// GetUserID возвращает ID пользователя из контекста
func GetUserID(c *gin.Context) uint {
	userID, exists := c.Get("userID")
	if !exists {
		return 0
	}
	return userID.(uint)
}
