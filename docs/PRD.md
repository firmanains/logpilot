# LogPilot — Product Requirements Document
**Version 3.1 | November 2026 | Author: Firman**

> **Changelog v3.0 → v3.1** (22 November 2026)
>
> Empat keputusan setelah review rencana. Pipeline inti tidak berubah.
>
> 1. **Laravel API diganti Go** (§4.6). Management API menjadi modul di dalam satu service `api` bersama Query API (§4.8). Endpoint `/internal/*` pindah ke listener terpisah yang tidak keluar lewat Ingress. Alasannya: satu bahasa, satu service lebih sedikit, dan sisi PHP ditunjukkan lewat handler Monolog (butir 2), bukan lewat control plane.
> 2. **Handler Monolog untuk PHP naik menjadi butir Stretch pertama** (§12), di atas ringkasan insiden dengan LLM.
> 3. **Perkakas platform dipasang satu per satu** (§6.4): urutan tetap, satu perkakas per sesi, dan boleh berhenti sebelum semuanya terpasang.
> 4. **Tulisan teknis pendek per masalah sulit** menjadi bagian dari deliverable (§7.5).
>
> Positioning proyek ditulis eksplisit di §1.4: backend engineer yang memahami platform dan DevOps.
>
> Ditunda: baris biaya VPS per bulan di README.

> **Changelog v2.0 → v3.0**
>
> v2.0 membetulkan arsitektur pipeline. v3.0 membetulkan lima kesalahan teknis yang masih tersisa, lalu menambah dua hal yang belum ada: sesuatu yang bisa **dilihat dan diklik** orang non-teknis, dan jalur belajar **DevOps/Kubernetes** yang sungguhan.
>
> **Diperbaiki:**
> 1. **Materialized view issue menghitung salah** (§5.3). Di `AggregatingMergeTree`, kolom biasa hasil `count()`, `min()`, `max()`, `argMin()` tidak dijumlahkan saat merge; ClickHouse hanya menyimpan salah satu nilainya. Hanya `uniqState` yang benar. Diganti tabel target eksplisit dengan `SimpleAggregateFunction` dan `-State`.
> 2. **Deduplikasi insert tidak pernah aktif** (§4.3). Untuk `MergeTree` non-replicated, `non_replicated_deduplication_window` bawaannya 0, jadi `insert_deduplication_token` diabaikan. Token dari rentang offset juga tidak stabil setelah crash karena batas batch ditentukan waktu. Ditambah setting tabel dan mekanisme high-water mark.
> 3. **DDL `logs` gagal dijalankan** (§4.3). `TTL timestamp + INTERVAL 30 DAY` ditolak untuk `DateTime64`; ini sudah tercatat di `PROBLEM.md` Issue #4 tapi PRD belum ikut diperbaiki. Partisi juga diubah dari bulanan ke harian supaya TTL 30 hari bisa membuang part utuh.
> 4. **Rate limiter memakai pola yang dikritik PRD ini sendiri** (§4.1 Layer 2). `INCRBY` lalu `EXPIRE` terpisah bisa meninggalkan key tanpa TTL, sehingga satu project terkena 429 selamanya. Diganti Lua atomic yang sama dengan counter alert.
> 5. **Timeline §8 sudah lewat dan salah rujukan** (`R-I3` dipakai untuk dua chunk berbeda). Diganti rencana berbasis milestone dengan exit demo.
>
> **Ditambahkan:**
> 6. **Scope tiers** (§2.3): Core → Showcase → Platform → Stretch, dengan aturan gerbang. Ini jawaban untuk keinginan "implement banyak hal" tanpa mengulang nasib timeline v2.0.
> 7. **Query API (Go)** (§4.8). Pintu baca tunggal yang memaksa filter `project_id`, sehingga keterbatasan multi-tenant di v2.0 selesai, sekaligus menjadi backend untuk Console.
> 8. **LogPilot Console** (§4.9). UI sendiri di Next.js: onboarding, Live Tail, Explorer, Issues, Trace waterfall. Grafana tetap dipakai untuk kesehatan sistem.
> 9. **Issue lifecycle dan jenis alert baru** (§5.5): status open/resolved/ignored, alert *new issue* dan *regression*.
> 10. **Platform & DevOps track** (§6): k3s di VPS, Helm, GitOps, ingress + TLS, operator Kafka/ClickHouse, autoscaling dari consumer lag, backup, dan live demo publik.
> 11. **Production readiness checklist** (§11): setiap butir punya bukti yang bisa ditunjukkan.
>
> Rincian implementasi ada di `docs/TODO.md` **PHASE S** (Showcase) dan **PHASE P** (Platform).

> **Changelog v1.0 → v2.0**
>
> Versi ini memperbaiki tiga kesalahan arsitektur di v1.0 dan menambahkan satu kapabilitas inti yang sebelumnya tidak ada.
>
> **Diperbaiki:**
> 1. **Alertmanager tidak bisa consume Kafka** (v1.0 §3.2 langkah 7). Alertmanager hanya menerima alert lewat HTTP `POST /api/v2/alerts`. Flow diubah jadi push-based.
> 2. **Sliding window counter bukan sliding, dan punya bug.** `INCR` + `EXPIRE` di setiap match membuat TTL ter-reset terus sehingga counter tidak pernah kedaluwarsa. Diganti dengan bucketed counter.
> 3. **Silent data loss di skema ClickHouse.** Payload menerima `metadata` bebas, tapi tabel hanya punya kolom `trace_id` dan `host`. Field lain hilang tanpa error.
>
> **Ditambahkan:**
> 4. **Trace context + error fingerprinting** (§5 — bagian baru). Ini kapabilitas yang menjawab pertanyaan "bagaimana membedakan 100.000 error yang terjadi bersamaan".
> 5. **SDK client** sebagai komponen resmi — tempat `trace_id` dilahirkan.
> 6. **Prometheus** ke dalam stack. v1.0 punya `/metrics` dan Alertmanager, tapi tidak ada yang men-scrape.
>
> Rincian implementasi per perubahan ada di `docs/TODO.md` **PHASE R**.

---

## 1. Overview

### 1.1 Project Summary

LogPilot adalah self-hosted centralized log ingestion dan alerting platform yang dirancang sebagai alternatif ringan dari Datadog dan Better Stack. Platform ini memungkinkan developer dan tim engineering untuk mengumpulkan log dari berbagai aplikasi, menyimpannya secara efisien, memvisualisasikannya via Grafana, dan menerima notifikasi otomatis ketika kondisi tertentu terpenuhi.

Yang membedakan LogPilot dari sekadar penampung log: platform ini **mengelompokkan error berdasarkan penyebabnya**, bukan hanya menyimpannya. Saat terjadi insiden dengan puluhan ribu error, LogPilot menampilkan berapa **jenis** masalah yang sedang terjadi, bukan daftar panjang yang tidak terbaca.

Project ini dibangun sebagai portofolio backend engineering dengan fokus pada event-driven architecture, distributed systems, dan production-grade observability stack.

**Siapa yang mengerjakan.** Seluruh kode, manifest, konfigurasi, dan operasi cluster di repo ini ditulis dan dijalankan oleh Firman sendiri. Tujuannya *get my hands dirty*: paham apa dan bagaimananya, bukan sekadar punya hasilnya. AI agent hanya mengajarkan, memandu, mengarahkan, dan me-review; ia tidak mengimplementasikan fitur, tidak menjalankan perintah di cluster, dan tidak memperbaiki bug atas nama Firman. Aturan lengkapnya ada di `CLAUDE.md`. Dokumen ini (PRD dan TODO) adalah satu-satunya bagian yang boleh disusun bersama AI, dan setiap DDL atau angka di dalamnya tetap harus dibuktikan sendiri lewat kriteria uji di §10.

### 1.2 Problem Statement

Tim engineering skala kecil hingga menengah menghadapi masalah berikut:

- Log dari banyak aplikasi tersebar di berbagai tempat — tidak ada satu tempat untuk query dan monitor semuanya.
- Setup ELK Stack atau Datadog terlalu kompleks atau mahal untuk tim kecil.
- Tidak ada alerting otomatis yang terhubung ke project management tool seperti ClickUp.
- Aplikasi yang sudah pakai OpenTelemetry tidak punya destination yang mudah di-self-host.
- **Saat insiden terjadi, log mentah tidak bisa dibaca.** 50.000 baris error dalam 2 menit tidak memberi tahu apa pun tentang ada berapa masalah yang sebenarnya terjadi.

### 1.3 Solution

LogPilot menyediakan:

- Single HTTP endpoint untuk menerima log dari aplikasi manapun, satuan maupun batch.
- SDK Go yang menangani trace context propagation dan pengiriman non-blocking.
- Kompatibilitas dengan OpenTelemetry lewat endpoint OTLP/HTTP.
- Kafka sebagai message bus yang memastikan tidak ada log hilang saat traffic spike.
- ClickHouse sebagai storage yang dioptimasi untuk log analytics.
- **Distributed tracing** — korelasi seluruh log satu request lewat `trace_id`, dan lokalisasi titik gagal lewat `span_id`.
- **Error fingerprinting** — pengelompokan otomatis error berdasarkan penyebab, bukan berdasarkan teks persisnya.
- **LogPilot Console** (Next.js): Live Tail, Explorer, Issues, dan Trace waterfall dalam satu UI, ditenagai **Query API** yang mengisolasi data per project.
- Grafana untuk dashboard kesehatan LogPilot sendiri (consumer lag, throughput, latency).
- Alert engine dengan bucketed sliding window counter, alert *new issue* dan *regression*, dan notifikasi ke ClickUp, Email, dan Slack.
- Management API (Go, satu service dengan Query API) sebagai control plane untuk project, API key, dan alert rules.
- **Live demo publik** yang berjalan di Kubernetes (k3s) dan di-deploy lewat GitOps.

### 1.4 Target Audience

| Persona | Kebutuhan | Cara Pakai LogPilot |
|---|---|---|
| Backend Developer | Debug error di production | Buka Issues, klik satu issue, telusuri satu trace utuh di waterfall |
| SRE / DevOps | Monitor kesehatan sistem | Setup alert rules, pantau consumer lag, terima notifikasi spike error |
| Engineering Manager | Visibilitas across services | Halaman Issues per project, berapa issue distinct per minggu |
| HR / Recruiter (non-teknis) | Menilai dalam 60 detik tanpa instal apa pun | Buka URL demo, lihat log mengalir, tekan "Simulasikan insiden", lihat 50.000 error menjadi 5 issue |
| Interviewer (teknis) | Menguji kedalaman | README dengan angka terukur, `DESIGN-DECISIONS.md`, repo GitOps, chaos test |

**Positioning (v3.1).** Proyek ini menampilkan pemiliknya sebagai backend engineer yang memahami platform dan DevOps. Backend adalah wajah depannya: pipeline, keandalan data, dan angka terukur. Platform adalah buktinya bahwa sistem yang sama bisa ia deploy, pulihkan, dan skalakan sendiri. Keduanya satu cerita, jadi tidak ada tier yang dibuang demi yang lain; yang diatur hanya urutan polesannya.

---

## 2. Goals & Non-Goals

### 2.1 Goals

- Membangun platform log ingestion yang functional dan dapat di-demo secara live.
- Mendemonstrasikan kemampuan event-driven architecture dengan Kafka fan-out pattern.
- Mengintegrasikan observability stack industri: ClickHouse, Grafana, Prometheus, Alertmanager.
- Mengimplementasikan alert engine dengan bucketed sliding window counter berbasis Redis.
- **Mengimplementasikan trace context (W3C) dan error grouping — kapabilitas yang membedakan platform observability dari sekadar log storage.**
- **Membuktikan setiap klaim performa dan ketahanan dengan angka hasil pengukuran, bukan target.**
- Menyediakan dua mode deployment: Docker Compose (development lokal) dan Kubernetes (k3s di VPS untuk live demo, kubeadm multi-node sebagai latihan).
- Membangun dummy application dan SDK sebagai log source untuk keperluan demo.
- **Menyediakan UI sendiri yang bisa dipahami orang non-teknis dalam 60 detik**, lewat URL publik tanpa instalasi.
- **Menjadi kendaraan belajar DevOps**: mengoperasikan sendiri cluster, deployment, TLS, autoscaling, backup, dan pemulihan untuk sistem stateful (Kafka, ClickHouse, PostgreSQL).

### 2.2 Non-Goals

- Bukan pengganti production-grade Datadog atau Splunk untuk enterprise.
- Tidak mendukung log dari binary/agent-based collection (Filebeat, Fluentd) pada versi pertama.
- Tidak ada billing, organisasi, RBAC bertingkat, atau SSO. Satu user memiliki project-nya sendiri; itu cukup untuk tujuan project ini.
- Tidak menyimpan metrics. Metrics tetap di Prometheus dan tidak diduplikasi ke ClickHouse.
- Console tidak membuat ulang dashboard kesehatan sistem. Grafana sudah menanganinya dan Console cukup menautkannya.
- Tidak ada full-text search engine seperti Elasticsearch. Pencarian teks memakai `tokenbf_v1` skip index ClickHouse, yang cukup untuk pola `LIKE '%keyword%'` tapi bukan pengganti relevance ranking.
- Tidak mengejar high availability di live demo. Cluster demo satu node; konfigurasi HA (3 broker, replika ClickHouse) didokumentasikan dan dilatih di lab, tidak dijalankan 24 jam.

> **Dicabut dari non-goals v2.0:** "isolasi multi-tenant di lapisan query belum ada" dan "frontend tidak perlu visualisasi sendiri". Keduanya sekarang ditangani Query API (§4.8) dan Console (§4.9). Grafana tetap terhubung langsung ke ClickHouse, tetapi hanya untuk operator LogPilot, bukan untuk tenant.

### 2.3 Scope Tiers

v2.0 menjadwalkan semuanya dalam 8 minggu dan tidak ada yang selesai tepat waktu. v3.0 menambah lebih banyak fitur, jadi urutannya harus dijaga dengan aturan, bukan niat.

| Tier | Isi | Exit demo (bukti tier ini selesai) |
|---|---|---|
| **1. Core** | Ingestor, Kafka, Consumer 1, ClickHouse, trace context, Consumer 2, Alertmanager, Dispatcher, Management API minimum, SDK, fingerprinting | `curl` satu log → muncul di ClickHouse; 50.000 error → 5 baris di query issue; alert masuk ke ClickUp |
| **2. Showcase** | Query API, Console (onboarding, Live Tail, Explorer, Issues, Trace waterfall), issue lifecycle, alert new issue/regression, tombol simulasi insiden, README dengan GIF | Orang non-teknis membuka Console di lokal dan memahami demo tanpa dijelaskan |
| **3. Platform** | k3s di VPS, Helm chart, ingress + TLS, GitOps, operator Kafka/ClickHouse, KEDA, kube-prometheus-stack, backup, live demo publik | URL publik ber-HTTPS; `git push` → ter-deploy sendiri; pod dibunuh → pulih tanpa kehilangan log |
| **4. Stretch** | Handler Monolog untuk PHP, ringkasan insiden dengan LLM, SLO burn-rate alert, cold storage ke object storage, dogfooding, cluster kubeadm multi-node | Masing-masing berdiri sendiri; tidak ada yang wajib |

**Aturan gerbang:**
1. Tier berikutnya baru dimulai setelah exit demo tier sebelumnya direkam (video pendek atau GIF di `docs/demo/`).
2. Pengecualian: pekerjaan Platform yang tidak bergantung pada fitur (cluster, ingress, TLS, GitOps) boleh berjalan paralel dengan Core sebagai selingan belajar. Yang di-deploy cukup Ingestor saja.
3. Butir Stretch tidak pernah menggeser butir tier di atasnya. Fitur yang paling "fancy" tetapi berdiri di atas pipeline yang belum jalan adalah persis yang dimaksud *level skipping* di `CLAUDE.md`.

---

## 3. System Architecture

### 3.1 High Level Architecture

| Layer | Komponen | Tanggung Jawab |
|---|---|---|
| Client Layer | LogPilot SDK (Go), OTel SDK | Membuat dan mempropagasi trace context, mengirim log secara non-blocking |
| Ingestion Layer | Go Ingestor | **Satu-satunya pintu masuk.** Auth, rate limit, validasi, enrichment, fingerprint, publish ke Kafka |
| Message Bus | Apache Kafka | Buffer log, fan-out ke multiple consumer group, guarantee no data loss |
| Processing Layer | Consumer 1 (Storage), Consumer 2 (Alert) | Simpan log ke ClickHouse, evaluasi alert rules secara real-time |
| **Query Layer** | **Go Query API** | **Baru di v3.0. Satu-satunya pintu baca untuk tenant.** Memaksa filter `project_id`, membatasi biaya query, menyediakan stream Live Tail |
| Presentation Layer | LogPilot Console (Next.js), Grafana | Console untuk tenant (log, issue, trace, konfigurasi). Grafana untuk operator (kesehatan sistem) |
| Alerting Layer | Alertmanager, Alert Dispatcher | Grouping/dedup/silencing alert, routing notifikasi |
| Observability Layer | Prometheus, kafka-exporter | Monitoring LogPilot itu sendiri — consumer lag, latency, error rate |
| Management Layer | Management API (Go, satu service `api` dengan Query API), PostgreSQL, Redis | Control plane: auth, project, API key, alert rules, notification config, status issue |
| **Platform Layer** | **Kubernetes (k3s), Helm, Argo CD, ingress-nginx, cert-manager** | **Baru di v3.0.** Menjalankan, memperbarui, dan memulihkan semua lapisan di atas |

**Catatan arsitektur — satu pintu baca.** Simetris dengan "satu pintu masuk" di sisi tulis. Di v2.0 tenant membaca lewat Grafana yang terhubung langsung ke ClickHouse dengan satu kredensial, sehingga tidak ada yang mencegah query ke `project_id` milik orang lain. Di v3.0 semua pembacaan tenant melewati Query API, yang mengambil `project_id` dari token dan tidak pernah dari parameter request.

**Catatan arsitektur — satu pintu masuk.** Di v1.0, OTel Collector mengirim langsung ke Kafka, sehingga melewati autentikasi, rate limiting, dan validasi sepenuhnya, dan tidak ada mekanisme yang menentukan `project_id` untuk jalur tersebut. Di v2.0, OTel Collector mengirim ke endpoint OTLP di Ingestor. Semua log — apa pun sumbernya — melewati pipeline penegakan yang sama.

### 3.2 Data Flow End-to-End

1. **Aplikasi** membuat `trace_id` di middleware paling awal (via SDK), atau melanjutkan `trace_id` dari header `traceparent` kalau request datang dari service lain.
2. Setiap log yang lahir selama request itu membawa `trace_id` yang sama dan `span_id` masing-masing.
3. SDK menampung log di buffer memory dan mengirimnya secara batch dan asynchronous ke Ingestor. Kalau buffer penuh, log dibuang — aplikasi user tidak pernah diblokir.
4. **Ingestor** menerima HTTP POST dengan API key di header, lalu menjalankan 6 layer pipeline (§4.1).
5. Ingestor publish enriched log ke Kafka topic `raw-logs` dengan partition key = `project_id`.
6. **Consumer 1 (Storage Worker)** poll Kafka, micro-batch 500 messages atau 1 detik, batch INSERT ke ClickHouse dengan `insert_deduplication_token`, commit offset setelah insert berhasil.
7. **Consumer 2 (Alert Evaluator)** poll Kafka secara batch dan independen, load alert rules dari cache, evaluasi kondisi, increment bucketed counter di Redis.
8. Jika total counter dalam window melewati threshold **dan** tidak ada cooldown aktif, Consumer 2 melakukan `POST /api/v2/alerts` ke **Alertmanager**, menulis salinan ke Kafka topic `alert-events` sebagai audit trail, lalu set cooldown di Redis.
9. **Alertmanager** melakukan grouping, deduplication, dan inhibition, lalu mengirim webhook ke Go Alert Dispatcher.
10. **Alert Dispatcher** mengambil notification config dari cache lokal (fallback ke Management API), lalu mengirim ke ClickUp API, SendGrid, dan Slack secara concurrent via goroutine.
11. **Developer** membuka Console: mulai dari halaman Issues (agregasi per fingerprint), klik satu issue, buka exemplar `trace_id` sebagai waterfall untuk melihat seluruh log satu request. Console memanggil **Query API**, yang menerjemahkannya menjadi SQL ClickHouse dengan filter `project_id` dari token.
12. Untuk **Live Tail**, Query API membaca topic `raw-logs` langsung dari Kafka (tanpa consumer group) dan meneruskannya ke browser lewat Server-Sent Events.
13. **Prometheus** men-scrape `/metrics` semua service dan kafka-exporter, sehingga kesehatan LogPilot sendiri terpantau di Grafana.

> **Perubahan dari v1.0:** langkah 8-9. v1.0 menyatakan Alertmanager consume topic `alert-events` dari Kafka. Alertmanager tidak memiliki Kafka consumer — satu-satunya jalur masuk adalah HTTP `POST /api/v2/alerts`.
>
> **Kenapa Alertmanager tetap dipakai** meski Consumer 2 bisa langsung memanggil Dispatcher: Alertmanager mengurus empat hal yang tidak sepele untuk ditulis sendiri — grouping (50 alert sejenis jadi 1 notifikasi), deduplication, inhibition (alert "service down" menahan alert "error rate tinggi" untuk service yang sama), dan silencing saat maintenance.
>
> Topic `alert-events` tetap ada, fungsinya berubah menjadi audit trail dan bahan replay, bukan jalur utama.

### 3.3 Technology Stack

| Komponen | Teknologi | Alasan Pemilihan |
|---|---|---|
| Client SDK | Go | Trace context propagation, buffered non-blocking sender |
| Go Ingestor | Go + Fiber | High throughput, low latency, cocok untuk HTTP ingestor |
| Consumer 1 & 2 | Go + sarama | Native Kafka client untuk Go, production battle-tested |
| Alert Dispatcher | Go | Concurrent goroutine untuk kirim ke multiple destination |
| Query API | Go + Fiber, SSE | **Baru di v3.0.** Jalur baca tenant; Live Tail lewat Server-Sent Events (satu arah, cukup HTTP biasa, lolos proxy tanpa konfigurasi WebSocket) |
| Management API | Go + Fiber, satu service `api` dengan Query API | **v3.1, menggantikan Laravel.** Control plane: sesi user, project, API key, alert rule. Menerbitkan JWT berumur pendek untuk jalur baca (§4.8) |
| Frontend | Next.js + TypeScript | Modern, SSR support, familiar di stack AHU |
| Message Bus | Apache Kafka (KRaft) | Durability, fan-out, replay capability, industry standard. KRaft mode — tanpa Zookeeper |
| Log Storage | ClickHouse | Columnar DB dioptimasi untuk log analytics, sangat cepat untuk append + query besar |
| Cache & Counter | Redis | Sub-millisecond lookup untuk auth, rate limiting, dan bucketed counter |
| Metadata DB | PostgreSQL | ACID untuk user, project, alert rules. Juga sumber kebenaran saat cache Redis miss |
| Visualization | Grafana | Industry standard, native ClickHouse plugin |
| Metrics | **Prometheus + kafka-exporter** | **Baru di v2.0.** Scrape `/metrics` semua service; consumer lag monitoring |
| Alert Routing | Alertmanager | Grouping, dedup, inhibition, silencing |
| OTel Support | OTel Collector → OTLP endpoint di Ingestor | Standard industri, tapi tetap lewat pipeline auth LogPilot |
| Load Testing | k6 | Mengukur throughput dan p50/p95/p99 secara reproducible |
| Container | Docker + Docker Compose | Development lokal: satu perintah menyalakan seluruh stack |
| Orchestration | Kubernetes: k3s (VPS), kind (lokal), kubeadm (lab) | k3s cukup ringan untuk satu VPS; kubeadm dipakai untuk latihan multi-node |
| Packaging | Helm | Satu chart untuk semua service LogPilot, values berbeda per environment |
| GitOps | Argo CD | Cluster menarik perubahan dari Git; CI tidak memegang kredensial cluster |
| Ingress & TLS | ingress-nginx + cert-manager (Let's Encrypt) | HTTPS otomatis untuk Console, Ingestor, dan Grafana |
| Operator | Strimzi (Kafka), Altinity (ClickHouse), CloudNativePG | Cara standar menjalankan sistem stateful di Kubernetes |
| Autoscaling | HPA (Ingestor), KEDA (consumer, dari consumer lag) | CPU bukan sinyal yang benar untuk consumer; lag-lah sinyalnya |
| CI/CD | GitHub Actions | Lint, test, build image, scan, lalu commit tag image ke repo GitOps |

---

## 4. Service Specifications

### 4.1 Go Ingestor Service

Service utama yang menjadi entry point semua log ke sistem LogPilot.

#### Endpoints

| Method | Path | Deskripsi |
|---|---|---|
| POST | `/v1/ingest` | Terima satu atau batch log dari aplikasi |
| POST | `/v1/logs` | Terima log format OTLP/HTTP JSON dari OTel Collector |
| GET | `/health` | Liveness probe — proses hidup |
| GET | `/ready` | Readiness probe — Kafka & Redis tersambung, antrean belum penuh |
| GET | `/metrics` | Prometheus metrics endpoint |

> **Kenapa `/health` dan `/ready` dipisah:** kalau liveness probe gagal, Kubernetes **membunuh** pod. Kalau readiness probe gagal, Kubernetes hanya **berhenti mengirim trafik** dan membiarkan pod pulih. Menggabungkan keduanya berarti pod yang sekadar sedang sibuk akan terus di-restart, yang justru memperparah kondisi overload.

#### Request Schema — Log Tunggal

```
POST /v1/ingest
Header: X-API-Key: logpilot_abc123xyz
Header: Content-Encoding: gzip   (opsional)

{
  "level": "ERROR",                       // required: DEBUG|INFO|WARN|ERROR|FATAL
  "message": "DB connection timeout",     // required
  "service": "sabh-api",                  // required
  "timestamp": "2026-07-17T05:12:00Z",    // required, ISO 8601

  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",  // optional, 32 hex
  "span_id": "00f067aa0ba902b7",                   // optional, 16 hex
  "parent_span_id": "a3ce929d0e0e4736",            // optional, 16 hex
  "user_id": "12345",                              // optional
  "session_id": "sess-abc",                        // optional

  "metadata": {                           // optional, bebas
    "host": "pod-xyz",
    "region": "jkt",
    "order_id": "ORD-9981"
  }
}
```

#### Request Schema — Batch

```json
{
  "logs": [ { ... }, { ... }, { ... } ]
}
```

Batas: maksimal **1.000 log per request**, body maksimal **5 MB**.

Rate limit dihitung **per log**, bukan per request — kalau tidak, batas 10.000/menit bisa dilewati hanya dengan membungkus 1.000 log dalam satu request.

#### Response Codes

| Code | Kondisi |
|---|---|
| `202 Accepted` | Semua log valid dan berhasil dipublish |
| `207 Multi-Status` | Sebagian valid. Body memuat `accepted`, `rejected`, dan daftar error per index |
| `400 Bad Request` | Semua log invalid, atau body tidak bisa di-parse |
| `401 Unauthorized` | API key tidak ada atau tidak valid |
| `429 Too Many Requests` | Melewati rate limit. Disertai header `Retry-After` |
| `503 Service Unavailable` | Kafka tidak bisa dihubungi, atau antrean internal > 95% penuh |

Contoh body `207`:
```json
{
  "accepted": 497,
  "rejected": 3,
  "errors": [
    { "index": 12, "field": "level", "reason": "invalid value: TRACE" },
    { "index": 88, "field": "timestamp", "reason": "not ISO 8601" },
    { "index": 341, "field": "trace_id", "reason": "must be 32 hex characters" }
  ]
}
```

#### Processing Pipeline (6 Layers)

| Layer | Proses | Gagal → Response |
|---|---|---|
| 1. Authentication | `GET api_key:{sha256(value)}` dari Redis → `project_id`. Kalau miss, fallback ke internal endpoint Management API (dibungkus `singleflight`), lalu isi ulang cache | 401 Unauthorized |
| 2. Rate Limiting | Lua atomic: `INCRBY rate:{project_id}:{menit} {jumlah_log}`, `EXPIRE` hanya saat key lahir. Batas diambil dari kuota project (default 10.000/menit) | 429 + `Retry-After` = sisa detik window |
| 3. Validation | Field wajib: level, message, service, timestamp. Format ISO 8601. Format trace_id (32 hex) & span_id (16 hex) kalau diisi. **Clock skew guard** | 400 + detail per field |
| 4. Enrichment | Tambah `project_id`, `ingested_at`, `ingestor_id`. Promosikan field yang dikenal dari `metadata` ke kolom sendiri, sisanya ke `attributes` (maks 32 key) | — |
| 5. **Fingerprint & Scrub** | **Baru di v2.0.** PII scrubbing (email, kartu kredit, JWT, NIK). Hitung `fingerprint` untuk level ERROR dan FATAL (§5.2) | — |
| 6. Publish | Kafka produce ke topic `raw-logs`, partition key = `project_id` | 503 Service Unavailable |

**Rate limiter (Layer 2), perbaikan v3.0.** v2.0 menulis `INCRBY` "dengan EXPIRE 60 detik" sebagai dua perintah, dan implementasi yang berjalan memang begitu: `INCR`, lalu `EXPIRE` bila hasilnya 1. Ini pola yang sama dengan yang ditolak di §4.4. Kalau proses mati atau koneksi putus di antara dua perintah itu, key hidup tanpa TTL, counter tidak pernah reset, dan project tersebut terkena 429 selamanya setelah log ke-10.000.

- Pakai skrip Lua yang sama dengan counter alert (`INCRBY`, lalu `EXPIRE` hanya bila key baru), dengan nomor menit di dalam nama key sebagai pengaman kedua.
- Rate limiter ini **fixed window, dan itu disengaja**. Boundary problem membolehkan lonjakan sampai 2x kuota di perbatasan menit. Untuk alert itu menghasilkan keputusan yang salah; untuk rate limit itu hanya berarti batasnya sedikit longgar, dan Kafka menyerap selisihnya. Keputusan ini dicatat di `docs/DESIGN-DECISIONS.md`.
- **Redis mati → fail open** dengan limiter in-memory per pod (kuota dibagi perkiraan jumlah pod). Menolak semua log karena komponen pembatasnya mati adalah kegagalan yang lebih buruk daripada kelebihan menerima selama beberapa menit.
- Kuota per project disimpan di PostgreSQL (`projects.rate_limit_per_minute`) dan ikut di-cache bersama hasil lookup API key, sehingga tidak ada lookup tambahan per request.

**Clock skew guard (Layer 3).** `timestamp` dikirim oleh client dan jam client tidak bisa dipercaya. Aturannya:
- Lebih dari **1 jam di masa depan** → tolak 400. Jam yang maju hampir selalu salah konfigurasi.
- Lebih dari **7 hari di masa lalu** → terima, tapi clamp, dan simpan nilai aslinya di `attributes['original_timestamp']`.
- `timestamp` = waktu kejadian menurut client; dipakai untuk tampilan, partisi, dan retensi. Karena sudah di-clamp, nilainya tidak bisa dipakai untuk mengakali retensi.
- **`ingested_at` = waktu server, sumber kebenaran untuk urutan kedatangan** dan untuk mengukur keterlambatan pipeline (`ingested_at - timestamp`).

**Kafka producer durability.** `acks=all`, `enable.idempotence=true`, `MaxOpenRequests=1`, kompresi LZ4, retry maksimal 5 dengan backoff.

Karena producer bersifat asynchronous, error publish tidak terdeteksi secara langsung per-request. Untuk mencegah Ingestor membalas `202` padahal Kafka sedang down, Ingestor memantau channel `Errors()` dan membuka **circuit breaker** ketika error rate melewati ambang — selama breaker terbuka, response menjadi `503`. Alternatifnya (sync producer) memberi kejujuran per-request tapi tidak realistis untuk target 1.000 req/detik; trade-off ini dicatat di `docs/DESIGN-DECISIONS.md`.

#### Backpressure & Load Shedding

Ingestor memiliki bounded queue antara HTTP handler dan Kafka producer:

| Kondisi antrean | Perilaku |
|---|---|
| < 80% | Normal |
| ≥ 80% | Buang log level DEBUG, tetap balas 202, catat `logs_shed_total` |
| ≥ 90% | Buang juga INFO |
| ≥ 95% | Balas 503 + `Retry-After`, **kecuali** ERROR dan FATAL |

ERROR dan FATAL tidak pernah dibuang — justru log itulah yang paling dibutuhkan saat sistem sedang bermasalah.

#### Enriched Payload (dikirim ke Kafka)

```json
{
  "level": "ERROR",
  "message": "DB connection timeout after 30s",
  "service": "sabh-api",
  "timestamp": "2026-07-17T05:12:00Z",

  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "span_id": "00f067aa0ba902b7",
  "parent_span_id": "a3ce929d0e0e4736",
  "user_id": "12345",
  "session_id": "sess-abc",
  "fingerprint": "9f2c4a7b1e8d3506",

  "attributes": { "region": "jkt", "order_id": "ORD-9981" },

  "project_id": "project-sabh",
  "host": "pod-xyz",
  "ingested_at": "2026-07-17T05:12:00.347Z",
  "ingestor_id": "ingestor-pod-3"
}
```

---

### 4.2 Kafka Configuration

| Topic | Partitions | Retention | Partition Key | Producer |
|---|---|---|---|---|
| `raw-logs` | 6 | 7 hari | `project_id` | Go Ingestor |
| `alert-events` | 1 | 1 hari | — | Consumer 2 (audit trail) |
| `raw-logs-dlq` | 1 | 30 hari | — | Consumer 1 (poison message) |
| `issue-events` | 1 | 7 hari | `project_id` | Consumer 2 (`new_issue`, `reopened`), dikonsumsi Management API (§5.5) |

**Consumer Groups:**
- `storage-workers` — Consumer 1, baca topic `raw-logs`, commit offset setelah batch insert berhasil.
- `alert-evaluators` — Consumer 2, baca topic `raw-logs` secara independen dari `storage-workers`.

**Konfigurasi durability.** Untuk demo satu broker, `replication.factor=1` dapat diterima. Konfigurasi produksi yang dituju:

| Setting | Nilai | Alasan |
|---|---|---|
| `acks` | `all` | Producer menunggu semua replika in-sync menulis |
| `replication.factor` | 3 | Toleransi kehilangan 1 broker |
| `min.insync.replicas` | 2 | Kalau hanya 1 replika sehat, tulis ditolak — memilih konsistensi daripada ketersediaan |
| `enable.idempotence` | `true` | Retry producer tidak menghasilkan duplikat di broker |
| `compression.type` | `lz4` | Rasio kompresi memadai dengan CPU rendah |

**Catatan partition key.** `project_id` sebagai partition key berarti satu project selalu masuk ke partition yang sama, sehingga urutan log per project terjaga. Konsekuensinya: kalau satu project jauh lebih besar dari yang lain, terjadi **hot partition** dan penambahan consumer tidak membantu untuk project tersebut. Mitigasi bila diperlukan: composite key `project_id + service`, atau `project_id + hash(trace_id)` bila urutan per-trace lebih penting daripada urutan per-project.

---

### 4.3 Consumer 1 — Storage Worker

Go service yang consume topic `raw-logs` dan melakukan batch insert ke ClickHouse.

#### Micro-batch Strategy

- Poll Kafka dengan batas: **500 messages ATAU 1 detik** (mana yang lebih dulu terpenuhi).
- **Batch dibentuk per partition Kafka**, bukan lintas partition. Satu batch = satu rentang offset yang berurutan dari satu partition.
- Satu query INSERT per batch ke ClickHouse, disertai `insert_deduplication_token`.
- Commit offset ke Kafka **HANYA** setelah INSERT berhasil.
- Jika INSERT gagal: jangan commit, retry dengan **exponential backoff + jitter** (1s → 2s → 4s → 8s, maksimal 30s, ditambah random 0–500ms).
- Setelah N kali gagal (default 10): pecah batch per-message untuk mengisolasi yang rusak, kirim yang rusak ke `raw-logs-dlq`, insert sisanya, commit offset.
- Saat menerima SIGTERM: flush batch yang masih di memory sebelum keluar.

> **Kenapa backoff pakai jitter:** retry dengan interval tetap membuat semua pod consumer menghantam ClickHouse pada detik yang sama persis, berulang-ulang (thundering herd). Jitter memecah barisan itu.
>
> **Kenapa retry harus dibatasi:** satu message rusak yang selalu menggagalkan INSERT akan membekukan seluruh partition selamanya kalau di-retry tanpa batas. Semua log lain di partition tersebut ikut macet meskipun tidak bermasalah. Ini disebut **poison message**, dan DLQ adalah pola standarnya.

#### Delivery Semantics

LogPilot menggunakan **at-least-once + idempotency**, bukan exactly-once.

| Semantics | Mekanisme | Konsekuensi |
|---|---|---|
| At-most-once | Commit offset sebelum proses | Data bisa hilang saat crash |
| **At-least-once** | Commit offset setelah proses | **Tidak ada yang hilang, tapi bisa duplikat** |
| Exactly-once | Transaksi lintas sistem | Mahal, dan tidak selalu mungkin lintas Kafka–ClickHouse |

Duplikat yang timbul dari at-least-once ditangani dengan dua mekanisme, karena ada dua skenario yang berbeda.

**Skenario 1: retry di dalam proses yang sama.** INSERT sebenarnya berhasil tetapi responsnya hilang (timeout), lalu worker mengulang batch yang persis sama. Ditangani `insert_deduplication_token`, dihitung deterministik dari `topic:partition:offset_awal:offset_akhir`. Batch yang sama menghasilkan token yang sama, sehingga insert ulang diabaikan.

> **Perbaikan v3.0.** Untuk `MergeTree` non-replicated, deduplikasi insert **mati secara bawaan**: `non_replicated_deduplication_window` bernilai 0, dan token diabaikan tanpa pesan apa pun. Tabel harus dibuat dengan `SETTINGS non_replicated_deduplication_window = 1000`. Tanpa ini, seluruh klaim F9 tidak punya dasar.

**Skenario 2: crash lalu restart, atau rebalance.** Worker mati setelah INSERT berhasil tetapi sebelum commit offset. Saat hidup lagi ia membaca ulang dari offset terakhir yang ter-commit, tetapi **batas batch-nya tidak lagi sama**: batch pertama dulu terpotong oleh timer 1 detik di offset 1.230, sekarang terpotong di 1.499 karena pesannya sudah menumpuk. Rentang offset berbeda, token berbeda, dan 231 log tersimpan dua kali. Token tidak bisa menolong di sini.

Penanganannya adalah menjadikan ClickHouse sumber kebenaran posisi baca (**high-water mark**):

- Setiap baris menyimpan `kafka_partition` dan `kafka_offset` asalnya.
- Setiap kali worker mendapat partition (startup atau rebalance), ia menanyakan `max(kafka_offset)` partition itu ke ClickHouse, lalu membuang pesan dengan offset ≤ nilai tersebut sebelum membentuk batch.
- Offset Kafka tetap di-commit seperti biasa, tetapi fungsinya turun menjadi petunjuk awal; yang menentukan adalah apa yang sudah benar-benar tersimpan.

Ini pola "simpan offset bersama datanya", cara standar mendapatkan efek exactly-once tanpa transaksi lintas sistem. Batas yang disadari: INSERT yang barisnya jatuh ke lebih dari satu partisi harian tidak atomik, sehingga crash tepat di tengahnya masih bisa meninggalkan sebagian baris. Besarnya celah ini diukur di chaos test (`R-S7`), bukan diasumsikan nol.

#### ClickHouse Schema

```sql
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
  fingerprint     String,

  host            LowCardinality(String),
  ingestor_id     LowCardinality(String),
  attributes      Map(LowCardinality(String), String),

  timestamp       DateTime64(3),
  ingested_at     DateTime64(3),

  kafka_partition UInt16,
  kafka_offset    UInt64,

  INDEX idx_trace  trace_id    TYPE bloom_filter(0.01)      GRANULARITY 1,
  INDEX idx_fp     fingerprint TYPE bloom_filter(0.01)      GRANULARITY 1,
  INDEX idx_user   user_id     TYPE bloom_filter(0.01)      GRANULARITY 1,
  INDEX idx_msg    message     TYPE tokenbf_v1(32768, 3, 0) GRANULARITY 1
) ENGINE = MergeTree()
PARTITION BY toDate(timestamp)
ORDER BY (project_id, service, level, timestamp)
TTL toDateTime(timestamp) + INTERVAL 30 DAY
SETTINGS non_replicated_deduplication_window = 1000,
         ttl_only_drop_parts = 1;
```

**Design decisions:**

- `MergeTree` — engine default ClickHouse, dioptimasi untuk bulk insert dan range query.
- `PARTITION BY toDate` — diubah dari bulanan di v2.0. Dengan retensi 30 hari dan partisi bulanan, tidak pernah ada partisi yang seluruh isinya kedaluwarsa pada saat yang sama, sehingga ClickHouse harus menulis ulang part untuk membuang baris lama. Partisi harian ditambah `ttl_only_drop_parts = 1` membuat TTL cukup menghapus part utuh, yang nyaris gratis. Batasnya: jumlah partisi aktif jadi sekitar 30, masih jauh di bawah ambang yang bermasalah (ratusan).
- `TTL toDateTime(timestamp)` — ekspresi TTL harus bertipe `Date` atau `DateTime`; `DateTime64` ditolak dengan `BAD_TTL_EXPRESSION` (`PROBLEM.md` Issue #4).
- `fingerprint String`, bukan `LowCardinality` — diubah dari v2.0. Jumlah fingerprint tumbuh seiring jumlah jenis error di semua tenant dan bisa melewati batas wajar `LowCardinality` (sekitar puluhan ribu nilai unik), di mana kamusnya justru memperlambat.
- `kafka_partition`, `kafka_offset` — dasar high-water mark di bagian Delivery Semantics, sekaligus memudahkan menelusuri satu baris kembali ke pesan asalnya.
- `ORDER BY (project_id, service, level, timestamp)` — diubah dari v1.0 yang hanya `(project_id, timestamp)`. `ORDER BY` di ClickHouse adalah primary index-nya, dan urutan kolomnya harus mengikuti pola query yang paling sering: "log project X, service Y, level ERROR, rentang waktu Z". Dengan urutan ini ClickHouse memotong sebagian besar data sebelum menyentuh timestamp.
- `LowCardinality(String)` — untuk kolom yang nilai uniknya sedikit tapi berulang jutaan kali. ClickHouse menyimpannya sebagai dictionary + integer: ukuran disk turun drastis dan perbandingan menjadi perbandingan integer. **Tidak** dipakai untuk `message`, `trace_id`, `user_id` karena nilainya hampir selalu unik — kamusnya akan lebih besar dari datanya.
- `Map(LowCardinality(String), String)` untuk `attributes` — **memperbaiki silent data loss di v1.0.** Sebelumnya payload menerima `metadata` bebas tapi tabel hanya punya kolom `trace_id` dan `host`; field lain hilang tanpa error apa pun. Field yang sering dipakai filter tetap dipromosikan jadi kolom sendiri; sisanya masuk Map.
- `CODEC(ZSTD(3))` di `message` — kolom teks terpanjang dan paling memakan disk.
- **Bloom filter skip index** — `trace_id` adalah nilai acak yang tidak mengikuti `ORDER BY`, sehingga query `WHERE trace_id = 'x'` tidak bisa memanfaatkan primary index dan akan memindai seluruh partisi. Bloom filter per granule memungkinkan ClickHouse melewati granule yang **pasti** tidak memuat nilai tersebut (false positive mungkin, false negative tidak). Ini yang mengubah pencarian satu trace dari hitungan detik menjadi milidetik.
- `TTL 30 DAY` — auto-delete log lama, tidak perlu cron job manual.

Semua DDL disimpan sebagai file bernomor di `migrations/clickhouse/`, bukan diketik manual di CLI.

---

### 4.4 Consumer 2 — Alert Evaluator

Go service yang consume topic `raw-logs` dan mengevaluasi setiap log terhadap alert rules yang aktif.

#### Alert Rule Schema

```json
{
  "id": "rule-001",
  "project_id": "project-sabh",
  "name": "High ERROR Rate",
  "type": "threshold",
  "condition": {
    "level": "ERROR",
    "service": "sabh-api"
  },
  "threshold": 10,
  "window_seconds": 300,
  "bucket_seconds": 10,
  "cooldown_seconds": 600
}
```

`type` baru di v3.0. `threshold` adalah perilaku v2.0 dan termasuk tier Core. Dua jenis lain (`new_issue`, `regression`) tidak memakai counter dan dijelaskan di §5.5.

#### Bucketed Sliding Window Counter (Redis)

> **Perbaikan dari v1.0.** Skema lama (`INCR` + `EXPIRE` setiap match) punya dua masalah:
>
> **1. Bukan sliding window — ini fixed window.** Window-nya kotak kaku: 10:00:00–10:04:59, reset, 10:05:00–10:09:59. Dengan threshold 10 per 5 menit, 9 error pada 10:04:50 dan 9 error pada 10:05:10 — 18 error dalam 20 detik, jelas sebuah insiden — **tidak akan pernah memicu alert**, karena masing-masing window hanya berisi 9. Ini disebut boundary problem.
>
> **2. TTL ter-reset terus.** `EXPIRE` dipanggil setiap kali ada match, sehingga setiap error baru memperpanjang umur key. Kalau error datang pelan tapi terus-menerus (1 error tiap 2 menit selama 3 jam), key tidak pernah kedaluwarsa dan counter menumpuk sampai 90. Alert akan terpicu sambil mengklaim "10 error dalam 5 menit" padahal kenyataannya 10 error tersebar dalam 3 jam. **Alert palsu.**
>
> Ditambah, `INCR` dan `EXPIRE` adalah dua perintah terpisah — kalau proses mati di antaranya, key ada tanpa TTL dan menjadi abadi.

**Pendekatan yang dipilih: bucketed counter.** Window dipecah menjadi bucket kecil, satu counter per bucket. Window 5 menit dengan bucket 10 detik = 30 bucket.

```
# Increment bucket saat ini (Lua, atomic — EXPIRE hanya saat key lahir):
key = alert_counter:{rule_id}:{project_id}:{floor(unix_now / bucket_seconds)}

local v = redis.call('INCR', KEYS[1])
if v == 1 then redis.call('EXPIRE', KEYS[1], ARGV[1]) end
return v

# Hitung total window: MGET 30 key terakhir dalam satu pipeline, jumlahkan.

# Jika total >= threshold:
SET cooldown:{rule_id}:{project_id} 1 EX {cooldown_seconds} NX
# Hanya proses yang berhasil SET (NX) yang boleh mengirim alert.
```

Perbandingan ketiga pendekatan:

| Pendekatan | Akurasi | Memory | Verdict |
|---|---|---|---|
| Fixed window (`INCR`+`EXPIRE`) | Buruk di batas window | O(1) | Ditolak — boundary problem |
| Sorted set (`ZADD`+`ZREMRANGEBYSCORE`) | Sempurna | **O(N)** per rule | Ditolak — 100.000 event = 100.000 entry |
| **Bucketed counter** | ±1 bucket (±10 detik) | **O(1)** — 30 key kecil | **Dipilih** |

Untuk keperluan alerting, ketidakakuratan ±10 detik sepenuhnya dapat diterima. Memory konstan tidak bisa ditawar.

#### Evaluasi Flow

1. Poll batch (100 message) dari Kafka.
2. Load alert rules dari memory cache (refresh tiap 30 detik dari Management API / PostgreSQL).
3. Untuk setiap log, cek terhadap setiap rule dengan `project_id` yang cocok.
4. Kelompokkan match per rule, lalu `INCRBY` sekali per rule per batch — bukan `INCR` per log.
5. **Selalu increment, tanpa syarat.** Hitung total window.
6. Kalau total < threshold → selesai.
7. Kalau total ≥ threshold → cek cooldown lewat `SET ... NX`:
   - Gagal set (cooldown aktif) → jangan kirim, catat metric `alerts_suppressed_total`
   - Berhasil set → `POST /api/v2/alerts` ke Alertmanager, tulis salinan ke `alert-events`
8. Commit offset ke Kafka.

> **Perubahan urutan dari v1.0.** v1.0 mengecek cooldown **sebelum** increment dan melewati increment bila cooldown aktif. Akibatnya counter berhenti menghitung selama cooldown, dan begitu cooldown habis penghitungan mulai dari nol — seluruh konteks error yang terjadi selama periode itu hilang, padahal justru di periode itulah insiden sedang berlangsung.
>
> Aturan yang benar: **cooldown menekan pengiriman notifikasi, bukan penghitungan.**
>
> **Kenapa `SET ... NX` dan bukan `GET` lalu `SET`:** kalau Consumer 2 di-scale menjadi 3 pod, ketiganya bisa menembus threshold pada saat yang hampir bersamaan dan mengirim 3 alert duplikat. `SET ... NX` bersifat atomic — hanya satu pod yang berhasil, dan hanya pod itu yang mengirim. Ini distributed lock sederhana.

#### Payload ke Alertmanager

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
    "sample_trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
    "fingerprint": "9f2c4a7b1e8d3506"
  },
  "startsAt": "2026-07-17T10:00:00Z"
}]
```

`labels` adalah identitas alert — Alertmanager menganggap dua alert dengan set label identik sebagai alert yang sama, dan inilah mekanisme deduplikasinya. `annotations` hanya informasi tampilan dan tidak ikut diperhitungkan.

---

### 4.5 Go Alert Dispatcher

Go service yang menerima webhook dari Alertmanager dan mendistribusikan notifikasi ke destination yang dikonfigurasi user.

| Destination | Method | Detail |
|---|---|---|
| ClickUp | POST `/api/v2/list/{list_id}/task` | Create task dengan priority Urgent, nama `[ALERT] {rule_name} - {project_id}`, body memuat sample message dan exemplar trace_id |
| SendGrid | POST `/v3/mail/send` | Email ke recipients yang dikonfigurasi per project |
| Slack | POST webhook URL | Message ke channel yang dikonfigurasi (opsional) |

Semua destination dikirim secara **concurrent** menggunakan Go goroutine dengan timeout 10 detik per destination via `context`.

**Notification config di-cache di memory dengan TTL 60 detik.** v1.0 menyatakan Dispatcher melakukan query ke Management API pada setiap alert secara sinkron — artinya kalau Management API sedang restart atau lambat, alert gagal terkirim tepat pada momen alert paling dibutuhkan. Komponen di jalur kritis tidak boleh bergantung secara sinkron pada komponen yang bukan jalur kritis.

Bila Management API tidak dapat dihubungi dan cache kosong, Dispatcher tetap mengirim ke destination default dari environment variable. Panggilan ke Management API dibatasi timeout 3 detik dan dilindungi circuit breaker.

---

### 4.6 Management API (Go) — Control Plane

> **v3.1.** Bagian ini sebelumnya ditulis untuk Laravel. Management API sekarang ditulis dengan Go dan berjalan dalam **satu service `api`** bersama Query API (§4.8): satu binary, satu image, dua kelompok route.

**Kenapa bukan Laravel lagi.** Laravel menambah satu bahasa, satu image, dan satu cara deploy hanya untuk CRUD project, API key, dan alert rule. Dengan Go, semua service memakai pola yang sama untuk konfigurasi, logging, metrik, health check, dan graceful shutdown. Sisi PHP proyek ini ditunjukkan oleh handler Monolog (§12), yang menempatkan aplikasi PHP sebagai pengirim log.

**Kenapa satu service dengan Query API, bukan dua.** Keduanya HTTP untuk tenant, memakai sesi dan token yang sama, dan sama-sama membaca PostgreSQL. Memisahkannya berarti satu Deployment, satu pipeline, dan satu jalur token antar-service tambahan tanpa beban yang membenarkannya. Batasnya dijaga di dalam kode: package `management` dan `query` tidak saling mengimpor, dan `query` hanya mempercayai JWT (§4.8), sehingga memisahkannya nanti cukup memindahkan package. Tanda sudah waktunya dipisah: Live Tail atau query berat mulai mengganggu latency endpoint management, atau keduanya butuh jumlah replika yang berbeda.

**Dua listener.** Port 8081 melayani `/api/*` dan `/v1/*` dan keluar lewat Ingress. Port 8082 melayani `/internal/*` dan tidak punya Ingress; hanya Ingestor dan Alert Dispatcher yang boleh menjangkaunya (NetworkPolicy, §6.3). NetworkPolicy bekerja per port, bukan per path, jadi endpoint internal yang berbagi port dengan endpoint publik tidak bisa dilindungi olehnya.

Management API adalah backend dari management UI. Tidak berada di data path log — hanya mengelola konfigurasi yang digunakan service lain, ditambah satu peran baru: **sumber kebenaran saat cache Redis miss**.

#### Endpoints

| Method | Path | Fungsi | Side Effect |
|---|---|---|---|
| POST | `/api/auth/register` | Register user baru | — |
| POST | `/api/auth/login` | Login, return session token (opaque, bisa dicabut) | — |
| POST | `/api/projects` | Buat project baru | Insert ke PostgreSQL |
| GET | `/api/projects` | List semua project milik user | — |
| POST | `/api/projects/{id}/api-keys` | Generate API key baru | Insert hash ke PostgreSQL + seed **hash** ke Redis |
| DELETE | `/api/projects/{id}/api-keys/{keyId}` | Revoke API key | Hapus dari PostgreSQL + hapus dari Redis |
| GET | `/api/projects/{id}/alert-rules` | List alert rules project | — |
| POST | `/api/projects/{id}/alert-rules` | Buat alert rule baru | Insert ke PostgreSQL, Consumer 2 baca ini tiap 30 detik |
| PUT | `/api/projects/{id}/alert-rules/{ruleId}` | Update alert rule | Update PostgreSQL |
| DELETE | `/api/projects/{id}/alert-rules/{ruleId}` | Hapus alert rule | Delete dari PostgreSQL |
| PUT | `/api/projects/{id}/notifications` | Set config notifikasi | Update PostgreSQL |
| GET | `/internal/projects/{id}/notifications` | Internal — untuk Alert Dispatcher | — |
| GET | `/internal/api-keys/{hash}` | **Baru.** Internal — fallback auth Ingestor saat Redis miss | — |
| POST | `/api/projects/{id}/query-token` | **v3.0.** Terbitkan JWT 15 menit untuk Query API (§4.8) | — |
| PUT | `/api/projects/{id}/issues/{fingerprint}` | **v3.0.** Ubah status issue: `open`, `resolved`, `ignored` | Update PostgreSQL + set `issue_resolved` di Redis |

Endpoint `/internal/*` dilayani listener internal (port 8082) dan tidak memakai sesi user; autentikasinya lewat internal secret header.

#### Sesi User

- Password di-hash dengan **bcrypt** (`golang.org/x/crypto/bcrypt`). Password dipilih manusia dan cenderung lemah, jadi fungsi hash yang sengaja lambat memang yang dibutuhkan di sini (bandingkan dengan API key di bawah).
- Session token: 32 byte dari `crypto/rand`, dikirim ke client sekali, dan disimpan di PostgreSQL sebagai SHA-256 (tabel `sessions`) beserta masa berlakunya. Logout dan ganti password menghapus barisnya, jadi sesi bisa dicabut seketika.
- Endpoint management memeriksa sesi ke PostgreSQL di setiap request. Trafiknya kecil, dan kemampuan mencabut lebih penting di sini daripada kecepatan. Jalur baca (§4.8) memakai JWT berumur pendek karena pertimbangannya terbalik.

#### Penyimpanan API Key

> **Perbaikan dari v1.0.** v1.0 menyimpan hash di PostgreSQL tapi **plaintext di Redis**. Karena Redis paling sering di-expose tanpa password di lingkungan pengembangan, satu kebocoran Redis berarti seluruh API key semua tenant bocor sekaligus — dan hash di PostgreSQL menjadi tidak ada gunanya.

- Plain key: `logpilot_` + 32 karakter acak dari **CSPRNG** (`crypto/rand`, bukan `math/rand`).
- Disimpan di PostgreSQL sebagai **SHA-256**.
- Disimpan di Redis sebagai **SHA-256** juga: `SET api_key:{sha256(plain)} {project_id}`.
- Ingestor melakukan hash terhadap key dari header, lalu lookup dengan hash tersebut.
- Plain key hanya ditampilkan sekali kepada user saat pembuatan.

**Kenapa SHA-256 dan bukan bcrypt:** bcrypt sengaja lambat (~100ms) untuk melawan brute-force terhadap password yang dipilih manusia dan cenderung lemah. API key ini 32 karakter acak dari CSPRNG — ruang tebakannya sudah terlalu besar untuk di-brute-force, sehingga perlambatan itu tidak memberi manfaat keamanan tambahan. Sementara fungsi ini berjalan pada **setiap request**; bcrypt akan meniadakan target 1.000 req/detik. SHA-256 (~1 mikrodetik) adalah pilihan yang tepat pada konteks ini.

#### Auth Fallback — Redis bukan Single Point of Failure

v1.0 membuat autentikasi Ingestor bergantung 100% pada Redis. Redis restart berarti seluruh platform menolak semua log dari semua tenant — dan Redis di-restart setiap kali container di-recreate.

Pola yang dipakai: **cache-aside dengan PostgreSQL sebagai sumber kebenaran.**

1. Lookup Redis. Hit → lanjut.
2. Miss → panggil `GET /internal/api-keys/{hash}`, dibungkus `singleflight` agar 1.000 request bersamaan hanya menghasilkan satu query.
3. Ketemu → tulis balik ke Redis (TTL 1 jam), lanjut.
4. Tidak ketemu → cache negatif `api_key_invalid:{hash}` TTL 30 detik, return 401.
5. Redis sendiri yang down (bukan sekadar miss) → gunakan in-memory cache di dalam pod sebagai lapisan terakhir (TTL 5 menit), naikkan alarm.

Dua jebakan yang ditangani: **thundering herd** (dicegah `singleflight`) dan **serangan key acak** yang setiap kali memicu query PostgreSQL (dicegah negative caching).

#### PostgreSQL Schema

| Tabel | Kolom Utama | Relasi |
|---|---|---|
| `users` | id, name, email, password, created_at | — |
| `sessions` | id, user_id, token_hash, expires_at, created_at | belongs to users (v3.1) |
| `projects` | id, user_id, name, slug, pii_scrubbing_enabled, rate_limit_per_minute, created_at | belongs to users |
| `api_keys` | id, project_id, name, key_hash, last_used_at, created_at | belongs to projects |
| `alert_rules` | id, project_id, name, type, condition (JSON), threshold, window_seconds, bucket_seconds, cooldown_seconds, is_active | belongs to projects |
| `notification_configs` | id, project_id, clickup_list_id, clickup_assignee_id, email_recipients (JSON), slack_webhook_url, created_at | belongs to projects |
| `alert_logs` | id, rule_id, project_id, triggered_at, fingerprint, exemplar_trace_id, dispatch_results (JSON) | belongs to alert_rules |
| `issues` | id, project_id, fingerprint, status, first_seen, resolved_at, resolved_by — unik per (project_id, fingerprint) | belongs to projects (v3.0, §5.5) |

`alert_logs` ditulis oleh **Alert Dispatcher** setelah seluruh destination selesai diproses, memuat hasil per destination (berhasil/gagal beserta alasannya). v1.0 mencantumkan tabel ini tanpa menetapkan siapa yang menulisnya.

**Migrasi PostgreSQL** berupa file SQL bernomor di `migrations/postgres/`, dijalankan dengan satu alat migrasi (goose atau golang-migrate; pilih satu dan catat alasannya di `DESIGN-DECISIONS.md`). Laravel dulu menyediakan ini lewat `artisan migrate`; sekarang alatnya harus dipilih sendiri.

---

### 4.7 LogPilot SDK (Go)

Komponen baru di v2.0. v1.0 menyebut SDK di demo script tapi tidak pernah menspesifikasikannya — padahal inilah tempat `trace_id` dilahirkan, sehingga tanpa SDK seluruh kapabilitas tracing tidak mungkin ada.

#### Public API

```go
log := logpilot.New(logpilot.Config{
    Endpoint: "http://localhost:8080",
    APIKey:   os.Getenv("LOGPILOT_KEY"),
    Service:  "sabh-api",
})
defer log.Close()

app.Use(logpilot.Middleware(log))   // membuat / melanjutkan trace context

log.Error(ctx, "payment gateway timeout", logpilot.Fields{"order_id": "ORD-9981"})
```

`trace_id`, `span_id`, `timestamp`, dan `service` diisi otomatis dari `context.Context`. User tidak perlu mengetahuinya.

#### Tanggung Jawab

| Modul | Fungsi |
|---|---|
| `trace.go` | Generate trace_id (16 byte → 32 hex) & span_id (8 byte → 16 hex), parse/format header `traceparent` |
| `context.go` | Simpan dan ambil trace context dari `context.Context` |
| `middleware.go` | Baca `traceparent`; ada → lanjutkan trace, buat span baru dengan parent dari header. Tidak ada → request ini root, buat trace_id baru |
| `client.go` | HTTP client wrapper yang otomatis menyisipkan header `traceparent` ke panggilan keluar |
| `sender.go` | Buffered channel (kapasitas 10.000) + goroutine pengirim, flush tiap 1 detik atau 100 log |

#### Prinsip Desain — Non-Blocking

Pengiriman log **tidak boleh pernah** memblokir aplikasi user. Bila buffer penuh, log **dibuang** dan counter `logpilot_sdk_dropped_total` dinaikkan.

Ini keputusan sadar: tool observability yang bisa menjatuhkan aplikasi yang diamatinya adalah tool yang gagal pada tugas utamanya. Lebih baik kehilangan log daripada kehilangan request user.

`Close()` melakukan flush terhadap sisa buffer saat aplikasi shutdown, dan retry ke Ingestor menggunakan backoff.

---

### 4.8 Go Query API (Tier Showcase)

Komponen baru di v3.0. Satu-satunya jalur baca untuk tenant, dan backend dari Console. Sejak v3.1 berjalan sebagai modul `query` di service `api` yang sama dengan Management API (§4.6).

**Kenapa jalur baca dijaga sebagai modul sendiri, bukan query langsung dari browser atau dari handler management:**

- **Isolasi tenant harus dipaksa di satu tempat.** `project_id` diambil dari token dan disisipkan ke setiap query oleh server. Client tidak pernah bisa mengirim `project_id` sendiri.
- **Query ke ClickHouse harus dibatasi biayanya.** Satu query tanpa rentang waktu bisa memindai seluruh tabel dan mengganggu jalur tulis.
- **Live Tail butuh koneksi yang hidup lama** (satu goroutine per penonton), yang cocok untuk Go.

#### Endpoints

| Method | Path | Fungsi |
|---|---|---|
| GET | `/v1/logs` | Cari log: filter `q`, `from`, `to`, `limit`, `cursor` |
| GET | `/v1/logs/tail` | Live Tail lewat Server-Sent Events, dengan filter `q` yang sama |
| GET | `/v1/logs/histogram` | Jumlah log per interval dan per level, untuk grafik batang di Explorer |
| GET | `/v1/issues` | Daftar issue (agregasi dari `issues_hourly`), digabung dengan status dari PostgreSQL |
| GET | `/v1/issues/{fingerprint}` | Detail satu issue: tren per jam, user terdampak, daftar exemplar trace |
| GET | `/v1/traces/{trace_id}` | Semua log dan span satu trace, sudah tersusun sebagai pohon |
| GET | `/health`, `/ready`, `/metrics` | Sama seperti service lain |

#### Autentikasi

Console login ke Management API dan mendapat session token (§4.6), lalu meminta `POST /api/projects/{id}/query-token`. Management API memeriksa kepemilikan project dan menerbitkan **JWT berumur 15 menit** berisi `project_id`, ditandatangani dengan kunci privat (EdDSA). Modul `query` memverifikasi tanda tangan dan masa berlaku dengan kunci publik, dan tidak pernah membaca tabel sesi di jalur request.

Karena kedua modul sekarang satu proses, pemisahan kunci privat dan publik belum memberi keamanan tambahan. Bentuk ini dipertahankan karena dua alasan: jalur baca tidak bergantung pada PostgreSQL per request (Live Tail dan Explorer tetap jalan saat PostgreSQL lambat), dan modul `query` bisa dipisah menjadi service sendiri tanpa mengganti format token.

Konsekuensi yang disadari: token tidak bisa dicabut sebelum kedaluwarsa. Umur 15 menit adalah batas kerugiannya. Alternatifnya (memeriksa sesi ke PostgreSQL di setiap query) membuat jalur baca ikut mati saat PostgreSQL mati.

#### Pengaman Query

| Pengaman | Nilai | Mencegah |
|---|---|---|
| `project_id` dari token, selalu di klausa `WHERE` pertama | wajib | Membaca data tenant lain |
| Query dibangun dengan parameter terikat, tidak pernah dengan menyambung string | wajib | SQL injection lewat `q` |
| Rentang waktu wajib, maksimal 7 hari per query | 7 hari | Memindai seluruh tabel |
| `limit` maksimal | 1.000 baris | Respons raksasa |
| User ClickHouse khusus, `readonly = 1` | wajib | Query yang mengubah data |
| `max_execution_time`, `max_rows_to_read`, `max_memory_usage` di profil user itu | 10 detik, 500 juta baris, 2 GB | Satu query menghabiskan server |
| Batas query bersamaan per project | 5 | Satu tenant memonopoli ClickHouse |

Pengaman terakhir dipasang di ClickHouse, bukan hanya di kode Go, supaya bug di Query API tidak bisa melewatinya.

#### Bahasa Query

Kotak pencarian menerima sintaks sederhana yang diterjemahkan menjadi SQL berparameter:

```
level:error service:sabh-api "connection timeout" user_id:12345 -region:sgp
```

| Bentuk | Arti | Menjadi |
|---|---|---|
| `kunci:nilai` | Kolom yang dikenal, atau `attributes['kunci']` | `level = ?` / `attributes[?] = ?` |
| `"teks"` atau kata lepas | Pencarian di `message` | `hasToken(message, ?)` atau `message ILIKE ?` |
| `-kunci:nilai` | Negasi | `!= ?` |
| Antar-term | Selalu `AND` | — |

Tidak ada `OR`, tanda kurung, atau regex di versi pertama. Parser ditulis tangan (tokenizer lalu daftar term), dan test case-nya ditulis sebelum parsernya, sama seperti fingerprinting.

#### Pagination

Keyset, bukan `OFFSET`: cursor memuat `(timestamp, kafka_partition, kafka_offset)` dari baris terakhir. `OFFSET 100000` memaksa ClickHouse membaca lalu membuang 100.000 baris di setiap halaman; keyset menambahkan `timestamp < ?` sehingga partisi harian dan granule yang lebih baru langsung dilewati. Pasangan partition dan offset dipakai sebagai pemecah seri karena unik per baris.

#### Live Tail

- Query API membaca `raw-logs` **tanpa consumer group**: hanya partition yang memuat `project_id` tersebut (dihitung dengan fungsi hash yang sama dengan producer), mulai dari offset terbaru. Tanpa group berarti tidak ada commit offset dan tidak ada rebalance; penonton hanyalah pengamat.
- Satu pembaca per partition dibagi ke semua penonton partition itu (fan-out di memori), supaya 50 penonton tidak menjadi 50 koneksi Kafka.
- Setiap penonton punya buffer terbatas (1.000 log). **Penonton lambat kehilangan log, tidak menahan yang lain**; event `dropped` dikirim supaya UI bisa menampilkan "n log dilewati".
- Batas: 5 koneksi tail per project, heartbeat tiap 15 detik, koneksi ditutup setelah 30 menit.
- Kenapa dari Kafka dan bukan polling ClickHouse: latensinya di bawah satu detik, dan Live Tail tetap berjalan saat ClickHouse mati. Itu sendiri demo yang bagus untuk manfaat message bus.

---

### 4.9 LogPilot Console (Tier Showcase)

Next.js + TypeScript. v2.0 membatasi frontend pada form konfigurasi; v3.0 menjadikannya wajah produk.

| Halaman | Isi | Sumber data |
|---|---|---|
| **Onboarding** | Buat project → API key tampil sekali → cuplikan `curl` dan SDK siap salin → indikator "menunggu log pertama…" yang berubah hijau saat log masuk | Management API, lalu Live Tail |
| **Live Tail** | Log mengalir real-time, warna per level, jeda/lanjut, filter | `/v1/logs/tail` |
| **Explorer** | Kotak query, pemilih rentang waktu, histogram per level, tabel log, panel detail dengan `attributes` | `/v1/logs`, `/v1/logs/histogram` |
| **Issues** | Daftar issue: pesan ternormalisasi, sparkline 24 jam, jumlah event, user terdampak, first/last seen, status | `/v1/issues` |
| **Issue detail** | Tren, contoh pesan asli, daftar exemplar trace, tombol Resolve/Ignore | `/v1/issues/{fp}`, Management API |
| **Trace** | Waterfall span dengan durasi, span gagal ditandai merah, log per span | `/v1/traces/{id}` |
| **Alerts** | Alert rules, riwayat alert dan hasil pengirimannya per destination | Management API |
| **Settings** | API keys, notification config, kuota | Management API |

**Aturan desain:**

- **Setiap tampilan punya URL yang bisa dibagikan.** Query, rentang waktu, dan baris terpilih tersimpan di URL. Menempelkan link ke rekan kerja adalah cara orang sungguhan memakai tool seperti ini.
- **Dari mana pun bisa menyeberang:** `trace_id` di baris log membuka halaman Trace; `fingerprint` membuka Issue; tombol "lihat log sekitar waktu ini" di Issue membuka Explorer dengan filter terisi.
- **Tidak ada grafik kesehatan sistem di Console.** Halaman status cukup menautkan ke Grafana.
- **Daftar panjang harus divirtualisasi.** Live Tail yang merender 10.000 baris DOM akan membekukan tab; hanya baris yang terlihat yang dirender.

**Mode demo.** Project `demo` bisa dibuka tanpa login dalam mode baca-saja. Di dalamnya ada tombol **"Simulasikan insiden"** yang menjalankan log generator mode `--chaos` selama 60 detik. Tombol ini dijaga cooldown global 5 menit (`SET ... NX` yang sama dengan cooldown alert), supaya pengunjung tidak bisa membanjiri server demo.

---

## 5. Trace Context & Error Grouping

*Bagian baru di v2.0. Ini kapabilitas yang membedakan LogPilot dari log storage biasa.*

### 5.1 Masalah yang Diselesaikan

Pertanyaannya sederhana: **100.000 user melakukan request pada waktu yang sama, ke service yang sama, dan semuanya mendapat error yang berbeda-beda. Bagaimana membedakan satu sama lain?**

Jawabannya berlapis, dan setiap lapisan menjawab pertanyaan yang berbeda.

| Lapisan | Menjawab | Mekanisme |
|---|---|---|
| 1. Correlation | "Baris log mana saja yang milik satu request?" | `trace_id` |
| 2. Causality | "Di titik mana request itu gagal?" | `span_id` + `parent_span_id` |
| 3. **Aggregation** | **"Sebenarnya ada berapa jenis masalah?"** | **`fingerprint`** |
| 4. Drill-down | "Seperti apa satu kasus konkretnya?" | exemplar `trace_id` per fingerprint |
| 5. Impact | "Berapa banyak user yang terdampak?" | `user_id` |

**Lapisan 3 adalah yang paling sering terlewat, dan justru yang paling menentukan.**

`trace_id` memang menyelesaikan masalah *membedakan*. Tapi setelah dibedakan, hasilnya adalah **100.000 trace_id berbeda** — dan tidak ada manusia yang bisa membuka 100.000 trace satu per satu. Sistemnya berhasil membedakan, tapi operator tetap tidak tahu apa yang sedang terjadi.

Analoginya: `trace_id` memberi 100.000 jarum yang masing-masing bernomor seri. Yang sebenarnya dibutuhkan adalah mengetahui bahwa dari 100.000 jarum itu, sebenarnya hanya ada **12 jenis jarum**.

### 5.2 Error Fingerprinting

Untuk setiap log ERROR dan FATAL, Ingestor menghitung sebuah hash dari **ciri struktural** error tersebut — bukan dari teks persisnya. Error yang berbeda isinya tapi sama penyebabnya menghasilkan hash yang sama.

Kuncinya ada di **normalisasi**:

```
"user 12345 not found in database"
"user 67890 not found in database"    →   "user <num> not found in database"
"user 11111 not found in database"
```

Tiga string berbeda menjadi satu fingerprint, satu issue dengan `count = 3`.

**Rumus:**
```
fingerprint = sha256(service + ":" + level + ":" + normalize(message) + ":" + top_stack_frame)[0:16]
```

**Aturan normalisasi:**

| Pola | Diganti |
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

**Yang tidak boleh dinormalisasi: nama service dan nama host tujuan.** `"timeout connecting to payment-svc"` dan `"timeout connecting to auth-svc"` harus menghasilkan fingerprint yang **berbeda** — sistem yang bermasalah tidak sama. Sebaliknya, `"disk full on /dev/sda1"` dan `"disk full on /dev/sdb2"` harus **sama**, karena penyebabnya identik.

Keseimbangan antara **over-grouping** (dua masalah berbeda tergabung, sehingga satu masalah tersembunyi) dan **under-grouping** (satu masalah terpecah jadi ribuan, sehingga tidak ada gunanya) adalah inti dari desain ini. Test case normalisasi adalah spesifikasi sesungguhnya dari fitur ini, dan ditulis sebelum implementasinya.

### 5.3 Issue Aggregation (Materialized View)

Agregasi per fingerprint dilakukan lewat Materialized View, yang di ClickHouse berperilaku sebagai **trigger saat INSERT** — bukan snapshot yang perlu di-refresh manual seperti di PostgreSQL. Hasilnya selalu terkini tanpa cron job.

> **Perbaikan v3.0.** DDL v2.0 menulis `count() AS event_count`, `min()`, `max()`, dan `argMin()` sebagai kolom biasa di dalam `AggregatingMergeTree`. Engine ini hanya menggabungkan kolom bertipe `AggregateFunction` atau `SimpleAggregateFunction`. Untuk kolom biasa, saat dua baris dengan kunci yang sama di-merge, ClickHouse **mengambil salah satu nilainya dan membuang yang lain**. Akibatnya: setiap batch insert menghasilkan satu baris per fingerprint per jam dengan hitungannya sendiri, lalu merge di latar belakang diam-diam membuang sebagian besar hitungan itu. Angka di panel Top Issues akan turun sendiri seiring waktu, dan berbeda setiap kali dilihat. Dari seluruh kolom di DDL lama, hanya `uniqState` yang benar.
>
> Ini kelas kesalahan yang sama dengan tiga temuan v2.0: tidak ada error, hasilnya hanya salah.

Tabel target dibuat eksplisit supaya tipe setiap kolom terlihat, lalu materialized view menulis ke sana:

```sql
CREATE TABLE logpilot.issues_hourly (
  project_id        LowCardinality(String),
  fingerprint       String,
  hour              DateTime,
  service           LowCardinality(String),
  level             LowCardinality(String),

  event_count       SimpleAggregateFunction(sum, UInt64),
  first_seen        SimpleAggregateFunction(min, DateTime64(3)),
  last_seen         SimpleAggregateFunction(max, DateTime64(3)),
  sample_message    AggregateFunction(argMin, String, DateTime64(3)),
  exemplar_trace_id AggregateFunction(argMin, String, DateTime64(3)),
  unique_users      AggregateFunction(uniq, String)
) ENGINE = AggregatingMergeTree()
PARTITION BY toYYYYMM(hour)
ORDER BY (project_id, fingerprint, hour, service, level)
TTL hour + INTERVAL 90 DAY;

CREATE MATERIALIZED VIEW logpilot.issues_mv TO logpilot.issues_hourly AS
SELECT
  project_id,
  fingerprint,
  toStartOfHour(timestamp)         AS hour,
  service,
  level,
  count()                          AS event_count,
  min(timestamp)                   AS first_seen,
  max(timestamp)                   AS last_seen,
  argMinState(message, timestamp)  AS sample_message,
  argMinState(trace_id, timestamp) AS exemplar_trace_id,
  uniqState(user_id)               AS unique_users
FROM logpilot.logs
WHERE level IN ('ERROR','FATAL')
GROUP BY project_id, fingerprint, hour, service, level;
```

Membacanya juga harus menggabungkan ulang, karena merge di latar belakang tidak dijamin sudah selesai:

```sql
SELECT
  fingerprint,
  sum(event_count)                AS events,
  uniqMerge(unique_users)         AS users,
  min(first_seen)                 AS first_seen,
  max(last_seen)                  AS last_seen,
  argMinMerge(sample_message)     AS sample_message,
  argMinMerge(exemplar_trace_id)  AS exemplar_trace_id
FROM logpilot.issues_hourly
WHERE project_id = {project_id} AND hour >= now() - INTERVAL 24 HOUR
GROUP BY fingerprint
ORDER BY events DESC
LIMIT 50;
```

Tiga hal yang perlu dipahami dari DDL ini:

- `SimpleAggregateFunction(sum, UInt64)` menyimpan angka biasa dan menjumlahkannya saat merge. Cukup untuk `sum`, `min`, `max`, yang hasil gabungannya bisa dihitung dari dua hasil sebelumnya.
- `AggregateFunction(argMin, ...)` dan `AggregateFunction(uniq, ...)` menyimpan **state**, bukan hasil. `argMin` perlu mengingat pasangan nilai dan timestamp; `uniq` perlu menyimpan sketsanya. Itu sebabnya ditulis dengan `-State` dan dibaca dengan `-Merge`.
- `service` dan `level` masuk ke kunci, bukan `any()`. Keduanya sudah bagian dari rumus fingerprint, jadi satu fingerprint selalu punya satu nilai untuk masing-masing.

`argMinState(x, timestamp)` mengambil nilai `x` dari baris dengan timestamp terkecil — inilah cara memperoleh exemplar. `uniqState` menyimpan sketsa sehingga jumlah user unik dapat digabungkan lintas jam tanpa menyimpan seluruh `user_id`. Catatan: `user_id` kosong ikut terhitung sebagai satu nilai; Query API mengurangi satu bila ada baris tanpa user.

Materialized view dipicu per blok insert. Blok yang ditolak deduplikasi tidak memicu view, jadi mekanisme di §4.3 juga menjaga hitungan issue tetap benar.

Trade-off yang disadari: MV menambah beban pada jalur tulis untuk menghemat besar pada jalur baca. Untuk sistem log — ditulis sekali, dibaca ribuan kali — pertukaran ini menguntungkan.

### 5.4 Hasil Akhir

Yang dilihat operator saat insiden bukan 100.000 baris, melainkan:

```
ISSUE                                    COUNT    USERS   FIRST SEEN
payment gateway timeout                  62.341   58.900  2 menit lalu
connection pool exhausted                31.204   30.100  2 menit lalu
user <num> not found                      4.102    4.102  8 jam lalu
invalid coupon code <str>                 2.353    2.201  3 hari lalu
```

Kesimpulannya langsung terbaca: ini bukan 100.000 masalah, ini **2 masalah** — dua teratas yang muncul bersamaan 2 menit lalu, kemungkinan besar saling berkaitan. Dua di bawahnya sudah lama ada dan bukan penyebab insiden ini.

Dari sini operator mengambil **satu exemplar trace_id** untuk menelusuri detailnya. Fingerprint untuk gambaran besar; trace_id untuk menyelam ke satu kasus.

### 5.5 Issue Lifecycle & Alert Berbasis Issue (Tier Showcase)

Agregasi di §5.3 menjawab "ada berapa jenis masalah". Yang belum terjawab: **mana yang baru, mana yang sudah ditangani, dan mana yang kambuh.** Tanpa itu, issue berumur tiga hari yang tidak penting tampil setara dengan issue yang lahir dua menit lalu.

**Status issue** disimpan di PostgreSQL (tabel `issues`: `project_id`, `fingerprint`, `status`, `resolved_at`, `first_seen`), bukan di ClickHouse. Status adalah data yang diubah manusia, jarang, dan butuh konsistensi; ClickHouse tidak dirancang untuk `UPDATE` per baris.

| Status | Arti | Tampil di daftar default |
|---|---|---|
| `open` | Sedang terjadi, belum ditangani | Ya |
| `resolved` | Dinyatakan selesai oleh user | Tidak |
| `ignored` | Diketahui dan sengaja diabaikan | Tidak, dan tidak pernah memicu alert |

**Dua jenis alert baru**, dievaluasi Consumer 2 pada setiap log ERROR/FATAL yang membawa fingerprint:

| Jenis | Terpicu ketika | Mekanisme |
|---|---|---|
| `new_issue` | Fingerprint ini belum pernah terlihat di project tersebut | `SET issue_seen:{project_id}:{fingerprint} 1 NX EX 30 hari`. Berhasil set berarti baru |
| `regression` | Fingerprint berstatus `resolved` muncul lagi | `SISMEMBER issue_resolved:{project_id} {fingerprint}`. Bila anggota: hapus dari set, kirim alert, tulis event `reopened` |

Dua-duanya satu perintah Redis per log error, tanpa counter dan tanpa window. `SET ... NX` juga menjamin hanya satu pod yang mengirim, sama seperti cooldown.

Perubahan status mengalir dua arah:
- User menekan Resolve di Console → Management API memperbarui PostgreSQL dan menambah fingerprint ke set `issue_resolved` di Redis.
- Consumer 2 mendeteksi regression → menulis event ke topic `issue-events` → Management API mengonsumsinya dan mengubah status kembali ke `open`. Consumer 2 tidak menulis langsung ke PostgreSQL; hanya control plane yang boleh.

Kenapa dua alert ini lebih berharga daripada threshold: alert threshold berbunyi saat sesuatu *banyak*. Alert `new_issue` berbunyi saat sesuatu *berbeda*, sering kali sesaat setelah deploy dan jauh sebelum jumlahnya menembus ambang.

---

## 6. Deployment & Platform

Bagian ini ditulis ulang di v3.0. v2.0 memperlakukan Kubernetes sebagai pajangan ("local showcase", manifest untuk portofolio). v3.0 memperlakukannya sebagai tempat live demo benar-benar berjalan, karena yang ingin dipelajari adalah **mengoperasikan**, bukan menulis YAML.

### 6.1 Tiga Environment

| Environment | Tech | Tujuan | Di mana |
|---|---|---|---|
| **dev** | Docker Compose | Menulis kode: satu perintah, seluruh stack hidup | Mac |
| **demo** | k3s satu node + Helm + Argo CD | Live demo publik ber-HTTPS, di-deploy dari Git | VPS |
| **lab** | kubeadm multi-node (Multipass atau beberapa VPS) | Latihan yang tidak mungkin di satu node: node mati, drain, HA Kafka | Mac / VPS sewaan sementara |

Environment `lab` memakai ulang `devops-kubernetes` Lab 2 (kubeadm, Calico, MetalLB, ingress-nginx, registry, MinIO); LogPilot menjadi beban kerja yang di-deploy ke atasnya, menggantikan aplikasi contoh.

### 6.2 Docker Compose Stack (dev)

Services yang di-define dalam `docker-compose.yml`:

**Custom services:**
- `go-ingestor` — port 8080
- `consumer-storage`
- `consumer-alert`
- `alert-dispatcher` — port 9090
- `api` — port 8081 (publik: Management API + Query API) dan 8082 (internal) — **v3.1, menggantikan `laravel-api` dan `query-api`**
- `nextjs-frontend` — port 3000

**Infrastructure:**
- `kafka` — apache/kafka:3.7.0, **KRaft mode (tanpa Zookeeper)**
- `clickhouse` — clickhouse/clickhouse-server
- `redis` — redis:7-alpine
- `postgres` — postgres:15-alpine
- `grafana` — grafana/grafana
- `alertmanager` — prom/alertmanager
- `prometheus` — prom/prometheus — **baru di v2.0**
- `kafka-exporter` — danielqsj/kafka-exporter — **baru di v2.0**, untuk consumer lag
- `otel-collector` — konfigurasi exporter diarahkan ke Ingestor, bukan ke Kafka

> v1.0 mencantumkan `zookeeper`; Kafka 3.7 dalam mode KRaft tidak memerlukannya, dan implementasi yang sudah berjalan memang memakai KRaft.

### 6.3 Kubernetes: Workload LogPilot (Tier Platform)

Semua service LogPilot dikemas dalam **satu Helm chart** (`deploy/helm/logpilot/`) dengan values per environment (`values-demo.yaml`, `values-lab.yaml`). Menulis chart sendiri, bukan manifest lepas, karena perbedaan antar-environment (jumlah replika, resource, hostname) persis masalah yang diselesaikan templating.

Setiap service wajib punya:

| Objek | Isi | Kenapa |
|---|---|---|
| `Deployment` | Liveness ke `/health`, readiness ke `/ready`, `resources.requests` dan `limits`, `securityContext` non-root dengan root filesystem read-only | Tanpa `requests`, scheduler tidak tahu kapasitas node; tanpa `limits`, satu pod bisa membuat node kehabisan memori |
| `Service` | ClusterIP | Hanya Ingestor, Console, port publik service `api`, dan Grafana yang keluar lewat Ingress |
| `ConfigMap` + `Secret` | Konfigurasi dan kredensial terpisah | Secret tidak pernah masuk Git dalam bentuk terbaca (lihat §6.6) |
| `PodDisruptionBudget` | `minAvailable: 1` untuk Ingestor | `kubectl drain` tidak boleh mematikan semua pod Ingestor sekaligus |
| `NetworkPolicy` | Port internal service `api` (8082, `/internal/*`) hanya bisa dijangkau Ingestor dan Dispatcher; ClickHouse hanya dari Consumer 1, Query API, dan Grafana | "Internal" harus dipaksa jaringan, bukan hanya disepakati |

**Graceful shutdown** berlaku untuk semua service: tangkap SIGTERM, tandai `/ready` gagal, tunggu 5 detik agar trafik berhenti mengalir, flush buffer, lalu keluar. `terminationGracePeriodSeconds` harus lebih panjang dari waktu flush terburuk Consumer 1.

**Autoscaling:**

| Service | Alat | Sinyal | Alasan |
|---|---|---|---|
| Ingestor | HPA | CPU, 2–6 pod | Beban Ingestor sebanding dengan CPU (parsing, hashing, kompresi) |
| Consumer 1 & 2 | **KEDA** | Consumer lag Kafka per group | Consumer yang tertinggal bisa saja ber-CPU rendah karena sedang menunggu ClickHouse. CPU bukan sinyalnya; lag-lah sinyalnya |

Batas atas replika consumer = jumlah partition (6). Pod ke-7 tidak mendapat partition dan hanya menganggur; ini batas yang ditentukan Kafka, bukan Kubernetes.

### 6.4 Kubernetes: Sistem Stateful

Bagian tersulit menjalankan LogPilot di Kubernetes bukan service Go-nya, melainkan Kafka, ClickHouse, dan PostgreSQL. Ketiganya dijalankan lewat **operator**:

| Sistem | Operator | Yang dipelajari |
|---|---|---|
| Kafka | Strimzi | `StatefulSet`, identitas pod yang stabil, `PersistentVolumeClaim` per broker, rolling update yang menunggu replika sinkron |
| ClickHouse | Altinity ClickHouse Operator | Konfigurasi sebagai resource, perubahan skema vs perubahan cluster |
| PostgreSQL | CloudNativePG | Backup terjadwal, point-in-time recovery, failover |
| Redis | Deployment biasa, tanpa persistence | Redis di sini cache dan counter; semua isinya bisa dibangun ulang (§4.6), jadi sengaja tidak diberi volume |

**Urutan pemasangan (v3.1).** Perkakas platform dipasang satu per satu, satu perkakas per sesi, dengan urutan ini:

1. ingress-nginx
2. cert-manager
3. Argo CD
4. kube-prometheus-stack
5. Strimzi (Kafka)
6. Altinity ClickHouse Operator
7. CloudNativePG
8. KEDA

Sebelum memasang yang berikutnya, dua hal harus ada: perkakas yang baru terpasang sudah dipakai oleh minimal satu beban kerja LogPilot, dan pemakaian memori node sebelum dan sesudahnya (`kubectl top`) tercatat di `docs/CAPACITY.md`. Dengan begitu, kalau ada yang rusak, penyebabnya adalah perkakas terakhir.

KEDA ditaruh terakhir karena ia butuh consumer dan metrik lag yang sudah berjalan. Kalau memori node habis sebelum daftar ini selesai, perkakas yang tersisa dijalankan di `lab` saja atau diganti cara yang lebih ringan (misalnya `StatefulSet` biasa tanpa operator). Itu dicatat sebagai keputusan di `DESIGN-DECISIONS.md`, bukan sebagai kegagalan.

Storage memakai `local-path` bawaan k3s di environment demo. Konsekuensinya dicatat: volume terikat ke satu node, jadi pod stateful tidak bisa pindah node. Di `lab`, Longhorn dipakai untuk merasakan bedanya.

**Backup dan pemulihan** adalah bagian dari definisi selesai, bukan tambahan:

| Data | Cara backup | Target | Diuji dengan |
|---|---|---|---|
| PostgreSQL | CloudNativePG ke object storage, harian + WAL | RPO ≤ 5 menit | Hapus cluster, pulihkan, login dengan user lama |
| ClickHouse | `clickhouse-backup` ke object storage, harian | RPO ≤ 24 jam (log 30 hari, bukan data transaksi) | Pulihkan ke namespace lain, bandingkan `count()` per hari |
| Kafka | Tidak di-backup | — | Kafka di sini buffer 7 hari; sumber kebenarannya ClickHouse |

Backup yang belum pernah dipulihkan dihitung sebagai belum ada.

### 6.5 Jalur Masuk, TLS, dan Live Demo Publik

- **ingress-nginx** sebagai satu-satunya jalur masuk; k3s dipasang tanpa Traefik bawaannya supaya sama dengan cluster AHU.
- **cert-manager** dengan Let's Encrypt (HTTP-01) menerbitkan dan memperpanjang sertifikat otomatis.
- Hostname: `app.<domain>` (Console), `ingest.<domain>` (Ingestor), `api.<domain>` (service `api`: Management API + Query API), `grafana.<domain>` (baca-saja, anonim, hanya dashboard System Health).
- Ingress untuk `/v1/logs/tail` butuh `proxy-buffering: off` dan timeout baca panjang; kalau tidak, SSE tertahan di buffer nginx dan Live Tail tampak macet.
- Rate limit di Ingress untuk endpoint login dan tombol simulasi insiden.

**Ukuran VPS.** Perkiraan awal kebutuhan memori seluruh stack di satu node:

| Komponen | Perkiraan RAM |
|---|---|
| k3s (control plane + agen) | 0,8 GB |
| Kafka (heap 512 MB) | 1,0 GB |
| ClickHouse | 1,5 GB |
| PostgreSQL, Redis | 0,3 GB |
| Prometheus, Grafana, Alertmanager, kafka-exporter | 0,7 GB |
| Argo CD, ingress-nginx, cert-manager, KEDA, operator | 1,2 GB |
| Service LogPilot (5 Go, Next.js) | 0,8 GB |
| **Total** | **± 6,3 GB** |

Jadi target VPS demo adalah **8 GB RAM / 4 vCPU**; VM 2 GB cukup untuk latihan k3s dan Ingestor saja. Tabel ini **hipotesis**, belum pengukuran: angka sebenarnya diambil dari `kubectl top` setelah berjalan seminggu dan dicatat di `docs/CAPACITY.md` (`R-S4`), termasuk selisihnya dari perkiraan ini.

### 6.6 CI/CD dan GitOps

```
push ke main
   │
   ▼
GitHub Actions: lint → unit test → integration test → build image → scan → push ghcr.io (tag = commit SHA)
   │
   ▼
commit otomatis: ubah image tag di repo `logpilot-deploy`
   │
   ▼
Argo CD (di dalam cluster) melihat perubahan → sinkronisasi → rolling update
```

- **Integration test** memakai testcontainers: Kafka, ClickHouse, dan Redis sungguhan di CI, bukan mock. Minimal satu test per service yang menyentuh dependensinya.
- **Scan image** dengan Trivy; build gagal untuk kerentanan kritis yang sudah ada perbaikannya.
- **Model pull, bukan push.** CI tidak memegang kubeconfig. Cluster-lah yang menarik dari Git, sehingga kebocoran secret CI tidak berarti akses ke cluster. Ini beda paling mendasar dari pipeline `kubectl set image` yang dipakai di AHU, dan layak dijelaskan di `DESIGN-DECISIONS.md`.
- **Git adalah sumber kebenaran.** Perubahan manual lewat `kubectl edit` akan dikembalikan Argo CD (self-heal). Rollback = `git revert`.
- **Secret** disimpan terenkripsi di Git dengan Sealed Secrets; hanya controller di dalam cluster yang bisa membukanya.
- **Migrasi skema** (ClickHouse dan PostgreSQL) dijalankan sebagai `Job` dengan hook pre-sync Argo CD, sebelum versi baru service naik. Migrasi harus kompatibel mundur satu versi supaya rolling update tidak memutus pod lama.

### 6.7 Observability untuk Platform

- **kube-prometheus-stack** menggantikan Prometheus dan Grafana yang dipasang manual: metrik node, pod, dan Kubernetes API tersedia tanpa konfigurasi.
- Service LogPilot di-scrape lewat `ServiceMonitor`, bukan daftar target statis.
- Alert untuk LogPilot sendiri didefinisikan sebagai `PrometheusRule`: consumer lag > 60 detik selama 5 menit, error rate Ingestor > 1%, DLQ bertambah, disk ClickHouse > 80%, sertifikat kedaluwarsa < 14 hari. Setiap alert punya tautan ke bagian runbook-nya di `docs/RUNBOOK.md`.
- **SLO LogPilot** (yang dijanjikan ke penggunanya): ketersediaan ingest 99,5% per 30 hari (rasio respons non-5xx di `/v1/ingest`), dan kesegaran data p99 < 10 detik (`ingested_at` sampai bisa di-query).

---

## 7. Demo Application

### 7.1 Dummy E-commerce API (Go)

Aplikasi HTTP yang mensimulasikan backend e-commerce, **menggunakan LogPilot SDK** (bukan HTTP call manual), sehingga trace context terbentuk secara alami.

| Endpoint | Behavior | Log yang Dihasilkan |
|---|---|---|
| GET `/products` | Selalu sukses | INFO: fetch products success |
| GET `/products/:id` | 80% sukses, 20% not found | INFO atau ERROR: product not found |
| POST `/checkout` | 70% sukses, 30% payment timeout | INFO atau ERROR: payment gateway timeout |
| GET `/health` | Selalu sukses | DEBUG: health check |

`/checkout` memanggil satu service dummy tambahan (payment) agar propagasi trace lintas service dapat didemonstrasikan secara nyata.

### 7.2 Log Generator Script (Go)

- Mode normal: 1 log setiap 200ms (5 logs/detik).
- Mix level: 60% INFO, 20% WARN, 15% ERROR, 5% FATAL.
- Setiap 2 menit: simulate error spike — 20 ERROR dalam 30 detik untuk memicu alert.
- **Mode `--chaos` (baru):** menghasilkan 50.000 error dari 5 penyebab berbeda, dengan variasi angka/ID di setiap pesan. Ini yang mendemonstrasikan fingerprinting.
- **Mode `--load`:** throughput tinggi untuk load testing.

### 7.3 Demo Script (Urutan Presentasi)

Ada dua naskah, karena dua penontonnya menilai hal yang berbeda.

**Naskah 60 detik (HR / recruiter, tanpa pendamping).** Harus bisa dijalankan sendiri dari URL publik:

1. Buka `app.<domain>` → masuk ke project `demo` tanpa login. Live Tail sudah bergerak.
2. Tekan **"Simulasikan insiden"** → layar dibanjiri baris merah. Satu kalimat di layar: *"50.000 error baru saja masuk. Bisa temukan masalahnya?"*
3. Klik **Issues** → 5 baris. Kalimat kuncinya: *"Bukan 50.000 masalah. Lima."*
4. Klik issue teratas → waterfall trace terbuka, satu span berwarna merah.
5. Banner kecil: *"Alert sudah terkirim 4 detik lalu"* dengan tangkapan layar task ClickUp.

Tidak ada langkah yang butuh penjelasan lisan. Kalau butuh, UI-nya yang harus diperbaiki, bukan naskahnya.

**Naskah 10 menit (interviewer teknis):**

1. Tunjukkan arsitektur di README — jelaskan setiap hop dan alasan keberadaannya.
2. Tunjukkan integrasi SDK di dummy app — 5–10 baris.
3. Jalankan naskah 60 detik di atas, kali ini sambil menjelaskan fingerprinting dan trace context.
4. Buka Explorer, ketik `level:error service:payment "timeout"`, bagikan URL-nya.
5. **Matikan pod ClickHouse** (`kubectl delete pod`). Tunjukkan: Live Tail tetap mengalir, consumer lag naik di Grafana, KEDA menambah consumer. Pod kembali → lag turun → `count()` sama dengan jumlah yang dikirim.
6. **Deploy lewat Git:** ubah satu baris, push, tunjukkan Argo CD menyinkronkan dan rolling update berjalan tanpa request gagal.
7. Tunjukkan `docs/BENCHMARKS.md` dan `chaos/scenarios.md`: setiap klaim di README punya angka di belakangnya.

### 7.4 README sebagai Halaman Depan

Kebanyakan orang tidak akan melewati layar pertama README. Layar pertama itu harus memuat, berurutan:

1. Satu kalimat: apa ini dan untuk siapa.
2. Tautan **Live Demo** dan GIF 15 detik naskah 60 detik di atas.
3. Diagram arsitektur (gambar, bukan ASCII).
4. Tabel 4–5 angka terukur dari `BENCHMARKS.md` (throughput, p99, waktu lookup trace, RTO), masing-masing dengan tautan ke cara pengukurannya.
5. Quick start: `docker compose up`, lalu `make demo`. Dua perintah, hasil terlihat di browser dalam 2 menit.

Setelah itu baru daftar fitur, keputusan desain, dan cara deploy ke Kubernetes. Urutannya mengikuti positioning di §1.4: pipeline backend dan angkanya dulu, lalu bagian platform sebagai bukti sistem yang sama dioperasikan sendiri (deploy dari Git, pemulihan, autoscaling), lalu tautan ke tulisan teknis (§7.5).

### 7.5 Tulisan Teknis (v3.1)

Setiap masalah sulit yang diselesaikan ditulis menjadi satu tulisan pendek (600–1.000 kata) di `docs/writing/` dan ditautkan dari README. Bentuknya tetap: masalahnya, kenapa pendekatan pertama salah, perbaikannya, dan angka atau test yang membuktikannya. Tulisan dibuat saat chunk-nya selesai, bukan dikumpulkan di akhir, dan ditulis oleh pemilik proyek (AI hanya me-review, sama seperti kodenya).

Kandidat pertama, semuanya sudah punya bahan di dokumen ini:

| Tulisan | Bahan |
|---|---|
| Deduplikasi insert pada pipeline at-least-once | §4.3, F9 |
| Kenapa materialized view issue menghitung salah | §5.3 |
| Rate limiter yang bisa mengunci satu tenant selamanya | §4.1 Layer 2 |
| Autoscaling consumer dari lag, bukan dari CPU | §6.3 |

---

## 8. Development Build Plan

**Status per 22 November 2026:**

| Phase | Status | Selesai |
|---|---|---|
| Phase 0 — Setup | ✅ Done | 2026-04-26 |
| Phase 1 — Ingestor (5 layer versi v1.0) | ✅ Done | 2026-05-17 |
| Phase 2 — Storage Worker | 🟡 Chunk 2.1–2.3 done; 2.4 sedang dikerjakan | — |
| Phase 3–7 | ⬜ Belum | — |
| Phase R — Revisi & Peningkatan | ⬜ Belum | — |
| Phase S — Showcase, Phase P — Platform | ⬜ Belum (baru di v3.0) | — |

**Kenapa tidak ada tanggal lagi.** Timeline v1.0 (19 April – 31 Mei) meleset. Timeline v2.0 (25 Juli – 20 September) juga meleset: empat bulan setelahnya, posisinya masih di Phase 2. Dua kali meleset dengan cara yang sama berarti masalahnya bukan di disiplin, melainkan di bentuk rencananya: jadwal mingguan untuk pekerjaan paruh waktu yang tidak teratur selalu kalah dari kenyataan.

v3.0 memakai **milestone dengan exit demo**. Satu milestone selesai ketika demonya bisa direkam, kapan pun itu. Kolom ukuran adalah perkiraan jumlah sesi (satu sesi = satu chunk, 30–90 menit), supaya kemajuan bisa dihitung tanpa kalender.

| # | Milestone | Tier | Isi (chunk di TODO) | Ukuran | Exit demo |
|---|---|---|---|---|---|
| M1 | **Log tersimpan** | Core | `R-J1`, `R-J2`, `R-J3`, Chunk 2.4–2.7, `R-I5` | ±12 sesi | `curl` satu log ber-`trace_id` → tampil di Grafana. Matikan worker di tengah batch → `count()` tetap tepat |
| M2 | **Terukur** | Core | `R-J6`, `R-I7`, `R-S3` (versi awal) | ±6 sesi | Dashboard System Health menampilkan throughput dan lag. Angka k6 pertama tercatat di `BENCHMARKS.md` |
| M3 | **Alert berbunyi** | Core | `R-I2`, Chunk 3.1–3.12 | ±16 sesi | 10 ERROR → task ClickUp muncul. Selama cooldown, counter tetap naik |
| M4 | **Bisa dipakai orang lain** | Core | Chunk 4.1–4.10 (minimum), `R-I4`, `R-I3`, Chunk 6.1–6.2 | ±20 sesi | Buat project dan API key lewat API, pasang SDK di dummy app, trace lintas dua service terlihat |
| M5 | **50.000 → 5** | Core | `R-S1`, `R-S2` | ±10 sesi | Mode `--chaos` → query `issues_hourly` mengembalikan 5 baris dengan hitungan yang benar setelah `OPTIMIZE` |
| M6 | **Ada wajahnya** | Showcase | `S-1` sampai `S-9` | ±25 sesi | Naskah 60 detik berjalan di lokal, dijalankan orang lain tanpa bantuan |
| M7 | **Hidup di internet** | Platform | `P-1` sampai `P-10` | ±20 sesi | URL publik ber-HTTPS; `git push` → ter-deploy; pod dibunuh → pulih |
| M8 | **Terbukti** | Core + Platform | `R-S3` (final), `R-S4`, `R-S7`, `R-S8`, README, video | ±10 sesi | Setiap angka di README punya baris di `BENCHMARKS.md` atau `chaos/scenarios.md` |

**Selingan Platform.** `P-1` sampai `P-4` (k3s, ingress, TLS, Argo CD) tidak bergantung pada fitur apa pun dan boleh dikerjakan kapan saja setelah M1, dengan Ingestor sebagai satu-satunya beban kerja. Ini cara menjaga belajar DevOps tetap berjalan tanpa menunggu M6.

**Chunk yang tidak memblokir apa pun** dan dapat disisipkan kapan saja: `R-J4` (batch endpoint), `R-J5` (clock skew), `R-I1` (bloom filter), `R-I6` (DLQ), `R-I8` (Kafka durability), `R-I9` (OTel), `R-I10` (PII), `R-I11` (performa Consumer 2), `R-S5` (backpressure), `R-S9` (sampling). `R-S6` (multi-tenancy) sekarang diselesaikan oleh `S-1` (Query API).

**Jalur pendek** bila interview datang lebih cepat: `R-J1` → `R-J2` → `R-I2` → `R-S1` → `R-S8`. Lima chunk ini menutup pertanyaan yang sudah pernah diterima, beserta kemampuan menjelaskannya. Jalur ini tidak menghasilkan apa pun yang bisa diklik; itu harga jalur pendek.

---

## 9. Repository Structure

```
logpilot/
├── services/
│   ├── ingestor/               # Go Ingestor Service
│   │   ├── main.go
│   │   ├── handler/            # ingest.go, otlp.go
│   │   ├── middleware/         # auth, rate limit, validation
│   │   ├── enricher/           # enricher.go, scrubber.go
│   │   ├── producer/           # Kafka producer
│   │   └── Dockerfile
│   │
│   ├── consumer-storage/       # Consumer 1
│   │   ├── consumer/
│   │   ├── storage/            # ClickHouse client
│   │   ├── dlq/                # dead letter queue
│   │   └── Dockerfile
│   │
│   ├── consumer-alert/         # Consumer 2
│   │   ├── consumer/
│   │   ├── evaluator/          # rule evaluation logic
│   │   ├── counter/            # bucketed window counter
│   │   ├── notifier/           # alertmanager client
│   │   └── Dockerfile
│   │
│   ├── alert-dispatcher/       # Go Alert Dispatcher
│   │   ├── handler/            # webhook receiver
│   │   ├── destinations/       # clickup, sendgrid, slack
│   │   └── Dockerfile
│   │
│   ├── api/                    # Go: Management API + Query API (v3.1)
│   │   ├── management/         # sesi, project, API key, alert rule, notifikasi, status issue
│   │   ├── internalapi/        # listener internal :8082 untuk Ingestor dan Dispatcher
│   │   ├── query/              # logs, tail (SSE), issues, traces
│   │   ├── auth/               # sesi, penerbitan dan verifikasi JWT
│   │   ├── lpql/               # parser bahasa query
│   │   ├── tail/               # pembaca Kafka + fan-out ke penonton
│   │   └── Dockerfile
│   │
│   └── frontend/               # Next.js — LogPilot Console
│
├── sdk/
│   └── go/                     # LogPilot SDK — trace context lahir di sini
│       ├── trace.go
│       ├── context.go
│       ├── middleware.go
│       ├── client.go
│       ├── sender.go
│       └── README.md
│
├── pkg/
│   └── fingerprint/            # dipakai bersama: normalize.go, fingerprint.go
│
├── migrations/
│   ├── clickhouse/             # 001_create_logs.sql, 002_..., run.sh
│   └── postgres/               # skema Management API (v3.1)
│
├── demo/
│   ├── dummy-app/              # Go e-commerce dummy (pakai SDK)
│   ├── dummy-payment/          # service kedua, untuk demo trace lintas service
│   └── log-generator/          # normal / --chaos / --load
│
├── loadtest/
│   └── ingest.js               # skenario k6
│
├── chaos/
│   └── scenarios.md            # hasil failure injection, RTO/RPO
│
├── deploy/
│   ├── docker-compose/         # environment dev
│   └── helm/logpilot/          # chart + values-demo.yaml, values-lab.yaml
│                               # (manifest Argo CD dan infra ada di repo logpilot-deploy)
│
├── config/
│   ├── grafana/                # dashboard JSON
│   ├── prometheus/             # prometheus.yml
│   ├── alertmanager/           # alertmanager.yml
│   └── otel-collector/         # otel-collector-config.yaml
│
├── .github/workflows/ci.yml
│
├── docs/
│   ├── PRD.md
│   ├── TODO.md
│   ├── DESIGN-DECISIONS.md     # kenapa X bukan Y, per komponen
│   ├── BENCHMARKS.md           # hasil load test, before/after optimasi
│   ├── CAPACITY.md             # proyeksi kapasitas dan biaya
│   ├── RUNBOOK.md              # satu bagian per alert: gejala, cara cek, cara pulih
│   ├── demo/                   # rekaman exit demo per milestone
│   └── INTERVIEW-NOTES.md      # jawaban tertulis untuk pertanyaan sulit
│
└── README.md
```

---

## 10. Success Metrics & Definition of Done

### 10.1 Functional Requirements

| # | Requirement | Acceptance Criteria |
|---|---|---|
| F1 | Log ingestion berfungsi | POST `/v1/ingest` dengan valid API key return 202, log muncul di ClickHouse dalam 5 detik |
| F2 | Authentication berfungsi | Request tanpa API key atau dengan key invalid return 401 |
| F3 | Rate limiting berfungsi | Log ke-10.001 dalam 60 detik return 429 |
| F4 | Query berfungsi | Query bisa filter by `level`, `service`, `trace_id`, time range, lewat Query API (tenant) dan Grafana (operator) |
| F5 | Alert trigger berfungsi | Setelah 10 ERROR dalam 5 menit, ClickUp task ter-create dan email terkirim |
| F6 | Cooldown berfungsi | Alert tidak trigger lagi dalam 10 menit, **tetapi counter tetap menghitung selama periode tersebut** |
| F7 | API key management | Generate API key via Management API, key langsung bisa digunakan di Ingestor |
| F8 | Alert rule management | Buat alert rule via Management API, rule aktif dalam 30 detik di Consumer 2 |
| F9 | **No data loss, no duplicate** | Matikan ClickHouse 1 menit — semua log yang masuk selama downtime tersimpan, **dan `count()` tepat sama dengan jumlah yang dikirim** (bukan lebih) |
| **F10** | **Trace correlation** | Satu request lintas 2 service menghasilkan log dengan `trace_id` identik dan `span_id` berbeda; query by `trace_id` mengembalikan seluruh log request tersebut secara urut |
| **F11** | **Error fingerprinting** | 1.000 error dengan pesan bervariasi dari 5 penyebab berbeda mengelompok menjadi tepat 5 fingerprint di halaman Issues |
| **F12** | **Poison message tidak memblok** | Log dengan timestamp di luar jangkauan masuk `raw-logs-dlq`, dan log lain di partition yang sama tetap mengalir |
| **F13** | **Auth tetap jalan saat Redis down** | Matikan Redis — log tetap diterima lewat fallback PostgreSQL; nyalakan lagi — cache terisi ulang otomatis |
| **F14** | **Batch ingest partial success** | Kirim 500 log dengan 3 invalid → response 207, 497 tersimpan, 3 dilaporkan beserta index dan alasannya |
| **F15** | **Load shedding** | Pada beban 5x kapasitas, pod tidak mati; DEBUG dibuang, ERROR dan FATAL tetap tersimpan seluruhnya |
| **F16** | **Hitungan issue stabil** (v3.0) | Kirim 10.000 error dari 5 penyebab dalam 20 batch, jalankan `OPTIMIZE TABLE issues_hourly FINAL` → `sum(event_count)` tetap 10.000, sebelum dan sesudah |
| **F17** | **Rate limit tidak pernah macet** (v3.0) | Tidak ada key `rate:*` tanpa TTL setelah Ingestor dibunuh 100 kali di tengah beban (`redis-cli --scan` + `TTL`) |
| **F18** | **Isolasi tenant** (Showcase) | Token project A tidak bisa membaca data project B lewat endpoint mana pun di Query API, termasuk dengan menyisipkan `project_id` di `q`. Diuji otomatis di CI |
| **F19** | **Query tidak bisa menjatuhkan ClickHouse** (Showcase) | Query tanpa rentang waktu ditolak 400; query yang melewati `max_execution_time` berakhir 408, ingest tetap berjalan |
| **F20** | **Live Tail** (Showcase) | Log tampil di browser < 2 detik setelah `202`. Penonton yang sengaja diperlambat tidak menahan penonton lain. Tetap mengalir saat ClickHouse dimatikan |
| **F21** | **Issue lifecycle** (Showcase) | Fingerprint baru memicu tepat satu alert `new_issue` walau Consumer 2 berjalan 3 pod. Issue yang di-resolve lalu muncul lagi berubah ke `open` dan memicu `regression` |
| **F22** | **Naskah 60 detik** (Showcase) | Dijalankan orang non-teknis dari URL tanpa penjelasan lisan; mereka bisa menyebutkan ada berapa masalah |
| **F23** | **Deploy lewat Git** (Platform) | Commit ke `main` sampai berjalan di cluster tanpa `kubectl` manual dan tanpa request gagal selama rolling update (diukur k6 yang berjalan selama deploy) |
| **F24** | **Pulih sendiri** (Platform) | `kubectl delete pod` pada Ingestor, Consumer 1, dan ClickHouse satu per satu → sistem pulih tanpa campur tangan, tanpa log hilang atau ganda |
| **F25** | **Pemulihan dari backup** (Platform) | PostgreSQL dan ClickHouse dipulihkan dari backup ke namespace kosong; hasilnya dicatat beserta lama waktunya |

### 10.2 Non-Functional Requirements

| Aspek | Target | Cara Verifikasi |
|---|---|---|
| Log ingestion latency | p95 < 100ms, **p99 < 200ms** dari HTTP request ke Kafka publish | k6 (`R-S3`) |
| Log visibility latency | < 5 detik dari ingest sampai muncul di Grafana | Manual + dashboard |
| Alert latency | < 30 detik dari kondisi terpenuhi ke notifikasi terkirim | End-to-end test |
| Throughput | Go Ingestor **terukur** ≥ 1.000 req/detik di single pod | k6, dicatat di `docs/BENCHMARKS.md` |
| Trace lookup | Query by `trace_id` < 500ms pada tabel ≥ 100 juta baris | `EXPLAIN indexes=1`, catat `rows read` sebelum/sesudah bloom filter |
| Consumer lag | < 10 detik dalam kondisi normal | Prometheus + kafka-exporter |
| Availability | Sistem tetap menerima log meskipun ClickHouse down (Kafka buffer) | Chaos test skenario 1 |
| Recovery | RTO < 2 menit, RPO = 0 untuk kegagalan ClickHouse | Chaos test, dicatat di `chaos/scenarios.md` |
| Query latency (Showcase) | Explorer p95 < 1 detik untuk rentang 24 jam pada tabel ≥ 100 juta baris; daftar Issues p95 < 300ms | k6 terhadap Query API |
| Live Tail latency (Showcase) | p95 < 2 detik dari `202` sampai tampil | Timestamp di payload vs waktu terima di browser |
| Deploy lead time (Platform) | < 10 menit dari `git push` sampai pod baru melayani trafik | Selisih waktu commit dan `Synced` di Argo CD |
| Jejak resource (Platform) | Seluruh stack demo ≤ 8 GB RAM | `kubectl top`, dicatat di `CAPACITY.md` |

> **Perubahan penting dari v1.0:** setiap angka di tabel ini disertai **cara verifikasinya**. Di v1.0, "1.000 req/detik" adalah target yang belum pernah diukur, sementara log generator hanya mengirim 5 log/detik — 200 kali lebih kecil dari klaimnya. Klaim performa yang tidak terukur lebih merugikan daripada tidak ada klaim sama sekali: satu pertanyaan lanjutan dari interviewer yang tidak menemukan bukti di baliknya akan membuat seluruh isi README ikut diragukan.

### 10.3 Definition of Done per Komponen

Sebuah komponen dinyatakan selesai bukan ketika kodenya jalan, melainkan ketika:

1. Fungsinya dapat dijelaskan beserta alasan desainnya
2. Mode kegagalannya di production teridentifikasi
3. Trade-off dari pendekatan yang dipilih dapat diuraikan
4. Pertanyaan **❓ Bisa jawab ini?** pada chunk PHASE R terkait dapat dijawab
5. Design Decision Record satu paragraf tertulis di `docs/DESIGN-DECISIONS.md`, memuat: apa yang dibangun, kenapa pendekatan ini dan bukan alternatifnya, apa yang akan diubah bila requirement berubah, dan **angka apa yang diukur**

---

## 11. Production Readiness Checklist

"Production ready" di sini berarti setiap butir di bawah punya **bukti yang bisa ditunjukkan**, bukan sekadar sudah dikerjakan. Butir tanpa bukti dihitung belum selesai.

| Area | Butir | Bukti | Tier |
|---|---|---|---|
| Keamanan | Semua jalur publik lewat HTTPS | `curl -I http://...` dialihkan; sertifikat diterbitkan cert-manager | Platform |
| Keamanan | Tidak ada secret terbaca di Git maupun image | Hasil `gitleaks` di CI; Sealed Secrets | Platform |
| Keamanan | API key disimpan sebagai hash di PostgreSQL dan Redis | Isi Redis hanya memuat `api_key:{sha256}` | Core |
| Keamanan | Endpoint `/internal/*` tidak terjangkau dari luar | `NetworkPolicy` + percobaan `curl` dari pod lain yang gagal | Platform |
| Keamanan | Container berjalan non-root, image tanpa kerentanan kritis | `securityContext`; laporan Trivy | Platform |
| Keandalan | Tidak ada log hilang atau ganda saat komponen mati | F9, F24, `chaos/scenarios.md` | Core |
| Keandalan | Satu pesan rusak tidak membekukan partition | F12 | Core |
| Keandalan | Graceful shutdown di semua service | Rolling update tanpa request gagal (F23) | Platform |
| Keandalan | Backup terjadwal dan pernah dipulihkan | F25 | Platform |
| Operasi | Setiap alert punya runbook | `docs/RUNBOOK.md`, tautan di anotasi alert | Platform |
| Operasi | SLO terdefinisi dan terpantau | Panel SLO di Grafana (§6.7) | Platform |
| Operasi | Log service LogPilot terstruktur (JSON) dan membawa `trace_id` | Contoh log dari satu request lintas Ingestor dan Consumer | Core |
| Delivery | Test otomatis di CI, termasuk integration test | Badge CI; minimal satu test per service terhadap dependensi asli | Core |
| Delivery | Migrasi skema bernomor, dijalankan otomatis, kompatibel mundur | `migrations/`; Job pre-sync | Platform |
| Delivery | Rollback teruji | `git revert` mengembalikan versi sebelumnya dalam < 10 menit | Platform |
| Data | Retensi diterapkan otomatis | TTL di DDL; partisi tertua ≤ 30 hari | Core |
| Data | PII di-scrub sebelum disimpan | Test scrubber; contoh log sebelum dan sesudah | Core |
| Kapasitas | Batas sistem diketahui dan tertulis | `BENCHMARKS.md` (titik jenuh), `CAPACITY.md` | Core |

## 12. Stretch (Tier 4)

Tidak ada yang wajib, dan tidak ada yang dimulai sebelum M7. Urutan baris adalah urutan prioritas. Dicatat supaya idenya tidak hilang dan tidak diselundupkan ke tier lain.

| Fitur | Nilainya | Yang dipelajari | Risiko |
|---|---|---|---|
| **Handler Monolog (PHP)** | Aplikasi PHP mengirim log ke LogPilot dengan satu baris konfigurasi. Butir Stretch pertama sejak v3.1 | Membuat paket Composer; relevan langsung dengan stack AHU | SDK kedua berarti beban pemeliharaan kedua |
| **Ringkasan insiden dengan LLM** | Tombol "Jelaskan issue ini": ringkasan pola error, kemungkinan penyebab, dan trace yang layak dibuka. Paling menarik perhatian orang non-teknis | Memanggil Claude API dari Go, membatasi konteks (contoh pesan + tren, bukan seluruh log), cache per fingerprint | Mengirim isi log ke pihak ketiga: hanya setelah PII scrubbing, opt-in per project, dan mati secara bawaan |
| **SLO burn-rate alert** | Alert berdasarkan laju habisnya error budget (multi-window), bukan ambang tetap | Konsep SRE yang paling sering ditanyakan untuk peran platform | Butuh data rasio sukses/gagal, bukan hanya hitungan error |
| **Cold storage** | Log > 7 hari pindah ke object storage (`TTL ... TO VOLUME`), retensi 90 hari tanpa menambah disk | Storage policy ClickHouse, trade-off biaya vs latensi query | Query ke data dingin lambat; harus terlihat di UI |
| **Dogfooding** | Service LogPilot mengirim log-nya sendiri ke LogPilot, project `logpilot-internal` | Menjadi pengguna produk sendiri | Lingkaran umpan balik: error Ingestor menghasilkan log yang dikirim ke Ingestor. Wajib lewat jalur terpisah dengan drop-on-failure |
| **Cluster kubeadm multi-node** | Kafka 3 broker dan ClickHouse 2 replika sungguhan di environment `lab` | `min.insync.replicas`, `ReplicatedMergeTree`, node drain, kegagalan node | Biaya sewa VM; cukup dinyalakan saat latihan |

---

*LogPilot PRD v3.1 | Firman | November 2026*
