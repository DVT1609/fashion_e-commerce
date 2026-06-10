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
	CtxWrite        *as.WritePolicy
	batchSize       int           // Số lượng phần tử tối đa trong 1 batch (1000)
	flushInterval   time.Duration // Thời gian chờ tối đa (10 giây)
}

func NewRegisterWorker(reader *kafka.Reader, db *gorm.DB, asClient *as.Client, asCtxWrite *as.WritePolicy) *RegisterWorker {
	return &RegisterWorker{
		KafkaReader:     reader,
		MysqlDB:         db,
		AerospikeClient: asClient,
		CtxWrite:        asCtxWrite,
		batchSize:       1000,
		flushInterval:   10 * time.Second,
	}
}

// Start kịch bản lắng nghe hàng đợi Kafka liên tục
func (w *RegisterWorker) WorkerRegister(ctx context.Context) {
	log.Println("Worker đăng ký đã sẵn sàng và đang lắng nghe Kafka...")

	// Tạo một channel nội bộ để trung chuyển các User đã băm Argon2id
	userChan := make(chan models.User, w.batchSize)

	// Mảng tạm để chứa lô dữ liệu chuẩn bị ghim vào MySQL
	batch := make([]models.User, 0, w.batchSize)

	// Khởi tạo bộ đếm thời gian 10 giây
	ticker := time.NewTicker(w.flushInterval)
	defer ticker.Stop()

	// Kích hoạt một Goroutine ngầm chuyên đọc tin nhắn từ Kafka và ném vào channel
	go func() {
		for {
			msg, err := w.KafkaReader.ReadMessage(ctx)
			if err != nil {
				log.Printf("Worker lỗi đọc Kafka: %v", err)
				continue
			}

			var user models.User
			if err := json.Unmarshal(msg.Value, &user); err != nil {
				log.Printf("Worker lỗi parse JSON: %v", err)
				continue
			}

			// TẦNG BĂM NẶNG: Thực hiện ngay khi vừa lấy ra khỏi Kafka
			finalArgon2Hash, err := security.HashPasswordArgon2(user.PasswordHash)
			if err != nil {
				log.Printf("Worker lỗi băm Argon2id cho email %s: %v", user.Email, err)
				continue
			}
			user.PasswordHash = finalArgon2Hash

			// Đẩy user đã xử lý xong vào channel để đợi gom lô
			userChan <- user
		}
	}()

	for {
		select {
		case user := <-userChan:
			// Nhận được 1 user từ channel, nạp ngay vào mảng batch
			batch = append(batch, user)

			// ĐIỀU KIỆN 1: Nếu mảng đạt đủ 1000 phần tử -> Kích hoạt ghi ngay
			if len(batch) >= w.batchSize {
				log.Printf("🚀 Đủ %d request, tiến hành Bulk Insert vào MySQL...", len(batch))
				w.flushBatch(batch)
				batch = make([]models.User, 0, w.batchSize) // Reset lại mảng trống
				ticker.Reset(w.flushInterval)               // Reset lại đồng hồ 10 giây
			}

		case <-ticker.C:
			// ĐIỀU KIỆN 2: Hết 10 giây, nếu mảng tạm đang có dữ liệu (dù chưa đủ 1000) -> Kích hoạt ghi
			if len(batch) > 0 {
				log.Printf("⏰ Hết 10 giây chờ, tiến hành ghi %d request còn lại vào MySQL...", len(batch))
				w.flushBatch(batch)
				batch = make([]models.User, 0, w.batchSize) // Reset lại mảng trống
			}
		}
	}

}

// Hàm thực thi Bulk Insert dữ liệu hàng loạt và xử lý lỗi trùng lặp
func (w *RegisterWorker) flushBatch(users []models.User) {
	// GORM tự động nhận diện mảng Slice và chuyển đổi thành câu lệnh Bulk Insert cực nhanh
	// Bạn có thể dùng CreateInBatches để GORM tự chia nhỏ ra nếu mảng quá lớn
	err := w.MysqlDB.Create(&users).Error
	if err != nil {
		// Nếu cả lô 1000 thằng bị lỗi (thường do dính 1 vài thằng trùng Unique Index)
		// Ta bắt buộc phải bóc tách từng thằng ra xử lý lẻ để cứu những thằng không trùng còn lại
		log.Println("⚠️ Lô Bulk Insert bị lỗi, chuyển sang cơ chế xử lý tách lẻ từng record...")
		w.handleBatchError(users)
		return
	}
	log.Printf("✅ Đã Bulk Insert thành công %d tài khoản vào MySQL.", len(users))
}

func (w *RegisterWorker) handleBatchError(users []models.User) {
	for _, user := range users {
		err := w.MysqlDB.Create(&user).Error
		if err != nil {
			if strings.Contains(err.Error(), "Error 1062") || strings.Contains(err.Error(), "Duplicate entry") {
				log.Printf("❌ Trùng lặp lọt lưới tại Kafka! Tiến hành vá Cache cho: %s", user.Email)
				w.compensateAerospikeCache(user.Email, user.Username)
				continue
			}
			log.Printf("Worker lỗi ghi MySQL lẻ: %v", err)
			continue
		}
	}
}

func (w *RegisterWorker) compensateAerospikeCache(email, username string) {
	keyEmail, _ := as.NewKey("fashion_e-commerce", "usersRegister", "pending:email:"+email)
	keyUsername, _ := as.NewKey("fashion_e-commerce", "usersRegister", "pending:username:"+username)

	bins := as.BinMap{
		"email":    email,
		"username": username,
		"status":   "confirmed_in_db",
	}

	_ = w.AerospikeClient.Put(w.CtxWrite, keyEmail, bins)
	_ = w.AerospikeClient.Put(w.CtxWrite, keyUsername, bins)
}
