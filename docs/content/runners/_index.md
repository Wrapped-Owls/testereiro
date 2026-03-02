---
title: "Runners"
weight: 4
menu:
  main:
    weight: 4
---

Runners encapsulate assertions and workflows executed through `engine.Execute(t, runner)`.

Main runner packages:

- `netoche`: HTTP request/response validation
- `bancoche`: SQL query/result validation
- `mongochecker` (provider module): Mongo query/result validation

You can also compose runners with `atores.MultiRunner`.
