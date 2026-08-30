---
title: "Database Lifecycle"
weight: 3
---

## What It Controls

`WithConnectionFactory` says *how* to open a database connection. `WithDatabaseLifecycle` says
*whether and how* to create a fresh database for each test and drop it afterwards.

```go
factory, err := puppetest.NewEngineFactory(
	puppetest.WithConnectionFactory(myPerformer),
	puppetest.WithDatabaseLifecycle(puppetest.PostgresLifecycle),
)
```

Without `WithDatabaseLifecycle` nothing creates a database, but the per-test name is still handed
to your performer as `conf.DBName`. Reuse is therefore your performer's decision: ignore `DBName`
to keep one database, or build a DSN that brings it into being on connect, as the in-memory SQLite
example does. Option order does not matter, because the lifecycle is built after all options have
run, over the root connection `WithConnectionFactory` opened.

The root connection must target a database *other* than the ones being created and dropped:
`postgres` or `template1` on Postgres, any existing schema on MySQL.

## Shipped Lifecycles

| Builder                         | Engine     | Notes                                                                |
|---------------------------------|------------|----------------------------------------------------------------------|
| `puppetest.MySQLLifecycle`      | MySQL      | `CREATE/DROP DATABASE IF [NOT] EXISTS`, backtick-quoted               |
| `puppetest.PostgresLifecycle`   | PostgreSQL | `pg_database` pre-check, terminates live backends before dropping     |
| `puppetest.SQLiteLifecycle(fn)` | SQLite     | File lifecycle, since SQLite has no `CREATE DATABASE`                 |

### Postgres

PostgreSQL has no `CREATE DATABASE ... IF NOT EXISTS`, so `PostgresLifecycle` queries
`pg_database` first and treats a `42P04` (`duplicate_database`) failure as success when a
concurrent creator wins the race. On teardown it terminates the sessions still attached before
dropping, because PostgreSQL refuses to drop a database holding live connections;
`pg_terminate_backend` is used rather than `DROP DATABASE ... WITH (FORCE)` so servers older than
13 keep working. Neither statement runs inside a transaction, which PostgreSQL rejects for
`CREATE DATABASE` and `DROP DATABASE`.

#### Concurrent creation depends on your driver

**On lib/pq, two tests creating the same database name at the same time are not protected, and one
of them fails.** The `pg_database` check and the `CREATE DATABASE` that follows it are separate
statements, so parallel callers can both pass the check. Only the `42P04` SQLSTATE of the losing
create closes that window, and puppetest reads it through an `interface{ SQLState() string }`:
pgx's `*pgconn.PgError` satisfies that, lib/pq's `*pq.Error` does not.

For a harness that provisions databases this is the main path rather than an edge case, since
`go test -p N` is the load it is built for. If you run parallel packages that can produce the same
test name, use pgx or give each package a distinct database name prefix.

### SQLite

A SQLite database is a file, so the lifecycle takes a resolver mapping the test database name to
its path. `Create` prepares the directory; `Drop` removes the file along with the `-wal` and
`-shm` sidecars, which otherwise leak between tests.

```go
puppetest.WithDatabaseLifecycle(
	puppetest.SQLiteLifecycle(func(dbName string) string {
		return filepath.Join(tempDir, dbName+".db")
	}),
)
```

Returning an empty path marks the database as memory-backed, leaving nothing to create or remove.

## Writing Your Own

`puppetest.DBLifecycle` is a two-method interface, so any engine can be supported without
changing the library:

```go
type DBLifecycle interface {
	Create(ctx context.Context, dbName string) error
	Drop(ctx context.Context, dbName string) error
}
```

Pass it through a `puppetest.DBLifecycleBuilder`, which receives the factory-owned root
connection:

```go
puppetest.WithDatabaseLifecycle(func(rootDB *sql.DB) puppetest.DBLifecycle {
	return mssqlLifecycle{rootDB: rootDB}
})
```

Quote identifiers with `puppetest.QuoteIdentifier` rather than `fmt.Sprintf`. A database name
reaches a lifecycle from the caller, and doubling the quote rune is what stops a crafted name from
closing the identifier and appending its own statement:

```go
quoted, err := puppetest.QuoteIdentifier(dbName, '"')  // your engine's quote rune
```

`QuoteIdentifier` assumes a symmetric delimiter escaped by doubling, which covers MySQL's
backtick and the double quote used by PostgreSQL, SQLite and the SQL standard. An engine with
asymmetric delimiters, such as SQL Server's `[name]`, needs its own escaping.

Normalizing and escaping are not the same thing. puppetest normalizes the database names it
generates itself, but a name reaching your lifecycle from a caller has had nothing done to it.

## Placeholder Style

`Engine.Seed` generates its own `INSERT` statements, and the bind marker is dialect-specific: MySQL
and SQLite use `?`, PostgreSQL uses `$1`, `$2`. `WithPlaceholderStyle` selects it, defaulting to `?`
so MySQL and SQLite consumers need no change.

```go
factory, err := puppetest.NewEngineFactory(
	puppetest.WithConnectionFactory(pgPerformer),
	puppetest.WithDatabaseLifecycle(puppetest.PostgresLifecycle),
	puppetest.WithPlaceholderStyle(puppetest.OrdinalPlaceholder),
)
```

The `bancoche` runner builds its own `WHERE` clauses, so it takes the same setting separately. It is
constructed from a `*sql.DB` rather than from an `Engine`, so the factory option cannot reach it:

```go
bancoche.NewMapQuery("jokers", filters,
	bancoche.WithPlaceholderStyle(puppetest.OrdinalPlaceholder),
)
```

A custom style is any `func(index int) string`, where `index` counts from one.

### Generate, never rewrite

`NewRawQuery` takes no placeholder setting, and that is deliberate. It draws the line the whole
package follows:

> puppetest picks the dialect for SQL it **constructs**, and never touches SQL you **wrote**.

`Engine.Seed` and `NewMapQuery` build their statements from a struct or a filter map, so they know
how many markers exist and where each one goes; emitting `$1` instead of `?` is a choice made while
assembling the string. `NewRawQuery` receives finished SQL, so adapting it would mean parsing it and
guessing at intent, and a rewriter has to know that a `?` inside a string literal or a comment is
not a bind marker. That is a class of bug this package declines to own.

For a contributor the rule reads as: a builder that assembles a statement may know about dialects; a
builder that receives one may not. If you add a query builder, that tells you which half it is in.
