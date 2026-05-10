package repositoryKafkaProducer

import (
	"fmt"
	"encoding/json"
	"github.com/DVT1609/fashion_e-commerce.git/internal/models"
	"context"
	"github.com/segmentio/kafka-go"
)

type UserKafkaProducerRepository struct {
	Writer *kafka.Writer
}

func NewUserKafkaProducerRepository(writer *kafka.Writer) *UserKafkaProducerRepository {
	return &UserKafkaProducerRepository{
		Writer: writer,
	}
}

func (repo *UserKafkaProducerRepository) ProduceRegisterMessage(userModel models.User) error {
	// 1. Serialize dữ liệu user thành JSON
	message, err := json.Marshal(userModel)
	if err != nil {
		return err
	}

	// 2. Tạo Key bằng cách nối Email và Username
	// Sử dụng fmt.Sprintf để nối chuỗi: "email-username"
	kafkaKey := fmt.Sprintf("%s-%s", userModel.Email, userModel.Username)

	// 3. Gửi tin nhắn kèm theo Key đã nối
	return repo.Writer.WriteMessages(context.Background(), kafka.Message{
		Key:  []byte(kafkaKey), // Sử dụng key đã tạo để phân vùng dữ liệu
		Value: message,
	})
}