package storage

import (
	"context"

	"github.com/ClickHouse/clickhouse-go/v2"
	"go.uber.org/zap"

	"github.com/firmanains/logpilot/services/consumer-storage/internal/domain"
)

// ClickHouseStore handles batch inserts into the logs table.
type ClickHouseStore struct {
	conn   clickhouse.Conn
	logger *zap.Logger
}

func New(addr, database string, logger *zap.Logger) (*ClickHouseStore, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{Database: database},
	})
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(context.Background()); err != nil {
		return nil, err
	}
	return &ClickHouseStore{conn: conn, logger: logger}, nil
}

// EnsureSchema creates the logs table if it does not exist.
// Safe to call on every startup.
func (s *ClickHouseStore) EnsureSchema(ctx context.Context) error {
	return s.conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS logs (
			project_id  String,
			service     String,
			level       String,
			message     String,
			trace_id    String,
			host        String,
			timestamp   DateTime64(3),
			ingested_at DateTime64(3)
		) ENGINE = MergeTree()
		PARTITION BY toYYYYMM(timestamp)
		ORDER BY (project_id, timestamp)
		TTL timestamp + INTERVAL 30 DAY
	`)
}

// BatchInsert writes a slice of logs in a single INSERT statement.
func (s *ClickHouseStore) BatchInsert(ctx context.Context, logs []domain.EnrichedLog) error {
	batch, err := s.conn.PrepareBatch(ctx,
		`INSERT INTO logs (project_id, service, level, message, trace_id, host, timestamp, ingested_at)`,
	)
	if err != nil {
		return err
	}

	for _, l := range logs {
		traceID, _ := l.Metadata["trace_id"].(string)
		host, _ := l.Metadata["host"].(string)
		if err := batch.Append(
			l.ProjectID, l.Service, l.Level, l.Message,
			traceID, host, l.Timestamp, l.IngestedAt,
		); err != nil {
			return err
		}
	}
	return batch.Send()
}

func (s *ClickHouseStore) Close() error {
	return s.conn.Close()
}
