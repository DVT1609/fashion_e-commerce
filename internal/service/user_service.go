package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"github.com/DVT1609/fashion_e-commerce.git/internal/models"
	"github.com/DVT1609/fashion_e-commerce.git/internal/repository/repositoryAerospike"
	"github.com/DVT1609/fashion_e-commerce.git/internal/repository/repositoryKafkaProducer"
	"github.com/DVT1609/fashion_e-commerce.git/internal/repository/repositoryMysql"
)

type UserService struct {
	KafkaProducer           *repositoryKafkaProducer.UserKafkaProducerRepository // Để đẩy vào hàng đợi
	UserMysqlRepository     *repositoryMysql.UserMysqlRepository
	UserAerospikeRepository *repositoryAerospike.UserAerospikeRepository
}

func NewUserService(kafkaProducerRepository *repositoryKafkaProducer.UserKafkaProducerRepository, mysqlrepository *repositoryMysql.UserMysqlRepository, aerospikeRepository *repositoryAerospike.UserAerospikeRepository) *UserService {
	return &UserService{
		KafkaProducer:           kafkaProducerRepository,
		UserMysqlRepository:     mysqlrepository,
		UserAerospikeRepository: aerospikeRepository,
	}
}

func (service *UserService) Register(registerRequest *models.RegisterRequest) error {

	// 1. Check Aerospike trước — chốt chặn tốc độ cao, tránh hit DB khi spam
	exists, err := service.UserAerospikeRepository.CheckUserExistsAerospike(registerRequest.Email, registerRequest.Username)
	if err != nil {
		return err
	}
	// Aerospike có record (pending hoặc confirmed_in_db) → chặn ngay, không tiếp tục
	if exists {
		return errors.New("Email hoặc username đã tồn tại hoặc đang được xử lý")
	}

	// 2. Aerospike không có → check MySQL để đảm bảo dữ liệu nhất quán
	existsInMySQL, err := service.UserMysqlRepository.CheckUserExistsMysql(registerRequest.Email, registerRequest.Username)
	if err != nil {
		return err
	}
	if existsInMySQL {
		return errors.New("Email hoặc username đã tồn tại")
	}

	// 2. Mã hóa mật khẩu (Bcrypt)
	// Phải băm mật khẩu trước khi lưu trữ hoặc đẩy vào Kafka để bảo mật thông tin người dùng
	// hashedPassword, err := security.HashPasswordArgon2(registerRequest.Password)
	// if err != nil {
	// 	log.Printf("Lỗi khi băm mật khẩu Argon2: %v", err)
	// 	return errors.New("không thể xử lý mật khẩu")
	// }

	// 🛠️ 2. BĂM MẬT KHẨU SIÊU NHẸ BẰNG SHA-256 TẠI API
	// Giải pháp này giúp bảo mật đường truyền sang Kafka mà không gây nghẽn CPU API
	hash := sha256.New()
	hash.Write([]byte(registerRequest.Password))
	passwordHash := hex.EncodeToString(hash.Sum(nil))

	// Chuẩn bị dữ liệu để lưu trữ (thay mật khẩu thô bằng bản đã băm)
	userModel := models.User{
		Username:     registerRequest.Username,
		FullName:     registerRequest.FullName,
		Email:        registerRequest.Email,
		Phone:        registerRequest.Phone,
		Address:      registerRequest.Address,
		PasswordHash: string(passwordHash),
	}

	// 3. Ghi tạm vào Aerospike để "giữ chỗ"
	// Giúp ngăn chặn các request đăng ký cùng lúc cho cùng một email trong khi chờ Kafka xử lý
	err = service.UserAerospikeRepository.CreateRecordRegister(userModel)
	if err != nil {
		return err
	}

	// 4. Đẩy vào Kafka Producer — trả Success cho client ngay khi push thành công
	err = service.KafkaProducer.ProduceRegisterMessage(userModel)
	if err != nil {
		// Xóa Aerospike reservation để user có thể thử lại
		if _, deleteErr := service.UserAerospikeRepository.DeleteRecordRegister(
			userModel.Email, userModel.Username,
		); deleteErr != nil {
			// log.Printf("Lỗi xóa Aerospike sau Kafka fail: %v", deleteErr)
		}
		return errors.New("Đăng ký thất bại do lỗi hệ thống, vui lòng thử lại sau")
	}

	return nil
}
