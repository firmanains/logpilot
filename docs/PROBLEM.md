# LogPilot — Problem Log
> Dokumentasi issue yang ditemukan selama development, penyebabnya, dan solusinya.

======================================================================================

## Issue #1 — Kafka Console Consumer Silent Hang dengan `--from-beginning`

### Symptom
Menjalankan command berikut dari dalam container Kafka menghasilkan cursor berkedip tanpa output apapun, tanpa error:

```bash
docker exec -it logpilot-kafka /opt/kafka/bin/kafka-console-consumer.sh \
  --bootstrap-server localhost:9092 \
  --topic raw-logs \
  --from-beginning
```

Padahal pesan sudah berhasil dipublish (Go ingestor return 202). Command `--partition 0 --offset earliest` berhasil menampilkan pesan.

### Root Cause
**Kafka Advertised Listeners misconfiguration.**

Kafka memiliki dua konsep berbeda:
- **Listener**: alamat tempat Kafka *mendengarkan* koneksi masuk
- **Advertised Listener**: alamat yang Kafka *beritahukan* ke client setelah connect

Konfigurasi di `docker-compose.yml`:
```yaml
KAFKA_ADVERTISED_LISTENERS: 'PLAINTEXT://kafka:9092,EXTERNAL://localhost:9093'
```

Ketika consumer di dalam container connect ke `localhost:9092` sebagai bootstrap, Kafka merespons dengan advertised listener `EXTERNAL://localhost:9093`. Consumer kemudian mencoba connect ke `localhost:9093` — tapi dari dalam container, `localhost` menunjuk ke container itu sendiri, bukan host Mac. Koneksi gagal.

`--from-beginning` menggunakan **consumer group protocol** (subscribe API) yang membutuhkan koneksi ke broker via advertised listener. Sedangkan `--partition 0 --offset earliest` menggunakan **manual partition assignment** yang bypass consumer group protocol sehingga tidak terpengaruh.

### Solution
Gunakan `kafka:9092` (PLAINTEXT listener — untuk internal Docker network) sebagai bootstrap server saat consume dari dalam container:

```bash
docker exec -it logpilot-kafka /opt/kafka/bin/kafka-console-consumer.sh \
  --bootstrap-server kafka:9092 \
  --topic raw-logs \
  --from-beginning
```

**Panduan listener:**
- `kafka:9092` (PLAINTEXT) → dipakai dari dalam Docker network (antar service, docker exec)
- `localhost:9093` (EXTERNAL) → dipakai dari host machine (Go ingestor yang jalan di local)

======================================================================================

## Issue #2 — Consumer Group Tidak Pernah Terbentuk (`__consumer_offsets` Gagal Dibuat)

### Symptom
Setelah fix Issue #1, consumer dengan `--from-beginning` dan `kafka:9092` tetap tidak menampilkan pesan. Broker log dipenuhi pesan yang terus berulang:

```
INFO Sent auto-creation request for Set(__consumer_offsets) to the active controller.
INFO Sent auto-creation request for Set(__consumer_offsets) to the active controller.
... (retry tiap ~500ms tanpa henti)
```

`kafka-consumer-groups.sh --list` mengembalikan hasil kosong meskipun consumer sedang running — artinya consumer group tidak pernah terbentuk.

### Root Cause
**Default replication factor `__consumer_offsets` adalah 3, tapi cluster hanya punya 1 broker.**

`__consumer_offsets` adalah internal topic Kafka yang wajib ada untuk menyimpan committed offset setiap consumer group. Kafka mencoba membuatnya secara otomatis saat pertama kali ada consumer group yang join.

Default `offsets.topic.replication.factor` = 3 (asumsi production cluster minimal 3 broker). Dengan hanya 1 broker, syarat ini tidak terpenuhi dan topic gagal dibuat. Tanpa `__consumer_offsets`, consumer group protocol tidak bisa berjalan — consumer terhubung ke broker tapi tidak bisa join group, sehingga diam tanpa error.

Dari dokumentasi Kafka:
> "Internal topic creation will fail until the cluster size meets this replication factor requirement."

### Solution
Tambahkan environment variable berikut ke Kafka service di `docker-compose.yml`:

```yaml
KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
```

Ini memberitahu Kafka bahwa cluster ini intentionally single-broker, sehingga `__consumer_offsets` bisa dibuat dengan 1 replica.

Restart container setelah perubahan:
```bash
docker-compose restart kafka
```

======================================================================================

## Issue #3 — consumer-storage Gagal Connect ke Kafka (`connection refused` di localhost:9092)

### Symptom
Menjalankan `go run cmd/main.go` di `consumer-storage` menghasilkan error:

```
failed to create consumer group: kafka: client has run out of available brokers to talk to: dial tcp [::1]:9092: connect: connection refused
```

### Root Cause
Default config di `consumer-storage` menggunakan `localhost:9092` sebagai Kafka broker address. Port 9092 adalah PLAINTEXT listener yang hanya bisa diakses dari dalam Docker network, bukan dari host machine. Go service yang jalan di host Mac harus pakai EXTERNAL listener.

### Solution
Ubah Kafka broker address di `.env` consumer-storage (atau default config) dari `localhost:9092` ke `127.0.0.1:9093`:

```
KAFKA_BROKERS=127.0.0.1:9093
```

**Panduan:**
- `kafka:9092` (PLAINTEXT) → dari dalam Docker network
- `127.0.0.1:9093` (EXTERNAL) → dari host machine (Go service yang jalan di local)

======================================================================================
