package testkit

import (
	"database/sql/driver"
	"io"
)

type stubSQLRows struct {
	columns []string
	rows    [][]driver.Value
	cursor  int
}

var _ driver.Rows = (*stubSQLRows)(nil)

func (r *stubSQLRows) Columns() []string {
	return r.columns
}

func (r *stubSQLRows) Close() error {
	r.cursor = len(r.rows)
	return nil
}

func (r *stubSQLRows) Next(dest []driver.Value) error {
	if r.cursor >= len(r.rows) {
		return io.EOF
	}

	copy(dest, r.rows[r.cursor])
	r.cursor++

	return nil
}
