---
title: "Examples"
weight: 6
menu:
  main:
    weight: 6
---

The repository includes end-to-end examples showing different integration-test stacks.

## webapi

Path: `examples/webapi`

- Pure HTTP example
- Uses `WithTestServer` + `netoche`
- No external database dependency

## sqlite

Path: `examples/sqlite`

- In-memory SQLite setup
- Uses `WithConnectionFactory` + `WithMigrationRunner` + `WithTestServerFromEngine`
- Seeds data with `engine.Seed(...)`
- Validates API with `netoche`

## testcontainers_mysql

Path: `examples/testcontainers_mysql`

- MySQL via testcontainers
- Uses factory hooks to close containers (`WithAfterFactoryClose`)
- Uses `bancoche` custom SQL validation

## mongodb_assert

Path: `examples/mongodb_assert`

- MongoDB via testcontainers
- Uses `mongotestage.WithMongoConnection`
- Seeds via `mongoseeder`
- Asserts API with `netoche` and DB state with `mongochecker`

## Run Examples

From repository root:

```bash
make examples
```

## Runtime Prerequisites

- `webapi`, `sqlite`: no container runtime required.
- `testcontainers_mysql`, `mongodb_assert`: require Docker (or a compatible Testcontainers runtime) available on the
  machine.
