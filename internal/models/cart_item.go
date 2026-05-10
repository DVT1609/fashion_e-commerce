package models

import (
	"time"
)

type CartItem struct {
	CartItemID     uint      `gorm:"column:cart_item_id;primaryKey" json:"id"`
	CartID         uint      `gorm:"column:cart_id;index" json:"cart_id"`
	ProductModelID uint      `gorm:"column:product_model_id;index" json:"product_model_id"`
	Quantity       int       `gorm:"column:quantity;not null;default:1" json:"quantity"`
	CreatedAt      time.Time `gorm:"column:create_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"create_at"`
	UpdatedAt      time.Time `gorm:"column:update_at;type:timestamp;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"update_at"`
}