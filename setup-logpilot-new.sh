#!/bin/bash
# ============================================================
# LogPilot — Clean Setup + Boilerplate Generator
# Run from inside your LOGPILOT root folder
# Usage: bash setup-logpilot.sh your-github-username
# ============================================================
set -e

GITHUB_USERNAME=${1:-"your-username"}
ROOT=$(pwd)

echo ""
echo "╔══════════════════════════════════════════╗"
echo "║   LogPilot — Project Setup & Boilerplate ║"
echo "╚══════════════════════════════════════════╝"
echo ""
echo "📍 Working directory: $ROOT"
echo "👤 GitHub username  : $GITHUB_USERNAME"
echo ""

# Helper: replace $GITHUB_USERNAME placeholder in a file (macOS + Linux safe)
replace_username() {
  perl -pi -e "s/GITHUB_USERNAME_PLACEHOLDER/$GITHUB_USERNAME/g" "$1"
}

# ─── CLEANUP ────────────────────────────────────────────────
echo "🧹 Cleaning up old artifacts..."
rm -rf logpilot init-project.sh setup-logpilot.sh
echo "✅ Cleaned"

# ═══════════════════════════════════════════════════════════
# FOLDER STRUCTURE
# ═══════════════════════════════════════════════════════════
echo ""
echo "📁 Creating folder structure..."

mkdir -p services/ingestor/{cmd,internal/{handler,middleware,enricher,producer,config,domain},pkg}
mkdir -p services/consumer-storage/{cmd,internal/{consumer,storage,config,domain},pkg}
mkdir -p services/consumer-alert/{cmd,internal/{consumer,evaluator,counter,producer,config,domain},pkg}
mkdir -p services/alert-dispatcher/{cmd,internal/{handler,destinations,config,domain},pkg}
mkdir -p demo/dummy-app/{cmd,internal/{handler,logger}}
mkdir -p demo/log-generator/cmd
mkdir -p deploy/docker-compose
mkdir -p config/{grafana,alertmanager,otel-collector}
mkdir -p .github/workflows

echo "✅ Folder structure created"

# ═══════════════════════════════════════════════════════════
# GO MODULES & DEPENDENCIES
# ═══════════════════════════════════════════════════════════
echo ""
echo "🐹 Initializing Go modules..."

init_go_service() {
  local path=$1
  local name=$2
  shift 2
  local deps=("$@")
  echo "  → $name"
  cd "$ROOT/$path"
  go mod init "github.com/$GITHUB_USERNAME/logpilot/$name"
  if [ ${#deps[@]} -gt 0 ]; then
    go get "${deps[@]}"
  fi
  go mod tidy
  cd "$ROOT"
}

init_go_service "services/ingestor" "services/ingestor" \
  "github.com/gofiber/fiber/v2" \
  "github.com/redis/go-redis/v9" \
  "github.com/IBM/sarama" \
  "github.com/joho/godotenv" \
  "go.uber.org/zap"

init_go_service "services/consumer-storage" "services/consumer-storage" \
  "github.com/IBM/sarama" \
  "github.com/ClickHouse/clickhouse-go/v2" \
  "github.com/joho/godotenv" \
  "go.uber.org/zap"

init_go_service "services/consumer-alert" "services/consumer-alert" \
  "github.com/IBM/sarama" \
  "github.com/redis/go-redis/v9" \
  "github.com/joho/godotenv" \
  "go.uber.org/zap"

init_go_service "services/alert-dispatcher" "services/alert-dispatcher" \
  "github.com/gofiber/fiber/v2" \
  "github.com/joho/godotenv" \
  "go.uber.org/zap"

init_go_service "demo/dummy-app" "demo/dummy-app" \
  "github.com/gofiber/fiber/v2" \
  "github.com/joho/godotenv"

init_go_service "demo/log-generator" "demo/log-generator" \
  "github.com/joho/godotenv"

echo "✅ Go modules initialized"

# ═══════════════════════════════════════════════════════════
# BOILERPLATE: INGESTOR SERVICE
# ═══════════════════════════════════════════════════════════
echo ""
echo "📝 Writing ingestor boilerplate..."

cat > services/ingestor/internal/domain/log.go << 'EOF'
package domain

import "time"

// LogLevel represents allowed log severity levels.
type LogLevel string

const (
	LevelDebug LogLevel = "DEBUG"
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
	LevelFatal LogLevel = "FATAL"
)

// ValidLevels for O(1) lookup during validation.
var ValidLevels = map[LogLevel]struct{}{
	LevelDebug: {}, LevelInfo: {}, LevelWarn: {},
	LevelError: {}, LevelFatal: {},
}

// IngestRequest is the raw payload received from the client.
type IngestRequest struct {
	Level     LogLevel               `json:"level"`
	Message   string                 `json:"message"`
	Service   string                 `json:"service"`
	Timestamp string                 `json:"timestamp"` // ISO 8601 string
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// EnrichedLog is the payload published to Kafka after enrichment.
type EnrichedLog struct {
	Level      LogLevel               `json:"level"`
	Message    string                 `json:"message"`
	Service    string                 `json:"service"`
	Timestamp  time.Time              `json:"timestamp"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	ProjectID  string                 `json:"project_id"`
	IngestedAt time.Time              `json:"ingested_at"`
	IngestorID string                 `json:"ingestor_id"`
}
EOF

cat > services/ingestor/internal/config/config.go << 'EOF'
package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config holds all runtime configuration for the ingestor.
type Config struct {
	Port         string
	RedisAddr    string
	KafkaBrokers []string
	KafkaTopic   string
	RateLimitMax int64
	IngestorID   string
}

// Load reads configuration from environment variables.
func Load() *Config {
	_ = godotenv.Load()
	return &Config{
		Port:         getEnv("INGESTOR_PORT", "8080"),
		RedisAddr:    getEnv("REDIS_ADDR", "localhost:6379"),
		KafkaBrokers: []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
		KafkaTopic:   getEnv("KAFKA_TOPIC_RAW_LOGS", "raw-logs"),
		RateLimitMax: 10000,
		IngestorID:   getHostname(),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getHostname() string {
	h, _ := os.Hostname()
	return h
}
EOF

cat > services/ingestor/internal/middleware/auth.go << 'EOF'
package middleware

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

const LocalKeyProjectID = "projectID"

// Auth validates X-API-Key header against Redis.
// Redis key pattern: api_key:{plain_key} → project_id
func Auth(rdb *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := c.Get("X-API-Key")
		if key == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing X-API-Key header",
			})
		}

		projectID, err := rdb.Get(context.Background(), fmt.Sprintf("api_key:%s", key)).Result()
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid api key",
			})
		}

		c.Locals(LocalKeyProjectID, projectID)
		return c.Next()
	}
}
EOF

cat > services/ingestor/internal/middleware/ratelimit.go << 'EOF'
package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

// RateLimit enforces per-project request limits.
// Window: 60 seconds. Redis key: rate:{project_id}
func RateLimit(rdb *redis.Client, maxReqPerMinute int64) fiber.Handler {
	return func(c *fiber.Ctx) error {
		projectID, _ := c.Locals(LocalKeyProjectID).(string)
		key := fmt.Sprintf("rate:%s", projectID)
		ctx := context.Background()

		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			return c.Next() // fail open if Redis is unavailable
		}
		if count == 1 {
			rdb.Expire(ctx, key, 60*time.Second)
		}
		if count > maxReqPerMinute {
			c.Set("Retry-After", "60")
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       "rate limit exceeded",
				"retry_after": 60,
			})
		}
		return c.Next()
	}
}
EOF

cat > services/ingestor/internal/enricher/enricher.go << 'EOF'
package enricher

import (
	"time"

	"github.com/GITHUB_USERNAME_PLACEHOLDER/logpilot/services/ingestor/internal/domain"
)

// Enricher adds server-side fields to incoming log requests.
type Enricher struct {
	ingestorID string
}

func New(ingestorID string) *Enricher {
	return &Enricher{ingestorID: ingestorID}
}

// Enrich creates an EnrichedLog from a raw request.
func (e *Enricher) Enrich(req domain.IngestRequest, projectID string) domain.EnrichedLog {
	ts, _ := time.Parse(time.RFC3339, req.Timestamp)
	return domain.EnrichedLog{
		Level:      req.Level,
		Message:    req.Message,
		Service:    req.Service,
		Timestamp:  ts,
		Metadata:   req.Metadata,
		ProjectID:  projectID,
		IngestedAt: time.Now().UTC(),
		IngestorID: e.ingestorID,
	}
}
EOF
replace_username "services/ingestor/internal/enricher/enricher.go"

cat > services/ingestor/internal/producer/kafka.go << 'EOF'
package producer

import (
	"encoding/json"

	"github.com/IBM/sarama"
	"go.uber.org/zap"

	"github.com/GITHUB_USERNAME_PLACEHOLDER/logpilot/services/ingestor/internal/domain"
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
EOF
replace_username "services/ingestor/internal/producer/kafka.go"

cat > services/ingestor/internal/handler/ingest.go << 'EOF'
package handler

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/GITHUB_USERNAME_PLACEHOLDER/logpilot/services/ingestor/internal/domain"
	"github.com/GITHUB_USERNAME_PLACEHOLDER/logpilot/services/ingestor/internal/enricher"
	"github.com/GITHUB_USERNAME_PLACEHOLDER/logpilot/services/ingestor/internal/middleware"
	"github.com/GITHUB_USERNAME_PLACEHOLDER/logpilot/services/ingestor/internal/producer"
)

// IngestHandler handles POST /v1/ingest.
type IngestHandler struct {
	enricher *enricher.Enricher
	producer *producer.KafkaProducer
	logger   *zap.Logger
}

func NewIngestHandler(e *enricher.Enricher, p *producer.KafkaProducer, log *zap.Logger) *IngestHandler {
	return &IngestHandler{enricher: e, producer: p, logger: log}
}

func (h *IngestHandler) Handle(c *fiber.Ctx) error {
	var req domain.IngestRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid json body",
		})
	}

	if errs := validate(req); len(errs) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "validation failed",
			"fields": errs,
		})
	}

	projectID, _ := c.Locals(middleware.LocalKeyProjectID).(string)
	enriched := h.enricher.Enrich(req, projectID)

	if err := h.producer.Publish(enriched); err != nil {
		h.logger.Error("kafka publish failed", zap.Error(err))
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "failed to process log, please retry",
		})
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"status": "accepted"})
}

func validate(req domain.IngestRequest) map[string]string {
	errs := map[string]string{}
	if req.Message == "" {
		errs["message"] = "required"
	}
	if req.Service == "" {
		errs["service"] = "required"
	}
	if _, ok := domain.ValidLevels[req.Level]; !ok {
		errs["level"] = fmt.Sprintf("must be one of DEBUG|INFO|WARN|ERROR|FATAL, got %q", req.Level)
	}
	if _, err := time.Parse(time.RFC3339, req.Timestamp); err != nil {
		errs["timestamp"] = "must be ISO 8601 format, e.g. 2026-04-17T05:12:00Z"
	}
	return errs
}
EOF
replace_username "services/ingestor/internal/handler/ingest.go"

cat > services/ingestor/cmd/main.go << 'EOF'
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/GITHUB_USERNAME_PLACEHOLDER/logpilot/services/ingestor/internal/config"
	"github.com/GITHUB_USERNAME_PLACEHOLDER/logpilot/services/ingestor/internal/enricher"
	"github.com/GITHUB_USERNAME_PLACEHOLDER/logpilot/services/ingestor/internal/handler"
	"github.com/GITHUB_USERNAME_PLACEHOLDER/logpilot/services/ingestor/internal/middleware"
	"github.com/GITHUB_USERNAME_PLACEHOLDER/logpilot/services/ingestor/internal/producer"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg := config.Load()
	logger.Info("ingestor starting", zap.String("port", cfg.Port))

	// ── Redis ────────────────────────────────────────────────
	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		logger.Fatal("redis connection failed", zap.Error(err))
	}
	logger.Info("redis connected", zap.String("addr", cfg.RedisAddr))

	// ── Kafka Producer ───────────────────────────────────────
	kafkaProducer, err := producer.New(cfg.KafkaBrokers, cfg.KafkaTopic, logger)
	if err != nil {
		logger.Fatal("kafka producer init failed", zap.Error(err))
	}
	defer kafkaProducer.Close()
	logger.Info("kafka producer connected")

	// ── Wire Dependencies ────────────────────────────────────
	logEnricher := enricher.New(cfg.IngestorID)
	ingestHandler := handler.NewIngestHandler(logEnricher, kafkaProducer, logger)

	// ── HTTP Server ──────────────────────────────────────────
	app := fiber.New(fiber.Config{
		AppName:      "LogPilot Ingestor",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "ingestor_id": cfg.IngestorID})
	})

	v1 := app.Group("/v1")
	v1.Post("/ingest",
		middleware.Auth(rdb),
		middleware.RateLimit(rdb, cfg.RateLimitMax),
		ingestHandler.Handle,
	)

	// ── Graceful Shutdown ────────────────────────────────────
	go func() {
		if err := app.Listen(":" + cfg.Port); err != nil {
			logger.Fatal("server error", zap.Error(err))
		}
	}()

	logger.Info("ingestor ready", zap.String("port", cfg.Port))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down...")
	app.Shutdown()
}
EOF
replace_username "services/ingestor/cmd/main.go"

cat > services/ingestor/.env << 'EOF'
INGESTOR_PORT=8080
REDIS_ADDR=localhost:6379
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC_RAW_LOGS=raw-logs
EOF

echo "  ✅ ingestor"

# ═══════════════════════════════════════════════════════════
# BOILERPLATE: CONSUMER STORAGE
# ═══════════════════════════════════════════════════════════
echo "📝 Writing consumer-storage boilerplate..."

cat > services/consumer-storage/internal/domain/log.go << 'EOF'
package domain

import "time"

// EnrichedLog mirrors the ingestor output — what we read from Kafka.
type EnrichedLog struct {
	Level      string                 `json:"level"`
	Message    string                 `json:"message"`
	Service    string                 `json:"service"`
	Timestamp  time.Time              `json:"timestamp"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	ProjectID  string                 `json:"project_id"`
	IngestedAt time.Time              `json:"ingested_at"`
	IngestorID string                 `json:"ingestor_id"`
}
EOF

cat > services/consumer-storage/internal/config/config.go << 'EOF'
package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	KafkaBrokers    []string
	KafkaTopic      string
	KafkaGroupID    string
	ClickHouseAddr  string
	ClickHouseDB    string
	BatchSize       int
	FlushIntervalMs int
}

func Load() *Config {
	_ = godotenv.Load()
	return &Config{
		KafkaBrokers:    []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
		KafkaTopic:      getEnv("KAFKA_TOPIC_RAW_LOGS", "raw-logs"),
		KafkaGroupID:    "storage-workers",
		ClickHouseAddr:  getEnv("CLICKHOUSE_ADDR", "localhost:9000"),
		ClickHouseDB:    getEnv("CLICKHOUSE_DB", "logpilot"),
		BatchSize:       500,
		FlushIntervalMs: 1000,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
EOF

cat > services/consumer-storage/internal/storage/clickhouse.go << 'EOF'
package storage

import (
	"context"

	"github.com/ClickHouse/clickhouse-go/v2"
	"go.uber.org/zap"

	"github.com/GITHUB_USERNAME_PLACEHOLDER/logpilot/services/consumer-storage/internal/domain"
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
EOF
replace_username "services/consumer-storage/internal/storage/clickhouse.go"

cat > services/consumer-storage/internal/consumer/consumer.go << 'EOF'
package consumer

import (
	"context"
	"encoding/json"
	"time"

	"github.com/IBM/sarama"
	"go.uber.org/zap"

	"github.com/GITHUB_USERNAME_PLACEHOLDER/logpilot/services/consumer-storage/internal/domain"
	"github.com/GITHUB_USERNAME_PLACEHOLDER/logpilot/services/consumer-storage/internal/storage"
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
EOF
replace_username "services/consumer-storage/internal/consumer/consumer.go"

cat > services/consumer-storage/cmd/main.go << 'EOF'
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/IBM/sarama"
	"go.uber.org/zap"

	"github.com/GITHUB_USERNAME_PLACEHOLDER/logpilot/services/consumer-storage/internal/config"
	"github.com/GITHUB_USERNAME_PLACEHOLDER/logpilot/services/consumer-storage/internal/consumer"
	"github.com/GITHUB_USERNAME_PLACEHOLDER/logpilot/services/consumer-storage/internal/storage"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg := config.Load()

	store, err := storage.New(cfg.ClickHouseAddr, cfg.ClickHouseDB, logger)
	if err != nil {
		logger.Fatal("clickhouse connection failed", zap.Error(err))
	}
	defer store.Close()

	if err := store.EnsureSchema(context.Background()); err != nil {
		logger.Fatal("schema creation failed", zap.Error(err))
	}
	logger.Info("clickhouse ready")

	kafkaCfg := sarama.NewConfig()
	kafkaCfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	kafkaCfg.Consumer.Offsets.AutoCommit.Enable = false

	cg, err := sarama.NewConsumerGroup(cfg.KafkaBrokers, cfg.KafkaGroupID, kafkaCfg)
	if err != nil {
		logger.Fatal("kafka consumer group init failed", zap.Error(err))
	}
	defer cg.Close()

	handler := consumer.NewHandler(store, logger, cfg.BatchSize, cfg.FlushIntervalMs)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for {
			if err := cg.Consume(ctx, []string{cfg.KafkaTopic}, handler); err != nil {
				logger.Error("consumer group error", zap.Error(err))
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()

	logger.Info("consumer-storage running",
		zap.String("group", cfg.KafkaGroupID),
		zap.String("topic", cfg.KafkaTopic),
	)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down...")
	cancel()
}
EOF
replace_username "services/consumer-storage/cmd/main.go"

cat > services/consumer-storage/.env << 'EOF'
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC_RAW_LOGS=raw-logs
CLICKHOUSE_ADDR=localhost:9000
CLICKHOUSE_DB=logpilot
EOF

echo "  ✅ consumer-storage"

# ═══════════════════════════════════════════════════════════
# BOILERPLATE: CONSUMER ALERT
# ═══════════════════════════════════════════════════════════
echo "📝 Writing consumer-alert boilerplate..."

cat > services/consumer-alert/internal/domain/alert.go << 'EOF'
package domain

import "time"

// AlertRule defines when an alert should fire.
type AlertRule struct {
	ID              string        `json:"id"`
	ProjectID       string        `json:"project_id"`
	Name            string        `json:"name"`
	Condition       RuleCondition `json:"condition"`
	Threshold       int64         `json:"threshold"`
	WindowSeconds   int           `json:"window_seconds"`
	CooldownSeconds int           `json:"cooldown_seconds"`
	IsActive        bool          `json:"is_active"`
}

// RuleCondition specifies which log fields must match.
type RuleCondition struct {
	Level   string `json:"level"`
	Service string `json:"service,omitempty"` // empty = match any service
}

// EnrichedLog mirrors what we consume from Kafka.
type EnrichedLog struct {
	Level     string `json:"level"`
	Message   string `json:"message"`
	Service   string `json:"service"`
	ProjectID string `json:"project_id"`
}

// AlertEvent is published to Kafka when a rule fires.
type AlertEvent struct {
	RuleID      string    `json:"rule_id"`
	ProjectID   string    `json:"project_id"`
	RuleName    string    `json:"rule_name"`
	TriggeredAt time.Time `json:"triggered_at"`
	LogSample   string    `json:"log_sample"`
}
EOF

cat > services/consumer-alert/internal/evaluator/evaluator.go << 'EOF'
package evaluator

import "github.com/GITHUB_USERNAME_PLACEHOLDER/logpilot/services/consumer-alert/internal/domain"

// Evaluate returns all active rules that match the given log.
func Evaluate(log domain.EnrichedLog, rules []domain.AlertRule) []domain.AlertRule {
	var matched []domain.AlertRule
	for _, rule := range rules {
		if !rule.IsActive {
			continue
		}
		if rule.ProjectID != log.ProjectID {
			continue
		}
		if rule.Condition.Level != log.Level {
			continue
		}
		if rule.Condition.Service != "" && rule.Condition.Service != log.Service {
			continue
		}
		matched = append(matched, rule)
	}
	return matched
}
EOF
replace_username "services/consumer-alert/internal/evaluator/evaluator.go"

cat > services/consumer-alert/internal/evaluator/evaluator_test.go << 'EOF'
package evaluator_test

import (
	"testing"

	"github.com/GITHUB_USERNAME_PLACEHOLDER/logpilot/services/consumer-alert/internal/domain"
	"github.com/GITHUB_USERNAME_PLACEHOLDER/logpilot/services/consumer-alert/internal/evaluator"
)

func TestEvaluate_MatchesCorrectProjectAndLevel(t *testing.T) {
	rules := []domain.AlertRule{
		{ID: "r1", ProjectID: "proj-a", Condition: domain.RuleCondition{Level: "ERROR"}, IsActive: true, Threshold: 10, WindowSeconds: 300, CooldownSeconds: 600},
		{ID: "r2", ProjectID: "proj-b", Condition: domain.RuleCondition{Level: "ERROR"}, IsActive: true, Threshold: 5, WindowSeconds: 60, CooldownSeconds: 120},
	}
	log := domain.EnrichedLog{ProjectID: "proj-a", Level: "ERROR", Service: "api"}
	matched := evaluator.Evaluate(log, rules)
	if len(matched) != 1 || matched[0].ID != "r1" {
		t.Errorf("expected rule r1 to match, got %v", matched)
	}
}

func TestEvaluate_InactiveRuleSkipped(t *testing.T) {
	rules := []domain.AlertRule{
		{ID: "r1", ProjectID: "proj-a", Condition: domain.RuleCondition{Level: "ERROR"}, IsActive: false},
	}
	log := domain.EnrichedLog{ProjectID: "proj-a", Level: "ERROR"}
	if matched := evaluator.Evaluate(log, rules); len(matched) != 0 {
		t.Error("inactive rule should not match")
	}
}

func TestEvaluate_ServiceFilterApplied(t *testing.T) {
	rules := []domain.AlertRule{
		{ID: "r1", ProjectID: "proj-a", Condition: domain.RuleCondition{Level: "ERROR", Service: "payment-svc"}, IsActive: true},
	}
	logNoMatch := domain.EnrichedLog{ProjectID: "proj-a", Level: "ERROR", Service: "api-gateway"}
	if matched := evaluator.Evaluate(logNoMatch, rules); len(matched) != 0 {
		t.Error("should not match different service")
	}
	logMatch := domain.EnrichedLog{ProjectID: "proj-a", Level: "ERROR", Service: "payment-svc"}
	if matched := evaluator.Evaluate(logMatch, rules); len(matched) != 1 {
		t.Error("should match exact service")
	}
}
EOF
replace_username "services/consumer-alert/internal/evaluator/evaluator_test.go"

cat > services/consumer-alert/internal/counter/sliding_window.go << 'EOF'
package counter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// SlidingWindow implements a Redis-based counter for alert rate evaluation.
type SlidingWindow struct {
	rdb *redis.Client
}

func New(rdb *redis.Client) *SlidingWindow {
	return &SlidingWindow{rdb: rdb}
}

// Increment increments the counter for a rule and returns the new count.
// TTL is set to windowSeconds on first increment.
func (sw *SlidingWindow) Increment(ctx context.Context, ruleID, projectID string, windowSeconds int) (int64, error) {
	key := fmt.Sprintf("alert_counter:%s:%s", ruleID, projectID)
	count, err := sw.rdb.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if count == 1 {
		sw.rdb.Expire(ctx, key, time.Duration(windowSeconds)*time.Second)
	}
	return count, nil
}

// HasCooldown returns true if a cooldown is active for this rule.
func (sw *SlidingWindow) HasCooldown(ctx context.Context, ruleID, projectID string) (bool, error) {
	key := fmt.Sprintf("cooldown:%s:%s", ruleID, projectID)
	n, err := sw.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// SetCooldown activates a cooldown for the given rule.
func (sw *SlidingWindow) SetCooldown(ctx context.Context, ruleID, projectID string, cooldownSeconds int) error {
	key := fmt.Sprintf("cooldown:%s:%s", ruleID, projectID)
	return sw.rdb.Set(ctx, key, 1, time.Duration(cooldownSeconds)*time.Second).Err()
}
EOF

cat > services/consumer-alert/internal/config/config.go << 'EOF'
package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	KafkaBrokers       []string
	KafkaTopicRawLogs  string
	KafkaTopicAlerts   string
	KafkaGroupID       string
	RedisAddr          string
	LaravelAPIURL      string
}

func Load() *Config {
	_ = godotenv.Load()
	return &Config{
		KafkaBrokers:      []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
		KafkaTopicRawLogs: getEnv("KAFKA_TOPIC_RAW_LOGS", "raw-logs"),
		KafkaTopicAlerts:  getEnv("KAFKA_TOPIC_ALERT_EVENTS", "alert-events"),
		KafkaGroupID:      "alert-evaluators",
		RedisAddr:         getEnv("REDIS_ADDR", "localhost:6379"),
		LaravelAPIURL:     getEnv("LARAVEL_API_URL", "http://localhost:8000"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
EOF

cat > services/consumer-alert/cmd/main.go << 'EOF'
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/GITHUB_USERNAME_PLACEHOLDER/logpilot/services/consumer-alert/internal/config"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg := config.Load()

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		logger.Fatal("redis connection failed", zap.Error(err))
	}
	logger.Info("redis connected")

	// TODO (Chunk 3.2): Initialize in-memory rule cache
	// TODO (Chunk 3.5): Initialize Kafka alert-events producer
	// TODO (Chunk 3.6): Start Kafka consumer group "alert-evaluators"

	logger.Info("consumer-alert running",
		zap.String("group", cfg.KafkaGroupID),
		zap.String("topic", cfg.KafkaTopicRawLogs),
	)

	_ = rdb

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down...")
}
EOF
replace_username "services/consumer-alert/cmd/main.go"

cat > services/consumer-alert/.env << 'EOF'
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC_RAW_LOGS=raw-logs
KAFKA_TOPIC_ALERT_EVENTS=alert-events
REDIS_ADDR=localhost:6379
LARAVEL_API_URL=http://localhost:8000
EOF

echo "  ✅ consumer-alert"

# ═══════════════════════════════════════════════════════════
# BOILERPLATE: ALERT DISPATCHER
# ═══════════════════════════════════════════════════════════
echo "📝 Writing alert-dispatcher boilerplate..."

cat > services/alert-dispatcher/internal/domain/alert.go << 'EOF'
package domain

import "time"

// AlertEvent received from Alertmanager webhook.
type AlertEvent struct {
	RuleID      string    `json:"rule_id"`
	ProjectID   string    `json:"project_id"`
	RuleName    string    `json:"rule_name"`
	TriggeredAt time.Time `json:"triggered_at"`
	LogSample   string    `json:"log_sample"`
}

// NotifConfig holds notification destinations for a project.
type NotifConfig struct {
	ClickUpListID     string   `json:"clickup_list_id"`
	ClickUpAssigneeID string   `json:"clickup_assignee_id"`
	EmailRecipients   []string `json:"email_recipients"`
	SlackWebhookURL   string   `json:"slack_webhook_url"`
}
EOF

cat > services/alert-dispatcher/internal/destinations/clickup.go << 'EOF'
package destinations

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/GITHUB_USERNAME_PLACEHOLDER/logpilot/services/alert-dispatcher/internal/domain"
)

const clickupBaseURL = "https://api.clickup.com/api/v2"

// ClickUp creates a task in ClickUp for the given alert event.
func ClickUp(apiKey string, event domain.AlertEvent, cfg domain.NotifConfig) error {
	if cfg.ClickUpListID == "" {
		return nil // not configured, skip
	}

	payload := map[string]interface{}{
		"name":     fmt.Sprintf("[ALERT] %s — %s", event.RuleName, event.ProjectID),
		"priority": 1, // Urgent
		"description": fmt.Sprintf(
			"Alert triggered at %s\n\nProject: %s\nRule: %s\n\nLog sample:\n%s",
			event.TriggeredAt.Format(time.RFC3339),
			event.ProjectID, event.RuleName, event.LogSample,
		),
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST",
		fmt.Sprintf("%s/list/%s/task", clickupBaseURL, cfg.ClickUpListID),
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("clickup returned status %d", resp.StatusCode)
	}
	return nil
}
EOF
replace_username "services/alert-dispatcher/internal/destinations/clickup.go"

cat > services/alert-dispatcher/internal/destinations/sendgrid.go << 'EOF'
package destinations

import (
	"fmt"
	"time"

	"github.com/GITHUB_USERNAME_PLACEHOLDER/logpilot/services/alert-dispatcher/internal/domain"
)

// SendGrid sends an alert notification via email.
// TODO (Chunk 3.10): implement using sendgrid-go SDK
func SendGrid(apiKey string, event domain.AlertEvent, cfg domain.NotifConfig) error {
	if len(cfg.EmailRecipients) == 0 {
		return nil
	}
	// Placeholder
	_ = fmt.Sprintf("[LogPilot Alert] %s triggered at %s for project %s. Log: %s",
		event.RuleName, event.TriggeredAt.Format(time.RFC3339), event.ProjectID, event.LogSample)
	return nil
}
EOF
replace_username "services/alert-dispatcher/internal/destinations/sendgrid.go"

cat > services/alert-dispatcher/internal/handler/webhook.go << 'EOF'
package handler

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/GITHUB_USERNAME_PLACEHOLDER/logpilot/services/alert-dispatcher/internal/destinations"
	"github.com/GITHUB_USERNAME_PLACEHOLDER/logpilot/services/alert-dispatcher/internal/domain"
)

// WebhookHandler receives Alertmanager webhooks and fans out notifications.
type WebhookHandler struct {
	clickupKey  string
	sendgridKey string
	logger      *zap.Logger
}

func NewWebhookHandler(clickupKey, sendgridKey string, logger *zap.Logger) *WebhookHandler {
	return &WebhookHandler{clickupKey: clickupKey, sendgridKey: sendgridKey, logger: logger}
}

func (h *WebhookHandler) Handle(c *fiber.Ctx) error {
	var event domain.AlertEvent
	if err := c.BodyParser(&event); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}

	// TODO (Chunk 4.10): fetch real NotifConfig from Laravel API by event.ProjectID
	cfg := domain.NotifConfig{}

	var wg sync.WaitGroup
	dispatch := func(name string, fn func() error) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			done := make(chan error, 1)
			go func() { done <- fn() }()
			select {
			case err := <-done:
				if err != nil {
					h.logger.Error("dispatch failed", zap.String("dest", name), zap.Error(err))
				} else {
					h.logger.Info("dispatched", zap.String("dest", name))
				}
			case <-time.After(10 * time.Second):
				h.logger.Warn("dispatch timeout", zap.String("dest", name))
			}
		}()
	}

	dispatch("clickup", func() error { return destinations.ClickUp(h.clickupKey, event, cfg) })
	dispatch("sendgrid", func() error { return destinations.SendGrid(h.sendgridKey, event, cfg) })
	wg.Wait()

	return c.JSON(fiber.Map{"status": "dispatched"})
}
EOF
replace_username "services/alert-dispatcher/internal/handler/webhook.go"

cat > services/alert-dispatcher/internal/config/config.go << 'EOF'
package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	LaravelAPIURL string
	ClickUpAPIKey string
	SendGridKey   string
}

func Load() *Config {
	_ = godotenv.Load()
	return &Config{
		Port:          getEnv("DISPATCHER_PORT", "9090"),
		LaravelAPIURL: getEnv("LARAVEL_API_URL", "http://localhost:8000"),
		ClickUpAPIKey: getEnv("CLICKUP_API_KEY", ""),
		SendGridKey:   getEnv("SENDGRID_API_KEY", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
EOF

cat > services/alert-dispatcher/cmd/main.go << 'EOF'
package main

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/GITHUB_USERNAME_PLACEHOLDER/logpilot/services/alert-dispatcher/internal/config"
	"github.com/GITHUB_USERNAME_PLACEHOLDER/logpilot/services/alert-dispatcher/internal/handler"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg := config.Load()
	webhookHandler := handler.NewWebhookHandler(cfg.ClickUpAPIKey, cfg.SendGridKey, logger)

	app := fiber.New(fiber.Config{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 15 * time.Second,
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
	app.Post("/webhook", webhookHandler.Handle)

	logger.Info("alert-dispatcher ready", zap.String("port", cfg.Port))
	app.Listen(":" + cfg.Port)
}
EOF
replace_username "services/alert-dispatcher/cmd/main.go"

cat > services/alert-dispatcher/.env << 'EOF'
DISPATCHER_PORT=9090
LARAVEL_API_URL=http://localhost:8000
CLICKUP_API_KEY=
SENDGRID_API_KEY=
EOF

echo "  ✅ alert-dispatcher"

# ═══════════════════════════════════════════════════════════
# BOILERPLATE: DEMO APPS
# ═══════════════════════════════════════════════════════════
echo "📝 Writing demo app boilerplate..."

cat > demo/dummy-app/cmd/main.go << 'EOF'
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

var (
	ingestorURL string
	apiKey      string
)

func main() {
	_ = godotenv.Load()
	ingestorURL = getEnv("INGESTOR_URL", "http://localhost:8080")
	apiKey = getEnv("API_KEY", "logpilot_changeme")

	app := fiber.New(fiber.Config{AppName: "LogPilot Dummy E-commerce"})

	app.Get("/health", func(c *fiber.Ctx) error {
		sendLog("DEBUG", "health check", "dummy-app", nil)
		return c.JSON(fiber.Map{"status": "ok"})
	})

	app.Get("/products", func(c *fiber.Ctx) error {
		sendLog("INFO", "fetch products success", "dummy-app", nil)
		return c.JSON(fiber.Map{"products": []string{"item-1", "item-2", "item-3"}})
	})

	app.Get("/products/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		if rand.Float32() < 0.2 {
			sendLog("ERROR", fmt.Sprintf("product not found: %s", id), "dummy-app", nil)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "product not found"})
		}
		sendLog("INFO", fmt.Sprintf("product fetched: %s", id), "dummy-app", nil)
		return c.JSON(fiber.Map{"id": id, "name": "Sample Product", "price": 99.99})
	})

	app.Post("/checkout", func(c *fiber.Ctx) error {
		if rand.Float32() < 0.3 {
			sendLog("ERROR", "payment gateway timeout", "dummy-app", map[string]interface{}{
				"trace_id": fmt.Sprintf("trace-%d", rand.Intn(99999)),
			})
			return c.Status(fiber.StatusGatewayTimeout).JSON(fiber.Map{"error": "payment gateway timeout"})
		}
		orderID := rand.Intn(99999)
		sendLog("INFO", fmt.Sprintf("checkout success, order %d", orderID), "dummy-app", nil)
		return c.JSON(fiber.Map{"status": "paid", "order_id": orderID})
	})

	port := getEnv("APP_PORT", "8081")
	app.Listen(":" + port)
}

func sendLog(level, message, service string, metadata map[string]interface{}) {
	payload := map[string]interface{}{
		"level":     level,
		"message":   message,
		"service":   service,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	if metadata != nil {
		payload["metadata"] = metadata
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", ingestorURL+"/v1/ingest", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)
	(&http.Client{Timeout: 3 * time.Second}).Do(req)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
EOF

cat > demo/dummy-app/.env << 'EOF'
APP_PORT=8081
INGESTOR_URL=http://localhost:8080
API_KEY=logpilot_changeme
EOF

cat > demo/log-generator/cmd/main.go << 'EOF'
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	ingestorURL := getEnv("INGESTOR_URL", "http://localhost:8080")
	apiKey := getEnv("API_KEY", "logpilot_changeme")

	log.Printf("📨 Log Generator → %s", ingestorURL)

	// Weighted level distribution: 60% INFO, 20% WARN, 15% ERROR, 5% FATAL
	levels := []string{
		"INFO", "INFO", "INFO", "INFO", "INFO", "INFO",
		"WARN", "WARN", "WARN", "WARN",
		"ERROR", "ERROR", "ERROR",
		"FATAL",
	}
	services := []string{"api-gateway", "checkout-svc", "product-svc", "payment-svc"}
	messages := map[string][]string{
		"INFO":  {"request processed", "user logged in", "cache hit", "db query ok", "session created"},
		"WARN":  {"high memory usage", "slow query detected", "cache miss", "retry attempt"},
		"ERROR": {"db connection timeout", "payment gateway timeout", "nil pointer", "disk write failed"},
		"FATAL": {"out of memory", "disk full", "core service unreachable"},
	}

	send := func(level, service, message string) {
		payload := map[string]interface{}{
			"level":     level,
			"message":   message,
			"service":   service,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"metadata":  map[string]interface{}{"trace_id": fmt.Sprintf("trace-%d", rand.Intn(99999))},
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", ingestorURL+"/v1/ingest", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", apiKey)
		(&http.Client{Timeout: 3 * time.Second}).Do(req)
	}

	normalTicker := time.NewTicker(200 * time.Millisecond) // 5 logs/sec
	spikeTicker := time.NewTicker(2 * time.Minute)

	for {
		select {
		case <-normalTicker.C:
			level := levels[rand.Intn(len(levels))]
			msgs := messages[level]
			send(level, services[rand.Intn(len(services))], msgs[rand.Intn(len(msgs))])

		case <-spikeTicker.C:
			log.Println("⚠️  Simulating error spike (20 ERRORs in 30s)...")
			go func() {
				for i := 0; i < 20; i++ {
					msgs := messages["ERROR"]
					send("ERROR", services[rand.Intn(len(services))], msgs[rand.Intn(len(msgs))])
					time.Sleep(1500 * time.Millisecond)
				}
				log.Println("✅ Error spike complete")
			}()
		}
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
EOF

cat > demo/log-generator/.env << 'EOF'
INGESTOR_URL=http://localhost:8080
API_KEY=logpilot_changeme
EOF

echo "  ✅ demo apps"

# ═══════════════════════════════════════════════════════════
# CONFIG & INFRA FILES
# ═══════════════════════════════════════════════════════════
echo "📝 Writing config and infra files..."

cat > config/alertmanager/alertmanager.yml << 'EOF'
global:
  resolve_timeout: 5m

route:
  group_by: ['alertname', 'project_id']
  group_wait: 10s
  group_interval: 30s
  repeat_interval: 1h
  receiver: 'logpilot-dispatcher'

receivers:
  - name: 'logpilot-dispatcher'
    webhook_configs:
      - url: 'http://host.docker.internal:9090/webhook'
        send_resolved: false
EOF

cat > deploy/docker-compose/docker-compose.yml << 'EOF'
version: '3.8'

networks:
  logpilot:
    driver: bridge

volumes:
  postgres_data:
  clickhouse_data:
  redis_data:
  grafana_data:

services:

  zookeeper:
    image: bitnami/zookeeper:3.8
    container_name: logpilot-zookeeper
    environment:
      - ALLOW_ANONYMOUS_LOGIN=yes
    networks:
      - logpilot

  kafka:
    image: bitnami/kafka:3.5
    container_name: logpilot-kafka
    depends_on:
      - zookeeper
    ports:
      - "9092:9092"
    environment:
      - KAFKA_CFG_ZOOKEEPER_CONNECT=zookeeper:2181
      - KAFKA_CFG_LISTENERS=PLAINTEXT://:9092
      - KAFKA_CFG_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092
      - ALLOW_PLAINTEXT_LISTENER=yes
    networks:
      - logpilot

  redis:
    image: redis:7-alpine
    container_name: logpilot-redis
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    networks:
      - logpilot

  postgres:
    image: postgres:15-alpine
    container_name: logpilot-postgres
    ports:
      - "5432:5432"
    environment:
      POSTGRES_USER: ${POSTGRES_USER:-logpilot}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-logpilot123}
      POSTGRES_DB: ${POSTGRES_DB:-logpilot}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - logpilot

  clickhouse:
    image: clickhouse/clickhouse-server:23.8
    container_name: logpilot-clickhouse
    ports:
      - "8123:8123"
      - "9000:9000"
    volumes:
      - clickhouse_data:/var/lib/clickhouse
    networks:
      - logpilot
    ulimits:
      nofile:
        soft: 262144
        hard: 262144

  grafana:
    image: grafana/grafana:10.2.0
    container_name: logpilot-grafana
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_PASSWORD:-admin}
      - GF_INSTALL_PLUGINS=grafana-clickhouse-datasource
    volumes:
      - grafana_data:/var/lib/grafana
    networks:
      - logpilot

  alertmanager:
    image: prom/alertmanager:v0.26.0
    container_name: logpilot-alertmanager
    ports:
      - "9093:9093"
    volumes:
      - ../../config/alertmanager/alertmanager.yml:/etc/alertmanager/alertmanager.yml
    command:
      - '--config.file=/etc/alertmanager/alertmanager.yml'
    extra_hosts:
      - "host.docker.internal:host-gateway"
    networks:
      - logpilot
EOF

cat > deploy/docker-compose/.env.example << 'EOF'
POSTGRES_USER=logpilot
POSTGRES_PASSWORD=logpilot123
POSTGRES_DB=logpilot
GRAFANA_PASSWORD=admin
EOF
cp deploy/docker-compose/.env.example deploy/docker-compose/.env

echo "  ✅ configs"

# ═══════════════════════════════════════════════════════════
# ROOT FILES
# ═══════════════════════════════════════════════════════════

cat > .gitignore << 'EOF'
# Env files (keep .env.example)
.env
!.env.example

# Go
*.exe
*.test
*.out
vendor/

# IDE
.idea/
.vscode/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Laravel
services/laravel-api/vendor/
services/laravel-api/storage/logs/

# Next.js
services/frontend/.next/
services/frontend/node_modules/
EOF

cat > README.md << 'EOF'
# LogPilot

> Self-hosted centralized log ingestion & alerting platform.
> A lightweight alternative to Datadog and Better Stack.

**Status:** 🚧 In Development

## Architecture

```
[ App ] ──► [ Go Ingestor ] ──► [ Kafka ] ──► [ Consumer Storage ] ──► [ ClickHouse ] ──► [ Grafana ]
                                          ╰──► [ Consumer Alert  ] ──► [ Alertmanager ] ──► [ Alert Dispatcher ] ──► ClickUp / Email / Slack
```

## Quick Start

```bash
# 1. Start all infrastructure
cd deploy/docker-compose && docker-compose up -d

# 2. Seed a test API key
docker exec logpilot-redis redis-cli SET api_key:test123 project-demo

# 3. Run ingestor
cd services/ingestor && go run cmd/main.go

# 4. Send a test log
curl -X POST http://localhost:8080/v1/ingest \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test123" \
  -d '{"level":"ERROR","message":"db timeout","service":"api","timestamp":"2026-04-17T05:12:00Z"}'
```

## Docs

- [PRD](docs/PRD.md) — Full product requirements
- [TODO](docs/TODO.md) — Development task list
EOF

# ═══════════════════════════════════════════════════════════
# GIT
# ═══════════════════════════════════════════════════════════
echo ""
echo "🗂️  Initializing git..."
git add .
git commit -m "chore: init logpilot with clean architecture boilerplate"

echo ""
echo "╔══════════════════════════════════════════════════════════╗"
echo "║                  ✅ Setup Complete!                      ║"
echo "╠══════════════════════════════════════════════════════════╣"
echo "║  services/ingestor         → 5-layer pipeline, ready    ║"
echo "║  services/consumer-storage → micro-batch ClickHouse     ║"
echo "║  services/consumer-alert   → sliding window + tests     ║"
echo "║  services/alert-dispatcher → concurrent dispatcher      ║"
echo "║  demo/dummy-app            → e-commerce with 80/70 prob ║"
echo "║  demo/log-generator        → 5 logs/sec + spike mode    ║"
echo "╚══════════════════════════════════════════════════════════╝"
echo ""
echo "Next steps:"
echo "  1. cd deploy/docker-compose && docker-compose up -d"
echo "  2. docker-compose ps   ← verify all services are Up"
echo "  3. Continue from Chunk 0.4 in docs/TODO.md"
echo ""
echo "Push ke GitHub:"
echo "  git remote add origin https://github.com/$GITHUB_USERNAME/logpilot.git"
echo "  git push -u origin main"
echo ""