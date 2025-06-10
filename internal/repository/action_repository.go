package repository

import (
	"fmt"
	"log"

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
	if err := r.db.Create(action).Error; err != nil {
		log.Printf("Error creating action: %v", err)
		return fmt.Errorf("failed to create action: %w", err)
	}
	return nil
}

// GetByID возвращает действие по ID
func (r *actionRepository) GetByID(id uint) (*models.Action, error) {
	var action models.Action
	if err := r.db.First(&action, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("action with id %d not found", id)
		}
		log.Printf("Error getting action by ID %d: %v", id, err)
		return nil, fmt.Errorf("failed to get action: %w", err)
	}
	return &action, nil
}

// GetByCollectionID возвращает действия коллекции
func (r *actionRepository) GetByCollectionID(collectionID uint) ([]*models.Action, error) {
	var actions []*models.Action

	// Добавляем логирование для отладки
	log.Printf("Fetching actions for collection ID: %d", collectionID)

	// Выполняем запрос с явным указанием полей и проверкой типа
	query := r.db.Where("collection_id = ? AND deleted_at IS NULL", collectionID).
		Order("\"order\" ASC"). // Экранируем 'order' так как это зарезервированное слово в SQL
		Find(&actions)

	if err := query.Error; err != nil {
		log.Printf("Error fetching actions for collection %d: %v", collectionID, err)
		return nil, fmt.Errorf("failed to get actions for collection %d: %w", collectionID, err)
	}

	log.Printf("Found %d actions for collection %d", len(actions), collectionID)

	// Проверяем каждое действие на корректность типа
	for i, action := range actions {
		if action.Type != models.ActionTypeTruth && action.Type != models.ActionTypeDare {
			log.Printf("Warning: Action %d has invalid type: %s", action.ID, action.Type)
			// Устанавливаем тип по умолчанию вместо возврата ошибки
			actions[i].Type = models.ActionTypeTruth
		}
	}

	return actions, nil
}

// Update обновляет данные действия
func (r *actionRepository) Update(action *models.Action) error {
	// Валидируем тип действия перед обновлением
	if action.Type != models.ActionTypeTruth && action.Type != models.ActionTypeDare {
		return fmt.Errorf("invalid action type: %s", action.Type)
	}

	if err := r.db.Save(action).Error; err != nil {
		log.Printf("Error updating action %d: %v", action.ID, err)
		return fmt.Errorf("failed to update action: %w", err)
	}
	return nil
}

// Delete удаляет действие (soft delete через GORM)
func (r *actionRepository) Delete(id uint) error {
	result := r.db.Delete(&models.Action{}, id)
	if err := result.Error; err != nil {
		log.Printf("Error deleting action %d: %v", id, err)
		return fmt.Errorf("failed to delete action: %w", err)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("action with id %d not found", id)
	}

	return nil
}

// BatchCreate сохраняет несколько действий в транзакции
func (r *actionRepository) BatchCreate(actions []*models.Action) error {
	if len(actions) == 0 {
		return nil
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		for i, action := range actions {
			// Валидируем тип действия
			if action.Type != models.ActionTypeTruth && action.Type != models.ActionTypeDare {
				return fmt.Errorf("invalid action type for action %d: %s", i, action.Type)
			}

			if err := tx.Create(action).Error; err != nil {
				log.Printf("Error creating action %d in batch: %v", i, err)
				return fmt.Errorf("failed to create action %d: %w", i, err)
			}
		}
		log.Printf("Successfully created %d actions in batch", len(actions))
		return nil
	})
}
