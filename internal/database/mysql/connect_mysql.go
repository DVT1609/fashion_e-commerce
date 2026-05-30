package mysql

import (
	"fmt"
	"log"
	"time"

	"github.com/DVT1609/fashion_e-commerce.git/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
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

	// 🛠️ CẤU HÌNH POOL: VỚI CÔNG THỨC POOL SIZE = CPU CORES * 2 + 1
	sqlDB, err := db.DB()
	if err == nil {
		// Số lượng kết nối mở tối đa tới DB, thường được tính theo công thức: CPU Cores * 2 + 1
		sqlDB.SetMaxOpenConns(9)

		// Số lượng kết nối rảnh tối đa được giữ lại trong pool để tái sử dụng
		sqlDB.SetMaxIdleConns(9)

		// Thời gian tối đa một kết nối có thể sống trong pool (tránh lỗi kết nối "lạc" phía MySQL)
		sqlDB.SetConnMaxLifetime(30 * time.Minute)
	}

	log.Printf("Kết nối cơ sở dữ liệu thành công 1")
	return db, nil

}
