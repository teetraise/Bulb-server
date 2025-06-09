package models

import (
    "time"

    "gorm.io/gorm"
)

// ActionType представляет тип действия (Правда или Действие)
type ActionType string

const (
    ActionTypeTruth ActionType = "truth"
    ActionTypeDare  ActionType = "dare"
)

// Action представляет действие в подборке
type Action struct {
    ID           uint           `gorm:"primaryKey" json:"id"`
    Text         string         `json:"text" gorm:"not null"`
    Type         ActionType     `json:"type" gorm:"type:varchar(20);not null;default:'truth'"` // Новое поле для типа
    CollectionID uint           `json:"collectionId"`
    Collection   Collection     `json:"-" gorm:"foreignKey:CollectionID"`
    Order        int            `json:"order" gorm:"default:0"` // Порядок действий в подборке
    CreatedAt    time.Time      `json:"createdAt"`
    UpdatedAt    time.Time      `json:"updatedAt"`
    DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}