---
title: "Seeding"
weight: 4
---

Testereiro supports two seeding paths.

## SQL Struct Seeding

`engine.Seed(...)` inserts struct values into SQL tables.

```go
type Game struct {
	ID    int    `db:"id"`
	Title string `db:"title"`
}

err := engine.Seed(Game{ID: 1, Title: "Hollow Knight"})
```

This requires a configured SQL database on the engine.

## Provider-Based Seeding

`engine.SeedWithProvider(...)` executes one or more `SeedProvider` values.

```go
seedDocs := []any{
	map[string]any{"_identity": "MegadwarfTron"},
	map[string]any{"_identity": "OptimadinPrime"},
}

err := engine.SeedWithProvider(
	mongoseeder.WithClearAndSeed("dungeonformers", seedDocs...),
)
```

`SeedProvider` is any type implementing:

```go
ExecuteSeed(engine *puppetest.Engine) error
```

Mongo examples use `providers/mongotest/pkg/mongoseeder` to seed collections.

## Seed Hooks

`WithBeforeEngineSeed` runs before both:

- `engine.Seed(...)`
- `engine.SeedWithProvider(...)`
