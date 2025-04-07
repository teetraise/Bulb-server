package services

import (
	"errors"
	"time"

	"github.com/KoLili12/bulb-server/internal/models"
	"github.com/KoLili12/bulb-server/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// UserService определяет методы сервиса пользователей
type UserService interface {
	Create(user *models.User) error
	GetByID(id uint) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	Update(user *models.User) error
	Delete(id uint) error
	List(page, pageSize int) ([]*models.User, int64, error)
	Authenticate(email, password string) (*models.User, error)
}

// userService реализует интерфейс UserService
type userService struct {
	userRepo repository.UserRepository
}

// NewUserService создает новый экземпляр сервиса пользователей
func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

// Create создает нового пользователя
func (s *userService) Create(user *models.User) error {
	// Проверяем, не существует ли уже пользователь с таким email
	existingUser, err := s.userRepo.GetByEmail(user.Email)
	if err == nil && existingUser != nil {
		return ErrEmailAlreadyExists
	}

	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	// Устанавливаем время создания и обновления
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// Сохраняем пользователя
	return s.userRepo.Create(user)
}

// GetByID возвращает пользователя по ID
func (s *userService) GetByID(id uint) (*models.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// GetByEmail возвращает пользователя по email
func (s *userService) GetByEmail(email string) (*models.User, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// Update обновляет данные пользователя
func (s *userService) Update(user *models.User) error {
	// Проверяем существование пользователя
	existingUser, err := s.userRepo.GetByID(user.ID)
	if err != nil {
		return ErrUserNotFound
	}

	// Если email был изменен, проверяем его уникальность
	if user.Email != existingUser.Email {
		userWithSameEmail, err := s.userRepo.GetByEmail(user.Email)
		if err == nil && userWithSameEmail != nil {
			return ErrEmailAlreadyExists
		}
	}

	// Если пароль был изменен, хешируем его
	if user.Password != "" && user.Password != existingUser.Password {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		user.Password = string(hashedPassword)
	} else {
		// Иначе используем существующий хешированный пароль
		user.Password = existingUser.Password
	}

	// Обновляем время изменения
	user.UpdatedAt = time.Now()

	// Сохраняем обновленного пользователя
	return s.userRepo.Update(user)
}

// Delete удаляет пользователя
func (s *userService) Delete(id uint) error {
	// Проверяем существование пользователя
	_, err := s.userRepo.GetByID(id)
	if err != nil {
		return ErrUserNotFound
	}

	// Удаляем пользователя
	return s.userRepo.Delete(id)
}

// List возвращает список пользователей с пагинацией
func (s *userService) List(page, pageSize int) ([]*models.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	return s.userRepo.List(offset, pageSize)
}

// Authenticate проверяет учетные данные пользователя и возвращает пользователя, если они верны
func (s *userService) Authenticate(email, password string) (*models.User, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Проверяем пароль
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}
