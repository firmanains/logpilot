package consumer

import (
	"context"
	"encoding/json"
	"time"

	"github.com/IBM/sarama"
	"go.uber.org/zap"

	"github.com/firmanains/logpilot/services/consumer-storage/internal/domain"
	"github.com/firmanains/logpilot/services/consumer-storage/internal/storage"
)

// Handler implements sarama.ConsumerGroupHandler with micro-batch logic.
type Handler struct {
	store           *storage.ClickHouseStore
	logger          *zap.Logger
	batchSize       int
	flushIntervalMs int
}

func NewHandler(store *storage.ClickHouseStore, logger *zap.Logger, batchSize, flushMs int) *Handler {
	return &Handler{store: store, logger: logger, batchSize: batchSize, flushIntervalMs: flushMs}
}

func (h *Handler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *Handler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *Handler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	batch := make([]domain.EnrichedLog, 0, h.batchSize)
	ticker := time.NewTicker(time.Duration(h.flushIntervalMs) * time.Millisecond)
	defer ticker.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		if err := h.store.BatchInsert(context.Background(), batch); err != nil {
			// Do NOT commit offset — Kafka will redeliver this batch
			h.logger.Error("clickhouse insert failed, will retry",
				zap.Int("batch_size", len(batch)), zap.Error(err))
			time.Sleep(5 * time.Second)
			return
		}
		h.logger.Info("batch inserted", zap.Int("count", len(batch)))
		session.Commit()
		batch = batch[:0]
	}

	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				flush()
				return nil
			}
			var log domain.EnrichedLog
			if err := json.Unmarshal(msg.Value, &log); err != nil {
				h.logger.Warn("unmarshal failed, skipping", zap.Error(err))
				session.MarkMessage(msg, "")
				continue
			}
			batch = append(batch, log)
			session.MarkMessage(msg, "")
			if len(batch) >= h.batchSize {
				flush()
			}

		case <-ticker.C:
			flush()

		case <-session.Context().Done():
			flush()
			return nil
		}
	}
}
