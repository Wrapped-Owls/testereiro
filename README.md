# Testereiro

Composable integration-test toolkit for Go.

Testereiro gives you a deterministic test lifecycle built around a factory/engine model:

- create one `EngineFactory` for a test package
- create one isolated `Engine` per test
- attach infra (DB, HTTP server, typed providers)
- execute reusable assertions with runners
- teardown automatically through lifecycle hooks

## Packages

- Core: `github.com/wrapped-owls/testereiro/puppetest`
- Mongo provider: `github.com/wrapped-owls/testereiro/providers/mongotestage`
- SQL assertion helpers: `github.com/wrapped-owls/testereiro/providers/siqeltestage`

## Requirements

- Go `1.24+`

## Install

```bash
go get github.com/wrapped-owls/testereiro/puppetest
```

## Quick Start

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

## Built-in Runners

- `netoche` (HTTP): request building, execution, status/body validation
- `bancoche` (SQL): query execution with row/count/custom validators

## Provider Model

- engine-level providers: attach typed resources to one test engine
- factory-level providers: shared resources with bind + teardown lifecycle
- provider seeding: `engine.SeedWithProvider(...)` for non-SQL setup

## Example Modules

- `examples/webapi`: HTTP-only tests
- `examples/sqlite`: in-memory SQLite + migration + HTTP
- `examples/testcontainers_mysql`: MySQL via testcontainers
- `examples/mongodb_assert`: Mongo via testcontainers + provider-backed checks

## Common Commands

From repository root:

```bash
make test
make test-race
make run-formatters
make run-lint
make examples
make docs-build
make docs-serve
```

## Documentation

Project docs live under `docs/` and cover:

- getting started
- core lifecycle and hooks
- runners (`netoche`, `bancoche`)
- providers (`mongotestage`, `siqeltestage`)
- examples and roadmap
