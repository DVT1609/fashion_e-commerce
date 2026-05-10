package repositorykafkaconsumer

import (
	"github.com/segmentio/kafka-go"
)

type UserKafkaConsumerRepository struct {
	Reader *kafka.Reader
}