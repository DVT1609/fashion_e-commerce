package main

import (
	"github.com/DVT1609/fashion_e-commerce.git/internal/database/aerospike"                 // Thay thế bằng đường dẫn thực tế đến package aerospike
	"github.com/DVT1609/fashion_e-commerce.git/internal/database/mysql"                     // Thay thế bằng đường dẫn thực tế đến package mysql
	"github.com/DVT1609/fashion_e-commerce.git/internal/handler"                            // Thay thế bằng đường dẫn thực tế đến package handler
	"github.com/DVT1609/fashion_e-commerce.git/internal/message_queue/kafka"                // Thay thế bằng đường dẫn thực tế đến package kafka
	"github.com/DVT1609/fashion_e-commerce.git/internal/models"                             // Thay thế bằng đường dẫn thực tế đến package models
	"github.com/DVT1609/fashion_e-commerce.git/internal/repository/repositoryAerospike"     // Thay thế bằng đường dẫn thực tế đến package repository aerospike
	"github.com/DVT1609/fashion_e-commerce.git/internal/repository/repositoryKafkaProducer" // Thay thế bằng đường dẫn thực tế đến package repository kafka producer
	"github.com/DVT1609/fashion_e-commerce.git/internal/repository/repositoryMysql"         // Thay thế bằng đường dẫn thực tế đến package repository mysql
	"github.com/DVT1609/fashion_e-commerce.git/internal/service"                            // Thay thế bằng đường dẫn thực tế đến package service
	as "github.com/aerospike/aerospike-client-go/v7"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

func main() {
	app := fiber.New()

	// Database Mysql and Migration
	// Kết nối đến MySQL
	db, err := mysql.ConnectMysql()
	if err != nil {
		panic("Không thể kết nối đến database: " + err.Error())
	}
	// Giả sử db là kết nối gorm.DB của bạn
	err = db.AutoMigrate(
		&models.User{},
		&models.Product{},
		&models.ProductModel{},
		&models.Cart{},
		&models.CartItem{},
		&models.Order{},
		&models.OrderItem{},
		&models.Review{},
	)

	if err != nil {
		panic("Không thể migrate database: " + err.Error())
	}

	// Khởi tạo validator
	validate := validator.New()

	// Kết nối đến Aerospike
	aerospikeClient, err := aerospike.ConnectAerospike()
	if err != nil {
		panic("Không thể kết nối đến Aerospike: " + err.Error())
	}

	// Cấu hình context cho Aerospike Repository
	ctxBase := as.NewPolicy()
	ctxWrite := as.NewWritePolicy(0, 300) // TTL 5 phút cho record tạm thời

	// Khởi tạo Producer Kafka
	writer := kafka.CreateProducer()

	// Khởi tạo Repository
	// Khởi tạo Repository Aerospike với client và context đã cấu hình
	userRepositoryAerospike := repositoryAerospike.UserAerospikeRepository{Client: aerospikeClient, CtxBase: ctxBase, CtxWrite: ctxWrite}
	// Khởi tạo Repository MySQL
	userRepositoryMysql := repositoryMysql.UserMysqlRepository{DB: db}
	// Khởi tạo Repository Kafka Producer (giả sử bạn đã có cấu hình Kafka)
	userRepositoryKafka := repositoryKafkaProducer.UserKafkaProducerRepository{Writer: writer}

	// Khởi tạo Service
	userService := service.NewUserService(&userRepositoryKafka, &userRepositoryMysql, &userRepositoryAerospike)

	// Khởi tạo Handler
	userHandler := handler.NewUserHandler(userService, validate)

	app.Post("/Register", userHandler.Register)

	app.Listen(":3004")
}
