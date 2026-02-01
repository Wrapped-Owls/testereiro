package dbrunner

import (
	"database/sql"
	"testing"
)

// WithCustomValidation allows a user-defined validation function on the result rows.
func WithCustomValidation(
	validate func(t testing.TB, rows *sql.Rows) error,
) Option {
	return func(r RunnerModifier) {
		r.AddValidator(&customValidator{validate: validate})
	}
}

type customValidator struct {
	validate func(t testing.TB, rows *sql.Rows) error
}

func (v *customValidator) Validate(t testing.TB, rows *sql.Rows) error {
	return v.validate(t, rows)
}
