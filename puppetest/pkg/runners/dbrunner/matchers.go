package dbrunner

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest/internal/stgctx"
)

// WithQuery sets the from and filters for the DB runner.
func WithQuery(table string, filters map[string]any) Option {
	return func(r RunnerModifier) {
		r.SetFrom(table)
		r.AddContextFilter(func(_ stgctx.RunnerContext) (map[string]any, error) {
			return filters, nil
		})
	}
}

// WithSubsequentQuery sets the from and a late-bound filter builder.
func WithSubsequentQuery[V any](table string, fn func(val V) map[string]any) Option {
	return func(r RunnerModifier) {
		lateFilter := func(ctx stgctx.RunnerContext) (map[string]any, error) {
			val, ok := stgctx.LoadFromCtx[V](ctx)
			if !ok {
				var zero V
				return nil, fmt.Errorf("failed to load value of type %T from storage", zero)
			}
			return fn(val), nil
		}
		r.SetFrom(table)
		r.AddContextFilter(lateFilter)
	}
}

// ExpectCount is a simple matcher to check the number of rows.
func ExpectCount(expectedAmount int) Option {
	return func(r RunnerModifier) {
		r.AddValidator(&countValidator{expected: expectedAmount})
	}
}

type countValidator struct {
	expected int
}

func (v *countValidator) SelectionFields() string {
	return "COUNT(*)"
}

func (v *countValidator) Validate(t testing.TB, rows *sql.Rows) error {
	if !rows.Next() {
		return fmt.Errorf("no records found")
	}

	var count int
	if err := rows.Scan(&count); err != nil {
		return fmt.Errorf("failed to scan count: %w", err)
	}

	if count != v.expected {
		t.Fatalf("expected count %d, got %d", v.expected, count)
	}

	return nil
}

// WithCustomValidation allows a user-defined validation function on the result rows.
func WithCustomValidation(
	selection string,
	validate func(t testing.TB, rows *sql.Rows) error,
) Option {
	return func(r RunnerModifier) {
		r.AddValidator(&customValidator{selection: selection, validate: validate})
	}
}

type customValidator struct {
	selection string
	validate  func(t testing.TB, rows *sql.Rows) error
}

func (v *customValidator) SelectionFields() string {
	if v.selection == "" {
		return "*"
	}
	return v.selection
}

func (v *customValidator) Validate(t testing.TB, rows *sql.Rows) error {
	return v.validate(t, rows)
}
