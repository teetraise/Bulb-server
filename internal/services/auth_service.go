package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/KoLili12/bulb-server/internal/models"
	"github.com/KoLili12/bulb-server/pkg/config"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

var (
	ErrExpiredToken = errors.New("token has expired")
	ErrInvalidToken = errors.New("token is invalid")
)

// TokenDetails содержит информацию о токенах
type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	AccessUuid   string
	RefreshUuid  string
	AtExpires    int64
	RtExpires    int64
}

// AccessTokenClaims определяет структуру полезной нагрузки JWT
type AccessTokenClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	UUID   string `json:"uuid"`
	jwt.RegisteredClaims
}

// RefreshTokenClaims определяет структуру полезной нагрузки для refresh token
type RefreshTokenClaims struct {
	UserID uint   `json:"user_id"`
	UUID   string `json:"uuid"`
	jwt.RegisteredClaims
}

// AuthService определяет методы для аутентификации и работы с токенами
type AuthService interface {
	CreateToken(user *models.User) (*TokenDetails, error)
	ValidateAccessToken(tokenString string) (*AccessTokenClaims, error)
	ValidateRefreshToken(tokenString string) (*RefreshTokenClaims, error)
}

// authService реализует интерфейс AuthService
type authService struct {
	config *config.Config
}

// NewAuthService создает новый экземпляр сервиса аутентификации
func NewAuthService(config *config.Config) AuthService {
	return &authService{
		config: config,
	}
}

// CreateToken создает токены доступа и обновления для пользователя
func (s *authService) CreateToken(user *models.User) (*TokenDetails, error) {
	td := &TokenDetails{}
	td.AtExpires = time.Now().Add(time.Hour * time.Duration(s.config.JWT.ExpiresIn)).Unix()
	td.AccessUuid = generateUUID()

	td.RtExpires = time.Now().Add(time.Hour * 24 * 7).Unix() // 7 дней
	td.RefreshUuid = generateUUID()

	// Создаем токен доступа
	atClaims := AccessTokenClaims{
		UserID: user.ID,
		Email:  user.Email,
		UUID:   td.AccessUuid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Unix(td.AtExpires, 0)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "bulb-api",
		},
	}
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	accessToken, err := at.SignedString([]byte(s.config.JWT.Secret))
	if err != nil {
		return nil, err
	}
	td.AccessToken = accessToken

	// Создаем токен обновления
	rtClaims := RefreshTokenClaims{
		UserID: user.ID,
		UUID:   td.RefreshUuid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Unix(td.RtExpires, 0)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "bulb-api",
		},
	}
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	refreshToken, err := rt.SignedString([]byte(s.config.JWT.Secret))
	if err != nil {
		return nil, err
	}
	td.RefreshToken = refreshToken

	return td, nil
}

// ValidateAccessToken проверяет и расшифровывает токен доступа
func (s *authService) ValidateAccessToken(tokenString string) (*AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*AccessTokenClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	if time.Unix(claims.ExpiresAt.Unix(), 0).Before(time.Now()) {
		return nil, ErrExpiredToken
	}

	return claims, nil
}

// ValidateRefreshToken проверяет и расшифровывает токен обновления
func (s *authService) ValidateRefreshToken(tokenString string) (*RefreshTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*RefreshTokenClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	if time.Unix(claims.ExpiresAt.Unix(), 0).Before(time.Now()) {
		return nil, ErrExpiredToken
	}

	return claims, nil
}

// generateUUID создает уникальный идентификатор для токена
func generateUUID() string {
	return uuid.New().String()
}
