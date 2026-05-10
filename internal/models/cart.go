package models

import (
	"time"
)

// ==========================================
// GIỎ HÀNG (Cart)
// ==========================================
type Cart struct {
	CartID    uint       `gorm:"column:cart_id;primaryKey" json:"id"`
	UserID    uint       `gorm:"column:user_id;uniqueIndex;not null" json:"user_id"`
	CreatedAt time.Time  `gorm:"column:create_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"create_at"`
	UpdatedAt time.Time  `gorm:"column:update_at;type:timestamp;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"update_at"`
	
	Items     []CartItem `gorm:"foreignKey:CartID" json:"items"`
}

