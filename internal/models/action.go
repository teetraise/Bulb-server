package models

import (
    "time"

    "gorm.io/gorm"
)

// Action представляет действие в подборке
type Action struct {
    ID           uint           `gorm:"primaryKey" json:"id"`
    Text         string         `json:"text"`
    CollectionID uint           `json:"collectionId"`
    Collection   Collection     `json:"-" gorm:"foreignKey:CollectionID"`
    Order        int            `json:"order" gorm:"default:0"` // Порядок действий в подборке
    CreatedAt    time.Time      `json:"createdAt"`
    UpdatedAt    time.Time      `json:"updatedAt"`
    DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}