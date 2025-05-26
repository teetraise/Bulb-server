package handlers

import (
	"net/http"
	"strconv"

	"github.com/KoLili12/bulb-server/internal/middleware"
	"github.com/KoLili12/bulb-server/internal/services"
	"github.com/gin-gonic/gin"
)

// UserHandler обрабатывает запросы, связанные с пользователями
type UserHandler struct {
	userService services.UserService
}

// NewUserHandler создает новый обработчик пользователей
func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetUserByID возвращает информацию о пользователе по ID (публичный endpoint)
func (h *UserHandler) GetUserByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user ID"})
		return
	}

	user, err := h.userService.GetByID(uint(id))
	if err != nil {
		if err == services.ErrUserNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get user"})
		return
	}

	// Возвращаем только публичную информацию о пользователе
	response := PublicUserResponse{
		ID:      user.ID,
		Name:    user.Name,
		Surname: user.Surname,
	}

	c.JSON(http.StatusOK, response)
}

// GetProfile возвращает профиль текущего пользователя
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	user, err := h.userService.GetByID(userID)
	if err != nil {
		if err == services.ErrUserNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get user"})
		return
	}

	response := UserResponse{
		ID:          user.ID,
		Name:        user.Name,
		Surname:     user.Surname,
		Email:       user.Email,
		Phone:       user.Phone,
		ImageURL:    user.ImageURL,
		Description: user.Description,
		CreatedAt:   user.CreatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateProfile обновляет профиль текущего пользователя
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Получаем текущего пользователя
	user, err := h.userService.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
		return
	}

	// Обновляем поля
	user.Name = req.Name
	user.Surname = req.Surname
	user.Email = req.Email
	user.Phone = req.Phone
	user.Description = req.Description

	// Сохраняем изменения
	if err := h.userService.Update(user); err != nil {
		if err == services.ErrEmailAlreadyExists {
			c.JSON(http.StatusConflict, ErrorResponse{Error: "Email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Message: "Profile updated successfully"})
}