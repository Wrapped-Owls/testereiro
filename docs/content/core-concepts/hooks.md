---
title: "Hooks"
weight: 2
---

Hooks let you customize behavior around engine and factory lifecycle events.

## Available Hook Options

Factory creation hooks:

- `WithBeforeEngineCreate`
- `WithAfterEngineCreate`

Engine execution hooks:

- `WithBeforeEngineRun`
- `WithAfterEngineRun`

Engine seeding hooks:

- `WithBeforeEngineSeed`

Engine teardown hooks:

- `WithBeforeEngineTeardown`
- `WithAfterEngineTeardown`

Factory close hooks:

- `WithBeforeFactoryClose`
- `WithAfterFactoryClose`

## Execution Order

- `before*` hooks run in registration order.
- `after*` hooks run in reverse registration order.
- Hook errors are joined and returned.

## Example

```go
factory, err := puppetest.NewEngineFactory(
	puppetest.WithBeforeEngineRun(func(evt *puppetest.EngineRunEvent) error {
		evt.TB.Log("about to run", evt.Runner)
		return nil
	}),
	puppetest.WithAfterFactoryClose(func(evt *puppetest.FactoryCloseEvent) error {
		// close external resources here (for example testcontainers)
		_ = evt.Factory
		return nil
	}),
)
```

A real use case from examples is closing testcontainers resources with `WithAfterFactoryClose`.
