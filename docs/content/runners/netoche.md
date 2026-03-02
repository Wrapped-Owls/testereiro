---
title: "Netoche (HTTP Runner)"
weight: 1
---

Package: `github.com/wrapped-owls/testereiro/puppetest/pkg/atores/netoche`

`netoche` builds HTTP requests, executes them, and validates status/body.

## Basic Usage

```go
runner := netoche.New(
	engine.BaseURL(),
	netoche.WithRequest(http.MethodGet, "/games", netoche.NoBody{}),
	netoche.ExpectStatus(http.StatusOK),
	netoche.ExpectBody(expectedGames),
)

err := engine.Execute(t, runner)
```

## Request Configuration

- `WithRequest(method, path, body)` for static request
- `WithSubsequentRequest` for request bodies generated from context
- `WithHeader`, `WithHeaderFromCtx`
- `WithPathParam`, `WithPathParamFromCtx`
- `WithRequestModifier` for custom request transformations

## Validation And Context Chaining

- `ExpectStatus(code)`
- `ExpectBody(expected)`
- `ExpectBodyWithComparator(expected, comparator)`
- `ExtractToState(extractor)` to save derived values for downstream steps

`ExpectBody` decodes JSON and stores decoded body on the runner context automatically.
