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
