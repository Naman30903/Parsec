Parsec — Distributed Log Processing & Alerting (skeleton)

This repository is a skeleton project for a distributed log processing and alerting system with real-time analytics.

Core design choices (configurable):
- Broker: Kafka (topic-per-tenant or keyed partitions)
- Processor: Go (low-latency, efficient concurrency)
- Storage: ClickHouse for large-scale analytics; Postgres (partitioned) for smaller scale
- Stateful processing: Kafka consumer offsets + Redis for ephemeral state; persist aggregated checkpoints to DB
- Alerting: Start with rule-based alerts; add ML anomaly detectors as a next step

This scaffold provides:
- Docker Compose for local dev (Kafka, Zookeeper, Redis, ClickHouse, Postgres)
- Minimal Go service skeleton under `cmd/processor` and `internal/*` packages
- Simple unit test and Makefile targets

Quick start (local dev):

1. Start infrastructure:

```bash
cd /Users/namanjain.3009/Documents/Parsec
make up
```

2. Build and run processor locally:

```bash
make build
./bin/processor
```

What to add next:
- Implement Kafka consumers and commit/seek semantics
- Add ClickHouse/Postgres connectors and schema migration
- Add Redis stateful logic and checkpoint persistence
- Implement alerting rules and an alerts delivery subsystem

See `Makefile` and `docker-compose.yml` for local bootstrapping.
