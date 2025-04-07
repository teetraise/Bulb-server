package services

import (
	"errors"
	"time"

	"github.com/KoLili12/bulb-server/internal/models"
	"github.com/KoLili12/bulb-server/internal/repository"
)

var (
	ErrCollectionNotFound = errors.New("collection not found")
	ErrInvalidUserID      = errors.New("invalid user ID")
	ErrNotCollectionOwner = errors.New("user is not the owner of this collection")
)

// CollectionService определяет методы сервиса коллекций
type CollectionService interface {
	Create(collection *models.Collection) error
	GetByID(id uint) (*models.Collection, error)
	GetByUserID(userID uint) ([]*models.Collection, error)
	GetTrending(limit int) ([]*models.Collection, error)
	Update(collection *models.Collection, userID uint) error
	Delete(id uint, userID uint) error
	List(page, pageSize int) ([]*models.Collection, int64, error)
	IncrementPlayCount(id uint) error
	AddAction(collectionID uint, action *models.Action) error
	GetActions(collectionID uint) ([]*models.Action, error)
	RemoveAction(actionID uint, userID uint) error
}

// collectionService реализует интерфейс CollectionService
type collectionService struct {
	collectionRepo repository.CollectionRepository
	actionRepo     repository.ActionRepository
	userRepo       repository.UserRepository
}

// NewCollectionService создает новый экземпляр сервиса коллекций
func NewCollectionService(
	collectionRepo repository.CollectionRepository,
	actionRepo repository.ActionRepository,
	userRepo repository.UserRepository,
) CollectionService {
	return &collectionService{
		collectionRepo: collectionRepo,
		actionRepo:     actionRepo,
		userRepo:       userRepo,
	}
}

// Create создает новую коллекцию
func (s *collectionService) Create(collection *models.Collection) error {
	// Проверяем существование пользователя
	_, err := s.userRepo.GetByID(collection.UserID)
	if err != nil {
		return ErrInvalidUserID
	}

	// Устанавливаем время создания и обновления
	now := time.Now()
	collection.CreatedAt = now
	collection.UpdatedAt = now
	collection.PlayCount = 0

	// Сохраняем коллекцию
	return s.collectionRepo.Create(collection)
}

// GetByID возвращает коллекцию по ID
func (s *collectionService) GetByID(id uint) (*models.Collection, error) {
	collection, err := s.collectionRepo.GetByID(id)
	if err != nil {
		return nil, ErrCollectionNotFound
	}

	// Получаем действия для коллекции
	actions, err := s.actionRepo.GetByCollectionID(id)
	if err != nil {
		return nil, err
	}
	collection.Actions = actions

	return collection, nil
}

// GetByUserID возвращает коллекции пользователя
func (s *collectionService) GetByUserID(userID uint) ([]*models.Collection, error) {
	// Проверяем существование пользователя
	_, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	return s.collectionRepo.GetByUserID(userID)
}

// GetTrending возвращает список популярных коллекций
func (s *collectionService) GetTrending(limit int) ([]*models.Collection, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.collectionRepo.GetTrending(limit)
}

// Update обновляет данные коллекции
func (s *collectionService) Update(collection *models.Collection, userID uint) error {
	// Проверяем существование коллекции
	existingCollection, err := s.collectionRepo.GetByID(collection.ID)
	if err != nil {
		return ErrCollectionNotFound
	}

	// Проверяем, что пользователь является владельцем коллекции
	if existingCollection.UserID != userID {
		return ErrNotCollectionOwner
	}

	// Обновляем только разрешенные поля
	existingCollection.Name = collection.Name
	existingCollection.Description = collection.Description
	if collection.ImageURL != "" {
		existingCollection.ImageURL = collection.ImageURL
	}
	existingCollection.UpdatedAt = time.Now()

	// Сохраняем обновленную коллекцию
	return s.collectionRepo.Update(existingCollection)
}

// Delete удаляет коллекцию
func (s *collectionService) Delete(id uint, userID uint) error {
	// Проверяем существование коллекции
	collection, err := s.collectionRepo.GetByID(id)
	if err != nil {
		return ErrCollectionNotFound
	}

	// Проверяем, что пользователь является владельцем коллекции
	if collection.UserID != userID {
		return ErrNotCollectionOwner
	}

	// Удаляем коллекцию
	return s.collectionRepo.Delete(id)
}

// List возвращает список коллекций с пагинацией
func (s *collectionService) List(page, pageSize int) ([]*models.Collection, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	return s.collectionRepo.List(offset, pageSize)
}

// IncrementPlayCount увеличивает счетчик запусков коллекции
func (s *collectionService) IncrementPlayCount(id uint) error {
	// Проверяем существование коллекции
	_, err := s.collectionRepo.GetByID(id)
	if err != nil {
		return ErrCollectionNotFound
	}

	return s.collectionRepo.IncrementPlayCount(id)
}

// AddAction добавляет новое действие в коллекцию
func (s *collectionService) AddAction(collectionID uint, action *models.Action) error {
	// Проверяем существование коллекции
	_, err := s.collectionRepo.GetByID(collectionID)
	if err != nil {
		return ErrCollectionNotFound
	}

	// Устанавливаем ID коллекции для действия
	action.CollectionID = collectionID

	// Устанавливаем порядок действия (если не указан)
	if action.Order == 0 {
		// Получаем существующие действия
		actions, err := s.actionRepo.GetByCollectionID(collectionID)
		if err != nil {
			return err
		}
		action.Order = len(actions) + 1
	}

	// Устанавливаем время создания и обновления
	now := time.Now()
	action.CreatedAt = now
	action.UpdatedAt = now

	// Сохраняем действие
	return s.actionRepo.Create(action)
}

// GetActions возвращает действия коллекции
func (s *collectionService) GetActions(collectionID uint) ([]*models.Action, error) {
	// Проверяем существование коллекции
	_, err := s.collectionRepo.GetByID(collectionID)
	if err != nil {
		return nil, ErrCollectionNotFound
	}

	return s.actionRepo.GetByCollectionID(collectionID)
}

// RemoveAction удаляет действие
func (s *collectionService) RemoveAction(actionID uint, userID uint) error {
	// Получаем действие
	action, err := s.actionRepo.GetByID(actionID)
	if err != nil {
		return err
	}

	// Получаем коллекцию
	collection, err := s.collectionRepo.GetByID(action.CollectionID)
	if err != nil {
		return ErrCollectionNotFound
	}

	// Проверяем, что пользователь является владельцем коллекции
	if collection.UserID != userID {
		return ErrNotCollectionOwner
	}

	// Удаляем действие
	return s.actionRepo.Delete(actionID)
}
