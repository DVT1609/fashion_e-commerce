package repositoryMysql

import (
	// "errors"
	"github.com/DVT1609/fashion_e-commerce.git/internal/models"
	"gorm.io/gorm"
)

type UserMysqlRepository struct {
	DB *gorm.DB
}

func (repo *UserMysqlRepository) CheckUserExistsMysql(email, username string) (bool, error) {
	var user models.User

	// Sử dụng UNION ALL để ép MySQL dùng riêng biệt Index của email và username.
	// Loại bỏ ORDER BY để tránh việc MySQL phải thực hiện thêm bước sort dữ liệu.
	// Limit 1 ở ngoài cùng giúp dừng truy vấn ngay khi tìm thấy bất kỳ sự trùng lặp nào.
	query := `
		(SELECT * FROM users WHERE email = ? LIMIT 1)
		UNION ALL
		(SELECT * FROM users WHERE username = ? LIMIT 1)
		LIMIT 1
	`

	err := repo.DB.Raw(query, email, username).Scan(&user).Error
	
	if err != nil {
		return false, err
	}

	// Trong GORM, khi dùng Raw().Scan(), nếu không tìm thấy bản ghi nào, 
	// nó sẽ không trả về lỗi gorm.ErrRecordNotFound mà sẽ trả về object rỗng.
	// Vì vậy ta kiểm tra ID (hoặc Email) của user để biết có tìm thấy hay không.
	if user.Email == "" && user.Username == "" {
		return false, nil
	}

	// Nếu tìm thấy bất kỳ trường nào, chứng tỏ tài khoản đã tồn tại
	return true, nil
}