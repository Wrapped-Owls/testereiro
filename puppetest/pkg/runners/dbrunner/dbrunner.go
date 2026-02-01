package dbrunner

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
	RunnerModifier interface {
		SetQueryBuilder(QueryBuilder)
		AddValidator(validator dbValidator)
	}
	// Option is a functional option for configuring the DbRunner.
	Option func(RunnerModifier)
)

// NewDbRunner creates a new DbRunner with the given options.
func NewDbRunner(db *sql.DB, opts ...Option) *DbRunner {
	r := &DbRunner{
		db: db,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

func (r *DbRunner) AddValidator(validator dbValidator) {
	r.validators = append(r.validators, validator)
}

func (r *DbRunner) SetQueryBuilder(qb QueryBuilder) {
	r.queryBuilder = qb
}

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
