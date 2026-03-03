---
title: "Roadmap"
weight: 9
menu:
  main:
    weight: 9
---

This roadmap tracks planned provider modules and core improvements.

## Near-Term

### Kafka Provider (`providers/kafkatestage`)

Goals:

- provider-seed API for message production using the same pattern as `mongotestage`:
  `engine.SeedWithProvider(...)`
- admin checks for topic metadata, consumer groups, lag, and offsets
- validation helpers for consumed messages (key/value/headers/partition/offset)
- deterministic test lifecycle with factory-level client management
- optional seed modes for producer behavior (single send, batch send, sync/async acks)

Initial API ideas:

- `WithKafkaConnection(...)`
- `kafkatestage.NewKafkaRunnerFromEngine(...)`
- `kafkaseeder.WithProduce(...)`
- `kafkaseeder.WithClearAndProduce(...)` (topic reset + publish flow when available)
- `kafkachecker.ExpectConsumed(...)`
- `kafkachecker.ExpectGroupLag(...)`

### Redis Provider (`providers/redistestage`)

Goals:

- seed helpers for strings/hashes/sets/zsets/streams
- direct assertions on key existence, TTL, values, and stream entries
- optional pub/sub and stream consumer-group checks
- cleanup controls per test engine (flush db / namespace strategy)

Initial API ideas:

- `WithRedisConnection(...)`
- `redisseeder.WithSet(...)`
- `redisseeder.WithHSet(...)`
- `redischecker.ExpectKey(...)`
- `redischecker.ExpectTTL(...)`

## Medium-Term

### gRPC Runner/Provider Extension

- reusable gRPC request runner in the same style as `netoche`
- optional provider for client/channel lifecycle
- typed response extraction into runner context

## Long-Term

### Message Contract Assertions

- reusable schema-aware checks (JSON schema / Avro / Protobuf)
- contract validation helpers integrated with provider runners

### Cross-Provider Scenario Runner

- orchestrate multi-step workflows across HTTP + DB + Kafka + Redis
- deterministic step state propagation and failure diagnostics

### Observability Test Helpers

- assertions for traces, metrics, and logs emitted during test execution
- provider modules for OTLP collectors and log sinks
