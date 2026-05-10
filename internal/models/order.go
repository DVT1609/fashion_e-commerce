package models

import (
	"time"
	"gorm.io/gorm"
	"github.com/google/uuid"
)

// ==========================================
// ĐƠN HÀNG (Orders)
// ==========================================
type Order struct {
	OrderID     uuid.UUID   `gorm:"column:order_id;type:char(36);primaryKey" json:"order_id"`
	UserID      uint        `gorm:"column:user_id;index" json:"user_id"`
	TotalAmount float64     `gorm:"column:total_amount;type:decimal(15,2)" json:"total_amount"`
	Status      string      `gorm:"column:status;type:enum('pending','processing','shipped','cancelled');default:'pending'" json:"status"`
	CreatedAt   time.Time   `gorm:"column:create_at;type:timestamp;default:CURRENT_TIMESTAMP;index" json:"create_at"`
	UpdatedAt   time.Time   `gorm:"column:update_at;type:timestamp;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"update_at"`
	
	OrderItems  []OrderItem `gorm:"foreignKey:OrderID" json:"order_items"`
}

func (o *Order) BeforeCreate(tx *gorm.DB) (err error) {
	o.OrderID = uuid.New()
	return
}