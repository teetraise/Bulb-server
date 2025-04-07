package handlers

import (
	"net/http"
	"time"

	"github.com/KoLili12/bulb-server/internal/models"
	"github.com/KoLili12/bulb-server/internal/services"
	"github.com/gin-gonic/gin"
)

// AuthHandler обрабатывает запросы, связанные с аутентификацией
type AuthHandler struct {
	userService services.UserService
	authService services.AuthService
}

// NewAuthHandler создает новый обработчик аутентификации
func NewAuthHandler(userService services.UserService, authService services.AuthService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		authService: authService,
	}
}

// Register обрабатывает запрос на регистрацию нового пользователя
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Создаем нового пользователя
	user := &models.User{
		Name:      req.Name,
		Surname:   req.Surname,
		Email:     req.Email,
		Password:  req.Password,
		Phone:     req.Phone,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Сохраняем пользователя
	if err := h.userService.Create(user); err != nil {
		if err == services.ErrEmailAlreadyExists {
			c.JSON(http.StatusConflict, ErrorResponse{Error: "Email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create user"})
		return
	}

	// Генерируем токены для пользователя
	td, err := h.authService.CreateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to generate token"})
		return
	}

	// Отправляем ответ с токенами
	c.JSON(http.StatusCreated, TokenResponse{
		AccessToken:  td.AccessToken,
		RefreshToken: td.RefreshToken,
		ExpiresAt:    time.Unix(td.AtExpires, 0),
	})
}

// Login обрабатывает запрос на вход пользователя
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Аутентифицируем пользователя
	user, err := h.userService.Authenticate(req.Email, req.Password)
	if err != nil {
		if err == services.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Authentication failed"})
		return
	}

	// Генерируем токены для пользователя
	td, err := h.authService.CreateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to generate token"})
		return
	}

	// Отправляем ответ с токенами
	c.JSON(http.StatusOK, TokenResponse{
		AccessToken:  td.AccessToken,
		RefreshToken: td.RefreshToken,
		ExpiresAt:    time.Unix(td.AtExpires, 0),
	})
}

// RefreshToken обрабатывает запрос на обновление токена
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Проверяем refresh token
	claims, err := h.authService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid refresh token"})
		return
	}

	// Получаем пользователя по ID из токена
	user, err := h.userService.GetByID(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "User not found"})
		return
	}

	// Генерируем новые токены
	td, err := h.authService.CreateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to generate token"})
		return
	}

	// Отправляем ответ с новыми токенами
	c.JSON(http.StatusOK, TokenResponse{
		AccessToken:  td.AccessToken,
		RefreshToken: td.RefreshToken,
		ExpiresAt:    time.Unix(td.AtExpires, 0),
	})
}
