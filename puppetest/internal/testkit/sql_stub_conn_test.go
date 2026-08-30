package testkit

import (
	"context"
	"database/sql/driver"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestStubSQLConn_Ping(t *testing.T) {
	t.Parallel()

	pingErr := errors.New("ping failed")
	testCases := []struct {
		name    string
		state   *SQLState
		wantErr error
	}{
		{name: "nil state returns nil", state: nil, wantErr: nil},
		{name: "returns state ping error", state: &SQLState{PingErr: pingErr}, wantErr: pingErr},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := (&stubSQLConn{state: testCase.state}).Ping(context.Background())
			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("expected ping error %v, got %v", testCase.wantErr, err)
			}
		})
	}
}

func TestStubSQLConn_ExecContext(t *testing.T) {
	t.Parallel()

	execErr := errors.New("exec failed")
	testCases := []struct {
		name         string
		state        *SQLState
		query        string
		wantErr      error
		wantRows     int64
		wantRecorded []string
	}{
		{
			name:         "records query and returns rows affected",
			state:        &SQLState{},
			query:        "SELECT 1",
			wantRows:     1,
			wantRecorded: []string{"SELECT 1"},
		},
		{
			name:         "records query and returns exec error",
			state:        &SQLState{ExecErr: execErr},
			query:        "BROKEN",
			wantErr:      execErr,
			wantRecorded: []string{"BROKEN"},
		},
		{
			name:         "nil state still returns rows affected",
			state:        nil,
			query:        "SELECT 2",
			wantRows:     1,
			wantRecorded: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result, err := (&stubSQLConn{state: testCase.state}).ExecContext(
				context.Background(),
				testCase.query,
				[]driver.NamedValue{},
			)
			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("expected exec error %v, got %v", testCase.wantErr, err)
			}
			if testCase.wantErr == nil {
				rows, rowsErr := result.RowsAffected()
				if rowsErr != nil {
					t.Fatalf("rows affected error: %v", rowsErr)
				}
				if rows != testCase.wantRows {
					t.Fatalf("expected rows affected %d, got %d", testCase.wantRows, rows)
				}
			}

			if testCase.state != nil {
				if got := testCase.state.ExecStatements(); !reflect.DeepEqual(got, testCase.wantRecorded) {
					t.Fatalf("expected recorded queries %v, got %v", testCase.wantRecorded, got)
				}
			}
		})
	}
}

func TestStubSQLConn_PrepareBeginAndClose(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		state        *SQLState
		wantCloseCnt int
		wantBeginErr string
	}{
		{
			name:         "with state increments close count",
			state:        &SQLState{},
			wantCloseCnt: 1,
			wantBeginErr: "transactions not supported",
		},
		{
			name:         "nil state close is no-op",
			state:        nil,
			wantCloseCnt: 0,
			wantBeginErr: "transactions not supported",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			conn := &stubSQLConn{state: testCase.state}

			stmt, err := conn.Prepare("SELECT 1")
			if err != nil {
				t.Fatalf("prepare error: %v", err)
			}
			if stmt == nil {
				t.Fatalf("expected non-nil stmt")
			}

			_, beginErr := conn.Begin()
			if beginErr == nil || !strings.Contains(beginErr.Error(), testCase.wantBeginErr) {
				t.Fatalf("expected begin error containing %q, got %v", testCase.wantBeginErr, beginErr)
			}

			if closeErr := conn.Close(); closeErr != nil {
				t.Fatalf("close error: %v", closeErr)
			}

			if testCase.state != nil {
				if got := testCase.state.CloseCount(); got != testCase.wantCloseCnt {
					t.Fatalf("expected close count %d, got %d", testCase.wantCloseCnt, got)
				}
			}
		})
	}
}

func TestStubSQLConn_QueryContext(t *testing.T) {
	t.Parallel()

	queryErr := errors.New("query failed")
	testCases := []struct {
		name         string
		state        *SQLState
		query        string
		args         []driver.NamedValue
		wantErr      error
		wantCols     []string
		wantRecorded []RecordedQuery
	}{
		{
			name: "records query with bound args and returns canned rows",
			state: &SQLState{
				QueryCols: []string{"exists"},
				QueryRows: [][]driver.Value{{true}},
			},
			query:    "SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)",
			args:     []driver.NamedValue{{Ordinal: 1, Value: "games"}},
			wantCols: []string{"exists"},
			wantRecorded: []RecordedQuery{
				{
					Statement: "SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)",
					Args:      []driver.Value{"games"},
				},
			},
		},
		{
			name:    "records query and returns query error",
			state:   &SQLState{QueryErr: queryErr},
			query:   "BROKEN",
			wantErr: queryErr,
			wantRecorded: []RecordedQuery{
				{Statement: "BROKEN", Args: []driver.Value{}},
			},
		},
		{
			name:  "nil state returns empty rows",
			state: nil,
			query: "SELECT 1",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			rows, err := (&stubSQLConn{state: testCase.state}).QueryContext(
				context.Background(),
				testCase.query,
				testCase.args,
			)
			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("expected query error %v, got %v", testCase.wantErr, err)
			}
			if testCase.wantErr == nil && !reflect.DeepEqual(rows.Columns(), testCase.wantCols) {
				t.Fatalf("expected columns %v, got %v", testCase.wantCols, rows.Columns())
			}

			if testCase.state == nil {
				return
			}
			if got := testCase.state.QueryCalls(); !reflect.DeepEqual(got, testCase.wantRecorded) {
				t.Fatalf("expected recorded queries %v, got %v", testCase.wantRecorded, got)
			}
		})
	}
}
