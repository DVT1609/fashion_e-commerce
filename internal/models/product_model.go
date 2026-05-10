package models

import (
	"time"
)

// ==========================================
// MẪU SẢN PHẨM (Product Models)
// ==========================================
type ProductModel struct {
	ProductModelID uint      `gorm:"column:product_model_id;primaryKey" json:"product_model_id"`
	ProductID      uint      `gorm:"column:product_id;index" json:"product_id"`
	SKU            string    `gorm:"column:sku;uniqueIndex;type:varchar(50)" json:"sku"`
	Size           string    `gorm:"column:size;type:varchar(10)" json:"size"`
	Color          string    `gorm:"column:color;type:varchar(20)" json:"color"`
	Price          float64   `gorm:"column:price;type:decimal(15,2)" json:"price"`
	Stock          int       `gorm:"column:stock;not null" json:"stock"`
	CreatedAt      time.Time `gorm:"column:create_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"create_at"`
	UpdatedAt      time.Time `gorm:"column:update_at;type:timestamp;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"update_at"`
}