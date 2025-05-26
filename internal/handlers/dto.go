package handlers

import "time"

// ErrorResponse представляет структуру ответа с ошибкой
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse представляет структуру успешного ответа
type SuccessResponse struct {
	Message string `json:"message"`
}

// RegisterRequest представляет структуру запроса на регистрацию
type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Surname  string `json:"surname" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Phone    string `json:"phone"`
}

// LoginRequest представляет структуру запроса на вход
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// TokenResponse представляет структуру ответа с токенами
type TokenResponse struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

// RefreshTokenRequest представляет структуру запроса на обновление токена
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// UserResponse представляет структуру ответа с данными пользователя
type UserResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Surname     string    `json:"surname"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone,omitempty"`
	ImageURL    string    `json:"imageUrl,omitempty"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

// UpdateProfileRequest представляет структуру запроса для обновления профиля
type UpdateProfileRequest struct {
	Name        string `json:"name" binding:"required"`
	Surname     string `json:"surname" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	Phone       string `json:"phone"`
	Description string `json:"description"`
}

// PublicUserResponse представляет публичную информацию о пользователе
type PublicUserResponse struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Surname string `json:"surname"`
}