package worker

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/DVT1609/fashion_e-commerce.git/internal/models"
	"github.com/DVT1609/fashion_e-commerce.git/internal/utils/security" // File password.go của Bạn
	as "github.com/aerospike/aerospike-client-go/v7"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)

type RegisterWorker struct {
	KafkaReader     *kafka.Reader
	MysqlDB         *gorm.DB
	AerospikeClient *as.Client
	AsCtxWrite      *as.WritePolicy
}

func NewRegisterWorker(reader *kafka.Reader, db *gorm.DB, asClient *as.Client, asCtxWrite *as.WritePolicy) *RegisterWorker {
	return &RegisterWorker{
		KafkaReader:     reader,
		MysqlDB:         db,
		AerospikeClient: asClient,
		AsCtxWrite:      asCtxWrite,
	}
}

// Start kịch bản lắng nghe hàng đợi Kafka liên tục
func (w *RegisterWorker) Start(ctx context.Context) {
	log.Println("Worker đăng ký đã sẵn sàng và đang lắng nghe Kafka...")

	for {
		// 1. Đọc message từ Kafka (Hàm này sẽ block luồng cho đến khi có tin nhắn mới)
		msg, err := w.KafkaReader.ReadMessage(ctx)
		if err != nil {
			log.Printf("Worker lỗi đọc message từ Kafka: %v", err)
			continue
		}

		// 2. Parse dữ liệu JSON ngược lại thành cấu trúc User
		var user models.User
		err = json.Unmarshal(msg.Value, &user)
		if err != nil {
			log.Printf("Worker lỗi parse JSON: %v", err)
			continue
		}

		// Xử lý logic đăng ký chính thức cho User này
		w.processRegistration(user)
	}
}

func (w *RegisterWorker) processRegistration(user models.User) {
	// 3. TẦNG BĂM NẶNG: Lấy chuỗi SHA-256 từ Kafka và băm chồng Argon2id lên
	// Việc này tốn CPU nhưng chạy ngầm ở Worker nên không làm chậm API phía trước
	finalArgon2Hash, err := security.HashPasswordArgon2(user.PasswordHash)
	if err != nil {
		log.Printf("Worker lỗi băm Argon2id cho email %s: %v", user.Email, err)
		return
	}
	user.PasswordHash = finalArgon2Hash // Ghi đè mã Argon2id vào struct

	// 4. CHỐT CHẶN CUỐI: INSERT dữ liệu chính thức vào MySQL
	err = w.MysqlDB.Create(&user).Error
	if err != nil {
		// Kiểm tra nếu trúng lỗi trùng lặp Unique Constraint (Mã lỗi 1062 trong MySQL)
		if strings.Contains(err.Error(), "Error 1062") || strings.Contains(err.Error(), "Duplicate entry") {
			log.Printf("⚠️ Phát hiện Message trùng lặp lọt lưới tại Kafka! Email: %s, Username: %s", user.Email, user.Username)
			
			// Thực hiện đồng bộ ngược (Vá lỗi Cache): Ghi đè dấu vết tồn tại vào Aerospike vĩnh viễn
			w.compensateAerospikeCache(user.Email, user.Username)
			return
		}

		// Nếu là các lỗi hệ thống khác (mất kết nối DB...), log lại để hệ thống retry sau
		log.Printf("Worker lỗi ghi MySQL: %v", err)
		return
	}

	log.Printf("✅ Worker đã xử lý đăng ký thành công cho Email: %s", user.Email)
}

// Hàm đồng bộ ngược để sửa sai cho lớp Cache Barrier phía trên
func (w *RegisterWorker) compensateAerospikeCache(email, username string) {
	keyEmail, _ := as.NewKey("fashion_e-commerce", "usersRegister", "pending:email:"+email)
	keyUsername, _ := as.NewKey("fashion_e-commerce", "usersRegister", "pending:username:"+username)

	bins := as.BinMap{
		"email":    email,
		"username": username,
		"status":   "confirmed_in_db", // Đánh dấu trạng thái đã chốt trong DB
	}

	// Ghi bù vào Aerospike với một TTL rất dài (ví dụ: 30 ngày) để chặn đứng các request spam sau này
	// Lưu ý: Bạn có thể tạo một Policy riêng với TTL dài hơn nếu muốn, ở đây tạm dùng CtxWrite hiện tại
	_ = w.AerospikeClient.Put(w.AsCtxWrite, keyEmail, bins)
	_ = w.AerospikeClient.Put(w.AsCtxWrite, keyUsername, bins)

	log.Printf("🔄 Đã vá lỗi Cache thành công! Đồng bộ trạng thái tồn tại của %s ngược lại Aerospike.", email)
}