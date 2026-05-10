package models

import (
	"time"
)

// ==========================================
// 1. NGƯỜI DÙNG (Users)
// ==========================================
type User struct {
	UserID       uint      `gorm:"column:user_id;primaryKey" json:"user_id"`
	Username     string    `gorm:"column:username;uniqueIndex;type:varchar(50);not null" json:"username"`
	Email        string    `gorm:"column:email;uniqueIndex;type:varchar(100);not null" json:"email"`
	PasswordHash string    `gorm:"column:password_hash;type:varchar(255);not null" json:"password_hash"`
	Role         string    `gorm:"column:role;type:varchar(20);default:'customer';index" json:"role"`
	FullName     string    `gorm:"column:full_name;type:varchar(100)" json:"full_name"`
	Phone        string    `gorm:"column:phone;type:varchar(10)" json:"phone"`
	Address      string    `gorm:"column:address;type:text" json:"address"`
	CreatedAt    time.Time `gorm:"column:create_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"create_at"`
	UpdatedAt    time.Time `gorm:"column:update_at;type:timestamp;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"update_at"`

	// Relationships (Bỏ tag column ở đây)
	Orders  []Order  `gorm:"foreignKey:UserID" json:"orders"`
	Cart    Cart     `gorm:"foreignKey:UserID" json:"cart"`
	Reviews []Review `gorm:"foreignKey:UserID" json:"reviews"`
}

// =====================================================
// 2. ĐẦU VÀO DỮ LIỆU ĐĂNG KÝ NGƯỜI DÙNG (InputRegister)
// =====================================================

type InputRegister struct {
	Username        string `json:"username" validate:"required,min=6"`
	FullName        string `json:"full_name" validate:"required,min=6"`
	Email           string `json:"email" validate:"required,email"`
	Phone           string `json:"phone" validate:"required,len=10,startswith=0,numeric"`
	Address         string `json:"address" validate:"required"`
	Password        string `json:"password" validate:"required,min=10"`
	PasswordConfirm string `json:"password_confirm" validate:"required,eqfield=Password"`
}

// ======================================================================================
// 3. YÊU CẦU ĐĂNG KÝ NGƯỜI DÙNG (RegisterRequest)
// dữ liệu đã được xác thực và kiểm tra ở tầng Handler, chuẩn bị để gửi vào tầng Service
// ======================================================================================

type RegisterRequest struct {
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Address  string `json:"address"`
	Password string `json:"password"`
}
