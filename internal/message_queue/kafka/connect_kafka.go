package kafka

import (
	"time"

	"github.com/segmentio/kafka-go"
)

func CreateProducer() *kafka.Writer {
	// Khởi tạo Writer (Producer)
	writer := &kafka.Writer{
		Addr:                   kafka.TCP("kafka:9092"), // Dùng tên service trong Docker
		Topic:                  "user-registered",
		Balancer:               &kafka.LeastBytes{},   // Thuật toán chia tải cho các Partition
		BatchSize:              10000,                  // Tối ưu số lượng message gửi mỗi lần
		BatchTimeout:           50 * time.Millisecond, // Tối ưu BatchTimeout: Chờ một chút để gom đủ 2000 tin nhắn
		AllowAutoTopicCreation: true,                  // Thêm dòng này để tự động tạo topic nếu chưa có
		Async:                  true,                  // Cho phép gửi song song nhiều luồng vào Kafka
	}
	return writer
}

func CreateConsumer() *kafka.Reader {
	// Khởi tạo Reader (Consumer)
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"kafka:9092"},
		GroupID:  "golang-group-1", // Group ID rất quan trọng khi đi làm
		Topic:    "user-registered",
		MaxBytes: 10e6, // 10MB
	})
	return reader
}
