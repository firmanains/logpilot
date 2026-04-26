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
