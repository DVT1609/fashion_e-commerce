package repositoryMysql

import (
	"errors"
	"github.com/DVT1609/fashion_e-commerce.git/internal/models"
	"gorm.io/gorm"
)

type UserMysqlRepository struct {
	DB *gorm.DB
}

func (repo *UserMysqlRepository) CheckUserExistsMysql(email, username string) (bool, error) {
	var user models.User
    // Truy vấn tìm user đầu tiên khớp với email hoặc username
	err := repo.DB.Model(&models.User{}).Where("email = ? OR username = ?", email, username).First(&user).Error
	
    if err != nil {
        // Trường hợp 1: Không tìm thấy bản ghi nào -> Tài khoản chưa tồn tại (Hợp lệ để đăng ký)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil 
		}
        // Trường hợp 2: Lỗi kết nối DB, lỗi cú pháp... -> Cần trả về lỗi để Service xử lý
		return false, err 
	}

    // Trường hợp 3: Không có lỗi (err == nil) -> Đã tìm thấy User -> Tài khoản ĐÃ tồn tại
	return true, nil
}