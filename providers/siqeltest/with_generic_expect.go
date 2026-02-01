package siqeltest

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vinovest/sqlx"

	"github.com/wrapped-owls/testereiro/puppetest/pkg/runners/dbrunner"
)

type DbSanitizer[O any] func(expected, actual *O) error

// WithExpect adds a validator that queries the database and compares the result with the expected object.
func WithExpect[O any](expected O, sanitizer DbSanitizer[O]) dbrunner.Option {
	return func(modifier dbrunner.RunnerModifier) {
		modifier.AddValidator(&expectValidator[O]{
			expected:  expected,
			sanitizer: sanitizer,
		})
	}
}

type expectValidator[O any] struct {
	expected  O
	sanitizer DbSanitizer[O]
}

func (v *expectValidator[O]) SelectionFields() string {
	return "*"
}

func (v *expectValidator[O]) Validate(t testing.TB, rows *sql.Rows) error {
	if !rows.Next() {
		return fmt.Errorf("no records found")
	}

	var destination O
	if err := sqlx.StructScan(rows, &destination); err != nil {
		return err
	}

	assert.Equal(t, v.expected, destination)
	return nil
}
