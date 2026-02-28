package testkit

import (
	"reflect"
	"strings"
	"testing"
)

func TestStubSQLStmt(t *testing.T) {
	tests := []struct {
		name         string
		wantNumInput int
		wantExecRows int64
		wantQueryErr string
		wantCloseErr bool
	}{
		{
			name:         "stmt defaults",
			wantNumInput: -1,
			wantExecRows: 1,
			wantQueryErr: "queries not supported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stmt := stubSQLStmt{}

			closeErr := stmt.Close()
			if tt.wantCloseErr && closeErr == nil {
				t.Fatalf("expected close error, got nil")
			}
			if !tt.wantCloseErr && closeErr != nil {
				t.Fatalf("expected nil close error, got %v", closeErr)
			}

			if got := stmt.NumInput(); got != tt.wantNumInput {
				t.Fatalf("expected NumInput=%d, got %d", tt.wantNumInput, got)
			}

			result, execErr := stmt.Exec(nil)
			if execErr != nil {
				t.Fatalf("exec error: %v", execErr)
			}
			rows, rowsErr := result.RowsAffected()
			if rowsErr != nil {
				t.Fatalf("rows affected error: %v", rowsErr)
			}
			if rows != tt.wantExecRows {
				t.Fatalf("expected rows affected %d, got %d", tt.wantExecRows, rows)
			}

			_, queryErr := stmt.Query(nil)
			if queryErr == nil || !strings.Contains(queryErr.Error(), tt.wantQueryErr) {
				t.Fatalf("expected query error containing %q, got %v", tt.wantQueryErr, queryErr)
			}

			if !reflect.DeepEqual(stmt, stubSQLStmt{}) {
				t.Fatalf("stmt should remain stateless after operations")
			}
		})
	}
}
