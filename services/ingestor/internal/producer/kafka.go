package producer

import (
	"encoding/json"

	"github.com/IBM/sarama"
	"github.com/firmanains/logpilot/services/ingestor/internal/domain"
	"go.uber.org/zap"
)

type KafkaProducer struct {
	producer sarama.AsyncProducer
	topic    string
	logger   *zap.Logger
}

func NewKafkaProducer(topic string, logger *zap.Logger, brokers []string) (*KafkaProducer, error) {
	cfg := sarama.NewConfig()
	cfg.Producer.Partitioner = sarama.NewHashPartitioner


	producer, err := sarama.NewAsyncProducer(brokers, cfg)
	if err != nil {
		logger.Fatal("failed to init producer", zap.Error(err))
		return nil, err
	}

	go func() {
		for err := range producer.Errors() {
			logger.Error("error while sending message:", zap.Error(err))
		}
	}()

	return &KafkaProducer{
		topic:    topic,
		logger:   logger,
		producer: producer,
	}, nil
}

func (kp *KafkaProducer) Publish(log *domain.EnrichedLog) error {
	byteLog, err := json.Marshal(log)
	if err != nil {
		kp.logger.Error("failed to encode Log", zap.Error(err))
		return err
	}

	msg := sarama.ProducerMessage{
		Topic: kp.topic,
		Key:   sarama.StringEncoder(log.ProjectID),
		Value: sarama.ByteEncoder(byteLog),
	}

	kp.producer.Input() <- &msg
	return nil
}
