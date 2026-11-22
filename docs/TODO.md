# LogPilot — TODO List
> **Cara baca:** Kerjakan dari atas ke bawah. Centang `[x]` saat selesai.
> Setiap chunk dirancang bisa diselesaikan dalam **30-90 menit**.
> Kalau cuma punya 30 menit, cukup selesaikan 1 chunk — it counts. ✅

---

> ## ⚠️ UPDATE 22 NOVEMBER 2026 (PRD v3.1) — Baca ini dulu
>
> Empat keputusan; rinciannya di changelog PRD v3.1.
>
> - **Phase 4 ditulis ulang untuk Go.** Laravel tidak dipakai lagi. Management API menjadi modul di service `services/api`, dan Query API (`S-1`) nanti masuk ke service yang sama sebagai modul `query` (PRD §4.6, §4.8). Nomor chunk 4.1–4.10 tetap.
> - **Handler Monolog untuk PHP** sekarang butir Stretch pertama (PRD §12).
> - **Perkakas platform dipasang satu per satu** dengan urutan di PRD §6.4. Berlaku untuk seluruh Phase P.
> - **Tulisan teknis:** setiap kali chunk yang memuat masalah sulit selesai, tulis satu tulisan pendek di `docs/writing/` (PRD §7.5).

> ## ⚠️ UPDATE NOVEMBER 2026 (PRD v3.0)
>
> PRD naik ke v3.0. Tiga hal berubah di file ini:
>
> **1. Koreksi teknis pada chunk yang belum dikerjakan** — ditandai `🔧 KOREKSI v3`. Kalau isi chunk dan PRD berbeda, **PRD v3.0 yang benar**:
> - `2.4` / `R-I5` — dedup token saja tidak cukup: batch harus per partition, tabel butuh `non_replicated_deduplication_window`, dan restart ditangani high-water mark (PRD §4.3).
> - `R-J2` — DDL: partisi harian, `TTL toDateTime(timestamp)`, `fingerprint String`, kolom `kafka_partition`/`kafka_offset`, dua `SETTINGS` (PRD §4.3).
> - `R-S1` / `R-S2` — DDL materialized view lama menghitung salah setelah merge. Pakai tabel target `issues_hourly` dengan `SimpleAggregateFunction` dan `-State` (PRD §5.3).
> - `1.5` (sudah selesai, perlu ditengok lagi) — rate limiter punya pola `INCR` lalu `EXPIRE` yang sama dengan yang ditolak di `R-I2` (PRD §4.1).
>
> **2. Dua phase baru di bawah Phase R:**
> - **PHASE S — Showcase**: Query API dan Console. Yang membuat project ini bisa diklik orang non-teknis.
> - **PHASE P — Platform**: k3s, Helm, GitOps, TLS, operator, autoscaling, backup. Jalur belajar DevOps.
>
> **3. Urutan kerja sekarang mengikuti milestone M1–M8 di PRD §8**, bukan tanggal. Aturan gerbangnya ada di PRD §2.3: Showcase baru dimulai setelah exit demo Core direkam. Pengecualian: `P-1` sampai `P-4` boleh jadi selingan kapan saja setelah M1.
>
> Chunk Phase S dan P di bawah baru berisi tujuan dan kriteria selesai. Blok **📖 Pahami dulu** dan **❓ Bisa jawab ini?** ditulis saat chunk itu akan dikerjakan, supaya isinya mengikuti apa yang saat itu sudah kamu kuasai.

> ## ⚠️ UPDATE JULI 2026
>
> Setelah review PRD (dan setelah pertanyaan interview soal **trace_id**), ada dua jenis perubahan di file ini:
>
> **1. Chunk lama yang isinya diperbaiki** — ditandai `🔧 REVISI`. Semuanya ada di chunk yang **belum dikerjakan**, jadi tinggal ikuti versi barunya. Yang direvisi: `2.4`, `2.5`, `3.3`, `3.6`, `3.7`, `4.6`.
>
> **2. Phase baru: [PHASE R — Revisi & Peningkatan](#-phase-r--revisi--peningkatan-hasil-review-prd)** — semua penambahan, dikelompokkan per **level knowledge**:
> - 🟢 **Junior** — konsepnya gampang, implementasinya nambah field / nambah `if`. Bisa langsung dikerjakan.
> - 🟡 **Intermediate** — butuh paham satu konsep distributed system dulu. Ini yang paling sering ditanya interviewer.
> - 🔴 **Senior** — konsep yang bikin kamu beda dari kandidat lain. Kerjakan setelah dua level di atas.
>
> **Urutan yang kusarankan:**
> `PHASE R Junior` → `Chunk 2.4 dst (Phase 2)` → `PHASE R Intermediate (bertahap sambil jalan)` → `Phase 3-7` → `PHASE R Senior`
>
> Alasannya: perubahan skema ClickHouse dan field trace context harus masuk **sebelum** kamu nulis batch insert di Chunk 2.4. Kalau tidak, nanti nulis ulang.

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
> 🔧 **KOREKSI v3** — Chunk ini sudah selesai, tapi tengok lagi `ratelimit.go` setelah mengerjakan `R-I2`: apa yang terjadi pada key `rate:{project}` kalau proses mati tepat setelah `INCR`? Spesifikasi barunya di PRD §4.1 (Rate limiter, Layer 2) dan kriteria ujinya F17.
- [x] Buat `middleware/ratelimit.go`
- [x] Logic: INCR `rate:{project_id}` di Redis, set EXPIRE 60 detik jika key baru
- [x] Jika counter > 10.000 → return 429 dengan header `Retry-After: 60`
- [x] Test: loop curl 10.001x (pakai script bash sederhana), pastikan yang ke-10.001 dapat 429
- [x] Commit: `feat(ingestor): layer 2 - rate limiting middleware`

### Chunk 1.6 — Layer 3: Payload Validation
- [x] Buat `handler/ingest.go` — define struct `IngestRequest`
- [x] Field required: `level`, `message`, `service`, `timestamp`
- [x] Validasi `level` harus salah satu dari: `DEBUG`, `INFO`, `WARN`, `ERROR`, `FATAL`
- [x] Validasi `timestamp` harus parseable sebagai ISO 8601
- [x] Jika ada field invalid → return 400 dengan detail field mana yang salah
- [x] Commit: `feat(ingestor): layer 3 - payload validation`

### Chunk 1.7 — Layer 4: Enrichment
- [x] Buat `enricher/enricher.go`
- [x] Logic: ambil `project_id` dari context, tambahkan `ingested_at` (time.Now()), `ingestor_id` (hostname)
- [x] Return struct `EnrichedLog` yang siap di-publish ke Kafka
- [x] Commit: `feat(ingestor): layer 4 - log enrichment`

### Chunk 1.8 — Kafka Producer Setup
- [x] Install sarama: `go get github.com/IBM/sarama`
- [x] Buat `producer/kafka.go` — inisialisasi async Kafka producer
- [x] Config: `raw-logs` topic, partition by `project_id` (implement `Partitioner` by hash)
- [x] Fungsi `Publish(log EnrichedLog) error` — marshal ke JSON, produce ke Kafka
- [x] Pastikan bisa connect ke Kafka di Docker Compose
- [x] Commit: `feat(ingestor): setup kafka async producer`

### Chunk 1.9 — Layer 5: Publish ke Kafka
- [x] Di `handler/ingest.go`, gabungkan semua layer:
  1. Panggil validator
  2. Panggil enricher
  3. Panggil producer.Publish()
  4. Return 202 jika berhasil; 503 jika Kafka gagal
- [x] Commit: `feat(ingestor): layer 5 - publish enriched log to kafka`

### Chunk 1.10 — Unit Tests Ingestor
- [x] Buat `middleware/auth_test.go` — test case: valid key, invalid key, missing key
- [x] Buat `middleware/ratelimit_test.go` — test case: under limit, at limit, over limit
- [x] Buat `handler/ingest_test.go` — test case: valid payload, missing field, invalid level, invalid timestamp
- [x] Jalankan `go test ./...` — semua pass
- [x] Commit: `test(ingestor): unit tests for all pipeline layers`

### Chunk 1.11 — Manual End-to-End Test Ingestor
- [x] Jalankan ingestor: `go run main.go`
- [x] Kirim log valid dengan curl:
  ```bash
  curl -X POST http://localhost:8080/v1/ingest \
    -H "Content-Type: application/json" \
    -H "X-API-Key: test123" \
    -d '{"level":"ERROR","message":"test error","service":"test-svc","timestamp":"2026-04-17T05:12:00Z"}'
  ```
- [x] Verifikasi di Kafka: gunakan console consumer, pastikan message masuk ke `raw-logs`
- [x] Commit: `feat(ingestor): phase 1 complete - ingestor e2e verified`

---

## 📦 PHASE 2 — Consumer 1: Storage Worker
*Target: Log dari Kafka tersimpan ke ClickHouse, bisa dilihat di Grafana.*

### Chunk 2.1 — Setup ClickHouse Schema
- [x] Masuk ke ClickHouse CLI: `docker exec -it <clickhouse> clickhouse-client`
- [x] Buat database: `CREATE DATABASE IF NOT EXISTS logpilot;`
- [x] Jalankan CREATE TABLE dari PRD Section 4.3
- [x] Verifikasi: `DESCRIBE TABLE logpilot.logs`
- [ ] Commit: `chore: clickhouse schema created`

### Chunk 2.2 — Setup Consumer Storage Project
- [x] Masuk ke `services/consumer-storage/`
- [x] Install dependencies: `go get github.com/IBM/sarama`, `go get github.com/ClickHouse/clickhouse-go/v2`
- [x] Buat `config/config.go` — load Kafka brokers, ClickHouse DSN
- [x] Commit: `chore(consumer-storage): setup project and dependencies`

### Chunk 2.3 — Kafka Consumer Setup
- [x] Buat `consumer/consumer.go` — implement `sarama.ConsumerGroupHandler`
- [x] Consumer group ID: `storage-workers`
- [x] Subscribe ke topic `raw-logs`
- [x] Jangan commit offset dulu di chunk ini — hanya polling dan print ke console dulu
- [x] Verifikasi: consumer bisa receive message yang dikirim ingestor
- [x] Commit: `feat(consumer-storage): kafka consumer group setup`

### Chunk 2.4 — ClickHouse Client & Batch Insert 🔧 REVISI
> 🔧 **KOREKSI v3** — Token dedup di bawah baru bekerja kalau (a) batch dibentuk per partition, dan (b) tabel dibuat dengan `non_replicated_deduplication_window`. Untuk kasus restart, token tidak cukup; lihat high-water mark di PRD §4.3. Struct insert juga perlu `kafka_partition` dan `kafka_offset`.
> **Prasyarat:** kerjakan `R-J1` dan `R-J2` dulu, supaya struct dan skema sudah final sebelum nulis insert.

- [ ] Buat `storage/clickhouse.go` — inisialisasi ClickHouse connection
- [ ] Fungsi `BatchInsert(ctx context.Context, logs []EnrichedLog) error` — satu query INSERT untuk batch
- [ ] Insert harus mencakup kolom baru: `trace_id`, `span_id`, `parent_span_id`, `user_id`, `session_id`, `attributes` (Map)
- [ ] **BARU** — Kirim `insert_deduplication_token` per batch (lihat `R-I5` untuk alasannya). Token = hash dari `topic:partition:offset_pertama:offset_terakhir`:
  ```go
  ctx = clickhouse.Context(ctx, clickhouse.WithSettings(map[string]any{
      "insert_deduplication_token": token,
  }))
  ```
- [ ] **BARU** — Pasang timeout di context insert (misal 30 detik), jangan biarkan menggantung selamanya
- [ ] Test insert manual dulu tanpa Kafka: hardcode 3 log, insert, query balik
- [ ] Test dedup: insert batch yang sama 2x dengan token yang sama → `SELECT count()` tetap 3, bukan 6
- [ ] Commit: `feat(consumer-storage): clickhouse batch insert with dedup token`

### Chunk 2.5 — Micro-batch Logic 🔧 REVISI
- [ ] Di `consumer/consumer.go`, implement micro-batch:
  - Kumpulkan messages dalam slice
  - Flush jika sudah 500 messages ATAU sudah 1 detik sejak message pertama
  - Setelah flush (INSERT ke ClickHouse): commit offset
  - Jika INSERT gagal: jangan commit, retry
- [ ] **REVISI** — Retry jangan "tunggu 5 detik" flat. Pakai **exponential backoff dengan jitter**: 1s → 2s → 4s → 8s → max 30s, ditambah random 0-500ms.
  *Kenapa:* kalau ClickHouse down dan kamu punya 3 pod consumer, retry flat bikin ketiganya hit ClickHouse di detik yang sama persis berulang-ulang (thundering herd). Jitter memecah barisan itu.
- [ ] **REVISI** — Batasi retry: setelah N kali gagal (misal 10), jangan retry selamanya. Kirim batch ke DLQ dan commit offset — lihat `R-I6`.
  *Kenapa:* kalau ada 1 message rusak yang selalu bikin INSERT gagal, retry tanpa batas akan **membekukan seluruh partition selamanya**. Semua log lain di partition itu ikut macet.
- [ ] **BARU** — Handle graceful shutdown: saat terima SIGTERM, flush batch yang masih di memory dulu baru exit. Kalau tidak, batch terakhir hilang setiap kali deploy.
- [ ] Commit: `feat(consumer-storage): micro-batch with backoff and bounded retry`

### Chunk 2.6 — End-to-End Test Storage Pipeline
- [ ] Jalankan ingestor + consumer-storage bersamaan
- [ ] Kirim 10 log via curl
- [ ] Verifikasi di ClickHouse: `SELECT * FROM logpilot.logs LIMIT 10;`
- [ ] Commit: `feat(consumer-storage): phase 2 complete - logs persisted to clickhouse`

### Chunk 2.8 — Profiling & Allocation Optimization
- [ ] Buat benchmark test untuk micro-batch pipeline: `go test -bench=. -benchmem`
- [ ] Run pprof: `go tool pprof` — identifikasi top allocations dan CPU hotspots
- [ ] Optimize: pre-allocate slice capacity di BatchInsert, pakai `sync.Pool` untuk reuse buffer
- [ ] Re-run benchmark, bandingkan allocs/op sebelum dan sesudah
- [ ] Target: kurangi allocs/op minimal 50% dari baseline
- [ ] Commit: `perf(consumer-storage): reduce heap allocations in batch pipeline`

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

### Chunk 3.3 — Redis Counter (Fixed Window) 🔧 REVISI BESAR

> **⚠️ Versi lama chunk ini punya 2 bug.** Baca dulu penjelasannya, jangan langsung ngoding.
>
> **Bug 1 — TTL ter-reset terus, counter tidak pernah expire.**
> Versi lama memanggil `EXPIRE` **setiap kali** ada log yang match. Artinya setiap error baru memperpanjang umur key 5 menit lagi. Kalau error datang pelan tapi terus-menerus (misal 1 error tiap 2 menit selama 3 jam), key itu **tidak pernah expire** dan counternya menumpuk sampai 90. Alert akan trigger dan mengklaim *"10 error dalam 5 menit"* padahal kenyataannya 10 error tersebar dalam 3 jam. **Alert palsu.**
> → Perbaikan: TTL hanya di-set **saat key pertama kali dibuat**, tidak pernah diperpanjang.
>
> **Bug 2 — INCR dan EXPIRE tidak atomic.**
> Kalau proses mati tepat di antara dua perintah itu, key sudah ada tapi belum punya TTL → key itu jadi abadi di Redis, counter tidak pernah reset.
> → Perbaikan: gabungkan jadi satu operasi atomic pakai Lua script.
>
> **Catatan nama:** ini namanya **fixed window**, bukan sliding window. Sliding window yang benar ada di `R-I2` — kerjakan itu setelah versi ini jalan. Jangan skip ke sana langsung, pahami yang gampang dulu.

- [ ] Buat `counter/window.go` (rename dari `sliding_window.go` — namanya jangan bohong)
- [ ] Fungsi `Increment(ctx, ruleID, projectID string, windowSeconds int) (int64, error)` — implement dengan Lua script:
  ```lua
  -- KEYS[1] = alert_counter:{ruleID}:{projectID}
  -- ARGV[1] = windowSeconds
  local v = redis.call('INCR', KEYS[1])
  if v == 1 then
    redis.call('EXPIRE', KEYS[1], ARGV[1])
  end
  return v
  ```
  *Baca:* `EXPIRE` hanya jalan kalau hasil INCR = 1, artinya key baru saja lahir. Seluruh script dijalankan Redis sebagai satu operasi tunggal → atomic.
- [ ] Load script pakai `redis.NewScript(...)` di go-redis, panggil dengan `.Run(ctx, client, keys, args)`
- [ ] Fungsi `HasCooldown(ctx, ruleID, projectID string) (bool, error)` — `EXISTS cooldown:{ruleID}:{projectID}`
- [ ] Fungsi `SetCooldown(ctx, ruleID, projectID string, cooldownSeconds int) error` — pakai `SET key 1 EX {cooldown} NX` (satu perintah, sudah atomic)
- [ ] **Test yang membuktikan bug lama sudah hilang:**
  - Set window 5 detik, threshold 10
  - Increment 1x, tunggu 3 detik, increment 1x lagi, tunggu 3 detik
  - Cek `TTL` key → harus sudah expire / counter balik ke 1, **bukan** 2
- [ ] Commit: `feat(consumer-alert): fixed-window counter with atomic lua script`

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

### Chunk 3.6 — Gabungkan Alert Pipeline 🔧 REVISI

> **⚠️ Urutan di versi lama salah.** Versi lama: cek cooldown → kalau ada cooldown, **skip sebelum increment**.
> Akibatnya counter berhenti menghitung selama cooldown. Begitu cooldown habis, counter mulai dari nol dan kamu **kehilangan seluruh konteks error yang terjadi selama 10 menit itu**. Padahal justru di periode itu insidennya sedang berlangsung.
>
> Aturan yang benar: **cooldown menekan pengiriman notifikasi, bukan penghitungan.** Counter jalan terus, selalu.

- [ ] Di consumer handler, untuk setiap message dari Kafka:
  1. Parse log
  2. Panggil `evaluator.Evaluate()`
  3. Untuk setiap matched rule:
     - **`Increment()` dulu — selalu, tanpa syarat**
     - Kalau counter < threshold → selesai, lanjut rule berikutnya
     - Kalau counter >= threshold → **baru** cek `HasCooldown()`
     - Kalau ada cooldown → jangan kirim, tapi catat metric `alerts_suppressed_total{rule_id}` (biar kelihatan di Grafana kalau ada insiden panjang yang terus tertahan)
     - Kalau tidak ada cooldown → publish alert, `SetCooldown()`
  4. Commit offset
- [ ] **BARU** — `SetCooldown` harus pakai `SET ... NX` dan cek return value-nya. Kalau ada 2 pod consumer-alert yang jalan bersamaan, keduanya bisa menembus threshold di saat yang sama dan mengirim 2 alert duplikat. Yang boleh kirim hanya pod yang berhasil men-set cooldown (`NX` return OK):
  ```go
  ok, err := rdb.SetNX(ctx, cooldownKey, 1, cooldownDur).Result()
  if !ok { return } // pod lain sudah kirim duluan
  publishAlert(...)
  ```
  *Ini namanya distributed lock sederhana. Pertanyaan interview yang sering muncul: "kalau consumer-nya di-scale jadi 3, apa yang rusak?"*
- [ ] Commit: `feat(consumer-alert): complete alert evaluation pipeline`

### Chunk 3.7 — Setup Alertmanager Config 🔧 REVISI PENTING

> **⚠️ Ada kesalahan arsitektur di PRD Section 3.2 langkah 7.** PRD bilang *"Alertmanager consume topic `alert-events`"*.
> **Alertmanager tidak bisa consume Kafka.** Dia tidak punya Kafka consumer sama sekali — satu-satunya cara memasukkan alert ke Alertmanager adalah HTTP `POST /api/v2/alerts`. Ini akan langsung ketahuan kalau ada interviewer yang pernah pakai Prometheus stack.
>
> **Lalu kenapa tetap pakai Alertmanager?** Karena dia mengurus 4 hal yang sayang kalau kamu tulis sendiri:
> - **Grouping** — 50 alert dari service yang sama digabung jadi 1 notifikasi
> - **Deduplication** — alert identik yang dikirim berkali-kali hanya diteruskan sekali
> - **Inhibition** — kalau alert "service down" aktif, alert "error rate tinggi" untuk service itu ditahan (percuma, sudah jelas penyebabnya)
> - **Silencing** — bisa dimatikan sementara saat maintenance
>
> **Flow yang benar:**
> `Consumer 2` → `POST /api/v2/alerts` ke Alertmanager → Alertmanager grouping/dedup → webhook ke `Alert Dispatcher` → ClickUp/Email/Slack
>
> Topic `alert-events` **tetap dipakai**, tapi fungsinya jadi audit trail + replay, bukan jalur utama ke Alertmanager.

- [ ] Update PRD Section 3.2 langkah 7 supaya sesuai flow di atas
- [ ] Di Consumer 2, tambah `notifier/alertmanager.go` — POST ke `http://alertmanager:9093/api/v2/alerts` dengan payload:
  ```json
  [{
    "labels": {
      "alertname": "HighErrorRate",
      "rule_id": "rule-001",
      "project_id": "project-sabh",
      "service": "sabh-api",
      "severity": "critical"
    },
    "annotations": {
      "summary": "10 ERROR dalam 5 menit",
      "sample_trace_id": "4bf92f3577b34da6a3ce929d0e0e4736"
    },
    "startsAt": "2026-07-23T10:00:00Z"
  }]
  ```
  *Catatan:* `labels` adalah identitas alert — Alertmanager menganggap dua alert dengan set label identik sebagai alert yang sama (inilah mekanisme dedup-nya). `annotations` isinya cuma informasi tampilan, tidak ikut dihitung.
- [ ] Buat `config/alertmanager/alertmanager.yml`:
  ```yaml
  global:
    resolve_timeout: 5m
  route:
    receiver: 'logpilot-dispatcher'
    group_by: ['project_id', 'service', 'alertname']
    group_wait: 10s        # tunggu 10s, siapa tau ada alert serumpun menyusul
    group_interval: 5m     # jeda antar notifikasi untuk grup yang sama
    repeat_interval: 4h    # kirim ulang kalau masih aktif
  receivers:
    - name: 'logpilot-dispatcher'
      webhook_configs:
        - url: 'http://alert-dispatcher:9090/webhook'
          send_resolved: true
  ```
- [ ] Update docker-compose untuk mount config ini ke alertmanager container
- [ ] Restart alertmanager, verifikasi config loaded di UI `http://localhost:9093`
- [ ] Test manual: `curl -XPOST http://localhost:9093/api/v2/alerts -d '[...]'` → alert muncul di UI Alertmanager
- [ ] **Pikirkan dan tulis jawabannya di catatan:** cooldown di Redis (Consumer 2) dan `group_interval`/`repeat_interval` (Alertmanager) sama-sama menahan notifikasi. Apakah ini dobel? Yang mana yang sebaiknya jadi sumber kebenaran? *(Ini pertanyaan interview yang bagus — jawaban singkatnya: cooldown mencegah alert dibuat berulang di level rule, Alertmanager mencegah notifikasi berulang di level pengiriman. Beda lapisan, tapi kamu harus bisa jelaskan kenapa keduanya perlu.)*
- [ ] Commit: `chore: alertmanager config with grouping and dedup`

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

## 🎛️ PHASE 4 — Management API (Go): Control Plane
*Target: Bisa manage projects, API keys, dan alert rules via REST API.*

> 🔧 **KOREKSI v3.1** — phase ini sebelumnya memakai Laravel. Sekarang Go, di service `services/api` yang nanti juga memuat Query API (`S-1`). Rancangannya ada di PRD §4.6.

### Chunk 4.1 — Setup service `api`
- [ ] Buat `services/api/` sebagai Go module baru dan daftarkan di `go.work`
- [ ] Dua listener: `:8081` untuk `/api/*` (publik) dan `:8082` untuk `/internal/*`
- [ ] `/health`, `/ready`, `/metrics`, konfigurasi dari env, logging JSON, graceful shutdown (pola yang sama dengan Ingestor)
- [ ] Koneksi PostgreSQL (`pgx`) dan Redis; `/ready` gagal kalau PostgreSQL tidak terjangkau
- [ ] Test: `curl localhost:8081/health` dan `curl localhost:8082/health` sama-sama 200
- [ ] Commit: `chore(api): service skeleton with public and internal listeners`

> **📖 Pahami dulu:** kenapa `/internal/*` diberi port sendiri, bukan sekadar prefix path. Jawabannya ada di PRD §4.6 bagian "Dua listener".

### Chunk 4.2 — Database Migration: Users, Sessions & Projects
- [ ] Pilih satu alat migrasi (goose atau golang-migrate) dan tulis alasannya satu paragraf di `docs/DESIGN-DECISIONS.md`
- [ ] `migrations/postgres/001_...sql`: tabel `users` (id, name, email unik, password, created_at)
- [ ] Tabel `sessions` (id, user_id FK, token_hash unik, expires_at, created_at)
- [ ] Tabel `projects`:
  ```
  id, user_id (FK), name, slug (unique), created_at, updated_at
  ```
- [ ] Jalankan migrasi, verifikasi tabel ada; jalankan sekali lagi dan pastikan tidak ada yang berubah
- [ ] Commit: `chore(api): users, sessions and projects migration`

### Chunk 4.3 — Database Migration: API Keys & Alert Rules
- [ ] Migrasi untuk `api_keys`:
  ```
  id, project_id (FK), name, key_hash, last_used_at, created_at
  ```
- [ ] Migrasi untuk `alert_rules`:
  ```
  id, project_id (FK), name, condition (JSON), threshold, window_seconds, cooldown_seconds, is_active (bool)
  ```
- [ ] Migrasi untuk `notification_configs` dan `alert_logs` sesuai PRD
- [ ] Jalankan migrasi, verifikasi semua tabel ada
- [ ] Commit: `chore(api): api_keys, alert_rules, notification_configs migrations`

### Chunk 4.4 — Auth Endpoints
- [ ] `POST /api/auth/register`: validasi email unik, hash password dengan bcrypt, return user
- [ ] `POST /api/auth/login`: cek password, buat session token 32 byte dari `crypto/rand`, simpan **SHA-256**-nya di `sessions` dengan `expires_at`, return token sekali
- [ ] Middleware sesi: hash token dari header, cari di `sessions`, tolak kalau tidak ada atau kedaluwarsa, taruh `user_id` di context
- [ ] `POST /api/auth/logout`: hapus baris sesi
- [ ] Test dengan curl: register → login → dapat token → logout → token yang sama ditolak
- [ ] Commit: `feat(api): register, login, logout and session middleware`

> **❓ Bisa jawab ini?** Kenapa session token disimpan sebagai hash, padahal ia acak dan bukan password? Kenapa endpoint management memakai sesi yang dicek ke database, sedangkan jalur baca nanti memakai JWT (PRD §4.6 "Sesi User", §4.8)?

### Chunk 4.5 — Project Endpoints
- [ ] `POST /api/projects`: validasi nama, generate slug dari nama, simpan dengan `user_id` dari context
- [ ] `GET /api/projects`: return hanya project milik user yang sedang login
- [ ] Semua route `/api/projects*` berada di belakang middleware sesi
- [ ] Test: user A tidak melihat project user B
- [ ] Commit: `feat(api): project endpoints`

### Chunk 4.6 — API Key Management
- [ ] `POST /api/projects/{id}/api-keys` dan `DELETE /api/projects/{id}/api-keys/{keyId}`
- [ ] Saat membuat key:
  - Generate plain key: `logpilot_` + random 32 char (**pakai CSPRNG** — `crypto/rand`, jangan `math/rand`)
  - Simpan `key_hash` = **SHA-256** ke PostgreSQL
  - 🔧 **REVISI** — Seed ke Redis pakai **hash**, bukan plain: `SET api_key:{sha256(plain_key)} {project_id}`
  - Return plain key ke user (HANYA sekali, tidak bisa dilihat lagi)
- [ ] 🔧 **REVISI** — Update `middleware/auth.go` di Ingestor (Chunk 1.4): hash dulu key dari header, baru lookup ke Redis
- [ ] Saat mencabut key: hapus dari PostgreSQL + hapus dari Redis (pakai hash sebagai key juga)
- [ ] Commit: `feat(api): api key generate and revoke`

> **📖 Kenapa direvisi:** versi lama simpan hash di PostgreSQL tapi **plaintext di Redis**. Kalau Redis bocor (dan Redis paling sering di-expose tanpa password di setup dev), semua API key semua tenant bocor sekaligus. Hash di PostgreSQL jadi tidak ada gunanya — kamu menutup pintu depan tapi jendelanya dibiarkan terbuka.
>
> **Kenapa SHA-256, bukan bcrypt?** bcrypt sengaja dibuat lambat (~100ms) untuk melawan brute-force pada *password* yang dipilih manusia dan biasanya lemah. API key kamu 32 karakter random dari CSPRNG — ruang tebakannya sudah terlalu besar untuk di-brute-force, jadi tidak butuh perlambatan itu. Dan di ingestor, fungsi ini jalan **setiap request** — bcrypt akan membunuh target 1.000 req/detik kamu. SHA-256 (~1 mikrodetik) adalah pilihan yang tepat di sini.
>
> *Ini pertanyaan interview yang sering muncul: "kenapa password pakai bcrypt tapi API key pakai SHA-256?" Sekarang kamu punya jawabannya.*

### Chunk 4.7 — Alert Rule Endpoints
- [ ] Endpoint alert rule: list, buat, ubah, hapus (`/api/projects/{id}/alert-rules`)
- [ ] Validasi: threshold dan window_seconds harus positif integer
- [ ] `condition` disimpan sebagai JSON column
- [ ] Commit: `feat(api): alert rule CRUD endpoints`

### Chunk 4.8 — Notification Config Endpoint
- [ ] Endpoint notification config: ubah dan baca
- [ ] `PUT /api/projects/{id}/notifications` — upsert config
- [ ] `GET /internal/projects/{id}/notifications` — endpoint untuk Alert Dispatcher, di listener internal `:8082` (tanpa sesi user, pakai internal secret header)
- [ ] Commit: `feat(api): notification config endpoints`

### Chunk 4.9 — Update Consumer Alert: Baca Rules dari DB
- [ ] Kembali ke `services/consumer-alert/`
- [ ] Update `evaluator/cache.go`: ganti hardcode rules dengan HTTP call ke `GET /internal/alert-rules/{project_id}`
  - Atau langsung query PostgreSQL — pilih yang lebih simple dulu
- [ ] Refresh tiap 30 detik (sudah ada skeleton dari Chunk 3.2)
- [ ] Test: buat rule via Management API → dalam 30 detik rule aktif di Consumer 2
- [ ] Commit: `feat(consumer-alert): load alert rules from management api`

### Chunk 4.10 — Update Alert Dispatcher: Baca Notif Config dari DB
- [ ] Kembali ke `services/alert-dispatcher/`
- [ ] Saat terima webhook dari Alertmanager: query `GET /internal/projects/{id}/notifications`
- [ ] Gunakan config tersebut untuk ClickUp list_id, email recipients, dll
- [ ] Commit: `feat(alert-dispatcher): load notification config from management api`

---

## 🖥️ PHASE 5 — Next.js Frontend
*Target: UI sederhana untuk manage project, API keys, dan alert rules.*

### Chunk 5.1 — Setup Next.js Project
- [ ] Di `services/frontend/`: `npx create-next-app@latest . --typescript --tailwind`
- [ ] Test: `npm run dev` → bisa akses `localhost:3000`
- [ ] Install axios atau fetch wrapper: `npm install axios`
- [ ] Setup base API URL ke service `api` (`http://localhost:8081`)
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
- [ ] Buat `Dockerfile` untuk service `api`
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

### Chunk 7.6 — Blog Post: Micro-batch Internals
- [ ] Tulis artikel: "How LogPilot's Micro-batch Ingestion Works"
- [ ] Cover: tradeoff latency vs throughput di berbagai batch size
- [ ] Include benchmark results dari Chunk 2.8 dan 8.4 (grafik atau tabel)
- [ ] Publish di dev.to atau medium
- [ ] Link di README dan GitHub profile

---

## 🔧 PHASE R — Revisi & Peningkatan (Hasil Review PRD)

*Semua penambahan hasil review ada di sini, dikelompokkan per level knowledge.*

**Cara pakai phase ini:**

- Setiap chunk punya blok **📖 Pahami dulu**. Baca sampai kamu bisa **menjelaskan ulang pakai kata-katamu sendiri** sebelum mulai ngoding. Kalau belum bisa, jangan lanjut — nanti kamu cuma copy-paste dan pas ditanya interviewer kamu blank.
- Kerjakan berurutan **Junior → Intermediate → Senior**. Level atas mengandalkan konsep dari level bawah.
- 🔥 = paling berdampak untuk interview. Kalau waktumu sempit, kerjakan yang ini duluan.
- Setiap chunk ada **"Bisa jawab ini?"** — pertanyaan yang harus bisa kamu jawab setelah selesai. Ini bahan latihan interview sekaligus cara mengecek kamu benar-benar paham.

---

# 🟢 LEVEL JUNIOR

*Konsepnya sederhana, implementasinya kebanyakan nambah field atau nambah validasi. Target: 1-2 minggu part-time.*

---

### Chunk R-J1 — Trace Context: trace_id, span_id, user_id 🔥
*Ini jawaban langsung untuk pertanyaan interview kemarin.*

> **📖 Pahami dulu — 5 hal:**
>
> **1. Masalahnya apa.** Satu request user bisa menghasilkan 8 baris log ("mulai checkout", "cek stok", "panggil payment", "timeout", "rollback", dst). Kalau 100.000 user request bersamaan, kamu punya 800.000 baris log yang teracak jadi satu. Tanpa penanda, mustahil tahu 8 baris mana yang milik satu orang.
>
> **2. trace_id itu apa.** Anggap saja nomor resi. Request masuk → dibuatkan satu ID acak → ID itu ditempel di **semua** log yang lahir dari request tersebut. Mau lewat 5 service pun, ID-nya sama. Query `WHERE trace_id = 'abc...'` → kamu dapat cerita utuh satu request, urut waktu.
>
> **3. Ini yang paling sering salah: trace_id TIDAK BOLEH dibuat oleh LogPilot.**
> Kalau Ingestor yang generate, sudah telat. Ingestor menerima log **satu per satu** dan tidak punya cara tahu bahwa log ke-3 dan log ke-7 itu berasal dari request yang sama — informasi itu sudah hilang di sisi aplikasi. trace_id harus lahir di **middleware paling awal aplikasi pengirim**, lalu dibawa ke mana-mana.
> *Kalau interviewer tanya "di mana trace_id dibuat?" dan kamu jawab "di ingestor", itu jawaban yang salah.*
>
> **4. span_id dan parent_span_id.** trace_id cuma bilang "request ini bermasalah". Dia tidak bilang **di titik mana**. Untuk itu setiap potongan kerja diberi `span_id`, dan mencatat siapa yang memanggilnya lewat `parent_span_id`. Susunannya jadi pohon:
> ```
> trace_id: 4bf92f35...
>   span A (POST /checkout)          parent: -
>     span B (query stok DB)         parent: A     ✅ 12ms
>     span C (call payment gateway)  parent: A     ❌ timeout 30s  ← ini biangnya
> ```
> Sekarang kamu tidak cuma tahu "checkout gagal", tapi "checkout gagal **di panggilan ke payment gateway**".
>
> **5. Format standar (W3C Trace Context).** Jangan bikin format sendiri. Standarnya:
> - `trace_id` = 16 byte → **32 karakter hex**, contoh `4bf92f3577b34da6a3ce929d0e0e4736`
> - `span_id` = 8 byte → **16 karakter hex**, contoh `00f067aa0ba902b7`
> - Dibawa antar-service lewat header `traceparent: 00-{trace_id}-{span_id}-01`
>
> Ikut standar ini artinya OpenTelemetry, Jaeger, Grafana Tempo, dan Datadog semuanya bisa baca data kamu tanpa konversi. Bikin format sendiri = mengunci diri sendiri.
>
> **6. trace_id ≠ identitas.** trace_id menjawab *"request mana"*, bukan *"siapa"*. Untuk pertanyaan "user mana saja yang kena?" kamu butuh `user_id` dan `session_id` sebagai field tersendiri.

- [ ] Update struct `IngestRequest` di `handler/ingest.go` — tambah field (semua optional):
  ```go
  TraceID      string `json:"trace_id"`
  SpanID       string `json:"span_id"`
  ParentSpanID string `json:"parent_span_id"`
  UserID       string `json:"user_id"`
  SessionID    string `json:"session_id"`
  ```
- [ ] Tambah validasi format di Layer 3 (Chunk 1.6):
  - Kalau `trace_id` diisi → wajib 32 karakter hex. Kalau tidak → 400 dengan pesan yang jelas
  - Kalau `span_id` diisi → wajib 16 karakter hex
  - Kalau kosong → **biarkan kosong, jangan generate sendiri** (lihat poin 3 di atas)
- [ ] Tambah Prometheus counter `logs_without_trace_id_total{service}` — supaya nanti kamu tahu aplikasi mana yang belum pakai SDK
- [ ] Update struct `EnrichedLog` di `enricher/enricher.go` dan JSON yang dikirim ke Kafka
- [ ] Update `handler/ingest_test.go`: test case trace_id valid, trace_id salah panjang, trace_id bukan hex, trace_id kosong
- [ ] Commit: `feat(ingestor): add w3c trace context fields`

> **❓ Bisa jawab ini?**
> - Kenapa trace_id tidak boleh di-generate di ingestor?
> - Apa bedanya trace_id, span_id, dan user_id? Kapan pakai yang mana?
> - Kalau satu request lewat 4 service, ada berapa trace_id dan berapa span_id?

---

### Chunk R-J2 — Migrasi Skema ClickHouse 🔥
> 🔧 **KOREKSI v3** — DDL di bawah masih versi v2.0. Pakai DDL di PRD §4.3: `PARTITION BY toDate(timestamp)`, `TTL toDateTime(timestamp) + INTERVAL 30 DAY` (versi di bawah gagal, lihat `PROBLEM.md` Issue #4), `fingerprint String`, kolom `kafka_partition`/`kafka_offset`, dan dua `SETTINGS`.

> **📖 Pahami dulu — 4 hal:**
>
> **1. Ada bug silent data loss di skema sekarang.** Payload ingest kamu menerima `metadata` bebas isinya apa saja (`user_id`, `host`, `region`, dll). Tapi tabel ClickHouse cuma punya kolom `trace_id` dan `host`. Semua field metadata lain **hilang begitu saja tanpa pesan error apa pun**. Ini jenis bug paling menyebalkan: tidak ada yang gagal, tidak ada log error, datanya cuma... tidak ada. Kamu baru sadar 3 bulan kemudian pas butuh.
>
> **2. `Map(String, String)` sebagai solusinya.** Satu kolom yang bisa menampung key-value bebas. Query-nya: `WHERE attributes['region'] = 'jkt'`. Field yang **sering dipakai untuk filter** (trace_id, user_id) tetap jadi kolom sendiri karena lebih cepat; sisanya masuk Map.
>
> **3. `LowCardinality(String)` — hemat besar, gratis.** ClickHouse itu columnar: satu kolom disimpan berurutan dalam satu file. Kolom `level` isinya cuma 5 nilai berbeda (`DEBUG`/`INFO`/`WARN`/`ERROR`/`FATAL`) tapi diulang jutaan kali. `LowCardinality` menyuruh ClickHouse menyimpan kamus `{1:"DEBUG", 2:"INFO", ...}` lalu isi kolomnya cukup angka. Hasilnya: ukuran disk turun drastis dan filter `WHERE level='ERROR'` jadi membandingkan integer, bukan string.
> **Aturannya:** pakai `LowCardinality` kalau nilai uniknya di bawah ~10.000. Untuk kolom yang nilainya hampir selalu unik (`message`, `trace_id`, `user_id`) **jangan** — kamusnya malah jadi lebih besar dari datanya.
>
> **4. Migration file, bukan ketik manual di CLI.** Chunk 2.1 kemarin kamu ketik `CREATE TABLE` langsung di CLI. Masalahnya: 3 bulan lagi kamu lupa skemanya versi berapa, dan orang lain yang clone repo tidak tahu harus jalankan apa. Simpan sebagai file bernomor di repo.

- [ ] Buat folder `migrations/clickhouse/`
- [ ] Pindahkan DDL yang sudah ada jadi `001_create_logs.sql` (dokumentasi keadaan sekarang)
- [ ] Tulis `002_add_trace_context.sql`. Karena datanya masih data test, **paling gampang DROP lalu CREATE ulang**:
  ```sql
  DROP TABLE IF EXISTS logpilot.logs;

  CREATE TABLE logpilot.logs (
    project_id      LowCardinality(String),
    service         LowCardinality(String),
    level           LowCardinality(String),
    message         String CODEC(ZSTD(3)),

    trace_id        String,
    span_id         String,
    parent_span_id  String,

    user_id         String,
    session_id      String,

    host            LowCardinality(String),
    ingestor_id     LowCardinality(String),
    attributes      Map(LowCardinality(String), String),

    timestamp       DateTime64(3),
    ingested_at     DateTime64(3)
  ) ENGINE = MergeTree()
  PARTITION BY toYYYYMM(timestamp)
  ORDER BY (project_id, service, level, timestamp)
  TTL timestamp + INTERVAL 30 DAY;
  ```
- [ ] Perhatikan `ORDER BY` berubah dari `(project_id, timestamp)` jadi `(project_id, service, level, timestamp)`.
  *Alasannya:* `ORDER BY` di ClickHouse itu **primary index**-nya, dan urutannya harus dari yang paling sering difilter dan paling sedikit nilai uniknya. Hampir semua query kamu bentuknya "log project X, service Y, level ERROR, jam sekian" — dengan urutan baru ClickHouse bisa memotong sebagian besar data sebelum menyentuh timestamp.
- [ ] `CODEC(ZSTD(3))` di kolom `message` — message itu teks panjang dan paling makan disk. ZSTD level 3 = rasio kompresi bagus dengan CPU yang masih murah.
- [ ] Buat script `migrations/run.sh` yang menjalankan semua file `.sql` berurutan
- [ ] Verifikasi: `DESCRIBE TABLE logpilot.logs` — pastikan semua kolom baru ada
- [ ] Test Map: insert 1 baris manual dengan `attributes = {'region':'jkt','pod':'ing-3'}`, lalu `SELECT attributes['region'] FROM logs`
- [ ] Commit: `chore(clickhouse): trace context, map attributes, lowcardinality`

> **❓ Bisa jawab ini?**
> - Kenapa `level` pakai LowCardinality tapi `message` tidak?
> - Apa efeknya kalau `ORDER BY` diurutkan terbalik (`timestamp` duluan)?
> - Apa itu silent data loss dan kenapa lebih bahaya dari error yang jelas?

---

### Chunk R-J3 — Enrichment: Pisahkan metadata ke attributes

> **📖 Pahami dulu:** sekarang aplikasi mengirim semuanya di dalam `metadata`. Yang perlu kamu lakukan adalah **promosi**: field yang sering dipakai filter diangkat jadi kolom sendiri, sisanya masuk `attributes`. Ini keputusan desain yang harus kamu bisa pertanggungjawabkan — kolom sendiri = query cepat tapi skema kaku; Map = fleksibel tapi query lebih lambat.

- [ ] Di `enricher/enricher.go`, buat daftar field yang "dipromosikan": `trace_id`, `span_id`, `parent_span_id`, `user_id`, `session_id`, `host`
- [ ] Field itu diambil dari `metadata` (kalau dikirim di sana) atau dari root payload, lalu **dikeluarkan** dari map
- [ ] Sisa `metadata` masuk ke `attributes`
- [ ] Batasi: maksimal 32 key di `attributes`, panjang key maks 128 char, panjang value maks 1024 char. Yang lewat → dipotong, dan catat metric `attributes_truncated_total`
  *Kenapa:* tanpa batas, satu aplikasi yang salah kirim seluruh request body sebagai metadata bisa meledakkan penyimpanan kamu.
- [ ] Unit test: metadata dengan field promosi, metadata kosong, metadata dengan 100 key
- [ ] Commit: `feat(ingestor): promote known fields, rest to attributes map`

---

### Chunk R-J4 — Endpoint Batch Ingest

> **📖 Pahami dulu:** PRD bilang endpoint menerima *"satu atau batch log"*, tapi schema-nya cuma satu objek. Ini harus diselesaikan sekarang karena SDK nanti (`R-I3`) akan mengirim batch — kalau tidak, satu HTTP request per baris log, dan overhead TCP + TLS handshake-nya jauh lebih mahal daripada log-nya sendiri.
>
> **Konsep baru: partial success.** Kalau kamu kirim 500 log dan 3 di antaranya invalid, apa yang kamu balas? 200 (bohong, 3 gagal) atau 400 (bohong juga, 497 berhasil)? Jawabannya **207 Multi-Status** — berhasil sebagian, dan body-nya menjelaskan mana yang gagal.

- [ ] Terima dua bentuk payload di `/v1/ingest`: satu objek (kompatibel dengan yang sudah ada) atau `{"logs": [ ... ]}`
- [ ] Batasi: maksimal 1.000 log per request, maksimal body 5 MB
- [ ] Dukung `Content-Encoding: gzip` (Fiber punya middleware `compress`)
- [ ] Response:
  - Semua valid → `202 Accepted`
  - Sebagian gagal → `207 Multi-Status` + body `{"accepted": 497, "rejected": 3, "errors": [{"index": 12, "field": "level", "reason": "invalid value"}]}`
  - Semua gagal → `400 Bad Request`
- [ ] Rate limit dihitung **per log**, bukan per request. Kalau tidak, orang bisa lolos limit dengan kirim 1 request berisi 1.000 log
- [ ] Update unit test
- [ ] Commit: `feat(ingestor): batch ingest with partial success`

---

### Chunk R-J5 — Clock Skew Guard

> **📖 Pahami dulu:** `timestamp` dikirim oleh **client**, dan jam client tidak bisa dipercaya. Kalau ada server yang NTP-nya mati dan jamnya mundur 3 bulan, ClickHouse akan menulis ke partisi bulan Maret. Efeknya: log itu tidak akan pernah muncul di dashboard "1 jam terakhir", dan kalau jamnya maju setahun, TTL 30 hari tidak akan menghapusnya selama setahun.
>
> **Aturannya:** `timestamp` (waktu kejadian menurut client) tetap disimpan apa adanya untuk konteks, tapi **`ingested_at` (waktu server) adalah sumber kebenaran** untuk urutan dan retensi.

- [ ] Di Layer 3 validasi, hitung selisih `timestamp` dengan waktu server
- [ ] Kalau `timestamp` lebih dari **1 jam di masa depan** → tolak 400 (jam yang maju hampir selalu salah konfigurasi)
- [ ] Kalau `timestamp` lebih dari **7 hari di masa lalu** → terima, tapi **clamp** dan catat aslinya di `attributes['original_timestamp']`
- [ ] Catat metric `clock_skew_seconds` sebagai histogram, dan `logs_clock_skewed_total{service}`
- [ ] Tulis di dokumentasi: `timestamp` = waktu client, `ingested_at` = waktu server, dan Grafana default pakai yang mana
- [ ] Commit: `feat(ingestor): clock skew detection and clamping`

> **❓ Bisa jawab ini?** Kalau log dari 2 server dengan jam berbeda 5 detik digabung dan diurutkan, urutannya masih benar tidak? Apa solusinya di sistem terdistribusi?

---

### Chunk R-J6 — Prometheus: Ada /metrics tapi Tidak Ada yang Baca

> **📖 Pahami dulu:** Ingestor kamu expose `GET /metrics` (PRD Section 4.1), dan docker-compose kamu punya `alertmanager` — tapi **tidak ada Prometheus**. Jadi tidak ada satu pun yang men-scrape metrics itu. Ini seperti memasang speedometer tapi tidak ada kacanya.
>
> **Konsep: pull vs push.** Prometheus itu **pull-based** — dia yang datang mengetuk `/metrics` setiap 15 detik, bukan aplikasi yang mengirim. Kelebihannya: kalau target mati, Prometheus langsung tahu (scrape gagal) tanpa perlu heartbeat terpisah.
>
> **4 tipe metric yang perlu kamu bedakan:**
> - **Counter** — hanya naik, tidak pernah turun. Contoh: total request. Dipakai untuk hitung rate.
> - **Gauge** — bisa naik-turun. Contoh: jumlah goroutine, ukuran batch saat ini.
> - **Histogram** — distribusi nilai dalam bucket. Contoh: latency. **Ini yang kamu butuhkan untuk p95/p99.**
> - **Summary** — mirip histogram tapi kuantil dihitung di sisi aplikasi. Jarang dipakai, agregasinya bermasalah.

- [ ] Tambah service `prometheus` (prom/prometheus:v2.48.0) ke docker-compose
- [ ] Buat `config/prometheus/prometheus.yml` — scrape target: ingestor, consumer-storage, consumer-alert, alert-dispatcher, interval 15s
- [ ] Install `github.com/prometheus/client_golang` di semua service Go
- [ ] Metrics minimal di **Ingestor**:
  - `logpilot_ingest_requests_total{project_id, status}` — Counter
  - `logpilot_ingest_logs_total{project_id, level}` — Counter
  - `logpilot_ingest_duration_seconds` — Histogram (ini sumber angka p50/p95/p99 kamu)
  - `logpilot_kafka_publish_errors_total` — Counter
- [ ] Metrics minimal di **Consumer Storage**:
  - `logpilot_batch_size` — Histogram
  - `logpilot_clickhouse_insert_duration_seconds` — Histogram
  - `logpilot_clickhouse_insert_errors_total` — Counter
- [ ] Tambah Prometheus sebagai datasource di Grafana
- [ ] Buat dashboard "LogPilot — System Health" dengan panel: request rate, error rate, p95 latency
- [ ] Commit: `feat: prometheus metrics across all go services`

> **❓ Bisa jawab ini?** Kenapa latency harus Histogram dan bukan Gauge? Kenapa rata-rata latency itu metrik yang menyesatkan?

---

### Chunk R-J7 — Span: Potongan Kerja Berdurasi

> **📖 Pahami dulu — apa itu span:**
>
> **1. span itu satu potongan kerja, dan punya durasi.** `trace_id` (R-J1) bilang "log-log ini semua milik satu request". Tapi dia tidak bilang **berapa lama** tiap langkah di dalamnya makan waktu. Span mengisi itu: span = satu langkah kerja yang punya **waktu mulai** dan **waktu selesai**. "Query stok ke DB" itu satu span. "Panggil payment gateway" itu span lain.
>
> **2. Satu request = satu trace, banyak span.** Request `POST /checkout` menghasilkan satu `trace_id`, tapi di dalamnya ada beberapa span berurutan atau bersarang: validasi input, cek stok, panggil payment, simpan order. Semuanya bernaung di bawah satu trace_id yang sama.
>
> **3. span_id dan parent_span_id membentuk pohon.** Tiap span punya `span_id` sendiri. Kalau span itu dipanggil dari dalam span lain, dia mencatat `parent_span_id` = span_id pemanggilnya. Span paling luar (span pertama, yang menerima request) tidak punya parent.
> ```
> trace_id: 4bf92f35...
>   span A "POST /checkout"        span_id=A  parent=-        durasi: 820ms
>     span B "cek stok DB"         span_id=B  parent=A        durasi: 12ms
>     span C "call payment"        span_id=C  parent=A        durasi: 790ms  ← paling lama
> ```
> Dari susunan ini kamu bisa langsung lihat: dari total 820ms, 790ms-nya habis di panggilan payment. Tanpa durasi per span, kamu cuma tahu "checkout lambat", tidak tahu **di mana**-nya lambat.
>
> **4. Kenapa ini harus dipahami sebelum R-I3 (SDK).** SDK nanti akan membuat span sungguhan dengan waktu mulai/selesai betulan, dan mempropagasikannya lewat header antar service. Kalau konsep span belum jelas di kepala kamu duluan, R-I3 jadi sekadar copy kode tanpa tahu apa yang sedang dibangun.

- [ ] Tambah field durasi ke pemahaman log: `start_time`, `end_time` (atau `duration_ms` yang dihitung dari keduanya)
- [ ] Buat helper sederhana di SDK: fungsi untuk menandai "mulai langkah" (catat waktu sekarang + generate span_id) dan "selesai langkah" (hitung selisih waktu, tulis durasi)
- [ ] Uji dengan satu request checkout: buat 3-4 span dalam satu `trace_id` (misal: terima request → cek stok → panggil payment → simpan order), pastikan semuanya punya `trace_id` yang sama, `span_id` masing-masing berbeda, dan `parent_span_id` yang benar
- [ ] Verifikasi durasi tiap span masuk akal (span anak tidak boleh lebih lama dari span induknya)
- [ ] Commit: `feat: span duration tracking for trace context`

> **❓ Bisa jawab ini?**
> - Kalau satu trace_id punya 4 span, apa yang membedakan satu span dari span lainnya selain namanya?
> - Kenapa span butuh waktu mulai DAN selesai, bukan cuma satu timestamp?
> - Kalau span C adalah anak dari span A, apa yang seharusnya benar soal durasi keduanya?

---

# 🟡 LEVEL INTERMEDIATE

*Di sini kamu mulai menyentuh konsep distributed systems yang sesungguhnya. Setiap chunk mengandung satu konsep yang sering ditanya interviewer. Target: 3-4 minggu part-time.*

---

### Chunk R-I1 — Bloom Filter Skip Index untuk trace_id 🔥

> **📖 Pahami dulu — kenapa trace_id butuh perlakuan khusus:**
>
> **1. `ORDER BY` di ClickHouse bukan index biasa.** Di MySQL/PostgreSQL, index itu B-tree yang menunjuk ke baris tertentu. Di ClickHouse, `ORDER BY` menghasilkan **sparse primary index**: data disimpan berurutan lalu dibagi jadi blok 8.192 baris (disebut **granule**), dan ClickHouse hanya mencatat nilai pertama tiap granule. Query `WHERE project_id='x' AND timestamp > y` cepat karena datanya memang tersusun urut begitu — ClickHouse tinggal melompat ke granule yang relevan.
>
> **2. Masalahnya, `trace_id` tidak ikut urutan itu.** trace_id adalah nilai acak yang tersebar merata di seluruh tabel. Query `WHERE trace_id = 'abc...'` tidak bisa memanfaatkan `ORDER BY` sama sekali → **ClickHouse membaca seluruh partisi**. Kalau satu bulan berisi 500 juta baris, satu pencarian trace bisa memakan detik-detik yang mahal. Padahal ini justru query yang paling sering kamu pakai saat debugging.
>
> **3. Bloom filter menyelesaikannya.** Bloom filter adalah struktur data probabilistik yang menjawab satu pertanyaan: *"apakah nilai X mungkin ada di kumpulan ini?"* dengan dua kemungkinan jawaban:
> - **"Pasti tidak ada"** → 100% akurat, boleh dilewati
> - **"Mungkin ada"** → bisa salah (false positive), harus dicek
>
> Yang penting: **false negative tidak mungkin terjadi**. Jadi kalau bloom filter bilang "tidak ada", kamu aman melewatinya.
>
> Cara kerjanya: nilai di-hash beberapa kali, tiap hash menyalakan satu bit di array bit. Untuk mengecek, hash lagi dan lihat apakah semua bit-nya menyala. Kalau ada satu saja yang mati → pasti belum pernah dimasukkan.
>
> **4. Di ClickHouse ini disebut "skip index".** Setiap granule punya bloom filter kecil. Saat query, ClickHouse cek dulu bloom filter tiap granule dan **melompati** granule yang pasti tidak punya trace_id itu. Dari membaca 500 juta baris, mungkin tinggal membaca 8.192 baris. Itu bedanya detik vs milidetik.
>
> **5. `tokenbf_v1` untuk pencarian teks.** Varian bloom filter yang memecah string jadi token (kata). Berguna untuk `WHERE message LIKE '%timeout%'`.

- [ ] Tulis `migrations/clickhouse/003_add_skip_indexes.sql`:
  ```sql
  ALTER TABLE logpilot.logs
    ADD INDEX idx_trace trace_id TYPE bloom_filter(0.01) GRANULARITY 1;
  ALTER TABLE logpilot.logs
    ADD INDEX idx_user user_id TYPE bloom_filter(0.01) GRANULARITY 1;
  ALTER TABLE logpilot.logs
    ADD INDEX idx_msg message TYPE tokenbf_v1(32768, 3, 0) GRANULARITY 1;

  ALTER TABLE logpilot.logs MATERIALIZE INDEX idx_trace;
  ALTER TABLE logpilot.logs MATERIALIZE INDEX idx_user;
  ALTER TABLE logpilot.logs MATERIALIZE INDEX idx_msg;
  ```
  *Catatan:* `0.01` = target false positive rate 1%. Makin kecil = makin akurat tapi makin besar index-nya. `MATERIALIZE INDEX` diperlukan untuk membangun index atas data yang **sudah** ada — kalau tidak, index hanya berlaku untuk data baru.
- [ ] **Buktikan efeknya, jangan cuma percaya.** Isi tabel dengan minimal 5 juta baris dummy (pakai `generateRandom` atau log-generator), lalu:
  ```sql
  EXPLAIN indexes = 1
  SELECT * FROM logpilot.logs WHERE trace_id = '4bf92f35...';
  ```
  Catat `rows read` **sebelum** dan **sesudah** index dipasang. Simpan angkanya — ini bahan README dan bahan cerita interview.
- [ ] Bandingkan juga waktu eksekusi query-nya
- [ ] Tulis hasilnya di `docs/BENCHMARKS.md`
- [ ] Commit: `perf(clickhouse): bloom filter skip index for high-cardinality lookup`

> **❓ Bisa jawab ini?**
> - Kenapa `ORDER BY (project_id, ...)` tidak membantu query `WHERE trace_id = x`?
> - Apa itu false positive dan false negative di bloom filter? Kenapa yang satu boleh terjadi dan yang satu tidak?
> - Kalau bloom filter sebagus itu, kenapa tidak dipasang di semua kolom?

---

### Chunk R-I2 — Sliding Window Counter yang Benar (Bucketed) 🔥

> **📖 Pahami dulu — tiga cara menghitung "N kejadian dalam M detik":**
>
> **Cara 1: Fixed window** *(yang kamu implement di Chunk 3.3)*
> Satu counter dengan TTL. Window-nya kotak-kotak kaku: 10:00:00–10:04:59, lalu reset, lalu 10:05:00–10:09:59.
> - ✅ Sangat murah: 1 key, 1 operasi
> - ❌ **Masalah batas window.** Threshold 10 per 5 menit. Ada 9 error di 10:04:50 dan 9 error lagi di 10:05:10. Dalam rentang 20 detik terjadi 18 error — jelas insiden. Tapi counter reset di 10:05:00, jadi masing-masing window cuma berisi 9. **Alert tidak pernah trigger.** Kebalikannya juga bisa terjadi: alert palsu karena dua window bersebelahan.
>
> **Cara 2: Sorted set (sliding window akurat)**
> Simpan timestamp setiap kejadian di Redis sorted set. Sebelum menghitung, buang yang lebih tua dari window.
> ```
> ZADD              key {now_ms} {uuid}
> ZREMRANGEBYSCORE  key 0 {now_ms - window_ms}
> ZCARD             key
> ```
> - ✅ Akurat sempurna, window benar-benar bergeser
> - ❌ **Memory O(N).** Kalau ada 100.000 error dalam window, kamu menyimpan 100.000 entry di Redis **per rule**. Di skala yang jadi pertanyaan interview kamu kemarin, ini meledak.
>
> **Cara 3: Bucketed counter** ⭐ *yang kusarankan*
> Pecah window jadi potongan kecil, satu counter per potongan. Window 5 menit dengan bucket 10 detik = 30 bucket.
> ```
> key = alert_counter:{rule_id}:{floor(unix_now / 10)}
> ```
> Untuk menghitung total: jumlahkan 30 bucket terakhir dalam satu Redis pipeline.
> - ✅ **Memory konstan** — berapa pun event-nya, tetap 30 key kecil per rule
> - ✅ Window bergeser tiap 10 detik, bukan melompat tiap 5 menit
> - ⚠️ Akurasi ±1 bucket (±10 detik). Untuk alerting, ini sepenuhnya bisa diterima
>
> **Ini bukan teori.** Cara ini yang dipakai rate limiter dan alerting engine di produksi nyata. Kalau di interview kamu bisa menjelaskan ketiga cara ini beserta trade-off memory-nya, kamu sudah menjawab pertanyaan yang biasanya dilempar ke kandidat senior.

- [ ] Buat `counter/bucketed.go`
- [ ] Konstanta `bucketSeconds = 10`, tapi bikin bisa dikonfigurasi per rule
- [ ] Fungsi `IncrementAndCount(ctx, ruleID, projectID string, windowSeconds int) (int64, error)`:
  1. Hitung bucket sekarang: `now / bucketSeconds`
  2. INCR bucket sekarang, set TTL = `windowSeconds + bucketSeconds` **hanya kalau key baru** (Lua, sama seperti Chunk 3.3)
  3. Hitung daftar key untuk `windowSeconds / bucketSeconds` bucket terakhir
  4. `MGET` semua key sekaligus (**satu round-trip**, bukan 30)
  5. Jumlahkan, abaikan nilai nil
- [ ] Pakai **pipeline** go-redis untuk langkah 2 dan 4:
  ```go
  pipe := rdb.Pipeline()
  incr := pipe.Incr(ctx, currentKey)
  pipe.Expire(ctx, currentKey, ttl)
  vals := pipe.MGet(ctx, keys...)
  _, err := pipe.Exec(ctx)
  ```
  *Kenapa penting:* tanpa pipeline, 30 bucket = 30 round-trip ke Redis. Kalau tiap round-trip 0,5ms, itu 15ms **per log**. Di 1.000 log/detik kamu langsung tenggelam. Dengan pipeline: satu round-trip.
- [ ] **Test kasus batas window** — ini test paling penting:
  - Window 60 detik, bucket 10 detik, threshold 10
  - Kirim 9 event di detik ke-55, lalu 9 event di detik ke-65
  - Fixed window: tidak trigger (bug). Bucketed: **harus trigger**
- [ ] Ganti pemanggilan di Chunk 3.6 dari `Increment()` ke `IncrementAndCount()`
- [ ] Tulis perbandingan ketiga pendekatan di `docs/DESIGN-DECISIONS.md`
- [ ] Commit: `feat(consumer-alert): bucketed sliding window counter`

> **❓ Bisa jawab ini?**
> - Gambarkan kasus di mana fixed window gagal mendeteksi insiden nyata.
> - Kenapa sorted set tidak dipakai padahal paling akurat?
> - Kalau bucket diperkecil jadi 1 detik, apa yang membaik dan apa yang memburuk?

---

### Chunk R-I3 — SDK Go: Tempat trace_id Lahir 🔥

> **Keputusan yang sudah diambil:** SDK ini dibuat sendiri, BUKAN pakai SDK OTel — justru supaya kamu belajar propagasi, batching, dan backpressure dengan tangan sendiri. Tapi formatnya tetap ikut W3C Trace Context supaya kompatibel dengan tooling lain (Jaeger, Tempo, OTel Collector).
>
> **Prasyarat:** `R-J7` (konsep span berdurasi). Kalau belum bisa jelaskan span = potongan kerja dengan start/end time tanpa buka dokumennya, balik ke situ dulu.

> **📖 Pahami dulu:** SDK disebut di PRD Section 6.3 (*"tunjukkan kode SDK integration — hanya 5-10 baris"*) tapi **tidak ada di spec maupun di struktur repo**. Padahal ini komponen yang menyelesaikan pertanyaan interview kemarin — karena inilah tempat trace_id dilahirkan dan dipropagasi.
>
> **Konsep 1: `context.Context` sebagai pembawa.** Di Go, `context.Context` adalah cara standar membawa nilai yang mengikuti satu request dari awal sampai akhir. Middleware membuat trace_id, menaruhnya di context, lalu setiap fungsi di bawahnya bisa mengambilnya tanpa perlu diteruskan sebagai parameter di setiap pemanggilan.
>
> **Konsep 2: propagasi antar service.** Kalau service A memanggil service B, trace_id harus ikut. Caranya lewat HTTP header `traceparent` (standar W3C). Service B membaca header itu, memakai trace_id yang sama, tapi **membuat span_id baru** dengan `parent_span_id` = span_id milik A. Begitulah pohonnya terbentuk.
>
> **Konsep 3: pengiriman harus non-blocking.** Ini krusial. Kalau SDK mengirim log secara sinkron, maka setiap kali LogPilot lambat, **aplikasi user ikut lambat**. Tool observability yang bisa menjatuhkan aplikasi yang diamatinya adalah tool yang gagal. Solusinya: buffered channel + goroutine pengirim di belakang, dan kalau buffer penuh → **buang log-nya**, jangan blokir.
> *Prinsipnya: lebih baik kehilangan log daripada kehilangan request user.*

- [ ] Buat folder `sdk/go/` di monorepo, `go mod init github.com/{username}/logpilot/sdk/go`
- [ ] `trace.go`:
  - `NewTraceID() string` — 16 byte random → 32 hex
  - `NewSpanID() string` — 8 byte random → 16 hex
  - `ParseTraceparent(header string) (traceID, spanID string, ok bool)`
  - `FormatTraceparent(traceID, spanID string) string` → `00-{trace}-{span}-01`
- [ ] `context.go` — simpan/ambil trace context dari `context.Context`
- [ ] `span.go` — **span berdurasi sungguhan, bukan cuma meneruskan field trace** (lanjutan `R-J7`):
  - `StartSpan(ctx context.Context, name string) (context.Context, *Span)` — catat waktu mulai, generate span_id baru, parent diambil dari span yang sedang aktif di context
  - `(*Span) End()` — catat waktu selesai, hitung durasi, kirim sebagai log/event lewat sender
- [ ] `middleware.go` — middleware Fiber & `net/http`:
  - Baca header `traceparent`. Kalau ada → pakai trace_id-nya, buat span_id baru, parent = span_id dari header
  - Kalau tidak ada → **request ini adalah root**, buat trace_id baru
  - Simpan ke context, dan tulis balik ke response header untuk debugging
- [ ] `client.go` — HTTP client wrapper yang otomatis menyisipkan header `traceparent` saat memanggil service lain
- [ ] `logger.go` — API yang dipakai user:
  ```go
  log := logpilot.New(logpilot.Config{
      Endpoint: "http://localhost:8080",
      APIKey:   os.Getenv("LOGPILOT_KEY"),
      Service:  "sabh-api",
  })
  log.Error(ctx, "payment gateway timeout", logpilot.Fields{"order_id": "123"})
  ```
  trace_id, span_id, timestamp, service diisi otomatis dari context — user tidak perlu tahu
- [ ] `sender.go` — buffered channel (kapasitas 10.000) + goroutine yang flush tiap 1 detik atau 100 log:
  - Kalau channel penuh → **drop**, increment counter `logpilot_sdk_dropped_total`
  - `Close()` yang mem-flush sisa buffer saat aplikasi shutdown
  - Retry dengan backoff kalau ingestor error
- [ ] Buat `sdk/go/README.md` — contoh integrasi 5-10 baris (ini yang kamu tunjukkan saat demo)
- [ ] Update `demo/dummy-app` (Chunk 6.1) supaya pakai SDK ini, bukan HTTP call manual
- [ ] Commit: `feat(sdk): go sdk with w3c trace context propagation`

> **❓ Bisa jawab ini?**
> - Kalau SDK mengirim log secara sinkron dan ingestor down, apa yang terjadi pada aplikasi user?
> - Kenapa buffer penuh lebih baik dijawab dengan "buang log" daripada "tunggu"?
> - Bagaimana trace_id berpindah dari service A ke service B?

---

### Chunk R-I4 — Redis sebagai Single Point of Failure

> **📖 Pahami dulu:** autentikasi Ingestor 100% bergantung pada Redis. Redis restart → semua key hilang → **seluruh platform menolak semua log dari semua tenant**. Ini bukan skenario teoretis; Redis di-restart setiap kali kamu deploy ulang atau container-nya di-recreate.
>
> **Pola yang dipakai: cache-aside dengan fallback.** Cache miss bukan berarti "tidak ada", melainkan "belum tahu — tanya ke sumber kebenaran". PostgreSQL adalah sumber kebenaran; Redis hanya lapisan cepat di depannya.
>
> **Dua jebakan yang harus kamu antisipasi:**
> - **Thundering herd** — Redis kosong, 1.000 request masuk bersamaan, semuanya cache miss, semuanya menghantam PostgreSQL di detik yang sama → PostgreSQL ikut tumbang. Solusi: `singleflight` (di Go: `golang.org/x/sync/singleflight`) yang menggabungkan permintaan identik yang bersamaan menjadi satu query.
> - **Negative caching** — kalau seseorang menyerang dengan API key acak, setiap key salah = satu query PostgreSQL. Solusi: cache hasil "key ini tidak valid" juga, dengan TTL pendek (30 detik).

- [ ] Tambah endpoint internal di Management API: `GET /internal/api-keys/{hash}` → return `project_id` atau 404
- [ ] Update `middleware/auth.go` di Ingestor:
  1. Lookup Redis
  2. Kalau **hit** → lanjut
  3. Kalau **miss** → panggil internal endpoint Management API (dibungkus `singleflight`)
  4. Kalau ketemu → tulis balik ke Redis dengan TTL 1 jam, lanjut
  5. Kalau tidak ketemu → cache negatif `api_key_invalid:{hash}` TTL 30 detik, return 401
- [ ] Kalau **Redis-nya sendiri yang down** (bukan cache miss) → jangan langsung tolak semua. Pakai in-memory cache di dalam pod sebagai lapisan terakhir (TTL 5 menit), dan naikkan alarm
- [ ] Tambah metric: `auth_cache_hits_total`, `auth_cache_misses_total`, `auth_fallback_total`, `auth_redis_errors_total`
- [ ] **Test:** matikan Redis (`docker stop redis`), kirim log → harus tetap diterima lewat fallback. Nyalakan lagi → Redis terisi ulang otomatis
- [ ] Commit: `feat(ingestor): postgres fallback for auth cache miss`

> **❓ Bisa jawab ini?**
> - Apa itu thundering herd dan bagaimana singleflight mencegahnya?
> - Kenapa hasil "tidak valid" juga perlu di-cache?
> - Apa bedanya "cache miss" dan "cache down"? Kenapa harus ditangani berbeda?

---

### Chunk R-I5 — Delivery Semantics & Deduplikasi
> 🔧 **KOREKSI v3** — `non_replicated_deduplication_window` bawaannya **0** (mati), bukan "kecil". Dan token hanya menutup retry dalam proses yang sama; untuk crash + restart batas batch berubah sehingga tokennya ikut berubah. Dua skenario dan penanganannya ada di PRD §4.3 (Delivery Semantics).

> **📖 Pahami dulu — tiga jaminan pengiriman:**
>
> - **At-most-once** — commit offset **sebelum** memproses. Kalau proses mati di tengah, message hilang. Cepat, tapi ada kehilangan data.
> - **At-least-once** — commit offset **setelah** memproses. Kalau mati setelah INSERT berhasil tapi sebelum commit, Kafka mengirim ulang batch yang sama → **data duplikat**. *Ini yang kamu pakai sekarang.*
> - **Exactly-once** — tidak ada yang hilang, tidak ada yang dobel. Mahal dan tidak selalu mungkin lintas sistem.
>
> **Yang perlu kamu sadari:** PRD kamu di F9 mengklaim *"no data loss"*, dan itu benar. Tapi diam soal duplikat — padahal MergeTree **tidak melakukan deduplikasi apa pun**. Jadi setelah crash, log yang sama akan muncul dua kali di Grafana dan hitungan error kamu jadi salah.
>
> **Solusi praktis: idempotency, bukan exactly-once.** Kalau operasi yang sama diulang menghasilkan hasil yang sama, maka duplikasi jadi tidak berbahaya. ClickHouse punya `insert_deduplication_token`: kalau token yang sama dikirim lagi dalam window tertentu, INSERT-nya diabaikan diam-diam. Token yang deterministik = `topic:partition:offset_awal:offset_akhir` — batch yang sama selalu menghasilkan token yang sama.

- [ ] Pastikan `insert_deduplication_token` di Chunk 2.4 sudah terpasang
- [ ] Perbesar `non_replicated_deduplication_window` di ClickHouse (default kecil untuk tabel non-replicated):
  ```sql
  ALTER TABLE logpilot.logs MODIFY SETTING non_replicated_deduplication_window = 1000;
  ```
- [ ] **Uji crash beneran:** jalankan consumer, kirim 10.000 log, `kill -9` consumer di tengah batch, jalankan lagi, hitung `SELECT count() FROM logs` → harus tepat 10.000, bukan lebih
- [ ] Update PRD F9 jadi: *"No data loss (at-least-once) dengan deduplikasi di sisi ClickHouse — tidak ada log hilang dan tidak ada duplikat"*
- [ ] Tulis penjelasan trade-off ketiga semantics di `docs/DESIGN-DECISIONS.md`
- [ ] Commit: `feat(consumer-storage): idempotent insert with dedup token`

> **❓ Bisa jawab ini?** Kenapa kamu memilih at-least-once + idempotency daripada mengejar exactly-once? Apa harga yang dibayar exactly-once?

---

### Chunk R-I6 — Dead Letter Queue (Poison Message)

> **📖 Pahami dulu:** kalau ada satu message yang **selalu** gagal diproses — JSON rusak, timestamp di luar jangkauan `DateTime64`, field kepanjangan — dan kamu retry tanpa batas, maka:
> 1. Consumer terus mengulang message yang sama
> 2. Offset tidak pernah maju
> 3. **Seluruh partition itu berhenti bergerak selamanya**
> 4. Semua log lain di partition tersebut ikut macet, meski tidak ada masalah dengannya
>
> Message seperti ini namanya **poison message**. Pola standarnya: coba beberapa kali, kalau tetap gagal → pindahkan ke **Dead Letter Queue** dan lanjut. Prinsipnya: satu message rusak tidak boleh menyandera seluruh pipeline.

- [ ] Buat topic `raw-logs-dlq` (1 partisi, retensi 30 hari)
- [ ] Di consumer-storage: kalau batch gagal setelah N retry, coba pecah batch jadi per-message untuk mengisolasi yang rusak
- [ ] Message yang rusak → publish ke DLQ dengan header tambahan: `original_topic`, `original_partition`, `original_offset`, `error_reason`, `failed_at`, `retry_count`
- [ ] Sisa batch yang sehat → insert normal, commit offset
- [ ] Metric `logpilot_dlq_messages_total{reason}` + alert kalau ada isinya
- [ ] Buat CLI kecil `cmd/dlq-inspect` untuk membaca dan (kalau sudah diperbaiki) me-replay message dari DLQ
- [ ] **Test:** kirim log dengan timestamp `9999-12-31T23:59:59Z` → harus masuk DLQ, dan log lain tetap mengalir
- [ ] Commit: `feat(consumer-storage): dead letter queue for poison messages`

---

### Chunk R-I7 — Consumer Lag Monitoring 🔥

> **📖 Pahami dulu:** **consumer lag adalah metrik paling penting di seluruh sistem ini**, dan sekarang belum ada sama sekali.
>
> Lag = `offset terakhir di broker` − `offset yang sudah di-commit consumer`. Artinya: berapa banyak message yang sudah masuk tapi belum sempat diproses.
>
> - Lag stabil di angka kecil → sehat, consumer mengejar
> - Lag naik terus → consumer lebih lambat dari producer. Ini **peringatan dini**: sistem belum rusak, tapi akan rusak
> - Lag = 0 terus-menerus → bisa jadi consumer-nya mati (tidak ada yang bertambah) — perhatikan bersama metrik throughput
>
> **Kenapa ini pertanyaan interview:** *"Gimana kamu tahu sistem kamu sehat?"* Jawaban "saya lihat log error" itu jawaban junior. Jawaban yang tepat: *"saya monitor consumer lag dan turunannya — kalau lag naik konsisten, saya tahu ada masalah 10 menit sebelum user komplain."*
>
> **Turunannya yang lebih berguna:** lag dalam **satuan waktu**, bukan jumlah message. Lag 50.000 message itu tidak bermakna; "consumer tertinggal 4 menit" itu langsung bisa ditindaklanjuti.

- [ ] Tambah `kafka-exporter` (danielqsj/kafka-exporter) ke docker-compose
- [ ] Konfigurasi Prometheus untuk scrape exporter tersebut
- [ ] Tambah metric di masing-masing consumer: `logpilot_consumer_lag_seconds` = `now - timestamp message yang sedang diproses`
- [ ] Panel Grafana di dashboard "System Health":
  - Lag per consumer group per partition
  - Lag dalam detik
  - Throughput: message masuk/detik vs message diproses/detik
- [ ] Aturan alert (di Prometheus, untuk LogPilot sendiri):
  - `lag > 100.000 message` selama 5 menit → warning
  - `lag_seconds > 300` → critical
  - Consumer group hilang dari metric → critical (consumer-nya mati)
- [ ] **Test:** hentikan consumer-storage 2 menit sambil log-generator jalan, lihat grafik lag naik, nyalakan lagi, lihat lag turun. **Rekam layarnya** — ini materi demo yang sangat kuat
- [ ] Commit: `feat: consumer lag monitoring and alerting`

> **❓ Bisa jawab ini?** Kalau lag naik terus, apa saja kemungkinan penyebabnya dan bagaimana urutan kamu mendiagnosisnya?

---

### Chunk R-I8 — Konfigurasi Durability Kafka

> **📖 Pahami dulu:** PRD mengklaim Kafka memberi *"guarantee no data loss"*, tapi tidak ada satu pun konfigurasi yang menjamin itu. Klaim tanpa konfigurasi adalah klaim kosong, dan interviewer yang paham Kafka akan menanyakannya.
>
> **`acks` — kapan producer dianggap berhasil:**
> - `acks=0` — kirim dan lupakan. Tercepat, tapi kalau broker mati saat itu juga, log hilang tanpa jejak
> - `acks=1` — tunggu leader menulis. Kalau leader mati sebelum follower menyalin, data hilang
> - `acks=all` — tunggu semua replika di ISR menulis. Paling aman, paling lambat
>
> **ISR (In-Sync Replicas)** — daftar replika yang up-to-date dengan leader. `min.insync.replicas=2` artinya: minimal 2 replika harus menulis, kalau tidak producer menerima error. Kombinasi `acks=all` + `min.insync.replicas=2` + `replication.factor=3` adalah setelan standar untuk "benar-benar tidak boleh hilang".
>
> **Idempotent producer** — tanpa ini, retry di sisi producer bisa menghasilkan duplikat di Kafka (message terkirim, ACK-nya yang hilang, producer kirim ulang). `enable.idempotence=true` membuat broker mendeteksi dan membuang pengiriman ulang tersebut.
>
> **Ada masalah lain di kode kamu sekarang:** producer di Chunk 1.8 pakai **async producer**, dan `Publish()` selalu return `nil`. Artinya Ingestor **selalu** membalas 202 ke client — bahkan ketika Kafka sedang down. Client mengira log-nya aman padahal hilang. Ini kebalikan dari yang dijanjikan PRD (503 kalau publish gagal). Sudah kamu catat sendiri di Chunk 8.1.5; sekarang saatnya diselesaikan.

- [ ] Konfigurasi sarama producer:
  ```go
  cfg.Producer.RequiredAcks = sarama.WaitForAll   // acks=all
  cfg.Producer.Idempotent = true
  cfg.Producer.Retry.Max = 5
  cfg.Producer.Retry.Backoff = 100 * time.Millisecond
  cfg.Net.MaxOpenRequests = 1                      // wajib kalau idempotent
  cfg.Producer.Compression = sarama.CompressionLZ4
  ```
- [ ] **Perbaiki masalah error yang tertelan.** Pilih salah satu, dan tulis alasannya:
  - **Opsi A (lebih aman):** pakai sync producer, balas 503 kalau gagal. Latency naik, tapi jujur ke client
  - **Opsi B (lebih cepat):** tetap async, tapi pantau channel `Errors()`. Kalau error rate melewati ambang, buka **circuit breaker** — balas 503 sampai Kafka pulih
  - Rekomendasiku: **Opsi B**, karena target 1.000 req/detik kamu sulit dicapai dengan sync producer. Tapi kamu harus bisa menjelaskan trade-off-nya
- [ ] Untuk demo satu broker, `replication.factor=1` tidak apa-apa — tapi **tulis konfigurasi produksinya** di `docs/DESIGN-DECISIONS.md` (RF=3, `min.insync.replicas=2`) beserta alasannya
- [ ] Tambah metric `kafka_produce_errors_total` dan `kafka_produce_duration_seconds`
- [ ] Commit: `feat(ingestor): kafka durability config and producer error handling`

> **❓ Bisa jawab ini?** Apa yang terjadi dengan `acks=all` + `RF=3` + `min.insync.replicas=2` kalau 2 dari 3 broker mati? *(Jawaban: producer berhenti menerima tulisan — sistem memilih konsistensi daripada ketersediaan. Ini pilihan sadar.)*

---

### Chunk R-I9 — Terima Log dari OTel Collector (Interop)

> **📖 Pahami dulu:** ini bukan sekadar menutup lubang keamanan — ini kapabilitas **interop** yang dipilih sengaja. Banyak aplikasi sudah pakai OpenTelemetry (SDK resmi atau auto-instrumentation) sebelum kenal LogPilot. Daripada memaksa mereka menulis ulang instrumentasinya pakai SDK LogPilot, LogPilot menerima kiriman dari **OTel Collector** lewat endpoint OTLP standar — aplikasi itu tidak perlu mengubah kodenya sama sekali.
>
> Tapi ini bukan berarti jalur ini boleh jadi jalur pintas. Di arsitektur sekarang, OTel Collector mengirim **langsung ke Kafka** — melewati Layer 1 (auth), Layer 2 (rate limit), dan Layer 3 (validasi) sepenuhnya. Pertanyaannya: siapa yang menentukan `project_id` untuk log yang lewat jalur ini? Tidak ada jawabannya di PRD. Ini bukan cuma soal keamanan (siapa pun yang menjangkau Kafka bisa menyuntik log atas nama project mana pun) — ini juga soal konsistensi: dua pintu masuk dengan aturan berbeda berarti dua kali beban pemeliharaan, dan salah satunya pasti akan tertinggal.
>
> **Prinsipnya: satu pintu masuk, satu tempat penegakan aturan** — kompatibel dengan OTel bukan berarti mengorbankan itu. Semua log, dari sumber manapun, tetap lewat auth, rate limit, dan validasi yang sama.

- [ ] Ubah `config/otel-collector/otel-collector-config.yaml`: ganti exporter Kafka jadi `otlphttp` yang menunjuk ke Ingestor:
  ```yaml
  exporters:
    otlphttp:
      endpoint: http://ingestor:8080
      headers:
        X-API-Key: ${LOGPILOT_API_KEY}
  ```
- [ ] Tambah endpoint `POST /v1/logs` di Ingestor yang menerima format **OTLP/HTTP JSON** (`resourceLogs` → `scopeLogs` → `logRecords`)
- [ ] Buat `handler/otlp.go` — konversi OTLP ke `IngestRequest`:
  - `severityNumber` → `level` (1-4=DEBUG, 5-8=INFO, 9-12=WARN, 13-16=ERROR, 17-24=FATAL)
  - `traceId`/`spanId` OTLP (base64 atau hex) → format hex kamu
  - `resource.attributes["service.name"]` → `service`
  - `timeUnixNano` → timestamp
  - Sisa attributes → `attributes`
- [ ] Semua 5 layer pipeline tetap berlaku untuk jalur ini
- [ ] Test: jalankan OTel Collector, kirim log OTLP, verifikasi masuk ClickHouse dengan `project_id` yang benar
- [ ] Commit: `feat(ingestor): otlp http endpoint, remove kafka bypass`

---

### Chunk R-I10 — PII Scrubbing

> **📖 Pahami dulu:** log sangat sering membawa data pribadi tanpa sengaja — email, nomor telepon, token JWT, nomor kartu, NIK. Begitu masuk sistem log terpusat, data itu tersimpan 30 hari dan bisa dilihat siapa pun yang punya akses Grafana. Ini masalah kepatuhan nyata (GDPR, UU PDP Indonesia), dan **hampir tidak pernah ada di project portofolio** — jadi ini pembeda yang murah tapi berkesan.

- [ ] Buat `enricher/scrubber.go` dengan pola regex untuk: email, nomor kartu kredit (+ validasi Luhn untuk mengurangi salah tangkap), JWT (`eyJ...`), IP address (opsional), NIK 16 digit, nomor HP Indonesia
- [ ] Ganti dengan penanda yang tetap informatif: `user@mail.com` → `[EMAIL_REDACTED]`
- [ ] **Compile regex sekali di package level**, jangan di dalam fungsi — kalau tidak, kamu meng-compile ulang regex 1.000 kali per detik
- [ ] Bisa dimatikan per project (kolom `pii_scrubbing_enabled` di tabel `projects`)
- [ ] Metric `pii_redactions_total{type}`
- [ ] Benchmark: ukur berapa tambahan latency-nya. Kalau > 1ms per log, optimasi (cek substring cepat dulu sebelum jalankan regex)
- [ ] Unit test dengan kasus jebakan: email di tengah kalimat, kartu kredit dengan spasi/strip, JWT di dalam JSON
- [ ] Commit: `feat(ingestor): pii scrubbing in enrichment layer`

---

### Chunk R-I11 — Performa Consumer 2 & Cache di Dispatcher

> **📖 Pahami dulu — dua masalah kinerja:**
>
> **1. Consumer 2 memproses satu per satu.** PRD Section 4.4 bilang *"Poll 1 message dari Kafka"*. Sementara Consumer 1 memproses 500 sekaligus. Kalau tiap message butuh beberapa round-trip Redis, di 1.000 log/detik Consumer 2 akan tertinggal jauh — dan alert kamu jadi terlambat, yang menghapus seluruh gunanya alerting.
>
> **2. Alert Dispatcher memanggil Management API setiap alert, secara sinkron.** Kalau Management API sedang restart atau lambat, alert gagal terkirim — tepat pada momen ketika alert paling dibutuhkan. Komponen di jalur kritis tidak boleh bergantung sinkron pada komponen yang bukan jalur kritis.

- [ ] Consumer 2: ubah jadi batch poll (100 message), kelompokkan berdasarkan rule, lalu increment sekali per rule per batch dengan `INCRBY` — bukan `INCR` berkali-kali
- [ ] Gabungkan semua operasi Redis dalam satu batch ke dalam satu pipeline
- [ ] Alert Dispatcher: cache notification config di memory dengan TTL 60 detik
- [ ] Kalau Management API down dan cache kosong → tetap kirim ke destination default dari env var, jangan diam saja
- [ ] Tambah timeout 3 detik + circuit breaker untuk panggilan ke Management API
- [ ] Benchmark Consumer 2 sebelum dan sesudah: berapa message/detik
- [ ] Commit: `perf(consumer-alert): batch evaluation with redis pipeline`

---

# 🔴 LEVEL SENIOR

*Ini bagian yang membedakan kamu dari kandidat lain. Kerjakan setelah dua level di atas selesai — bukan karena kodenya sulit, tapi karena konsepnya bertumpu pada semua yang di atas.*

---

### Chunk R-S1 — Error Fingerprinting & Grouping 🔥🔥
> 🔧 **KOREKSI v3** — Kolom `fingerprint` bertipe `String`, bukan `LowCardinality`. DDL materialized view di chunk ini menghitung salah setelah merge; pakai `issues_hourly` + `issues_mv` di PRD §5.3, dan uji dengan F16.
*Ini jawaban yang sebenarnya untuk pertanyaan interview kemarin.*

> **📖 Pahami dulu — ini bagian terpenting di seluruh dokumen ini:**
>
> **Pertanyaan interviewer kemarin belum selesai dijawab dengan trace_id.**
>
> Pertanyaannya: *"kalau 100.000 user hit di waktu yang sama, service yang sama, semua kena error berbeda — gimana cara bedain masing-masing?"*
>
> trace_id memang jawaban yang benar untuk **membedakan**. Tapi coba pikirkan sampai ke ujungnya: sekarang kamu punya **100.000 trace_id berbeda**. Lalu apa? Kamu buka satu per satu? Tidak ada manusia yang bisa. Kamu berhasil membedakan 100.000 error, tapi kamu tetap tidak tahu **apa yang sedang terjadi**.
>
> Analoginya: trace_id memberimu 100.000 jarum yang masing-masing punya nomor seri. Yang sebenarnya kamu butuhkan adalah mengetahui bahwa dari 100.000 jarum itu, sebenarnya cuma ada **12 jenis jarum**.
>
> **Itulah error fingerprinting.**
>
> **Cara kerjanya.** Untuk setiap error, hitung sebuah hash dari ciri-ciri strukturalnya — bukan dari isi persisnya. Error yang **berbeda isinya tapi sama penyebabnya** akan menghasilkan hash yang sama.
>
> Kuncinya ada di **normalisasi**. Bandingkan:
> ```
> "user 12345 not found in database"
> "user 67890 not found in database"
> "user 11111 not found in database"
> ```
> Tiga string berbeda. Tapi setelah angka diganti placeholder:
> ```
> "user <num> not found in database"
> ```
> Ketiganya identik → satu fingerprint → **satu issue dengan count 3**, bukan tiga error terpisah.
>
> **Yang perlu dinormalisasi:** angka, UUID, alamat email, IP address, timestamp, path file dengan hash, hex string panjang, segmen URL yang variabel (`/user/123/order/456` → `/user/<id>/order/<id>`), durasi (`timeout after 30.2s` → `timeout after <duration>`).
>
> **Rumus fingerprint:**
> ```
> fingerprint = sha256(service + ":" + level + ":" + normalize(message) + ":" + top_stack_frame)
> ```
> Ambil 16 karakter pertama saja — cukup untuk menghindari tabrakan, tapi enak dibaca.
>
> **Hasil akhirnya:** 100.000 error tadi berubah jadi tampilan seperti ini:
> ```
> ISSUE                                    COUNT    USERS   FIRST SEEN
> payment gateway timeout                  62.341   58.900  2 menit lalu
> connection pool exhausted                31.204   30.100  2 menit lalu
> user <num> not found                      4.102    4.102  8 jam lalu
> invalid coupon code <str>                 2.353    2.201  3 hari lalu
> ```
> Sekarang kamu bisa langsung melihat: **ini bukan 100.000 masalah, ini 2 masalah** (dua yang teratas, yang muncul barengan 2 menit lalu — kemungkinan besar saling terkait). Dua yang di bawah sudah lama ada dan bukan penyebab insiden ini.
>
> Baru dari sini kamu ambil **satu contoh trace_id** per fingerprint untuk menelusuri detailnya. Fingerprint untuk melihat gambaran besar; trace_id untuk menyelam ke satu kasus.
>
> **Ini persis cara Sentry bekerja.** Kalau di interview kamu bisa bilang:
>
> > *"trace_id untuk menelusuri satu request. Tapi di skala 100 ribu, saya butuh fingerprinting dulu untuk pengelompokan — kalau tidak, saya cuma punya 100 ribu jarum tanpa tahu ada berapa jenis jarum. Setelah dikelompokkan, biasanya tinggal 10-20 issue nyata, dan saya ambil satu exemplar trace per issue untuk drill-down."*
>
> ...itu bukan jawaban yang diharapkan interviewer. Itu jawaban yang membuat interviewer berhenti dan mikir.

- [ ] Buat package `pkg/fingerprint/` (dipakai bersama oleh ingestor dan consumer)
- [ ] `normalize.go` — fungsi `Normalize(msg string) string` dengan aturan berurutan:
  | Pola | Diganti jadi |
  |---|---|
  | UUID | `<uuid>` |
  | Email | `<email>` |
  | IP address | `<ip>` |
  | Hex ≥ 8 karakter | `<hex>` |
  | Timestamp ISO 8601 | `<ts>` |
  | Angka (integer/desimal) | `<num>` |
  | String dalam kutip | `<str>` |
  | Path file | `<path>` |
  | Segmen URL variabel | `<id>` |
- [ ] **Compile semua regex sekali di package level** — ini jalan di jalur panas
- [ ] `fingerprint.go` — `Compute(service, level, message, stackTop string) string` → sha256, ambil 16 hex pertama
- [ ] **Unit test dengan kasus jebakan** (bagian ini yang menentukan kualitasnya):
  - Over-grouping: `"disk full on /dev/sda1"` vs `"disk full on /dev/sdb2"` — harus **sama**? Ya, penyebabnya sama
  - Under-grouping: `"timeout connecting to payment-svc"` vs `"timeout connecting to auth-svc"` — harus **beda**? Ya, sistem yang bermasalah berbeda. Artinya nama service **tidak boleh** dinormalisasi
  - `"failed after 3 retries"` vs `"failed after 5 retries"` → sama
  - Message kosong, message 10 KB, message dengan emoji dan karakter non-ASCII
- [ ] Hitung fingerprint di **enrichment layer** Ingestor, hanya untuk level ERROR dan FATAL (hemat CPU)
- [ ] Tambah kolom `fingerprint LowCardinality(String)` + bloom filter index di ClickHouse
- [ ] Buat Materialized View untuk agregasi issue:
  ```sql
  CREATE MATERIALIZED VIEW logpilot.issues_mv
  ENGINE = AggregatingMergeTree()
  ORDER BY (project_id, fingerprint, hour)
  AS SELECT
    project_id,
    fingerprint,
    toStartOfHour(timestamp) AS hour,
    any(service)             AS service,
    any(level)               AS level,
    argMin(message, timestamp)  AS sample_message,
    argMin(trace_id, timestamp) AS exemplar_trace_id,
    count()                  AS event_count,
    uniqState(user_id)       AS unique_users,
    min(timestamp)           AS first_seen,
    max(timestamp)           AS last_seen
  FROM logpilot.logs
  WHERE level IN ('ERROR','FATAL')
  GROUP BY project_id, fingerprint, hour;
  ```
  *`argMin(x, timestamp)` = ambil nilai x dari baris dengan timestamp terkecil — inilah cara mendapat exemplar. `uniqState` menyimpan sketsa HyperLogLog sehingga jumlah user unik bisa digabung lintas jam tanpa menyimpan semua user_id.*
- [ ] Panel Grafana **"Top Issues"**: fingerprint, sample_message, count, unique users, first/last seen, exemplar trace_id — diurutkan berdasarkan count
- [ ] Buat panel kedua yang menerima variabel `trace_id` sehingga klik dari panel pertama langsung membuka seluruh log satu trace
- [ ] **Tambahkan ke demo script (PRD Section 6.3) sebagai langkah baru:**
  1. Jalankan log-generator dengan mode `--chaos` yang menghasilkan 50.000 error dari 5 penyebab berbeda
  2. Tunjukkan tampilan log mentah — puluhan ribu baris, tidak terbaca
  3. Buka panel Top Issues — 5 baris, jelas terbaca
  4. Klik satu issue → seluruh trace satu request terbuka
  5. Kalimat penutupnya: *"dari 50.000 error jadi 5 issue dalam satu klik"*
- [ ] Commit: `feat: error fingerprinting and issue grouping`

> **❓ Bisa jawab ini?**
> - Kenapa nama service tidak boleh dinormalisasi, tapi angka boleh?
> - Apa itu over-grouping dan under-grouping? Mana yang lebih berbahaya dan kenapa?
> - Kenapa exemplar trace_id diambil dengan `argMin` dan bukan `any`?

---

### Chunk R-S2 — Materialized View untuk Pra-agregasi
> 🔧 **KOREKSI v3** — Sebelum mulai: kenapa `count() AS event_count` di dalam `AggregatingMergeTree` memberi angka yang turun sendiri? Jawabannya dan DDL yang benar ada di PRD §5.3.

> **📖 Pahami dulu:** dashboard Grafana yang menghitung `count() GROUP BY minute` di atas 500 juta baris akan memindai seluruh partisi **setiap kali dibuka, setiap 10 detik auto-refresh, oleh setiap orang**. Ini cara paling cepat membuat ClickHouse kamu tumbang.
>
> **Materialized View di ClickHouse bukan seperti di PostgreSQL.** Di PostgreSQL, MV adalah snapshot yang harus di-refresh manual. Di ClickHouse, MV adalah **trigger saat INSERT**: setiap batch yang masuk ke tabel sumber otomatis diagregasi dan ditulis ke tabel tujuan. Selalu terkini, tanpa cron job.
>
> **Trade-off yang harus kamu sadari:** MV menambah beban di jalur tulis (setiap INSERT jadi lebih berat) untuk menghemat besar di jalur baca. Untuk sistem log — yang tulisnya sekali, bacanya ribuan kali — pertukaran ini hampir selalu menguntungkan. Tapi kamu harus bisa mengatakannya sebagai keputusan sadar, bukan kebetulan.

- [ ] Buat MV `logs_per_minute` — agregasi count per project/service/level per menit
- [ ] Buat MV `service_health_5m` — total, error count, error rate, p95 (via `quantileState`) per 5 menit
- [ ] Ubah panel Grafana dari query tabel mentah ke MV
- [ ] Ukur dan catat: waktu query dan `rows read` sebelum vs sesudah
- [ ] Ukur juga dampaknya ke throughput INSERT — jujur soal harganya
- [ ] Tulis di `docs/DESIGN-DECISIONS.md`
- [ ] Commit: `perf(clickhouse): materialized views for dashboard queries`

---

### Chunk R-S3 — Load Test dengan Angka Nyata 🔥

> **📖 Pahami dulu:** PRD kamu mengklaim *"1.000 req/detik di single pod"*. Sekarang itu masih **target, bukan hasil pengukuran**. Log generator kamu cuma mengirim 5 log/detik — 200 kali lebih kecil dari klaim.
>
> Interviewer akan menanyakan angka itu. Ada dua kemungkinan jawaban:
> - *"Targetnya 1.000 req/detik"* → dia tahu kamu belum pernah mengukur
> - *"Saya ukur dengan k6: 1.240 req/s dengan p99 47ms, bottleneck-nya ada di JSON marshalling, setelah saya pakai sync.Pool naik jadi 1.850"* → percakapannya langsung berubah level
>
> **Kenapa p99, bukan rata-rata.** Rata-rata menyembunyikan bencana. Kalau 99 request selesai 10ms dan 1 request 5 detik, rata-ratanya 60ms — terlihat baik-baik saja. Padahal 1% pengguna kamu menunggu 5 detik. p99 = "99% request lebih cepat dari ini". Itu angka yang benar-benar dirasakan pengguna.

- [ ] Install k6 dan buat `loadtest/ingest.js`
- [ ] Skenario 1 — **ramp-up**: 0 → 2.000 VU dalam 5 menit, cari titik patahnya
- [ ] Skenario 2 — **sustained**: 1.000 req/s selama 10 menit, cek apakah memory bertambah terus (kebocoran memori baru kelihatan di tes panjang)
- [ ] Skenario 3 — **spike**: dari 100 langsung ke 5.000 req/s, lihat perilaku saat kaget
- [ ] Skenario 4 — **batch**: kirim batch 500 log per request, bandingkan throughput per-log dengan mode satuan
- [ ] Catat untuk setiap skenario: req/s, p50, p95, p99, error rate, CPU, memory, dan consumer lag
- [ ] Jalankan `pprof` saat load test berjalan, identifikasi 3 hotspot teratas
- [ ] Optimasi 3 hotspot itu, ukur ulang, catat perbaikannya
- [ ] Tulis `docs/BENCHMARKS.md` lengkap dengan tabel dan grafik, spesifikasi mesin, dan versi setiap komponen
- [ ] Tempel ringkasannya di README — **angka nyata di README lebih meyakinkan daripada daftar teknologi yang panjang**
- [ ] Commit: `docs: load test results and performance analysis`

---

### Chunk R-S4 — Capacity Planning & Model Biaya

> **📖 Pahami dulu:** *"berapa biayanya kalau di-scale?"* adalah pertanyaan yang membedakan engineer dari engineer yang dipercaya mengambil keputusan. Sebagian besar kandidat portofolio tidak pernah menghitungnya sama sekali.
>
> Perhitungannya sederhana, yang penting kamu **pernah melakukannya**:
> ```
> Ukuran log rata-rata           : ~400 byte JSON
> Volume                          : 1.000 log/detik
>                                 = 86,4 juta log/hari
>                                 = 34,5 GB/hari mentah
> Rasio kompresi ClickHouse       : ~10:1 untuk log (ZSTD + kolom LowCardinality)
>                                 = ~3,5 GB/hari tersimpan
> Retensi 30 hari                 = ~105 GB
> Buffer Kafka 7 hari             = ~240 GB (Kafka kompresinya lebih rendah, LZ4)
> ```
> Angka-angka ini yang mendasari keputusan: berapa besar disk, kapan mulai butuh sharding, apakah retensi 30 hari masuk akal, apakah tiering ke S3 perlu.

- [ ] Ukur ukuran log rata-rata yang sebenarnya dari data kamu, jangan menebak
- [ ] Ukur rasio kompresi ClickHouse yang sebenarnya:
  ```sql
  SELECT
    formatReadableSize(sum(data_uncompressed_bytes)) AS raw,
    formatReadableSize(sum(data_compressed_bytes))   AS compressed,
    round(sum(data_uncompressed_bytes) / sum(data_compressed_bytes), 2) AS ratio
  FROM system.columns WHERE table = 'logs';
  ```
- [ ] Buat tabel proyeksi untuk 100 / 1.000 / 10.000 log per detik: disk ClickHouse, disk Kafka, RAM, dan perkiraan biaya VPS
- [ ] Tentukan pada volume berapa arsitektur sekarang mulai patah, dan apa langkah berikutnya (sharding ClickHouse, tiering ke S3, retensi berjenjang)
- [ ] Bandingkan dengan harga Datadog pada volume yang sama — ini menghubungkan langsung ke *problem statement* di PRD kamu
- [ ] Tulis `docs/CAPACITY.md`
- [ ] Commit: `docs: capacity planning and cost model`

---

### Chunk R-S5 — Backpressure & Degradasi Bertahap

> **📖 Pahami dulu:** setiap sistem punya batas. Yang membedakan sistem yang matang bukan batasnya berapa, tapi **apa yang terjadi ketika batas itu terlampaui**.
>
> - Sistem yang buruk: menerima semua, antrean membengkak, memori habis, **OOM kill, semuanya mati sekaligus**
> - Sistem yang baik: mendeteksi kelebihan beban, **menolak sebagian permintaan dengan sopan**, tetap hidup, pulih sendiri saat beban turun
>
> Ini namanya **load shedding** — membuang sebagian beban untuk menyelamatkan sisanya. Berlawanan dengan intuisi (kok malah menolak?) tapi inilah yang membedakan degradasi bertahap dari kegagalan total.
>
> **Konsep terkait: prioritas.** Kalau harus membuang, buang yang paling tidak berharga dulu. Log DEBUG boleh dibuang; log FATAL jangan pernah.

- [ ] Tambah bounded queue di Ingestor antara handler HTTP dan Kafka producer
- [ ] Kalau antrean > 80% penuh → mulai buang log level DEBUG, balas 202 tapi catat `logs_shed_total{level}`
- [ ] Kalau > 90% → buang juga INFO
- [ ] Kalau > 95% → balas `503` dengan header `Retry-After`, kecuali untuk ERROR dan FATAL
- [ ] **Jangan pernah membuang ERROR/FATAL** — itu justru log yang paling berharga saat sistem sedang bermasalah
- [ ] Tambah `/health` (liveness: proses hidup) dan `/ready` (readiness: siap menerima trafik — Kafka tersambung, Redis tersambung, antrean belum penuh) sebagai dua endpoint terpisah
  *Kenapa dipisah:* kalau liveness gagal, Kubernetes **membunuh** pod-nya. Kalau readiness gagal, Kubernetes cuma **berhenti mengirim trafik** tapi pod dibiarkan pulih. Menyatukan keduanya berarti pod yang cuma sedang sibuk akan terus-menerus di-restart — memperparah masalah.
- [ ] Pasang graceful shutdown di semua service: tangkap SIGTERM, tandai `/ready` gagal, tunggu 5 detik, flush buffer, keluar
- [ ] **Test:** kirim 5.000 req/s ke pod yang cuma sanggup 1.000. Verifikasi: pod tidak mati, ERROR tetap masuk semua, DEBUG dibuang, metric `logs_shed_total` naik
- [ ] Commit: `feat(ingestor): load shedding and graceful degradation`

> **❓ Bisa jawab ini?** Kenapa menolak request bisa lebih baik daripada menerimanya? Apa bedanya liveness dan readiness probe, dan apa akibatnya kalau tertukar?

---

### Chunk R-S6 — Isolasi Multi-Tenant di Grafana

> **📖 Pahami dulu:** Grafana tersambung **langsung** ke ClickHouse dengan satu kredensial. Artinya siapa pun yang bisa membuka Grafana bisa mengetik `SELECT * FROM logs WHERE project_id = 'punya-orang-lain'`. Untuk demo tidak masalah — tapi kalau interviewer bertanya *"gimana kalau ini dipakai 2 tim berbeda?"* kamu harus punya jawaban.
>
> Menyadari batasan sistemmu sendiri dan bisa menjelaskan cara menutupnya, itu justru sinyal kematangan. Yang buruk adalah tidak sadar sama sekali.

- [ ] Pilih dan implementasikan salah satu (tulis alasan pemilihannya):
  - **A — ClickHouse row policy:** satu user ClickHouse per tenant dengan `CREATE ROW POLICY ... USING project_id = 'x'`. Paling kuat, tapi tidak praktis kalau tenant-nya banyak
  - **B — Query proxy:** service kecil di depan ClickHouse yang menyuntikkan filter `project_id` berdasarkan token. Paling fleksibel, tapi kamu harus menulis dan menjaganya sendiri
  - **C — Grafana per organisasi:** pakai fitur Org bawaan Grafana + datasource terpisah per tenant. Paling cepat dibangun, paling merepotkan untuk banyak tenant
- [ ] Kalau memilih tidak mengimplementasikan: **tulis eksplisit di PRD Section 2.2 (Non-Goals)** beserta rencana penanganannya di masa depan
- [ ] Tambah `project_id` sebagai variabel dashboard di Grafana dan pastikan semua panel memakainya
- [ ] Tulis analisis ketiga opsi di `docs/DESIGN-DECISIONS.md`
- [ ] Commit: `feat: multi-tenant isolation for grafana access`

---

### Chunk R-S7 — Chaos Test: Buktikan Klaim di PRD

> **📖 Pahami dulu:** PRD kamu membuat beberapa klaim ketahanan (F9: *"matikan ClickHouse 1 menit, semua log tetap tersimpan"*; NFR: *"sistem tetap menerima log meskipun ClickHouse down"*). Klaim yang belum pernah diuji bukan klaim, itu harapan.
>
> Uji satu per satu, **rekam hasilnya**, dan pastikan hasilnya sesuai atau perbaiki sistemnya sampai sesuai.

- [ ] Buat `chaos/scenarios.md` — untuk setiap skenario tulis: apa yang dimatikan, perilaku yang **diharapkan**, perilaku yang **benar-benar terjadi**, dan perbaikan apa yang dilakukan
- [ ] Skenario 1 — **ClickHouse mati 2 menit** saat 500 log/s mengalir. Harapan: nol log hilang, consumer lag naik lalu pulih
- [ ] Skenario 2 — **Redis mati**. Harapan: auth tetap jalan lewat fallback (`R-I4`)
- [ ] Skenario 3 — **Kafka mati**. Harapan: Ingestor membalas 503, bukan 202 palsu (`R-I8`)
- [ ] Skenario 4 — **Management API mati** saat alert terpicu. Harapan: alert tetap terkirim dari cache (`R-I11`)
- [ ] Skenario 5 — **Consumer di-`kill -9`** di tengah batch. Harapan: tidak ada log hilang, tidak ada duplikat (`R-I5`)
- [ ] Skenario 6 — **Disk ClickHouse penuh**. Harapan: gagal dengan jelas, tidak diam-diam kehilangan data
- [ ] Skenario 7 — **Network partition** antara consumer dan Kafka (pakai `tc netem` atau `docker network disconnect`)
- [ ] Untuk setiap skenario: catat waktu pemulihan (RTO) dan data yang hilang (RPO)
- [ ] Rekam salah satunya sebagai video pendek — ini bagian demo yang paling berkesan
- [ ] Commit: `docs: chaos testing scenarios and results`

---

### Chunk R-S8 — Persiapan Interview: Tulis Jawabannya 🔥

> **📖 Pahami dulu:** kamu bisa tahu jawabannya di kepala dan tetap gagal menyampaikannya di ruang interview — itu hal yang sangat umum dan tidak ada hubungannya dengan kemampuan teknis. Menulis jawabannya lebih dulu memaksa kamu menemukan bagian mana yang sebenarnya masih kabur.
>
> Tulis di `docs/INTERVIEW-NOTES.md`. Setiap jawaban maksimal 5 kalimat. Kalau tidak bisa diringkas jadi 5 kalimat, berarti kamu belum benar-benar paham.

- [ ] **Pertanyaan yang sudah pernah ditanya** (siapkan jawaban lengkapnya): *"100.000 user hit bersamaan, service sama, error berbeda semua — gimana membedakannya?"*
  Struktur jawaban: trace_id untuk korelasi → span untuk mencari titik gagal → fingerprint untuk pengelompokan → exemplar untuk drill-down → user_id untuk mengukur dampak
- [ ] Kenapa Kafka, bukan langsung tulis ke ClickHouse?
- [ ] Kenapa ClickHouse, bukan Elasticsearch atau PostgreSQL?
- [ ] Bagaimana kamu menjamin tidak ada log yang hilang? Kalau duplikat?
- [ ] Kenapa dua consumer group terpisah, bukan satu consumer yang mengerjakan keduanya?
- [ ] Apa yang terjadi kalau Ingestor di-scale jadi 10 pod? Apa yang rusak?
- [ ] Kenapa partition key `project_id`? Apa yang terjadi kalau satu project jauh lebih besar dari yang lain? *(hot partition — siapkan jawabannya)*
- [ ] Bagaimana kamu tahu sistem ini sehat? Tiga metrik pertama yang kamu lihat?
- [ ] Apa bagian paling lemah dari sistem ini menurutmu sendiri? *(pertanyaan jebakan — "tidak ada" adalah jawaban terburuk)*
- [ ] Kalau harus mengulang dari nol, apa yang kamu ubah?
- [ ] Kenapa sliding window, bukan fixed window? Kenapa bucketed, bukan sorted set?
- [ ] Berapa biayanya kalau volume-nya 10x?
- [ ] **Latihan mengucapkannya keras-keras.** Rekam suaramu, dengarkan, perbaiki bagian yang tersendat
- [ ] Commit: `docs: interview preparation notes`

---

### Chunk R-S9 — Sampling: Head vs Tail

> **📖 Pahami dulu:** di volume tinggi, kamu **tidak** menyimpan setiap trace — cuma sebagian (sampel). Kalau tidak, storage dan biaya membengkak sia-sia, karena sebagian besar trace itu "request sukses, semua normal, tidak ada yang perlu dilihat".
>
> **Head sampling** — keputusan diambil **di awal** request, sebelum tahu hasilnya. Murah: cukup lempar dadu (misal "simpan 10% acak") begitu request masuk, tidak perlu menahan apa pun. Tapi **buta** — kamu bisa saja membuang justru trace yang berakhir error, karena keputusannya diambil sebelum errornya terjadi.
>
> **Tail sampling** — keputusan diambil **setelah** request selesai. Simpan semua yang error atau lambat, plus sebagian kecil acak dari sisanya yang normal (untuk baseline). Jauh lebih berguna karena kamu tidak pernah kehilangan trace yang justru paling ingin kamu lihat. Tapi lebih mahal: kamu harus **menahan** seluruh span dari satu trace di memory sampai request-nya selesai, baru memutuskan buang atau simpan. Ini salah satu tugas utama **OTel Collector** dalam arsitektur observability modern — dan ini alasan bagus kenapa Collector ada di arsitektur LogPilot, bukan cuma sebagai jembatan protokol (lihat `R-I9`).
>
> **Kenapa kamu tidak membangun tail sampling sendiri:** menahan trace di memory butuh state per-trace yang hidup selama durasi request, plus logic penggabungan span lintas waktu — ini pekerjaan yang sudah diselesaikan dengan baik oleh OTel Collector `tail_sampling` processor. Menulis ulang itu sendiri tidak menambah pembelajaran baru dibanding chunk lain di Phase R — ini keputusan **delegasi yang sadar**, bukan kemalasan.
>
> **Chunk ini tidak memblokir apa pun** — sisipkan kapan saja.

- [ ] Implementasi head sampling sederhana di SDK: simpan `X%` trace secara acak (bisa dikonfigurasi), TAPI **selalu** simpan kalau ada log level ERROR atau FATAL di dalam trace itu, berapa pun persentase sampling-nya
- [ ] Tulis di `docs/DESIGN-DECISIONS.md`: kenapa tail sampling secara konseptual lebih baik dari head sampling, dan kenapa kamu menyerahkan implementasinya ke OTel Collector daripada membangunnya sendiri di SDK/Ingestor
- [ ] Commit: `feat(sdk): head sampling with always-sample-on-error`

> **❓ Bisa jawab ini?**
> - Kalau head sampling menyimpan 1% trace secara acak, dan errornya cuma terjadi di 0,5% request, apa risikonya?
> - Kenapa tail sampling butuh menahan trace di memory, sementara head sampling tidak?
> - Kalau kamu diminta membangun tail sampling sendiri (bukan pakai Collector), komponen apa yang paling sulit — dan kenapa?

---

## 🎬 PHASE S — Showcase (PRD v3.0 §4.8, §4.9, §5.5)

> Mulai setelah M5 (fingerprinting jalan). Sebelum itu, Console tidak punya apa-apa untuk ditampilkan.

### Chunk S-1 — Query API: kerangka + isolasi tenant 🔥
- [ ] Modul `query` di `services/api` (service yang sama dengan Management API, PRD v3.1 §4.8); `/health`, `/ready`, `/metrics` sudah ada dari Chunk 4.1
- [ ] Verifikasi JWT (EdDSA, kunci publik saja); `project_id` hanya dari token
- [ ] User ClickHouse `readonly` dengan profil batas (`max_execution_time`, `max_rows_to_read`, `max_memory_usage`)
- [ ] Endpoint Management API `POST /api/projects/{id}/query-token`
- [ ] Test: token project A tidak bisa membaca project B (F18)
- **Selesai kalau:** F18 lulus di CI. Ini sekaligus menutup `R-S6`.

### Chunk S-2 — Bahasa query (LPQL)
- [ ] Tulis test case dulu: `level:error service:api "timeout" -region:sgp`
- [ ] Tokenizer → daftar term → SQL berparameter (tidak ada penyambungan string)
- **Selesai kalau:** test injeksi (`q=' OR 1=1 --`) menghasilkan pencarian teks biasa, bukan SQL.

### Chunk S-3 — `GET /v1/logs` + histogram + keyset pagination
- [ ] Rentang waktu wajib, maksimal 7 hari; `limit` maksimal 1.000 (F19)
- [ ] Cursor dari `(timestamp, kafka_partition, kafka_offset)`
- **Selesai kalau:** halaman ke-100 sama cepatnya dengan halaman pertama (ukur, catat di `BENCHMARKS.md`).

### Chunk S-4 — Live Tail (SSE) 🔥
- [ ] Pembaca Kafka tanpa consumer group, satu per partition, fan-out di memori
- [ ] Buffer per penonton 1.000 log; penonton lambat di-drop, event `dropped` dikirim
- [ ] Heartbeat, batas koneksi per project, batas umur koneksi
- **Selesai kalau:** F20 lulus, termasuk bagian "ClickHouse dimatikan, tail tetap jalan".

### Chunk S-5 — Endpoint Issues + Trace
- [ ] `/v1/issues`, `/v1/issues/{fp}` dari `issues_hourly` (pakai `-Merge`)
- [ ] `/v1/traces/{id}` mengembalikan pohon span, bukan daftar datar

### Chunk S-6 — Console: Onboarding + Live Tail
- [ ] Wizard project → API key → cuplikan → "menunggu log pertama…"
- [ ] Live Tail dengan daftar tervirtualisasi

### Chunk S-7 — Console: Explorer + Issues + Trace waterfall
- [ ] Semua state di URL (bisa dibagikan)
- [ ] Tautan silang: `trace_id` → Trace, `fingerprint` → Issue

### Chunk S-8 — Issue lifecycle + alert `new_issue` / `regression` 🔥
- [ ] Tabel `issues` di PostgreSQL, endpoint ubah status
- [ ] Consumer 2: `SET ... NX` untuk issue baru, `SISMEMBER` untuk regression, publish ke `issue-events`
- **Selesai kalau:** F21 lulus dengan Consumer 2 berjalan 3 instance.

### Chunk S-9 — Mode demo + README 🔥
- [ ] Project `demo` baca-saja tanpa login; tombol "Simulasikan insiden" dengan cooldown global
- [ ] README: GIF 15 detik, diagram, tabel angka, quick start dua perintah
- **Selesai kalau:** F22 — orang non-teknis menjalankan naskah 60 detik sendirian.

---

## ⚙️ PHASE P — Platform & DevOps (PRD v3.0 §6)

> 🔧 **KOREKSI v3.1** — perkakas platform dipasang satu per satu, satu per sesi, dengan urutan di PRD §6.4. Jangan memasang yang berikutnya sebelum yang terakhir dipakai satu beban kerja dan pemakaian memorinya tercatat di `docs/CAPACITY.md`.

> `P-1` sampai `P-4` boleh dikerjakan kapan saja setelah M1, dengan Ingestor sebagai satu-satunya beban kerja. Sisanya setelah M6.
> Bahan pendamping: `~/Developer/Belajar/devops-kubernetes/` (Lab 1–3).

### Chunk P-1 — k3s di VPS
- [ ] Pasang k3s tanpa Traefik; firewall hanya 22, 80, 443
- [ ] `kubeconfig` di Mac, `kubectl get nodes` dari laptop lewat SSH tunnel (port 6443 tidak dibuka ke publik)
- **Selesai kalau:** bisa menjelaskan beda k3s dengan cluster kubeadm di AHU (apa yang digabung, apa yang diganti).

### Chunk P-2 — Dockerfile + Helm chart untuk Ingestor
- [ ] Image multi-stage, non-root, ukuran dicatat
- [ ] Chart: Deployment (probe, requests/limits, securityContext), Service, ConfigMap, Secret
- **Selesai kalau:** `helm upgrade --install` dua kali berturut-turut tidak mengubah apa pun.

### Chunk P-3 — ingress-nginx + cert-manager
- [ ] Domain mengarah ke VPS; sertifikat Let's Encrypt terbit otomatis
- **Selesai kalau:** `curl https://ingest.<domain>/health` → 200 dengan sertifikat sah.

### Chunk P-4 — GitOps dengan Argo CD 🔥
- [ ] Repo `logpilot-deploy`; Argo CD menyinkronkan dari sana
- [ ] CI: build → push ghcr.io → commit tag image ke repo deploy
- [ ] Sealed Secrets untuk semua kredensial
- **Selesai kalau:** F23 — `git push` sampai berjalan tanpa `kubectl`; `kubectl edit` manual dikembalikan sendiri.

### Chunk P-5 — Sistem stateful lewat operator
- [ ] Strimzi (Kafka KRaft 1 broker), Altinity (ClickHouse), CloudNativePG, Redis tanpa volume
- [ ] Migrasi skema sebagai Job pre-sync
- **Selesai kalau:** node di-reboot → semua data masih ada, consumer melanjutkan dari offset terakhir.

### Chunk P-6 — Seluruh service + NetworkPolicy + PodDisruptionBudget
- [ ] `/internal/*` tidak terjangkau dari pod yang tidak berhak (buktikan dengan `curl` yang gagal)

### Chunk P-7 — Autoscaling: HPA + KEDA 🔥
- [ ] HPA Ingestor dari CPU; KEDA consumer dari lag
- **Selesai kalau:** k6 menaikkan beban → jumlah consumer naik mengikuti lag, lalu turun lagi. Rekam grafiknya.

### Chunk P-8 — kube-prometheus-stack + alert + runbook
- [ ] `ServiceMonitor` per service, `PrometheusRule` (§6.7), panel SLO
- [ ] `docs/RUNBOOK.md`: satu bagian per alert

### Chunk P-9 — Backup dan pemulihan 🔥
- [ ] PostgreSQL (CloudNativePG) dan ClickHouse (`clickhouse-backup`) ke object storage
- **Selesai kalau:** F25 — dipulihkan ke namespace kosong, lama waktunya dicatat.

### Chunk P-10 — Uji kegagalan di cluster
- [ ] `kubectl delete pod` pada Ingestor, Consumer 1, ClickHouse; `kubectl drain` node (di `lab`)
- **Selesai kalau:** F24, hasilnya masuk `chaos/scenarios.md`.

---

## 🔮 PHASE 8 — Future Features (Post-MVP)
*Dikerjakan setelah semua phase selesai dan sistem sudah bisa di-demo.*

### Chunk 8.1 — Loki Storage Support
- [ ] Buat `services/consumer-loki/` — consumer group baru `storage-loki-workers`
- [ ] Subscribe ke topic `raw-logs` secara independen (fan-out pattern)
- [ ] Insert log ke Loki via HTTP push API
- [ ] Update Grafana: tambah Loki sebagai datasource alternatif
- [ ] User bisa pilih query dari ClickHouse atau Loki

### Chunk 8.1.5 — Kafka Producer: Error Monitoring Enhancement
- [ ] Enhance goroutine error handler di `producer/kafka.go` — saat ini error hanya di-log
- [ ] Tambah counter/metric untuk track jumlah produce error
- [ ] Pertimbangkan dead letter queue atau retry mechanism untuk failed messages
- [ ] Note: `Publish()` selalu return nil (async) — caller tidak tau kalau Kafka down post-init

### Chunk 8.2 — Alert Dispatcher: Semaphore untuk Concurrent Dispatch
- [ ] Ganti `sync.WaitGroup` murni dengan semaphore untuk batasi concurrent call ke tiap destination (ClickUp/SendGrid/Slack)
- [ ] Semaphore lebih tepat dari WaitGroup karena mengontrol berapa goroutine yang boleh hit API bersamaan
- [ ] Pertahankan timeout 10 detik per destination via context
- [ ] Note: berbeda dengan worker pool — goroutine sudah di-spawn per alert, semaphore hanya sebagai throttle

### Chunk 8.3 — Consumer Storage: Worker Pool Pattern
- [ ] Refactor consumer storage menggunakan worker pool/queue pattern
- [ ] Worker pool lebih cocok dari fan-out biasa karena log terus mengalir dan worker harus tetap hidup
- [ ] Batasi jumlah worker yang concurrent insert ke ClickHouse

### Chunk 8.4 — Custom Ring Buffer untuk Micro-batch
- [ ] Implement fixed-size ring buffer di `consumer-storage/buffer/ring_buffer.go`
- [ ] Struktur: head/tail pointer dengan `sync/atomic` — jangan pakai mutex
- [ ] Ganti slice biasa di micro-batch logic dengan ring buffer ini
- [ ] Benchmark: bandingkan throughput dan allocs/op vs slice append sebelumnya
- [ ] Commit: `feat(consumer-storage): lock-free ring buffer for batch ingestion`

---

## 📊 Progress Tracker

### Phase Utama

| Phase | Status | Selesai Tgl |
|---|---|---|
| Phase 0 — Setup | ✅ Done | 2026-04-26 |
| Phase 1 — Ingestor | ✅ Done | 2026-05-17 |
| Phase 2 — Storage Worker | 🟡 In progress | — |
| Phase 3 — Alert Pipeline | ⬜ Not started | — |
| Phase 4 — Management API (Go) | ⬜ Not started | — |
| Phase 5 — Frontend | ⬜ Not started | — |
| Phase 6 — Demo App | ⬜ Not started | — |
| Phase 7 — Polish | ⬜ Not started | — |
| Phase 2 Chunk 2.8 — Profiling | ⬜ Not started | — |
| Phase 8 Chunk 8.4 — Ring Buffer | ⬜ Not started | — |
| Phase 7 Chunk 7.6 — Blog Post | ⬜ Not started | — |

### Phase R — 🟢 Junior

| Chunk | Konsep utama | Status |
|---|---|---|
| R-J1 — Trace Context 🔥 | W3C trace context, propagasi | ⬜ |
| R-J2 — Migrasi Skema ClickHouse 🔥 | LowCardinality, Map, migration file | ⬜ |
| R-J3 — Promote metadata → attributes | desain skema fleksibel vs kaku | ⬜ |
| R-J4 — Batch Ingest | partial success, HTTP 207 | ⬜ |
| R-J5 — Clock Skew Guard | jam client tidak bisa dipercaya | ⬜ |
| R-J6 — Prometheus | pull vs push, counter/gauge/histogram | ⬜ |
| R-J7 — Span Berdurasi | span, span_id/parent_span_id, waktu mulai-selesai | ⬜ |

### Phase R — 🟡 Intermediate

| Chunk | Konsep utama | Status |
|---|---|---|
| R-I1 — Bloom Filter Index 🔥 | sparse index, granule, probabilistic DS | ⬜ |
| R-I2 — Bucketed Sliding Window 🔥 | fixed vs sliding, memory trade-off | ⬜ |
| R-I3 — SDK Go 🔥 | context propagation, non-blocking sender | ⬜ |
| R-I4 — Redis Fallback | cache-aside, thundering herd, singleflight | ⬜ |
| R-I5 — Delivery Semantics | at-least-once, idempotency | ⬜ |
| R-I6 — Dead Letter Queue | poison message | ⬜ |
| R-I7 — Consumer Lag 🔥 | metrik kesehatan sistem terdistribusi | ⬜ |
| R-I8 — Kafka Durability | acks, ISR, idempotent producer | ⬜ |
| R-I9 — Terima Log OTel (Interop) | satu pintu masuk, satu penegakan | ⬜ |
| R-I10 — PII Scrubbing | kepatuhan data | ⬜ |
| R-I11 — Performa Consumer 2 | batching, pipeline, circuit breaker | ⬜ |

### Phase R — 🔴 Senior

| Chunk | Konsep utama | Status |
|---|---|---|
| R-S1 — Error Fingerprinting 🔥🔥 | normalisasi, grouping, exemplar | ⬜ |
| R-S2 — Materialized View | pra-agregasi, trade-off write vs read | ⬜ |
| R-S3 — Load Test 🔥 | p50/p95/p99, profiling | ⬜ |
| R-S4 — Capacity & Cost | proyeksi kapasitas | ⬜ |
| R-S5 — Backpressure | load shedding, liveness vs readiness | ⬜ |
| R-S6 — Multi-tenant Isolation | row policy, query proxy | ⬜ |
| R-S7 — Chaos Test | RTO/RPO, membuktikan klaim | ⬜ |
| R-S8 — Interview Prep 🔥 | artikulasi | ⬜ |
| R-S9 — Sampling Head vs Tail | trade-off cost vs completeness, kenapa didelegasikan ke Collector | ⬜ |

### Phase S — 🎬 Showcase

| Chunk | Konsep utama | Status |
|---|---|---|
| S-1 — Query API + isolasi tenant 🔥 | JWT, satu pintu baca, batas query | ⬜ |
| S-2 — Bahasa query | tokenizer, SQL berparameter | ⬜ |
| S-3 — Logs + keyset pagination | cursor vs OFFSET | ⬜ |
| S-4 — Live Tail 🔥 | SSE, fan-out, slow consumer | ⬜ |
| S-5 — Issues + Trace API | `-Merge`, pohon span | ⬜ |
| S-6 — Console: Onboarding + Tail | virtualisasi daftar | ⬜ |
| S-7 — Console: Explorer + Issues + Trace | state di URL | ⬜ |
| S-8 — Issue lifecycle 🔥 | `SET NX`, event antar-service | ⬜ |
| S-9 — Mode demo + README 🔥 | naskah 60 detik | ⬜ |

### Phase P — ⚙️ Platform

| Chunk | Konsep utama | Status |
|---|---|---|
| P-1 — k3s di VPS | distribusi Kubernetes, akses aman ke API | ⬜ |
| P-2 — Dockerfile + Helm | templating, requests/limits | ⬜ |
| P-3 — Ingress + TLS | ingress controller, ACME | ⬜ |
| P-4 — GitOps 🔥 | pull vs push, self-heal, sealed secrets | ⬜ |
| P-5 — Operator stateful | StatefulSet, PVC, CRD | ⬜ |
| P-6 — NetworkPolicy + PDB | isolasi jaringan, disruption | ⬜ |
| P-7 — HPA + KEDA 🔥 | sinyal scaling yang benar | ⬜ |
| P-8 — Monitoring + runbook | ServiceMonitor, SLO | ⬜ |
| P-9 — Backup & restore 🔥 | RPO/RTO yang diuji | ⬜ |
| P-10 — Uji kegagalan | self-healing, drain | ⬜ |

---

*Update status: ⬜ Not started → 🟡 In progress → ✅ Done*

---

## 🗺️ Urutan Pengerjaan yang Kusarankan

| Tahap | Kerjakan | Kenapa urutannya begini |
|---|---|---|
| **Sekarang** | `R-J1`, `R-J2`, `R-J3` | Skema dan struct harus final **sebelum** nulis batch insert. Kalau tidak, Chunk 2.4-2.6 harus ditulis ulang |
| **Sebelum Phase 2** | `R-J7` (konsep span) | Prasyarat konseptual `R-I3` (SDK) — lebih murah dipahami sekarang selagi masih di level Junior, sebelum nanti diimplementasikan sungguhan di `R-I3` |
| **Lanjut** | `Chunk 2.4` → `2.5` → `2.6` → `2.7` | Selesaikan Phase 2 dengan skema yang sudah benar |
| **Selingan** | `R-J6` (Prometheus) | Kamu butuh metric ini untuk semua pengukuran setelahnya |
| **Sebelum Phase 3** | `R-I2` (bucketed window) | Supaya Chunk 3.3-3.6 langsung dikerjakan dengan versi yang benar, sekali jalan |
| **Phase 3** | `Chunk 3.1-3.12` + `R-I7` (lag) | — |
| **Setelah Phase 3** | `R-I5`, `R-I6`, `R-I8` | Ketahanan pipeline — ini gugusan konsep yang saling berkaitan, enak dipelajari sekaligus |
| **Phase 4-5** | `Chunk 4.x`, `5.x` + `R-I4`, `R-J5` | — |
| **Sebelum Phase 6** | `R-I3` (SDK) | Dummy app harus pakai SDK, bukan HTTP call manual |
| **Phase 6** | `Chunk 6.x` + `R-S1` (fingerprinting) 🔥 | Fingerprinting harus masuk demo script |
| **Phase 7** | `Chunk 7.x` + `R-S3`, `R-S4`, `R-S7`, `R-S8` | Angka nyata dan cerita yang siap diucapkan |
| **Sisa** | `R-J4`, `R-I1`, `R-I9`, `R-I10`, `R-I11`, `R-S2`, `R-S5`, `R-S6`, `R-S9` | Penting, tapi tidak memblokir apa pun. Sisipkan kapan saja |

> **Kalau waktumu sangat terbatas dan interview sudah dekat:**
> `R-J1` → `R-J2` → `R-I2` → `R-S1` → `R-S8`
> Lima chunk ini saja sudah menutup pertanyaan yang pernah kamu dapat, plus kemampuan menjelaskannya.
