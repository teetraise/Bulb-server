package repository

import (
	"github.com/KoLili12/bulb-server/internal/models"
	"gorm.io/gorm"
)

// ActionRepository определяет методы для работы с действиями в базе данных
type ActionRepository interface {
	Create(action *models.Action) error
	GetByID(id uint) (*models.Action, error)
	GetByCollectionID(collectionID uint) ([]*models.Action, error)
	Update(action *models.Action) error
	Delete(id uint) error
	BatchCreate(actions []*models.Action) error
}

// actionRepository реализует интерфейс ActionRepository
type actionRepository struct {
	db *gorm.DB
}

// NewActionRepository создает новый экземпляр репозитория действий
func NewActionRepository(db *gorm.DB) ActionRepository {
	return &actionRepository{
		db: db,
	}
}

// Create сохраняет новое действие в базе данных
func (r *actionRepository) Create(action *models.Action) error {
	return r.db.Create(action).Error
}

// GetByID возвращает действие по ID
func (r *actionRepository) GetByID(id uint) (*models.Action, error) {
	var action models.Action
	if err := r.db.First(&action, id).Error; err != nil {
		return nil, err
	}
	return &action, nil
}

// GetByCollectionID возвращает действия коллекции
func (r *actionRepository) GetByCollectionID(collectionID uint) ([]*models.Action, error) {
	var actions []*models.Action
	if err := r.db.Where("collection_id = ?", collectionID).Order("order ASC").Find(&actions).Error; err != nil {
		return nil, err
	}
	return actions, nil
}

// Update обновляет данные действия
func (r *actionRepository) Update(action *models.Action) error {
	return r.db.Save(action).Error
}

// Delete удаляет действие (soft delete через GORM)
func (r *actionRepository) Delete(id uint) error {
	return r.db.Delete(&models.Action{}, id).Error
}

// BatchCreate сохраняет несколько действий в транзакции
func (r *actionRepository) BatchCreate(actions []*models.Action) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, action := range actions {
			if err := tx.Create(action).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
