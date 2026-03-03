---
title: "Mongotestage"
weight: 1
---

Module: `github.com/wrapped-owls/testereiro/providers/mongotestage`

`mongotestage` wires a Mongo client at factory level and binds a per-engine database provider.

## Install

```bash
go get github.com/wrapped-owls/testereiro/providers/mongotestage
```

## Factory Setup

```go
factory, err := puppetest.NewEngineFactory(
	mongotestage.WithMongoConnection(mongotestage.ConnectionConfig{
		Host: "localhost",
		Port: 27017,
	}),
)
```

Or reuse an existing client:

- `mongotestage.WithMongoClient(client)`
- `WithMongoClient` still registers provider teardown on `factory.Close()`, so the client is disconnected when the
  factory closes.

## Accessing Resources

- `mongotestage.ClientFromFactory(factory)`
- `mongotestage.DatabaseFromEngine(engine)`

Engine database binding uses `engine.DBName()` internally, so each test engine maps to its own Mongo database name.

## Mongo Checker Runner

Package:
`github.com/wrapped-owls/testereiro/providers/mongotestage/pkg/mongochecker`

Use with:

- `mongotestage.NewMongoRunnerFromEngine(engine, opts...)`
- query options like `WithFindOneQuery`, `WithAggregateQuery`, `WithCountQuery`
- validators like `ExpectDoc`, `ExpectDocs`, `ExpectCount`, `WithCustomValidation`

## Mongo Seeder

Package:
`github.com/wrapped-owls/testereiro/providers/mongotestage/pkg/mongoseeder`

Use with `engine.SeedWithProvider(...)`:

```go
err := engine.SeedWithProvider(
	mongoseeder.WithClearAndSeed("dungeonformers", docs...).
		WithClientBulkWriteSeedMode(),
)
```

Seeding modes:

- `SeedModeInsertMany`
- `SeedModeClientBulkWrite`
