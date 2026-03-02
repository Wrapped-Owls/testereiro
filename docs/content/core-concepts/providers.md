---
title: "Providers"
weight: 3
---

Providers are typed resources stored on engines and factories.

Use them to attach external clients or custom resources without global state.

## Provider Keys

Create keys with generics:

```go
var httpClientKey = puppetest.NewProviderKey[http.Client]()
var taggedKey = puppetest.NewTaggedProviderKey[sql.DB]("readonly")
```

## Engine-Level Providers

Set and retrieve from an engine:

```go
var cfgKey = puppetest.NewProviderKey[MyConfig]()

cfg := &MyConfig{BaseURL: "http://localhost"}
_ = puppetest.SetProvider(engine, cfgKey, cfg, func(ctx context.Context, c *MyConfig) error {
	return nil
})

cfgPtr, ok := puppetest.Provider[MyConfig](engine, cfgKey)
```

## Factory-Level Providers

Factory providers are stored once, then optionally bound to each newly-created engine.

```go
var factoryProviderKey = puppetest.NewProviderKey[MyProvider]()
var engineProviderKey = puppetest.NewProviderKey[MyProvider]()

err := puppetest.RegisterFactoryProvider(
	factory,
	factoryProviderKey,
	provider,
	func(ctx context.Context, e *puppetest.Engine, v *MyProvider) error {
		return puppetest.SetProvider(e, engineProviderKey, v, nil)
	},
	func(ctx context.Context, v *MyProvider) error {
		return v.Close(ctx)
	},
)
```

## Resolution API

Any type implementing `ProviderResolver` can use:

```go
v, ok := puppetest.ResolveProvider[T](resolver, key)
```

This is used internally by both engine and factory retrieval helpers.
