# LogPilot — Product Requirements Document
**Version 1.0 | April 2026 | Author: Firman**

---

## 1. Overview

### 1.1 Project Summary

LogPilot adalah self-hosted centralized log ingestion dan alerting platform yang dirancang sebagai alternatif ringan dari Datadog dan Better Stack. Platform ini memungkinkan developer dan tim engineering untuk mengumpulkan log dari berbagai aplikasi, menyimpannya secara efisien, memvisualisasikannya via Grafana, dan menerima notifikasi otomatis ketika kondisi tertentu terpenuhi.

Project ini dibangun sebagai portofolio backend engineering dengan fokus pada event-driven architecture, distributed systems, dan production-grade observability stack.

### 1.2 Problem Statement

Tim engineering skala kecil hingga menengah menghadapi masalah berikut:

- Log dari banyak aplikasi tersebar di berbagai tempat — tidak ada satu tempat untuk query dan monitor semuanya.
- Setup ELK Stack atau Datadog terlalu kompleks atau mahal untuk tim kecil.
- Tidak ada alerting otomatis yang terhubung ke project management tool seperti ClickUp.
- Aplikasi yang sudah pakai OpenTelemetry tidak punya destination yang mudah di-self-host.

### 1.3 Solution

LogPilot menyediakan:

- Single HTTP endpoint untuk menerima log dari aplikasi manapun.
- Kompatibilitas dengan OpenTelemetry Collector sebagai sumber log alternatif.
- Kafka sebagai message bus yang memastikan tidak ada log hilang saat traffic spike.
- ClickHouse sebagai storage yang dioptimasi untuk log analytics.
- Grafana untuk visualisasi dan query log secara real-time.
- Alert engine dengan sliding window counter dan notifikasi ke ClickUp, Email, dan Slack.
- Management UI berbasis Next.js + Laravel API untuk konfigurasi project, API key, dan alert rules.

### 1.4 Target Audience

| Persona | Kebutuhan | Cara Pakai LogPilot |
|---|---|---|
| Backend Developer | Debug error di production | Query log di Grafana, filter by service dan level |
| SRE / DevOps | Monitor kesehatan sistem | Setup alert rules, terima notifikasi spike error |
| Engineering Manager | Visibilitas across services | Dashboard Grafana per project |
| Interviewer / Recruiter | Lihat portofolio teknikal | Demo live: kirim log, lihat di Grafana, trigger alert |

---

## 2. Goals & Non-Goals

### 2.1 Goals

- Membangun platform log ingestion yang functional dan dapat di-demo secara live.
- Mendemonstrasikan kemampuan event-driven architecture dengan Kafka fan-out pattern.
- Mengintegrasikan observability stack industri: ClickHouse, Grafana, Alertmanager.
- Mengimplementasikan alert engine dengan sliding window counter berbasis Redis.
- Menyediakan dua mode deployment: Docker Compose (VPS) dan Kubernetes (local showcase).
- Membangun dummy application sebagai log source untuk keperluan demo.

### 2.2 Non-Goals

- Bukan pengganti production-grade Datadog atau Splunk untuk enterprise.
- Tidak mendukung log dari binary/agent-based collection (seperti Filebeat atau Fluentd) pada versi pertama.
- Tidak ada billing atau multi-tenancy isolasi tingkat tinggi pada versi pertama.
- Frontend Next.js tidak perlu memiliki visualisasi chart sendiri — Grafana menangani ini.

---

## 3. System Architecture

### 3.1 High Level Architecture

| Layer | Komponen | Tanggung Jawab |
|---|---|---|
| Ingestion Layer | Go Ingestor, OTel Collector | Menerima log dari sumber manapun, validasi, enrich, publish ke Kafka |
| Message Bus | Apache Kafka | Buffer log, fan-out ke multiple consumer group, guarantee no data loss |
| Processing Layer | Consumer 1 (Storage), Consumer 2 (Alert) | Simpan log ke ClickHouse, evaluasi alert rules secara real-time |
| Presentation Layer | Grafana, Next.js, Alertmanager | Visualisasi log, management UI, routing notifikasi |
| Management Layer | Laravel API, PostgreSQL, Redis | Control plane: auth, project, API key, alert rules, notification config |

### 3.2 Data Flow End-to-End

1. Aplikasi kirim HTTP POST ke Go Ingestor dengan API key di header.
2. Ingestor validasi API key via Redis, cek rate limit, validasi payload, enrich dengan `project_id` dan `ingested_at`.
3. Ingestor publish enriched log ke Kafka topic `raw-logs` dengan partition key = `project_id`.
4. Consumer 1 (Storage Worker) poll Kafka, micro-batch 500 messages atau 1 detik, batch INSERT ke ClickHouse, commit offset.
5. Consumer 2 (Alert Evaluator) poll Kafka satu per satu, load alert rules dari cache/PostgreSQL, evaluasi kondisi, increment Redis sliding window counter.
6. Jika counter melewati threshold, Consumer 2 publish ke Kafka topic `alert-events` dan set cooldown di Redis.
7. Alertmanager consume `alert-events`, kirim webhook ke Go Alert Dispatcher.
8. Alert Dispatcher query notification config dari Laravel API, kirim ke ClickUp API, SendGrid, dan Slack secara concurrent via goroutine.
9. Developer query log di Grafana yang connect langsung ke ClickHouse.

### 3.3 Technology Stack

| Komponen | Teknologi | Alasan Pemilihan |
|---|---|---|
| Go Ingestor | Go + Fiber | High throughput, low latency, cocok untuk HTTP ingestor |
| Consumer 1 & 2 | Go + sarama | Native Kafka client untuk Go, production battle-tested |
| Alert Dispatcher | Go | Concurrent goroutine untuk kirim ke multiple destination |
| Laravel API | PHP Laravel + Sanctum | Familiar, rapid development untuk management/control plane |
| Frontend | Next.js + TypeScript | Modern, SSR support, familiar di stack AHU |
| Message Bus | Apache Kafka | Durability, fan-out, replay capability, industry standard |
| Log Storage | ClickHouse | Columnar DB dioptimasi untuk log analytics, sangat cepat untuk append + query besar |
| Cache & Rate Limit | Redis | Sub-millisecond lookup untuk auth dan rate limiting |
| Metadata DB | PostgreSQL | ACID untuk user, project, alert rules |
| Visualization | Grafana | Industry standard, native ClickHouse plugin |
| Alert Routing | Alertmanager | Part of Prometheus stack, production battle-tested |
| OTel Support | OTel Collector | Standard industri observability, forward ke Kafka |
| Container | Docker + Docker Compose | Deployment di VPS untuk live demo |
| Orchestration | Kubernetes + Minikube | Local showcase, K8s manifests untuk portofolio |
| CI/CD | GitHub Actions | Otomatis build image dan deploy |

---

## 4. Service Specifications

### 4.1 Go Ingestor Service

Service utama yang menjadi entry point semua log ke sistem LogPilot.

#### Endpoints

| Method | Path | Deskripsi |
|---|---|---|
| POST | `/v1/ingest` | Terima satu atau batch log dari aplikasi |
| GET | `/health` | Health check untuk Kubernetes liveness probe |
| GET | `/metrics` | Prometheus metrics endpoint |

#### Request Schema

```
POST /v1/ingest
Header: X-API-Key: logpilot_abc123xyz

{
  "level": "ERROR",       // required: DEBUG|INFO|WARN|ERROR|FATAL
  "message": "DB connection timeout",  // required
  "service": "sabh-api",  // required
  "timestamp": "2026-04-17T05:12:00Z", // required, ISO 8601
  "metadata": {           // optional
    "trace_id": "abc-def-123",
    "host": "pod-xyz",
    "user_id": "12345"
  }
}
```

#### Processing Pipeline (5 Layers)

| Layer | Proses | Gagal → Response |
|---|---|---|
| 1. Authentication | GET `api_key:{value}` dari Redis, dapat `project_id` | 401 Unauthorized |
| 2. Rate Limiting | INCR `rate:{project_id}` dengan EXPIRE 60 detik, cek < 10.000/menit | 429 Too Many Requests + Retry-After header |
| 3. Validation | Cek field wajib: level, message, service, timestamp dan format ISO 8601 | 400 Bad Request + detail field yang invalid |
| 4. Enrichment | Tambah `project_id`, `ingested_at` (server time), `ingestor_id` (hostname pod) | — |
| 5. Publish | Kafka produce ke topic `raw-logs`, partition key = `project_id` | 503 Service Unavailable |

#### Enriched Payload (dikirim ke Kafka)

```json
{
  "level": "ERROR",
  "message": "DB connection timeout",
  "service": "sabh-api",
  "timestamp": "2026-04-17T05:12:00Z",
  "metadata": { "trace_id": "abc-def-123" },
  "project_id": "project-sabh",
  "ingested_at": "2026-04-17T05:12:00.347Z",
  "ingestor_id": "ingestor-pod-3"
}
```

---

### 4.2 Kafka Configuration

| Topic | Partitions | Retention | Partition Key | Producer |
|---|---|---|---|---|
| `raw-logs` | 6 | 7 hari | `project_id` | Go Ingestor |
| `alert-events` | 1 | 1 hari | — | Consumer 2 |

**Consumer Groups:**
- `storage-workers` — Consumer 1, baca topic `raw-logs`, commit offset setelah batch insert berhasil.
- `alert-evaluators` — Consumer 2, baca topic `raw-logs` secara independen dari `storage-workers`.

---

### 4.3 Consumer 1 — Storage Worker

Go service yang consume topic `raw-logs` dan melakukan batch insert ke ClickHouse.

#### Micro-batch Strategy

- Poll Kafka dengan batas: **500 messages ATAU 1 detik** (mana yang lebih dulu terpenuhi).
- Satu query INSERT untuk seluruh batch ke ClickHouse.
- Commit offset ke Kafka **HANYA** setelah INSERT berhasil.
- Jika INSERT gagal: jangan commit offset, tunggu 5 detik, retry. Kafka akan deliver ulang batch yang sama.
- Efek: tidak ada log yang hilang meskipun ClickHouse down sementara.

#### ClickHouse Schema

```sql
CREATE TABLE logs (
  project_id   String,
  service      String,
  level        String,
  message      String,
  trace_id     String,
  host         String,
  timestamp    DateTime64(3),
  ingested_at  DateTime64(3)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (project_id, timestamp)
TTL timestamp + INTERVAL 30 DAY;
```

**Design decisions:**
- `MergeTree` — engine default ClickHouse, dioptimasi untuk bulk insert dan range query.
- `PARTITION BY toYYYYMM` — data dipisah per bulan, query bulan tertentu tidak scan semua data.
- `ORDER BY (project_id, timestamp)` — query "log project X dalam rentang waktu Y" sangat cepat.
- `TTL 30 DAY` — auto-delete log lama, tidak perlu cron job manual.

---

### 4.4 Consumer 2 — Alert Evaluator

Go service yang consume topic `raw-logs` dan mengevaluasi setiap log terhadap alert rules yang aktif.

#### Alert Rule Schema

```json
{
  "id": "rule-001",
  "project_id": "project-sabh",
  "name": "High ERROR Rate",
  "condition": {
    "level": "ERROR",
    "service": "sabh-api"
  },
  "threshold": 10,
  "window_seconds": 300,
  "cooldown_seconds": 600
}
```

#### Sliding Window Counter (Redis)

```
# Setiap kali log match kondisi rule:
INCR  alert_counter:{rule_id}:{project_id}
EXPIRE alert_counter:{rule_id}:{project_id} {window_seconds}

# Jika counter >= threshold DAN tidak ada cooldown:
# 1. Publish ke Kafka topic alert-events
# 2. Set cooldown:
SET    cooldown:{rule_id}:{project_id} 1
EXPIRE cooldown:{rule_id}:{project_id} {cooldown_seconds}
```

#### Evaluasi Flow

1. Poll 1 message dari Kafka.
2. Load alert rules dari memory cache (refresh tiap 30 detik dari PostgreSQL).
3. Untuk setiap rule dengan `project_id` yang cocok: cek apakah log match kondisi.
4. Jika match: cek cooldown di Redis.
5. Jika tidak ada cooldown: increment counter di Redis.
6. Jika counter >= threshold: publish ke `alert-events`, set cooldown.
7. Commit offset ke Kafka.

---

### 4.5 Go Alert Dispatcher

Go service yang menerima webhook dari Alertmanager dan mendistribusikan notifikasi ke destination yang dikonfigurasi user.

| Destination | Method | Detail |
|---|---|---|
| ClickUp | POST `/api/v2/list/{list_id}/task` | Create task dengan priority Urgent, nama [ALERT] + detail |
| SendGrid | POST `/v3/mail/send` | Email ke recipients yang dikonfigurasi per project |
| Slack | POST webhook URL | Message ke channel yang dikonfigurasi (opsional) |

Semua destination dikirim secara **concurrent** menggunakan Go goroutine dengan timeout 10 detik per destination.

---

### 4.6 Laravel API — Management Control Plane

Laravel API adalah backend dari management UI. Tidak berada di data path log sama sekali — hanya mengelola konfigurasi yang digunakan service lain.

#### Endpoints

| Method | Path | Fungsi | Side Effect |
|---|---|---|---|
| POST | `/api/auth/register` | Register user baru | — |
| POST | `/api/auth/login` | Login, return Sanctum token | — |
| POST | `/api/projects` | Buat project baru | Insert ke PostgreSQL |
| GET | `/api/projects` | List semua project milik user | — |
| POST | `/api/projects/{id}/api-keys` | Generate API key baru | Insert hash ke PostgreSQL + seed plain key ke Redis |
| DELETE | `/api/projects/{id}/api-keys/{keyId}` | Revoke API key | Hapus dari PostgreSQL + hapus dari Redis |
| GET | `/api/projects/{id}/alert-rules` | List alert rules project | — |
| POST | `/api/projects/{id}/alert-rules` | Buat alert rule baru | Insert ke PostgreSQL, Consumer 2 baca ini tiap 30 detik |
| PUT | `/api/projects/{id}/alert-rules/{ruleId}` | Update alert rule | Update PostgreSQL |
| DELETE | `/api/projects/{id}/alert-rules/{ruleId}` | Hapus alert rule | Delete dari PostgreSQL |
| PUT | `/api/projects/{id}/notifications` | Set config notifikasi | Update PostgreSQL, Alert Dispatcher baca ini |
| GET | `/internal/projects/{id}/notifications` | Internal endpoint untuk Alert Dispatcher | — |

#### PostgreSQL Schema

| Tabel | Kolom Utama | Relasi |
|---|---|---|
| `users` | id, name, email, password, created_at | — |
| `projects` | id, user_id, name, slug, created_at | belongs to users |
| `api_keys` | id, project_id, name, key_hash, last_used_at, created_at | belongs to projects |
| `alert_rules` | id, project_id, name, condition (JSON), threshold, window_seconds, cooldown_seconds, is_active | belongs to projects |
| `notification_configs` | id, project_id, clickup_list_id, clickup_assignee_id, email_recipients (JSON), slack_webhook_url, created_at | belongs to projects |
| `alert_logs` | id, rule_id, project_id, triggered_at, dispatch_results (JSON) | belongs to alert_rules |

---

## 5. Deployment Strategy

### 5.1 Dua Mode Deployment

| Mode | Tech | Tujuan | Environment |
|---|---|---|---|
| Docker Compose | Docker + Compose | Self-hosted sederhana, live demo saat interview | VPS Production |
| Kubernetes | Minikube / OrbStack | Production-grade showcase dengan HPA dan rolling update | Local M4 |

### 5.2 Docker Compose Stack

Services yang di-define dalam `docker-compose.yml`:

- `go-ingestor` — custom image, port 8080
- `consumer-storage` — custom image
- `consumer-alert` — custom image
- `alert-dispatcher` — custom image
- `laravel-api` — custom image, port 8000
- `nextjs-frontend` — custom image, port 3000
- `kafka` + `zookeeper` — bitnami/kafka
- `clickhouse` — clickhouse/clickhouse-server
- `redis` — redis:alpine
- `postgres` — postgres:15
- `grafana` — grafana/grafana
- `alertmanager` — prom/alertmanager

### 5.3 Kubernetes Setup (Local Showcase)

Setiap custom service memiliki Kubernetes manifest di folder `k8s/`:

- `deployment.yaml` — define pod, image, environment variables dari ConfigMap/Secret
- `service.yaml` — ClusterIP untuk internal communication, LoadBalancer untuk Ingestor
- `hpa.yaml` — Horizontal Pod Autoscaler khusus untuk Go Ingestor (scale 2-10 pods berdasarkan CPU)

Infrastructure dependencies (Kafka, ClickHouse, Redis, PostgreSQL, Grafana, Alertmanager) di-deploy menggunakan Helm Charts.

### 5.4 CI/CD Pipeline (GitHub Actions)

1. Developer push ke branch `main` di GitHub.
2. GitHub Actions trigger: run unit tests untuk setiap service Go.
3. Build Docker image untuk setiap custom service.
4. Push image ke GitHub Container Registry (ghcr.io) dengan tag berdasarkan commit SHA.
5. Update image tag di Kubernetes deployment manifest.
6. (Opsional) `kubectl apply` otomatis ke cluster.

---

## 6. Demo Application

### 6.1 Dummy E-commerce API (Go)

Aplikasi HTTP sederhana yang mensimulasikan backend e-commerce dengan berbagai jenis log.

| Endpoint | Behavior | Log yang Dihasilkan |
|---|---|---|
| GET `/products` | Selalu sukses | INFO: fetch products success |
| GET `/products/:id` | 80% sukses, 20% not found | INFO atau ERROR: product not found |
| POST `/checkout` | 70% sukses, 30% payment timeout | INFO atau ERROR: payment gateway timeout |
| GET `/health` | Selalu sukses | DEBUG: health check |

### 6.2 Log Generator Script (Go)

Script terpisah yang mengirim log random ke Ingestor secara otomatis untuk mensimulasikan traffic dan trigger alert.

- Kirim 1 log setiap 200ms (5 logs/detik) dalam kondisi normal.
- Setiap 2 menit: simulate error spike — kirim 20 ERROR logs dalam 30 detik untuk trigger alert.
- Mix level: 60% INFO, 20% WARN, 15% ERROR, 5% FATAL.

### 6.3 Demo Script (Urutan Presentasi)

1. Tunjukkan arsitektur di README atau diagram — jelaskan komponennya.
2. Tunjukkan kode SDK integration di dummy app — hanya 5-10 baris.
3. Hit beberapa endpoint dummy app, buka Grafana — log muncul dalam < 1 detik.
4. Tunjukkan query di Grafana: filter by `level=ERROR`, filter by service.
5. Jalankan log generator — tunjukkan traffic masuk secara real-time di Grafana.
6. Tunggu error spike — alert trigger, tunjukkan ClickUp task ter-create otomatis.
7. Tunjukkan email notifikasi masuk.
8. (Opsional) Tunjukkan K8s manifests dan video Minikube deployment.

---

## 7. Development Build Plan

**Estimasi waktu:** 5-6 minggu, part-time (weekend + malam hari)
**Mulai:** Sabtu, 19 April 2026

| Minggu | Milestone | Deliverable | Tech Focus |
|---|---|---|---|
| 1 (19-26 Apr) | Foundation & Ingestor | Monorepo setup, Docker Compose jalan, Go Ingestor selesai dengan semua 5 layer pipeline | Go, Kafka, Redis, Docker |
| 2 (27 Apr - 3 Mei) | Storage Pipeline | Consumer 1 selesai, ClickHouse schema dibuat, log bisa masuk dari Ingestor sampai tersimpan di ClickHouse, Grafana connect | Go, ClickHouse, Grafana |
| 3 (4-10 Mei) | Alert Pipeline | Consumer 2 selesai dengan sliding window counter, Alertmanager config, Go Alert Dispatcher selesai dengan ClickUp + Email integration | Go, Redis, Alertmanager, ClickUp API, SendGrid |
| 4 (11-17 Mei) | Management Layer | Laravel API semua endpoint selesai, PostgreSQL schema, API key seed ke Redis, Next.js frontend halaman utama | PHP Laravel, PostgreSQL, Next.js |
| 5 (18-24 Mei) | Demo App & Integration | Dummy e-commerce app selesai, log generator script, end-to-end test semua komponen, Kubernetes manifests | Go, Docker, Kubernetes |
| 6 (25-31 Mei) | Polish & Documentation | README lengkap dengan arsitektur diagram, demo video direkam, GitHub Actions CI/CD, final testing | Documentation, GitHub Actions |

### 7.1 Detail Minggu 1

**Sabtu 19 April — Setup foundation:**
- Init monorepo structure di GitHub.
- Setup `docker-compose.yml` dengan Kafka, Redis, PostgreSQL, ClickHouse.
- Pastikan semua infrastructure service bisa jalan dengan `docker-compose up`.
- Init Go module untuk ingestor service.

**Minggu 20 April — Build Ingestor Layer 1-3:**
- Implement HTTP server dengan Fiber.
- Layer 1: Authentication via Redis lookup.
- Layer 2: Rate limiting dengan Redis counter.
- Layer 3: Payload validation.

**Selasa-Jumat 21-25 April — Build Ingestor Layer 4-5 + Testing:**
- Layer 4: Enrichment logic.
- Layer 5: Kafka producer publish ke `raw-logs`.
- Unit test untuk setiap layer.
- Manual test dengan curl: kirim log, verifikasi di Kafka console.

---

## 8. Repository Structure

```
logpilot/
├── services/
│   ├── ingestor/               # Go Ingestor Service
│   │   ├── main.go
│   │   ├── handler/
│   │   ├── middleware/         # auth, rate limit, validation
│   │   ├── enricher/
│   │   ├── producer/           # Kafka producer
│   │   └── Dockerfile
│   │
│   ├── consumer-storage/       # Consumer 1
│   │   ├── main.go
│   │   ├── consumer/
│   │   ├── storage/            # ClickHouse client
│   │   └── Dockerfile
│   │
│   ├── consumer-alert/         # Consumer 2
│   │   ├── main.go
│   │   ├── consumer/
│   │   ├── evaluator/          # rule evaluation logic
│   │   ├── counter/            # Redis sliding window
│   │   └── Dockerfile
│   │
│   ├── alert-dispatcher/       # Go Alert Dispatcher
│   │   ├── main.go
│   │   ├── handler/            # webhook receiver
│   │   ├── destinations/       # clickup, sendgrid, slack
│   │   └── Dockerfile
│   │
│   ├── laravel-api/            # PHP Laravel
│   │   ├── app/
│   │   ├── routes/api.php
│   │   └── Dockerfile
│   │
│   └── frontend/               # Next.js
│       ├── pages/
│       └── Dockerfile
│
├── demo/
│   ├── dummy-app/              # Go e-commerce dummy
│   └── log-generator/          # Go log spam script
│
├── deploy/
│   ├── docker-compose/
│   │   ├── docker-compose.yml
│   │   └── .env.example
│   └── kubernetes/
│       ├── ingestor/
│       ├── consumer-storage/
│       ├── consumer-alert/
│       ├── alert-dispatcher/
│       ├── laravel-api/
│       └── helm/
│
├── config/
│   ├── grafana/                # dashboard JSON
│   ├── alertmanager/           # alertmanager.yml
│   └── otel-collector/         # otel-collector-config.yaml
│
├── .github/workflows/
│   └── ci.yml                  # GitHub Actions
│
├── docs/
│   ├── PRD.md
│   └── TODO.md
│
└── README.md
```

---

## 9. Success Metrics & Definition of Done

### 9.1 Functional Requirements

| # | Requirement | Acceptance Criteria |
|---|---|---|
| F1 | Log ingestion berfungsi | POST `/v1/ingest` dengan valid API key return 202, log muncul di ClickHouse dalam 5 detik |
| F2 | Authentication berfungsi | Request tanpa API key atau dengan key invalid return 401 |
| F3 | Rate limiting berfungsi | Request ke-10.001 dalam 60 detik return 429 |
| F4 | Grafana query berfungsi | Query log di Grafana bisa filter by `project_id`, `level`, `service`, time range |
| F5 | Alert trigger berfungsi | Setelah 10 ERROR dalam 5 menit, ClickUp task ter-create dan email terkirim |
| F6 | Cooldown berfungsi | Alert tidak trigger lagi dalam 10 menit setelah trigger pertama |
| F7 | API key management | Generate API key via Laravel API, key langsung bisa digunakan di Ingestor |
| F8 | Alert rule management | Buat alert rule via Laravel API, rule aktif dalam 30 detik di Consumer 2 |
| F9 | No data loss | Matikan ClickHouse 1 menit, nyalakan kembali — semua log yang masuk selama downtime tersimpan |

### 9.2 Non-Functional Requirements

| Aspek | Target |
|---|---|
| Log ingestion latency | < 100ms dari HTTP request ke Kafka publish |
| Log visibility latency | < 5 detik dari ingest ke muncul di Grafana |
| Alert latency | < 30 detik dari kondisi terpenuhi ke notifikasi terkirim |
| Throughput | Go Ingestor mampu handle 1.000 req/detik di single pod |
| Availability | Sistem tetap menerima log meskipun ClickHouse down (Kafka buffer) |

---

*LogPilot PRD v1.0 | Firman | April 2026 | Confidential — Portfolio Use Only*
