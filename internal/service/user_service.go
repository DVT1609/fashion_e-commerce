package service

import (
	"errors"
	"log"

	"github.com/DVT1609/fashion_e-commerce.git/internal/models"
	"github.com/DVT1609/fashion_e-commerce.git/internal/repository/repositoryAerospike"
	"github.com/DVT1609/fashion_e-commerce.git/internal/repository/repositoryKafkaProducer"
	"github.com/DVT1609/fashion_e-commerce.git/internal/repository/repositoryMysql"
	"github.com/gofiber/fiber/v3"
	"golang.org/x/crypto/bcrypt"
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

func (service *UserService) Register(ctx fiber.Ctx, registerRequest *models.RegisterRequest) error {

	// 1. Check Aerospike (Email/Username?Password) để chặn request trùng lặp ngay lập tức
	// Đây là "chốt chặn" tốc độ cao giúp hệ thống không bị quá tải bởi các yêu cầu spam
	exists, err := service.UserAerospikeRepository.CheckUserExistsAerospike(registerRequest.Email, registerRequest.Username)
	if err != nil {
		log.Printf("Lỗi kiểm tra tồn tại dữ liệu chưa trong aerospike: %v", err)
		return err
	}

	if !exists {
		// Nếu không tồn tại, tiếp tục quy trình check trong MySQL để đảm bảo dữ liệu nhất quán
		existsInMySQL, err := service.UserMysqlRepository.CheckUserExistsMysql(registerRequest.Email, registerRequest.Username)
		if err != nil {
			log.Printf("Lỗi kiểm tra tồn tại dữ liệu trong mysql: %v", err)
			return err
		}

		if existsInMySQL == true {
			return errors.New("Email hoặc username đã tồn tại")
		}
	}

	// 2. Mã hóa mật khẩu (Bcrypt)
	// Phải băm mật khẩu trước khi lưu trữ hoặc đẩy vào Kafka để bảo mật thông tin người dùng
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registerRequest.Password), bcrypt.MinCost)
	if err != nil {
		return err
	}

	// Chuẩn bị dữ liệu để lưu trữ (thay mật khẩu thô bằng bản đã băm)
	userModel := models.User{
		Username:     registerRequest.Username,
		FullName:     registerRequest.FullName,
		Email:        registerRequest.Email,
		Phone:        registerRequest.Phone,
		Address:      registerRequest.Address,
		PasswordHash: string(hashedPassword),
	}

	// 3. Ghi tạm vào Aerospike để "giữ chỗ"
	// Giúp ngăn chặn các request đăng ký cùng lúc cho cùng một email trong khi chờ Kafka xử lý
	err = service.UserAerospikeRepository.CreateRecordRegister(userModel)
	if err != nil {
		return err
	}

	// 4. Đẩy vào Kafka Producer
	// Sau khi ném vào Kafka, ta có thể trả về Success cho khách hàng ngay lập tức
	err = service.KafkaProducer.ProduceRegisterMessage(userModel)
	if err != nil {
		// Nếu Kafka lỗi, cần xóa "chỗ" đã giữ trong Aerospike để người dùng có thể thử lại
		deleteSuccess, deleteErr := service.UserAerospikeRepository.DeleteRecordRegister(userModel.Email, userModel.Username)
		if deleteErr != nil {
			log.Printf("Lỗi khi xóa record tạm thời trong Aerospike: %v", deleteErr)
		}
		if deleteSuccess {
			ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Đăng ký thất bại do lỗi hệ thống, vui lòng thử lại sau",
			})
		}
	}

	return nil
}
