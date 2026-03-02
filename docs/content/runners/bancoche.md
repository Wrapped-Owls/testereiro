---
title: "Bancoche (SQL Runner)"
weight: 2
---

Package: `github.com/wrapped-owls/testereiro/puppetest/pkg/atores/bancoche`

`bancoche` runs SQL queries and validates rows.

## Basic Usage

```go
runner := bancoche.New(
	engine.DB(),
	bancoche.WithMapQuery("jokers", map[string]any{"rarity": "Common"}),
	bancoche.ExpectCount(2, true),
)

err := engine.Execute(t, runner)
```

## Query Builders

- `WithMapQuery(table, filter)`
- `WithMapQueryFromCtx(table, fn)`
- `WithQuery(bancoche.NewRawQuery(query, args...))`

## Validators

- `ExpectCount(expected, countRows)`
- `WithCustomValidation(func(t testing.TB, rows *sql.Rows) error)`

`WithCustomValidation` is useful when row scanning logic is domain-specific.

`ExpectCount` behavior:

- `countRows=true`: counts the number of returned rows.
- `countRows=false`: expects the query to return a `COUNT(*)`-style single value and scans it.
