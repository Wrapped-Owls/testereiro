package dbrunner

import (
	"database/sql"
	"fmt"
	"maps"
	"strings"
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest/internal/stgctx"
)

type (
	dbValidator interface {
		SelectionFields() string
		Validate(t testing.TB, rows *sql.Rows) error
	}
	filterFromContext func(stgctx.RunnerContext) (map[string]any, error)
)

// DbRunner is a test runner for database assertions.
type DbRunner struct {
	db          *sql.DB
	from        string
	lateFilters []filterFromContext
	validators  []dbValidator
}

type (
	RunnerModifier interface {
		SetFrom(string)
		AddValidator(validator dbValidator)
		AddContextFilter(filterFromContext)
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

func (r *DbRunner) Run(t testing.TB, rCtx stgctx.RunnerContext) error {
	var filters map[string]any
	for _, filterResolver := range r.lateFilters {
		newFilter, err := filterResolver(rCtx)
		if err != nil {
			t.Fatalf("Failed to build late filters: %v", err)
		}
		maps.Copy(filters, newFilter)
	}

	var (
		where = make([]string, 0, len(filters))
		args  = make([]any, 0, len(filters))
	)
	for filterKey, filterValue := range filters {
		where = append(where, fmt.Sprintf("%s = ?", filterKey))
		args = append(args, filterValue)
	}
	for _, v := range r.validators {
		query := fmt.Sprintf(
			"SELECT %s FROM %s WHERE %s",
			v.SelectionFields(), r.from, strings.Join(where, " AND "),
		)

		rows, err := r.db.Query(query, args...)
		if err != nil {
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

func (r *DbRunner) AddValidator(validator dbValidator) {
	r.validators = append(r.validators, validator)
}

func (r *DbRunner) AddContextFilter(newFilter filterFromContext) {
	r.lateFilters = append(r.lateFilters, newFilter)
}

func (r *DbRunner) SetFrom(s string) {
	r.from = s
}
