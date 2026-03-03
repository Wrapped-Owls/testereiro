---
title: "Getting Started"
weight: 2
menu:
  main:
    weight: 2
---

## Requirements

- Go `1.24+`
- A test framework using `testing` (`go test`)

## Install Core

```bash
go get github.com/wrapped-owls/testereiro/puppetest
```

## Minimal Engine Setup

```go
package mytests

import (
	"log/slog"
	"os"
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest"
)

var NewEngine func(t testing.TB) *puppetest.Engine

func TestMain(m *testing.M) {
	factory, err := puppetest.NewEngineFactory()
	if err != nil {
		slog.Error("failed to create engine factory", slog.String("error", err.Error()))
		os.Exit(1)
	}

	NewEngine = factory.NewEngine
	code := m.Run()

	if err = factory.Close(); err != nil {
		slog.Error("failed to close engine factory", slog.String("error", err.Error()))
		os.Exit(1)
	}

	os.Exit(code)
}
```

## First Test

```go
func TestSmoke(t *testing.T) {
	engine := NewEngine(t)
	if engine.BaseURL() != "" {
		t.Fatal("expected no server in this setup")
	}
}
```

## Add an HTTP Server Extension

```go
factory, err := puppetest.NewEngineFactory(
	puppetest.WithExtensions(
		puppetest.WithTestServer(myHandler()),
	),
)
if err != nil {
	panic(err)
}
```

Then use `engine.BaseURL()` from your tests.

## Add a Database Connection

```go
factory, err := puppetest.NewEngineFactory(
	puppetest.WithConnectionFactory(mySQLPerformer, true),
)
if err != nil {
	panic(err)
}
```

`mySQLPerformer` must match `puppetest.ConnectionPerformer`.

## Next Steps

- Read [Engine And Factory](/core-concepts/engine-and-factory/)
- Read [Hooks](/core-concepts/hooks/)
- Read [Runners](/runners/)
- Read [Providers](/providers/)
