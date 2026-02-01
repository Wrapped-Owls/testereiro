package dbrunner

import (
	"database/sql"
	"fmt"
	"testing"
)

// ExpectCount is a simple matcher to check the number of rows.
func ExpectCount(expectedAmount int, countRows bool) Option {
	return func(r RunnerModifier) {
		r.AddValidator(&countValidator{expected: expectedAmount, scanCount: !countRows})
	}
}

type countValidator struct {
	expected  int
	scanCount bool
}

func (v *countValidator) scanCountColumn(rows *sql.Rows) (count int, err error) {
	if err = rows.Scan(&count); err != nil {
		return -1, fmt.Errorf("failed to scan count: %w", err)
	}

	return count, nil
}

func (v *countValidator) countRows(rows *sql.Rows) int {
	var count int
	for rows.Next() {
		count++
	}

	return count
}

func (v *countValidator) Validate(t testing.TB, rows *sql.Rows) error {
	var count int
	if v.scanCount {
		var err error
		if count, err = v.scanCountColumn(rows); err != nil {
			t.Fatalf("failed to scan count: %s", err.Error())
			return err
		}
	} else {
		count = v.countRows(rows)
	}

	if count != v.expected {
		t.Fatalf("expected count %d, got %d", v.expected, count)
	}

	return nil
}
