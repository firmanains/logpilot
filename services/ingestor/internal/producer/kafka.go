package producer

import (
	"encoding/json"

	"github.com/IBM/sarama"
	"go.uber.org/zap"

	"github.com/firmanains/logpilot/services/ingestor/internal/domain"
)

// KafkaProducer wraps a sarama async producer.
type KafkaProducer struct {
	producer sarama.AsyncProducer
	topic    string
	logger   *zap.Logger
}

func New(brokers []string, topic string, logger *zap.Logger) (*KafkaProducer, error) {
	cfg := sarama.NewConfig()
	cfg.Producer.RequiredAcks = sarama.WaitForLocal
	cfg.Producer.Compression = sarama.CompressionSnappy
	cfg.Producer.Return.Successes = false
	cfg.Producer.Return.Errors = true

	p, err := sarama.NewAsyncProducer(brokers, cfg)
	if err != nil {
		return nil, err
	}

	kp := &KafkaProducer{producer: p, topic: topic, logger: logger}
	go kp.drainErrors()
	return kp, nil
}

// Publish serializes and sends a log to Kafka, partitioned by project_id.
func (kp *KafkaProducer) Publish(log domain.EnrichedLog) error {
	payload, err := json.Marshal(log)
	if err != nil {
		return err
	}
	kp.producer.Input() <- &sarama.ProducerMessage{
		Topic: kp.topic,
		Key:   sarama.StringEncoder(log.ProjectID),
		Value: sarama.ByteEncoder(payload),
	}
	return nil
}

func (kp *KafkaProducer) Close() error {
	return kp.producer.Close()
}

func (kp *KafkaProducer) drainErrors() {
	for err := range kp.producer.Errors() {
		kp.logger.Error("kafka producer error",
			zap.String("topic", err.Msg.Topic),
			zap.Error(err.Err),
		)
	}
}
