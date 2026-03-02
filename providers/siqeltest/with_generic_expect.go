package siqeltest

import (
	"database/sql"
	"fmt"
	"reflect"
	"testing"

	"github.com/vinovest/sqlx"

	"github.com/wrapped-owls/testereiro/puppetest/pkg/atores/bancoche"
)

type (
	DbSanitizer[O any]  func(expected, actual *O) error
	DbComparator[O any] func(t testing.TB, expected, actual O) bool
)

// WithExpect adds a validator that queries the database and compares the result with the expected object.
func WithExpect[O any](expected O, sanitizer ...DbSanitizer[O]) bancoche.Option {
	var selectedSanitizer DbSanitizer[O]
	if len(sanitizer) > 0 {
		selectedSanitizer = sanitizer[0]
	}
	return func(modifier bancoche.RunnerModifier) {
		modifier.AddValidator(&expectValidator[O]{
			expected:   expected,
			sanitizer:  selectedSanitizer,
			comparator: defaultComparator[O],
		})
	}
}

func WithExpectWithComparator[O any](expected O, comparator DbComparator[O]) bancoche.Option {
	return func(modifier bancoche.RunnerModifier) {
		modifier.AddValidator(&expectValidator[O]{
			expected:   expected,
			comparator: comparator,
		})
	}
}

type expectValidator[O any] struct {
	expected   O
	sanitizer  DbSanitizer[O]
	comparator DbComparator[O]
}

func (v *expectValidator[O]) SelectionFields() string {
	return "*"
}

func (v *expectValidator[O]) Validate(t testing.TB, rows *sql.Rows) error {
	var destinations []O
	if err := sqlx.StructScan(rows, &destinations); err != nil {
		return err
	}
	if len(destinations) == 0 {
		return fmt.Errorf("no records found")
	}
	destination := destinations[0]

	expected := v.expected
	if v.sanitizer != nil {
		if err := v.sanitizer(&expected, &destination); err != nil {
			return fmt.Errorf("failed to sanitize database result: %w", err)
		}
	}

	if v.comparator != nil && !v.comparator(t, expected, destination) {
		return fmt.Errorf("database result did not match expected value")
	}
	return nil
}

func defaultComparator[O any](t testing.TB, expected, actual O) bool {
	if reflect.DeepEqual(expected, actual) {
		return true
	}

	t.Errorf("database result mismatch: expected=%#v actual=%#v", expected, actual)
	return false
}
