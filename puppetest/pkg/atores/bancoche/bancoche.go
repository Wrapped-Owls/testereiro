// Package bancoche (Banco + Fantoche) is a database puppet for validating state and executing queries.
package bancoche

import (
	"database/sql"
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest/internal/stgctx"
)

type (
	// QueryBuilder is an interface for building SQL queries.
	QueryBuilder interface {
		Build(ctx stgctx.RunnerContext) (query string, args []any, err error)
	}

	dbValidator interface {
		Validate(t testing.TB, rows *sql.Rows) error
	}
	filterFromContext func(stgctx.RunnerContext) (map[string]any, error)
)

// DbRunner is a test runner for database assertions.
type DbRunner struct {
	db           *sql.DB
	queryBuilder QueryBuilder
	validators   []dbValidator
}

type (
	// RunnerModifier receives option-driven mutations for a DbRunner.
	RunnerModifier interface {
		// SetQueryBuilder sets the query builder used by the runner.
		SetQueryBuilder(QueryBuilder)
		// AddValidator appends a database validator to the runner.
		AddValidator(validator dbValidator)
	}
	// Option is a functional option for configuring the DbRunner.
	Option func(RunnerModifier)
)

// New creates a new DbRunner with the given options.
func New(db *sql.DB, opts ...Option) *DbRunner {
	r := &DbRunner{
		db: db,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// AddValidator appends a validator to be executed during Run.
func (r *DbRunner) AddValidator(validator dbValidator) {
	r.validators = append(r.validators, validator)
}

// SetQueryBuilder replaces the query builder used by Run.
func (r *DbRunner) SetQueryBuilder(qb QueryBuilder) {
	r.queryBuilder = qb
}

// Run executes the configured query and validators.
func (r *DbRunner) Run(t testing.TB, rCtx stgctx.RunnerContext) error {
	for _, v := range r.validators {
		query, args, err := r.queryBuilder.Build(rCtx)
		if err != nil {
			t.Fatalf("Failed to build query: %v", err)
			return err
		}

		var rows *sql.Rows
		if rows, err = r.db.Query(query, args...); err != nil {
			t.Fatalf("Failed to execute query %q: %v", query, err)
			return err
		}

		if err = v.Validate(t, rows); err != nil {
			t.Fatalf("Database validation failed: %v", err)
		}
		_ = rows.Close()
	}
	return nil
}
