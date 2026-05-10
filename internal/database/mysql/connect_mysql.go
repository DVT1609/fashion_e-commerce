package mysql

import(
	"log"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/driver/mysql"
	"github.com/DVT1609/fashion_e-commerce.git/config"
)

func ConnectMysql() (*gorm.DB, error) {
	config := config.LoadConfigMysql()

	// Sử dụng fmt.Sprintf để lắp ghép các biến vào chuỗi
	// %s đại diện cho một chuỗi (string)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.MysqlUser,
		config.MysqlPassword,
		config.MysqlHost,
		config.MysqlDatabase,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatalf("Lỗi kết nối cơ sở dữ liệu 1: %v", err)
		return nil, err
	}

	log.Printf("Kết nối cơ sở dữ liệu thành công 1")
	return db, nil

}
