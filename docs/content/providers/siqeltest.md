---
title: "Siqeltestage"
weight: 2
---

Module: `github.com/wrapped-owls/testereiro/providers/siqeltestage`

`siqeltestage` adds typed object comparison validators for `bancoche` SQL assertions.

## Install

```bash
go get github.com/wrapped-owls/testereiro/providers/siqeltestage
```

## Usage

```go
runner := bancoche.New(
	engine.DB(),
	bancoche.WithMapQuery("games", map[string]any{"id": 1}),
	siqeltestage.WithExpect(expectedGame),
)

err := engine.Execute(t, runner)
```

## Custom Comparator

```go
siqeltestage.WithExpectWithComparator(expected, func(t testing.TB, expected, actual MyType) bool {
	return reflect.DeepEqual(expected, actual)
})
```

You can also pass a sanitizer to normalize values before comparison.
