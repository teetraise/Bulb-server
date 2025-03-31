package models

import (
    "time"

    "gorm.io/gorm"
)

// Collection представляет подборку в системе
type Collection struct {
    ID          uint           `gorm:"primaryKey" json:"id"`
    Name        string         `json:"name"`
    Description string         `json:"description"`
    ImageURL    string         `json:"imageUrl"`
    UserID      uint           `json:"userId"`
    User        User           `json:"user" gorm:"foreignKey:UserID"`
    Actions     []Action       `json:"actions,omitempty" gorm:"foreignKey:CollectionID"`
    PlayCount   int            `json:"playCount" gorm:"default:0"`
    CreatedAt   time.Time      `json:"createdAt"`
    UpdatedAt   time.Time      `json:"updatedAt"`
    DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}