package models

import (
	"time"
	"github.com/google/uuid"
)

// ==========================================
// 5. ĐÁNH GIÁ (Reviews)
// ==========================================
type Review struct {
	ReviewID  uint      `gorm:"column:review_id;primaryKey" json:"id"`
	UserID    uint      `gorm:"column:user_id;index" json:"user_id"`
	ProductID uint      `gorm:"column:product_id;index" json:"product_id"`
	OrderID   uuid.UUID `gorm:"column:order_id;type:char(36);index" json:"order_id"`
	Rating    int       `gorm:"column:rating;type:tinyint;not null" json:"rating"`
	Comment   string    `gorm:"column:comment;type:text" json:"comment"`
	CreatedAt time.Time `gorm:"column:create_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"create_at"`
	UpdatedAt time.Time `gorm:"column:update_at;type:timestamp;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"update_at"`
}