package models

import (
	"github.com/google/uuid"
)

// ==========================================
// SẢN PHẨM TRONG ĐƠN HÀNG (Order Items)
// ==========================================
type OrderItem struct {
	OrderItemID    uint      `gorm:"column:order_item_id;primaryKey" json:"order_item_id"`
	OrderID        uuid.UUID `gorm:"column:order_id;type:char(36);index" json:"order_id"`
	ProductModelID uint      `gorm:"column:product_model_id;index" json:"product_model_id"`
	Quantity       int       `gorm:"column:quantity;not null" json:"quantity"`
	PriceAtPurchase float64  `gorm:"column:price_at_purchase;type:decimal(15,2)" json:"price_at_purchase"`
}