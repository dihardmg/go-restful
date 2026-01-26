package models

import (
	"time"

	"gorm.io/gorm"
)

// Product represents a product entity
type Product struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null" binding:"required"`
	Description string    `json:"description" gorm:"type:text"`
	Price       float64   `json:"price" gorm:"not null" binding:"required,gt=0"`
	Stock       int       `json:"stock" gorm:"not null;default:0" binding:"gte=0"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName specifies the table name for Product model
func (Product) TableName() string {
	return "products"
}
