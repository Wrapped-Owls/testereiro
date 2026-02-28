package siqeltest

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
)

type testRecord struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

func TestExpectValidator_Validate(t *testing.T) {
	tests := []struct {
		name      string
		rows      [][]driver.Value
		validator *expectValidator[testRecord]
		wantErr   string
		assertion func(*testing.T, bool)
	}{
		{
			name: "success",
			rows: [][]driver.Value{{int64(1), "Optimus"}},
			validator: &expectValidator[testRecord]{
				expected:   testRecord{ID: 1, Name: "Optimus"},
				comparator: defaultComparator[testRecord],
			},
		},
		{
			name: "no records",
			rows: [][]driver.Value{},
			validator: &expectValidator[testRecord]{
				expected: testRecord{ID: 1},
			},
			wantErr: "no records found",
		},
		{
			name: "sanitizer error",
			rows: [][]driver.Value{{int64(1), "Optimus"}},
			validator: &expectValidator[testRecord]{
				expected: testRecord{ID: 1, Name: "Optimus"},
				sanitizer: func(expected, actual *testRecord) error {
					return errors.New("sanitize failed")
				},
				comparator: defaultComparator[testRecord],
			},
			wantErr: "failed to sanitize database result",
		},
		{
			name: "custom comparator false",
			rows: [][]driver.Value{{int64(1), "Optimus"}},
			validator: &expectValidator[testRecord]{
				expected: testRecord{ID: 1, Name: "Optimus"},
				comparator: func(testing.TB, testRecord, testRecord) bool {
					return false
				},
			},
			wantErr: "database result did not match expected value",
		},
		{
			name: "sanitizer adjusts values",
			rows: [][]driver.Value{{int64(1), "Optimus"}},
			validator: &expectValidator[testRecord]{
				expected: testRecord{ID: 1, Name: "Any"},
				sanitizer: func(expected, actual *testRecord) error {
					expected.Name = "normalized"
					actual.Name = "normalized"
					return nil
				},
				comparator: defaultComparator[testRecord],
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rows := openRows(t, []string{"id", "name"}, tc.rows)
			err := tc.validator.Validate(t, rows)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !contains(err.Error(), tc.wantErr) {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestExpectValidator_SelectionFields(t *testing.T) {
	validator := &expectValidator[testRecord]{}
	if got := validator.SelectionFields(); got != "*" {
		t.Fatalf("unexpected selection fields: %q", got)
	}
}

func TestDefaultComparator(t *testing.T) {
	if !defaultComparator(t, testRecord{ID: 1}, testRecord{ID: 1}) {
		t.Fatal("expected comparator to return true for equal values")
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

var registerDriverOnce sync.Once

func openRows(t *testing.T, columns []string, data [][]driver.Value) *sql.Rows {
	t.Helper()
	registerDriverOnce.Do(func() {
		sql.Register("siqeltest-driver", &queryDriver{})
	})

	dsn := addFixture(columns, data)
	db, err := sql.Open("siqeltest-driver", dsn)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	rows, err := db.QueryContext(context.Background(), "SELECT 1")
	if err != nil {
		t.Fatalf("failed to query rows: %v", err)
	}
	t.Cleanup(func() { _ = rows.Close() })
	return rows
}

type fixture struct {
	columns []string
	data    [][]driver.Value
}

var (
	fixturesMu sync.Mutex
	fixtures   = map[string]fixture{}
)

func addFixture(columns []string, data [][]driver.Value) string {
	fixturesMu.Lock()
	defer fixturesMu.Unlock()
	id := fmt.Sprintf("fixture-%d", len(fixtures)+1)
	fixtures[id] = fixture{columns: columns, data: data}
	return id
}

func getFixture(dsn string) fixture {
	fixturesMu.Lock()
	defer fixturesMu.Unlock()
	return fixtures[dsn]
}

type queryDriver struct{}

func (d *queryDriver) Open(name string) (driver.Conn, error) {
	f := getFixture(name)
	return &queryConn{fixture: f}, nil
}

type queryConn struct {
	fixture fixture
}

func (c *queryConn) Prepare(
	string,
) (driver.Stmt, error) {
	return nil, errors.New("not implemented")
}
func (c *queryConn) Close() error { return nil }

func (c *queryConn) Begin() (driver.Tx, error) { return nil, errors.New("not implemented") }

func (c *queryConn) QueryContext(
	context.Context,
	string,
	[]driver.NamedValue,
) (driver.Rows, error) {
	return &queryRows{columns: c.fixture.columns, data: c.fixture.data}, nil
}

type queryRows struct {
	columns []string
	data    [][]driver.Value
	index   int
}

func (r *queryRows) Columns() []string { return r.columns }
func (r *queryRows) Close() error      { return nil }

func (r *queryRows) Next(dest []driver.Value) error {
	if r.index >= len(r.data) {
		return io.EOF
	}
	row := r.data[r.index]
	r.index++
	for i := range row {
		dest[i] = row[i]
	}
	return nil
}
