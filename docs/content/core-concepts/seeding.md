---
title: "Seeding"
weight: 5
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

### Table Names

The table is the snake_case type name made plural: `Game` seeds `games`, `JobCategory` seeds
`job_categories` (a consonant before the final `y` takes `ies`), `ApiKey` seeds `api_keys`, and a
name ending in `x` is left as it is, so `JobOutbox` seeds `job_outbox`.

A type whose table does not follow that rule names it with a `TableName() string` method:

```go
type UsageMonthly struct {
	Month string `db:"month"`
}

func (UsageMonthly) TableName() string {
	return "usage_monthly"
}
```

A type you cannot add a method to, such as a generated row, gets one through a defined type over
it: `type UsageMonthly gen.UsageMonthly` keeps the fields and their `db` tags.

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

Mongo examples use `providers/mongotestage/pkg/mongoseeder` to seed collections.

## Seed Hooks

`WithBeforeEngineSeed` runs before both:

- `engine.Seed(...)`
- `engine.SeedWithProvider(...)`
