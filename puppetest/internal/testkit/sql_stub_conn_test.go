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
	pingErr := errors.New("ping failed")
	tests := []struct {
		name    string
		state   *SQLState
		wantErr error
	}{
		{name: "nil state returns nil", state: nil, wantErr: nil},
		{name: "returns state ping error", state: &SQLState{PingErr: pingErr}, wantErr: pingErr},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := (&stubSQLConn{state: tt.state}).Ping(context.Background())
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected ping error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestStubSQLConn_ExecContext(t *testing.T) {
	execErr := errors.New("exec failed")
	tests := []struct {
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := (&stubSQLConn{state: tt.state}).ExecContext(
				context.Background(),
				tt.query,
				[]driver.NamedValue{},
			)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected exec error %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr == nil {
				rows, rowsErr := result.RowsAffected()
				if rowsErr != nil {
					t.Fatalf("rows affected error: %v", rowsErr)
				}
				if rows != tt.wantRows {
					t.Fatalf("expected rows affected %d, got %d", tt.wantRows, rows)
				}
			}

			if tt.state != nil {
				if got := tt.state.ExecStatements(); !reflect.DeepEqual(got, tt.wantRecorded) {
					t.Fatalf("expected recorded queries %v, got %v", tt.wantRecorded, got)
				}
			}
		})
	}
}

func TestStubSQLConn_PrepareBeginAndClose(t *testing.T) {
	tests := []struct {
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := &stubSQLConn{state: tt.state}

			stmt, err := conn.Prepare("SELECT 1")
			if err != nil {
				t.Fatalf("prepare error: %v", err)
			}
			if stmt == nil {
				t.Fatalf("expected non-nil stmt")
			}

			_, beginErr := conn.Begin()
			if beginErr == nil || !strings.Contains(beginErr.Error(), tt.wantBeginErr) {
				t.Fatalf("expected begin error containing %q, got %v", tt.wantBeginErr, beginErr)
			}

			if closeErr := conn.Close(); closeErr != nil {
				t.Fatalf("close error: %v", closeErr)
			}

			if tt.state != nil {
				if got := tt.state.CloseCount(); got != tt.wantCloseCnt {
					t.Fatalf("expected close count %d, got %d", tt.wantCloseCnt, got)
				}
			}
		})
	}
}
