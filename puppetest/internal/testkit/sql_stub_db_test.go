package testkit

import (
	"context"
	"reflect"
	"testing"
)

func TestOpenStubDB(t *testing.T) {
	tests := []struct {
		name          string
		dsn           string
		state         *SQLState
		execQuery     string
		wantCloseAtLe int
		wantRecorded  []string
	}{
		{
			name:          "opens db with pre-registered state",
			dsn:           "db-with-state",
			state:         &SQLState{},
			execQuery:     "SELECT 1",
			wantCloseAtLe: 1,
			wantRecorded:  []string{"SELECT 1"},
		},
		{
			name:          "opens db with nil state and creates one internally",
			dsn:           "db-with-nil-state",
			state:         nil,
			execQuery:     "SELECT 2",
			wantCloseAtLe: 0,
			wantRecorded:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ResetSQLRegistry()
			db := OpenStubDB(t, tt.dsn, tt.state)

			if err := db.PingContext(context.Background()); err != nil {
				t.Fatalf("ping error: %v", err)
			}
			if _, err := db.Exec(tt.execQuery); err != nil {
				t.Fatalf("exec error: %v", err)
			}
			if err := db.Close(); err != nil {
				t.Fatalf("close error: %v", err)
			}

			opened := OpenedDSNs()
			if len(opened) != 1 || opened[0] != tt.dsn {
				t.Fatalf("expected opened dsns [%q], got %v", tt.dsn, opened)
			}

			if tt.state != nil {
				if got := tt.state.CloseCount(); got < tt.wantCloseAtLe {
					t.Fatalf("expected close count >= %d, got %d", tt.wantCloseAtLe, got)
				}
				if got := tt.state.ExecStatements(); !reflect.DeepEqual(got, tt.wantRecorded) {
					t.Fatalf("expected recorded queries %v, got %v", tt.wantRecorded, got)
				}
			}
		})
	}
}
