---
title: "Testereiro"
weight: 1
menu:
  main:
    identifier: home
    weight: 1
---

Testereiro is a composable test engine toolkit for Go.

It gives you a deterministic lifecycle for integration tests:

- build an `EngineFactory`
- create an isolated `Engine` per test
- attach infrastructure (DB, HTTP server, providers)
- run assertions through reusable runners

## Packages

- Core: `github.com/wrapped-owls/testereiro/puppetest`
- Mongo provider: `github.com/wrapped-owls/testereiro/providers/mongotestage`
- SQL assertion helpers: `github.com/wrapped-owls/testereiro/providers/siqeltestage`

## What You Can Do

- Run HTTP assertions with `netoche` (`puppetest/pkg/atores/netoche`)
- Run SQL assertions with `bancoche` (`puppetest/pkg/atores/bancoche`)
- Seed SQL structs with `engine.Seed(...)`
- Seed provider-backed systems with `engine.SeedWithProvider(...)`
- Inject and teardown typed providers per engine or per factory
- Hook into create/run/seed/teardown/close lifecycle events

Use the sidebar to start with setup, then core concepts, runners, providers, and real examples.
