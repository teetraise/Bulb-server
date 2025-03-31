package models

import (
	"time"

	"gorm.io/gorm"
)

// User представляет модель пользователя в системе
type User struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `json:"name"`
	Surname     string         `json:"surname"`
	Email       string         `json:"email" gorm:"uniqueIndex"`
	Password    string         `json:"-"` // Не отдаем пароль в JSON-ответах
	Phone       string         `json:"phone"`
	ImageURL    string         `json:"imageUrl"`
	Description string         `json:"description"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
