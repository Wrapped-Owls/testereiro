---
title: "Project Structure"
weight: 7
menu:
  main:
    weight: 7
---

## Repository Layout

- `puppetest/`: core engine, lifecycle, and built-in runners
- `providers/`: integration modules (`mongotestage`, `siqeltestage`)
- `examples/`: usage examples and integration-style tests
- `docs/`: Hugo documentation site

## Core Subpackages

Inside `puppetest/pkg/atores`:

- `netoche`: HTTP runner
- `bancoche`: SQL runner

Utility package:

- `puppetest/pkg/strnormalizer`

## Build And Test Commands

From repository root:

```bash
make test
make test-race
make run-formatters
make run-lint
make install-tools
make examples
make docs-build
make docs-serve
make docs-theme
```
