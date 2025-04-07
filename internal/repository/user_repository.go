package repository

import (
	"github.com/KoLili12/bulb-server/internal/models"
	"gorm.io/gorm"
)

// CollectionRepository определяет методы для работы с коллекциями в базе данных
type CollectionRepository interface {
	Create(collection *models.Collection) error
	GetByID(id uint) (*models.Collection, error)
	GetByUserID(userID uint) ([]*models.Collection, error)
	GetTrending(limit int) ([]*models.Collection, error)
	Update(collection *models.Collection) error
	Delete(id uint) error
	List(offset, limit int) ([]*models.Collection, int64, error)
	IncrementPlayCount(id uint) error
}

// collectionRepository реализует интерфейс CollectionRepository
type collectionRepository struct {
	db *gorm.DB
}

// NewCollectionRepository создает новый экземпляр репозитория коллекций
func NewCollectionRepository(db *gorm.DB) CollectionRepository {
	return &collectionRepository{
		db: db,
	}
}

// Create сохраняет новую коллекцию в базе данных
func (r *collectionRepository) Create(collection *models.Collection) error {
	return r.db.Create(collection).Error
}

// GetByID возвращает коллекцию по ID
func (r *collectionRepository) GetByID(id uint) (*models.Collection, error) {
	var collection models.Collection
	if err := r.db.Preload("Actions").First(&collection, id).Error; err != nil {
		return nil, err
	}
	return &collection, nil
}

// GetByUserID возвращает коллекции пользователя
func (r *collectionRepository) GetByUserID(userID uint) ([]*models.Collection, error) {
	var collections []*models.Collection
	if err := r.db.Where("user_id = ?", userID).Find(&collections).Error; err != nil {
		return nil, err
	}
	return collections, nil
}

// GetTrending возвращает список популярных коллекций
func (r *collectionRepository) GetTrending(limit int) ([]*models.Collection, error) {
	var collections []*models.Collection
	if err := r.db.Order("play_count DESC").Limit(limit).Find(&collections).Error; err != nil {
		return nil, err
	}
	return collections, nil
}

// Update обновляет данные коллекции
func (r *collectionRepository) Update(collection *models.Collection) error {
	return r.db.Save(collection).Error
}

// Delete удаляет коллекцию (soft delete через GORM)
func (r *collectionRepository) Delete(id uint) error {
	return r.db.Delete(&models.Collection{}, id).Error
}

// List возвращает список коллекций с пагинацией
func (r *collectionRepository) List(offset, limit int) ([]*models.Collection, int64, error) {
	var collections []*models.Collection
	var count int64

	// Получаем общее количество коллекций
	if err := r.db.Model(&models.Collection{}).Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// Получаем список коллекций с пагинацией
	if err := r.db.Offset(offset).Limit(limit).Find(&collections).Error; err != nil {
		return nil, 0, err
	}

	return collections, count, nil
}

// IncrementPlayCount увеличивает счетчик запусков коллекции
func (r *collectionRepository) IncrementPlayCount(id uint) error {
	return r.db.Model(&models.Collection{}).Where("id = ?", id).
		UpdateColumn("play_count", gorm.Expr("play_count + ?", 1)).Error
}

// UserRepository определяет методы для работы с пользователями в базе данных
type UserRepository interface {
	Create(user *models.User) error
	GetByID(id uint) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	Update(user *models.User) error
	Delete(id uint) error
	List(offset, limit int) ([]*models.User, int64, error)
}

// userRepository реализует интерфейс UserRepository
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository создает новый экземпляр репозитория пользователей
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

// Create сохраняет нового пользователя в базе данных
func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// GetByID возвращает пользователя по ID
func (r *userRepository) GetByID(id uint) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail возвращает пользователя по email
func (r *userRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// Update обновляет данные пользователя
func (r *userRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

// Delete удаляет пользователя (soft delete через GORM)
func (r *userRepository) Delete(id uint) error {
	return r.db.Delete(&models.User{}, id).Error
}

// List возвращает список пользователей с пагинацией
func (r *userRepository) List(offset, limit int) ([]*models.User, int64, error) {
	var users []*models.User
	var count int64

	// Получаем общее количество пользователей
	if err := r.db.Model(&models.User{}).Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// Получаем список пользователей с пагинацией
	if err := r.db.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, count, nil
}
