package models

import (
	"time"
)// ==========================================
// SẢN PHẨM (Products)
// ==========================================
type Product struct {
	ProductID   uint           `gorm:"column:product_id;primaryKey" json:"product_id"`
	Name        string         `gorm:"column:name;type:varchar(255);index" json:"name"`
	Description string         `gorm:"column:description;type:text" json:"description"`
	IsFlashSale bool           `gorm:"column:is_flash_sale;default:false;index" json:"is_flash_sale"`
	CreatedAt   time.Time      `gorm:"column:create_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"create_at"`
	UpdatedAt   time.Time      `gorm:"column:update_at;type:timestamp;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"update_at"`
	
	// Relationships
	Models      []ProductModel `gorm:"foreignKey:ProductID" json:"variants"`
	Reviews     []Review       `gorm:"foreignKey:ProductID" json:"reviews"`
}





