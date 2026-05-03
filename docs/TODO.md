# LogPilot — TODO List
> **Cara baca:** Kerjakan dari atas ke bawah. Centang `[x]` saat selesai.
> Setiap chunk dirancang bisa diselesaikan dalam **30-90 menit**.
> Kalau cuma punya 30 menit, cukup selesaikan 1 chunk — it counts. ✅

---

## 🏁 PHASE 0 — Project Setup
*Target: Monorepo jalan, semua tools siap, bisa langsung nulis kode.*

### Chunk 0.1 — Inisialisasi Repo
- [x] Buat repo baru di GitHub: `logpilot`
- [x] Clone ke lokal, buat struktur folder sesuai PRD Section 8
- [x] Buat `README.md` kosong dulu (nanti diisi di Phase 6)
- [x] Commit awal: `chore: init monorepo structure`

### Chunk 0.2 — Setup Go Workspace
- [x] Di dalam `services/ingestor/`, jalankan `go mod init github.com/{username}/logpilot/ingestor`
- [x] Ulangi untuk `consumer-storage`, `consumer-alert`, `alert-dispatcher`
- [x] Di `demo/dummy-app/`, jalankan `go mod init github.com/{username}/logpilot/demo/dummy-app`
- [x] Di `demo/log-generator/`, jalankan `go mod init github.com/{username}/logpilot/demo/log-generator`

### Chunk 0.3 — Setup Docker Compose Infrastructure
- [x] Buat `deploy/docker-compose/docker-compose.yml` dengan services berikut (tanpa custom services dulu):
  - `kafka` (apache/kafka:3.7.0, KRaft mode — tanpa zookeeper)
  - `redis` (redis:7-alpine)
  - `postgres` (postgres:15-alpine)
  - `clickhouse` (clickhouse/clickhouse-server:23.8)
  - `grafana` (grafana/grafana:10.2.0)
  - `alertmanager` (prom/alertmanager:v0.26.0)
- [x] Buat `deploy/docker-compose/.env.example` dengan semua variable yang dibutuhkan
- [x] Jalankan `docker-compose up -d` dan pastikan semua service healthy
- [x] Verifikasi: `docker-compose ps` — semua status `Up`

### Chunk 0.4 — Verifikasi Koneksi Infrastructure
- [x] Test Redis: `docker exec -it <redis_container> redis-cli ping` → harus reply `PONG`
- [x] Test Kafka: masuk ke Kafka container, buat topic `raw-logs` manual dulu:
  ```
  kafka-topics.sh --create --topic raw-logs --partitions 6 --replication-factor 1 --bootstrap-server localhost:9092
  kafka-topics.sh --create --topic alert-events --partitions 1 --replication-factor 1 --bootstrap-server localhost:9092
  ```
- [x] Test ClickHouse: `docker exec -it <clickhouse_container> clickhouse-client` → bisa masuk CLI
- [x] Test PostgreSQL: connect via TablePlus atau psql, pastikan bisa create database
- [x] Buat database di PostgreSQL: `CREATE DATABASE logpilot;`
- [x] Commit: `chore: docker-compose infra stack ready`

---

## 🚀 PHASE 1 — Go Ingestor Service
*Target: HTTP endpoint `/v1/ingest` bisa menerima log, validasi, dan publish ke Kafka.*

### Chunk 1.1 — Setup HTTP Server
- [x] Masuk ke `services/ingestor/`
- [x] Install Fiber: `go get github.com/gofiber/fiber/v2`
- [x] Buat `main.go` — setup Fiber app, register routes (kosong dulu)
- [x] Buat route `GET /health` yang return `{"status": "ok"}`
- [x] Test: `go run main.go` → curl `localhost:8080/health` → harus dapat response
- [x] Commit: `feat(ingestor): setup fiber http server with health endpoint`

### Chunk 1.2 — Struktur Folder & Config
- [x] Buat folder: `handler/`, `middleware/`, `enricher/`, `producer/`, `config/`
- [x] Buat `config/config.go` — load env variables (Redis URL, Kafka brokers, port)
- [x] Install godotenv untuk local dev: `go get github.com/joho/godotenv`
- [x] Buat `.env` file di folder ingestor (jangan di-commit, tambahkan ke `.gitignore`)
- [x] Commit: `chore(ingestor): setup folder structure and config loader`

### Chunk 1.3 — Redis Client Setup
- [x] Install go-redis: `go get github.com/redis/go-redis/v9`
- [x] Buat `config/redis.go` — inisialisasi Redis client, ping saat startup
- [x] Pastikan ingestor bisa connect ke Redis yang ada di Docker Compose
- [x] Commit: `feat(ingestor): setup redis client`

### Chunk 1.4 — Layer 1: Authentication Middleware
- [x] Buat `middleware/auth.go`
- [x] Logic: ambil header `X-API-Key`, GET dari Redis key `api_key:{value}`, dapat `project_id`
- [x] Jika key tidak ada di Redis → return 401 JSON `{"error": "unauthorized"}`
- [x] Jika ada → simpan `project_id` di Fiber context (`c.Locals`)
- [x] Register middleware ke route `/v1/ingest`
- [x] Test manual: seed dulu satu key ke Redis — `SET api_key:test123 project-test`
- [x] Curl dengan key valid → lanjut; tanpa key → 401
- [x] Commit: `feat(ingestor): layer 1 - api key authentication middleware`

### Chunk 1.5 — Layer 2: Rate Limiting Middleware
- [ ] Buat `middleware/ratelimit.go`
- [ ] Logic: INCR `rate:{project_id}` di Redis, set EXPIRE 60 detik jika key baru
- [ ] Jika counter > 10.000 → return 429 dengan header `Retry-After: 60`
- [ ] Test: loop curl 10.001x (pakai script bash sederhana), pastikan yang ke-10.001 dapat 429
- [ ] Commit: `feat(ingestor): layer 2 - rate limiting middleware`

### Chunk 1.6 — Layer 3: Payload Validation
- [ ] Buat `handler/ingest.go` — define struct `IngestRequest`
- [ ] Field required: `level`, `message`, `service`, `timestamp`
- [ ] Validasi `level` harus salah satu dari: `DEBUG`, `INFO`, `WARN`, `ERROR`, `FATAL`
- [ ] Validasi `timestamp` harus parseable sebagai ISO 8601
- [ ] Jika ada field invalid → return 400 dengan detail field mana yang salah
- [ ] Commit: `feat(ingestor): layer 3 - payload validation`

### Chunk 1.7 — Layer 4: Enrichment
- [ ] Buat `enricher/enricher.go`
- [ ] Logic: ambil `project_id` dari context, tambahkan `ingested_at` (time.Now()), `ingestor_id` (hostname)
- [ ] Return struct `EnrichedLog` yang siap di-publish ke Kafka
- [ ] Commit: `feat(ingestor): layer 4 - log enrichment`

### Chunk 1.8 — Kafka Producer Setup
- [ ] Install sarama: `go get github.com/IBM/sarama`
- [ ] Buat `producer/kafka.go` — inisialisasi async Kafka producer
- [ ] Config: `raw-logs` topic, partition by `project_id` (implement `Partitioner` by hash)
- [ ] Fungsi `Publish(log EnrichedLog) error` — marshal ke JSON, produce ke Kafka
- [ ] Pastikan bisa connect ke Kafka di Docker Compose
- [ ] Commit: `feat(ingestor): setup kafka async producer`

### Chunk 1.9 — Layer 5: Publish ke Kafka
- [ ] Di `handler/ingest.go`, gabungkan semua layer:
  1. Panggil validator
  2. Panggil enricher
  3. Panggil producer.Publish()
  4. Return 202 jika berhasil; 503 jika Kafka gagal
- [ ] Commit: `feat(ingestor): layer 5 - publish enriched log to kafka`

### Chunk 1.10 — Unit Tests Ingestor
- [ ] Buat `middleware/auth_test.go` — test case: valid key, invalid key, missing key
- [ ] Buat `middleware/ratelimit_test.go` — test case: under limit, at limit, over limit
- [ ] Buat `handler/ingest_test.go` — test case: valid payload, missing field, invalid level, invalid timestamp
- [ ] Jalankan `go test ./...` — semua pass
- [ ] Commit: `test(ingestor): unit tests for all pipeline layers`

### Chunk 1.11 — Manual End-to-End Test Ingestor
- [ ] Jalankan ingestor: `go run main.go`
- [ ] Kirim log valid dengan curl:
  ```bash
  curl -X POST http://localhost:8080/v1/ingest \
    -H "Content-Type: application/json" \
    -H "X-API-Key: test123" \
    -d '{"level":"ERROR","message":"test error","service":"test-svc","timestamp":"2026-04-17T05:12:00Z"}'
  ```
- [ ] Verifikasi di Kafka: gunakan console consumer, pastikan message masuk ke `raw-logs`
- [ ] Commit: `feat(ingestor): phase 1 complete - ingestor e2e verified`

---

## 📦 PHASE 2 — Consumer 1: Storage Worker
*Target: Log dari Kafka tersimpan ke ClickHouse, bisa dilihat di Grafana.*

### Chunk 2.1 — Setup ClickHouse Schema
- [ ] Masuk ke ClickHouse CLI: `docker exec -it <clickhouse> clickhouse-client`
- [ ] Buat database: `CREATE DATABASE IF NOT EXISTS logpilot;`
- [ ] Jalankan CREATE TABLE dari PRD Section 4.3
- [ ] Verifikasi: `DESCRIBE TABLE logpilot.logs`
- [ ] Commit: `chore: clickhouse schema created`

### Chunk 2.2 — Setup Consumer Storage Project
- [ ] Masuk ke `services/consumer-storage/`
- [ ] Install dependencies: `go get github.com/IBM/sarama`, `go get github.com/ClickHouse/clickhouse-go/v2`
- [ ] Buat `config/config.go` — load Kafka brokers, ClickHouse DSN
- [ ] Commit: `chore(consumer-storage): setup project and dependencies`

### Chunk 2.3 — Kafka Consumer Setup
- [ ] Buat `consumer/consumer.go` — implement `sarama.ConsumerGroupHandler`
- [ ] Consumer group ID: `storage-workers`
- [ ] Subscribe ke topic `raw-logs`
- [ ] Jangan commit offset dulu di chunk ini — hanya polling dan print ke console dulu
- [ ] Verifikasi: consumer bisa receive message yang dikirim ingestor
- [ ] Commit: `feat(consumer-storage): kafka consumer group setup`

### Chunk 2.4 — ClickHouse Client & Batch Insert
- [ ] Buat `storage/clickhouse.go` — inisialisasi ClickHouse connection
- [ ] Fungsi `BatchInsert(logs []EnrichedLog) error` — satu query INSERT untuk batch
- [ ] Test insert manual dulu tanpa Kafka: hardcode 3 log, insert, query balik
- [ ] Commit: `feat(consumer-storage): clickhouse batch insert`

### Chunk 2.5 — Micro-batch Logic
- [ ] Di `consumer/consumer.go`, implement micro-batch:
  - Kumpulkan messages dalam slice
  - Flush jika sudah 500 messages ATAU sudah 1 detik sejak message pertama
  - Setelah flush (INSERT ke ClickHouse): commit offset
  - Jika INSERT gagal: jangan commit, tunggu 5 detik, proses ulang
- [ ] Commit: `feat(consumer-storage): micro-batch strategy implemented`

### Chunk 2.6 — End-to-End Test Storage Pipeline
- [ ] Jalankan ingestor + consumer-storage bersamaan
- [ ] Kirim 10 log via curl
- [ ] Verifikasi di ClickHouse: `SELECT * FROM logpilot.logs LIMIT 10;`
- [ ] Commit: `feat(consumer-storage): phase 2 complete - logs persisted to clickhouse`

### Chunk 2.7 — Connect Grafana ke ClickHouse
- [ ] Buka Grafana: `http://localhost:3000` (default admin/admin)
- [ ] Install plugin: ClickHouse datasource (via Grafana UI atau provisioning)
- [ ] Tambah datasource baru → ClickHouse → isi host, database `logpilot`
- [ ] Test query sederhana: `SELECT * FROM logs ORDER BY timestamp DESC LIMIT 100`
- [ ] Simpan sebagai dashboard "LogPilot - All Logs"
- [ ] Commit: `chore: grafana connected to clickhouse`

---

## 🚨 PHASE 3 — Consumer 2: Alert Evaluator + Alert Dispatcher
*Target: Error spike trigger alert → ClickUp task ter-create dan email terkirim.*

### Chunk 3.1 — Setup Consumer Alert Project
- [ ] Masuk ke `services/consumer-alert/`
- [ ] Install dependencies: sarama, go-redis
- [ ] Buat struktur folder: `consumer/`, `evaluator/`, `counter/`, `config/`
- [ ] Consumer group ID: `alert-evaluators` (berbeda dari storage-workers!)
- [ ] Subscribe ke `raw-logs` — sama dengan consumer-storage tapi independent
- [ ] Commit: `chore(consumer-alert): setup project structure`

### Chunk 3.2 — Alert Rule Struct & Cache
- [ ] Buat `evaluator/rules.go` — define struct `AlertRule` sesuai PRD Section 4.4
- [ ] Buat `evaluator/cache.go` — in-memory cache untuk alert rules
- [ ] Fungsi `LoadRules()` — placeholder dulu, return hardcoded 1 rule untuk testing
- [ ] Background goroutine: refresh rules dari "database" tiap 30 detik (hardcode dulu)
- [ ] Commit: `feat(consumer-alert): alert rule struct and in-memory cache`

### Chunk 3.3 — Redis Sliding Window Counter
- [ ] Buat `counter/sliding_window.go`
- [ ] Fungsi `Increment(ruleID, projectID string, windowSeconds int) (int64, error)`
  - INCR key `alert_counter:{ruleID}:{projectID}`
  - EXPIRE key dengan `windowSeconds`
  - Return nilai counter setelah increment
- [ ] Fungsi `HasCooldown(ruleID, projectID string) (bool, error)`
  - GET key `cooldown:{ruleID}:{projectID}`
  - Return true jika key ada
- [ ] Fungsi `SetCooldown(ruleID, projectID string, cooldownSeconds int) error`
  - SET + EXPIRE key `cooldown:{ruleID}:{projectID}`
- [ ] Test manual fungsi-fungsi ini tanpa Kafka
- [ ] Commit: `feat(consumer-alert): redis sliding window counter`

### Chunk 3.4 — Evaluasi Logic
- [ ] Buat `evaluator/evaluator.go`
- [ ] Fungsi `Evaluate(log EnrichedLog, rules []AlertRule) []AlertRule`
  - Return rules yang di-match oleh log ini
- [ ] Matching logic: `project_id` cocok DAN `level` cocok DAN `service` cocok (jika di-set)
- [ ] Commit: `feat(consumer-alert): log matching evaluation logic`

### Chunk 3.5 — Kafka Producer untuk Alert Events
- [ ] Buat `producer/kafka.go` — produce ke topic `alert-events`
- [ ] Struct `AlertEvent`: rule_id, project_id, rule_name, triggered_at, log sample
- [ ] Commit: `feat(consumer-alert): alert event kafka producer`

### Chunk 3.6 — Gabungkan Alert Pipeline
- [ ] Di consumer handler, untuk setiap message dari Kafka:
  1. Parse log
  2. Panggil `evaluator.Evaluate()`
  3. Untuk setiap matched rule:
     - Cek `HasCooldown()` → skip jika ada
     - `Increment()` → cek apakah >= threshold
     - Jika threshold tercapai: publish ke `alert-events`, `SetCooldown()`
  4. Commit offset
- [ ] Commit: `feat(consumer-alert): complete alert evaluation pipeline`

### Chunk 3.7 — Setup Alertmanager Config
- [ ] Buat `config/alertmanager/alertmanager.yml`:
  ```yaml
  global:
    resolve_timeout: 5m
  route:
    receiver: 'logpilot-dispatcher'
  receivers:
    - name: 'logpilot-dispatcher'
      webhook_configs:
        - url: 'http://alert-dispatcher:9090/webhook'
  ```
- [ ] Update docker-compose untuk mount config ini ke alertmanager container
- [ ] Restart alertmanager, verifikasi config loaded
- [ ] Commit: `chore: alertmanager config for logpilot dispatcher`

### Chunk 3.8 — Setup Alert Dispatcher Project
- [ ] Masuk ke `services/alert-dispatcher/`
- [ ] Install Fiber, go-resty atau net/http
- [ ] Buat `handler/webhook.go` — terima POST dari Alertmanager
- [ ] Parse payload Alertmanager, extract alert info
- [ ] Commit: `chore(alert-dispatcher): project setup and webhook handler`

### Chunk 3.9 — ClickUp Integration
- [ ] Daftar ClickUp API key (gratis)
- [ ] Buat `destinations/clickup.go`
- [ ] Fungsi `CreateTask(alert AlertEvent, config NotifConfig) error`
  - POST ke `https://api.clickup.com/api/v2/list/{list_id}/task`
  - Task name: `[ALERT] {rule_name} - {project_id}`
  - Priority: Urgent (1)
  - Body: include log detail
- [ ] Test dengan hardcode list_id dulu
- [ ] Commit: `feat(alert-dispatcher): clickup task creation`

### Chunk 3.10 — SendGrid Email Integration
- [ ] Daftar SendGrid, dapat API key (free tier cukup)
- [ ] Install `go get github.com/sendgrid/sendgrid-go`
- [ ] Buat `destinations/sendgrid.go`
- [ ] Fungsi `SendEmail(alert AlertEvent, config NotifConfig) error`
- [ ] Test kirim email ke email kamu sendiri
- [ ] Commit: `feat(alert-dispatcher): sendgrid email notification`

### Chunk 3.11 — Concurrent Dispatch
- [ ] Di dispatcher handler, jalankan semua destinations secara concurrent:
  ```go
  var wg sync.WaitGroup
  wg.Add(2) // clickup + email
  go func() { defer wg.Done(); clickup.CreateTask(...) }()
  go func() { defer wg.Done(); sendgrid.SendEmail(...) }()
  wg.Wait()
  ```
- [ ] Timeout 10 detik per goroutine menggunakan context
- [ ] Commit: `feat(alert-dispatcher): concurrent multi-destination dispatch`

### Chunk 3.12 — End-to-End Test Alert Pipeline
- [ ] Jalankan semua: ingestor + consumer-alert + alert-dispatcher + alertmanager
- [ ] Kirim 10+ ERROR logs dalam waktu < 5 menit (sesuai threshold rule hardcoded)
- [ ] Verifikasi: ClickUp task ter-create, email masuk
- [ ] Test cooldown: kirim 10 ERROR lagi — alert tidak trigger (cooldown aktif)
- [ ] Commit: `feat: phase 3 complete - alert pipeline e2e verified`

---

## 🎛️ PHASE 4 — Laravel API: Management Control Plane
*Target: Bisa manage projects, API keys, dan alert rules via REST API.*

### Chunk 4.1 — Setup Laravel Project
- [ ] Di `services/laravel-api/`: `composer create-project laravel/laravel .`
- [ ] Install Sanctum: `composer require laravel/sanctum`, publish config
- [ ] Konfigurasi `.env`: DB_CONNECTION=pgsql, DB_DATABASE=logpilot, dll
- [ ] Test: `php artisan serve` → bisa akses `http://localhost:8000`
- [ ] Commit: `chore(laravel-api): laravel project setup with sanctum`

### Chunk 4.2 — Database Migration: Users & Projects
- [ ] Buat migration untuk tabel `users` (sudah ada default Laravel, cukup review)
- [ ] Buat migration untuk tabel `projects`:
  ```
  id, user_id (FK), name, slug (unique), created_at, updated_at
  ```
- [ ] Jalankan `php artisan migrate`
- [ ] Commit: `chore(laravel-api): users and projects migration`

### Chunk 4.3 — Database Migration: API Keys & Alert Rules
- [ ] Buat migration untuk `api_keys`:
  ```
  id, project_id (FK), name, key_hash, last_used_at, created_at
  ```
- [ ] Buat migration untuk `alert_rules`:
  ```
  id, project_id (FK), name, condition (JSON), threshold, window_seconds, cooldown_seconds, is_active (bool)
  ```
- [ ] Buat migration untuk `notification_configs` dan `alert_logs` sesuai PRD
- [ ] Jalankan migrate, verifikasi semua tabel ada
- [ ] Commit: `chore(laravel-api): api_keys, alert_rules, notification_configs migrations`

### Chunk 4.4 — Auth Endpoints
- [ ] Buat `AuthController` dengan method `register` dan `login`
- [ ] Register: validasi email unik, hash password, return user
- [ ] Login: validasi credentials, `createToken()` Sanctum, return token
- [ ] Register route di `routes/api.php`
- [ ] Test dengan Postman/curl: register → login → dapat token
- [ ] Commit: `feat(laravel-api): auth register and login endpoints`

### Chunk 4.5 — Project Endpoints
- [ ] Buat `ProjectController` dengan method `index` dan `store`
- [ ] `store`: validasi nama, generate slug (dari nama), simpan dengan `user_id` dari auth
- [ ] `index`: return hanya project milik user yang sedang login
- [ ] Protect route dengan `auth:sanctum` middleware
- [ ] Commit: `feat(laravel-api): project CRUD endpoints`

### Chunk 4.6 — API Key Management
- [ ] Buat `ApiKeyController` dengan method `store` dan `destroy`
- [ ] `store`:
  - Generate plain key: `logpilot_` + random 32 char
  - Simpan `key_hash` (sha256 atau bcrypt) ke PostgreSQL
  - Seed plain key ke Redis: `SET api_key:{plain_key} {project_id}`
  - Return plain key ke user (HANYA sekali, tidak bisa dilihat lagi)
- [ ] `destroy`: hapus dari PostgreSQL + hapus dari Redis
- [ ] Commit: `feat(laravel-api): api key generate and revoke`

### Chunk 4.7 — Alert Rule Endpoints
- [ ] Buat `AlertRuleController` dengan method `index`, `store`, `update`, `destroy`
- [ ] Validasi: threshold dan window_seconds harus positif integer
- [ ] `condition` disimpan sebagai JSON column
- [ ] Commit: `feat(laravel-api): alert rule CRUD endpoints`

### Chunk 4.8 — Notification Config Endpoint
- [ ] Buat `NotificationController` dengan method `update` dan `show`
- [ ] `PUT /api/projects/{id}/notifications` — upsert config
- [ ] `GET /internal/projects/{id}/notifications` — endpoint untuk Alert Dispatcher (tanpa auth Sanctum, pakai internal secret header)
- [ ] Commit: `feat(laravel-api): notification config endpoints`

### Chunk 4.9 — Update Consumer Alert: Baca Rules dari DB
- [ ] Kembali ke `services/consumer-alert/`
- [ ] Update `evaluator/cache.go`: ganti hardcode rules dengan HTTP call ke `GET /internal/alert-rules/{project_id}`
  - Atau langsung query PostgreSQL — pilih yang lebih simple dulu
- [ ] Refresh tiap 30 detik (sudah ada skeleton dari Chunk 3.2)
- [ ] Test: buat rule via Laravel API → dalam 30 detik rule aktif di Consumer 2
- [ ] Commit: `feat(consumer-alert): load alert rules from laravel api`

### Chunk 4.10 — Update Alert Dispatcher: Baca Notif Config dari DB
- [ ] Kembali ke `services/alert-dispatcher/`
- [ ] Saat terima webhook dari Alertmanager: query `GET /internal/projects/{id}/notifications`
- [ ] Gunakan config tersebut untuk ClickUp list_id, email recipients, dll
- [ ] Commit: `feat(alert-dispatcher): load notification config from laravel api`

---

## 🖥️ PHASE 5 — Next.js Frontend
*Target: UI sederhana untuk manage project, API keys, dan alert rules.*

### Chunk 5.1 — Setup Next.js Project
- [ ] Di `services/frontend/`: `npx create-next-app@latest . --typescript --tailwind`
- [ ] Test: `npm run dev` → bisa akses `localhost:3000`
- [ ] Install axios atau fetch wrapper: `npm install axios`
- [ ] Setup base API URL ke Laravel (`http://localhost:8000`)
- [ ] Commit: `chore(frontend): next.js project setup`

### Chunk 5.2 — Halaman Login & Register
- [ ] Buat `pages/login.tsx` — form email + password
- [ ] On submit: POST ke `/api/auth/login`, simpan token ke localStorage
- [ ] Buat `pages/register.tsx` — form name + email + password
- [ ] Redirect ke `/projects` setelah login berhasil
- [ ] Commit: `feat(frontend): login and register pages`

### Chunk 5.3 — Halaman Projects
- [ ] Buat `pages/projects/index.tsx` — list semua project user
- [ ] Button "New Project" → modal atau form inline
- [ ] Setiap project card: nama + link ke detail project
- [ ] Commit: `feat(frontend): projects list page`

### Chunk 5.4 — Halaman API Keys
- [ ] Buat `pages/projects/[id]/api-keys.tsx`
- [ ] List semua API keys (tampilkan nama + tanggal dibuat + last_used_at)
- [ ] Button "Generate New Key" → tampilkan modal dengan plain key (one-time display)
- [ ] Button "Revoke" per key
- [ ] Commit: `feat(frontend): api key management page`

### Chunk 5.5 — Halaman Alert Rules
- [ ] Buat `pages/projects/[id]/alerts.tsx`
- [ ] List alert rules dengan status aktif/nonaktif
- [ ] Form buat rule baru: name, level filter, service filter, threshold, window, cooldown
- [ ] Toggle is_active per rule
- [ ] Commit: `feat(frontend): alert rules management page`

### Chunk 5.6 — Halaman Notification Config
- [ ] Buat `pages/projects/[id]/notifications.tsx`
- [ ] Form: ClickUp list ID, ClickUp assignee ID, email recipients (comma separated), Slack webhook
- [ ] Save button → PUT ke API
- [ ] Commit: `feat(frontend): notification config page`

---

## 🎪 PHASE 6 — Demo Application
*Target: Ada dummy app yang bisa digunakan untuk demo live.*

### Chunk 6.1 — Dummy E-commerce App
- [ ] Masuk ke `demo/dummy-app/`
- [ ] Install Fiber: `go get github.com/gofiber/fiber/v2`
- [ ] Implement 4 endpoints sesuai PRD Section 6.1:
  - `GET /products`
  - `GET /products/:id` (random 80/20 sukses/not found)
  - `POST /checkout` (random 70/30 sukses/timeout)
  - `GET /health`
- [ ] Setiap response, kirim log ke Ingestor via HTTP POST (gunakan `go-resty` atau `net/http`)
- [ ] Commit: `feat(demo): dummy e-commerce api with log integration`

### Chunk 6.2 — Log Generator Script
- [ ] Masuk ke `demo/log-generator/`
- [ ] Kirim 1 log setiap 200ms secara default
- [ ] Mix level: 60% INFO, 20% WARN, 15% ERROR, 5% FATAL (random distribution)
- [ ] Setiap 2 menit: spike mode — kirim 20 ERROR dalam 30 detik
- [ ] Config via env: `INGESTOR_URL`, `API_KEY`, `PROJECT_ID`
- [ ] Commit: `feat(demo): log generator script with error spike simulation`

### Chunk 6.3 — Grafana Dashboard
- [ ] Di Grafana, buat dashboard "LogPilot Demo":
  - Panel 1: Log volume per menit (time series)
  - Panel 2: Error rate (% ERROR dari total)
  - Panel 3: Log table (sortable by timestamp, filterable by level)
  - Panel 4: Top services by log volume
- [ ] Export dashboard JSON ke `config/grafana/dashboard.json`
- [ ] Commit: `chore: grafana demo dashboard`

---

## 📝 PHASE 7 — Polish & Documentation
*(Dikerjakan setelah semua phase sebelumnya selesai)*

### Chunk 7.1 — README Utama
- [ ] Buat `README.md` yang mencakup:
  - Badge: Go version, License
  - Screenshot/GIF demo
  - Architecture diagram (bisa pakai Mermaid)
  - Quick start: `git clone` + `docker-compose up`
  - Link ke PRD

### Chunk 7.2 — Dockerfile per Service
- [ ] Buat `Dockerfile` untuk setiap Go service (multi-stage build)
- [ ] Buat `Dockerfile` untuk Laravel API
- [ ] Buat `Dockerfile` untuk Next.js frontend
- [ ] Update `docker-compose.yml` untuk include custom services

### Chunk 7.3 — GitHub Actions CI
- [ ] Buat `.github/workflows/ci.yml`
- [ ] Jobs: lint, test (Go), build Docker image
- [ ] Trigger: push ke `main` dan pull request

### Chunk 7.4 — Kubernetes Manifests
*(Bahan belajar — kerjakan setelah Docker Compose fully working)*
- [ ] Buat `deployment.yaml` untuk Go Ingestor (dengan HPA 2-10 pods)
- [ ] Buat manifests untuk semua custom services
- [ ] Setup Helm untuk infrastructure dependencies
- [ ] Test di Minikube atau OrbStack

### Chunk 7.5 — Demo Video & Final Polish
- [ ] Record demo video mengikuti urutan dari PRD Section 6.3
- [ ] Upload ke YouTube (unlisted) atau LinkedIn
- [ ] Link di README dan profil GitHub

---

## 📊 Progress Tracker

| Phase | Status | Selesai Tgl |
|---|---|---|
| Phase 0 — Setup | ✅ Done | 2026-04-26 |
| Phase 1 — Ingestor | ⬜ Not started | — |
| Phase 2 — Storage Worker | ⬜ Not started | — |
| Phase 3 — Alert Pipeline | ⬜ Not started | — |
| Phase 4 — Laravel API | ⬜ Not started | — |
| Phase 5 — Frontend | ⬜ Not started | — |
| Phase 6 — Demo App | ⬜ Not started | — |
| Phase 7 — Polish | ⬜ Not started | — |

---

*Update status: ⬜ Not started → 🟡 In progress → ✅ Done*
