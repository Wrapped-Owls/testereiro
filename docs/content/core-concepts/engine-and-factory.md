---
title: "Engine And Factory"
weight: 1
---

## EngineFactory

`EngineFactory` is the owner of shared test resources.

Key responsibilities:

- configure DB creation (`WithConnectionFactory`)
- apply engine extensions (`WithExtensions`)
- register factory-level providers (`RegisterFactoryProvider`)
- run factory close lifecycle (`Close`)

Create once (commonly in `TestMain`) and reuse for all tests in the package.

## Engine

`Engine` is created per test via `factory.NewEngine(t)`.

It exposes:

- `BaseURL()` for test HTTP server address
- `DB()` / `DBName()` for SQL resources
- `Context()` for test-scoped context
- `Execute(t, runner)` for runner execution
- `Seed(...)` and `SeedWithProvider(...)` for data setup
- `Teardown()` for resource cleanup

## Lifecycle (High Level)

When `NewEngine(t)` is called:

1. engine is initialized
2. factory providers are bound into engine providers
3. extensions run in order
4. `t.Cleanup` teardown is attached

On cleanup:

1. DB-specific teardown (if configured by DB factory)
2. `engine.Teardown()` closes server/DB/providers

On factory close:

1. factory close hooks run
2. DB factory and factory providers are torn down

## Minimal Factory Example

```go
factory, err := puppetest.NewEngineFactory(
	puppetest.WithConnectionFactory(myPerformer, true),
	puppetest.WithExtensions(
		puppetest.WithMigrationRunner(migrationsFS),
		puppetest.WithTestServerFromEngine(func(e *puppetest.Engine) (http.Handler, error) {
			return NewHandler(e.DB()), nil
		}),
	),
)
```

This pattern matches the sqlite example.
