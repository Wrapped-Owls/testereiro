package testkit

import (
	"reflect"
	"testing"
)

func TestSQLState_Recorders(t *testing.T) {
	tests := []struct {
		name        string
		queries     []string
		closeCalls  int
		wantQueries []string
		wantClose   int
	}{
		{
			name:        "records queries and close count",
			queries:     []string{"CREATE TABLE t(id INT)", "INSERT INTO t VALUES(1)"},
			closeCalls:  2,
			wantQueries: []string{"CREATE TABLE t(id INT)", "INSERT INTO t VALUES(1)"},
			wantClose:   2,
		},
		{
			name:        "empty state remains empty",
			queries:     nil,
			closeCalls:  0,
			wantQueries: []string{},
			wantClose:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &SQLState{}
			for _, query := range tt.queries {
				state.recordExec(query)
			}
			for i := 0; i < tt.closeCalls; i++ {
				state.recordClose()
			}

			if got := state.ExecStatements(); !reflect.DeepEqual(got, tt.wantQueries) {
				t.Fatalf("expected queries %v, got %v", tt.wantQueries, got)
			}
			if got := state.CloseCount(); got != tt.wantClose {
				t.Fatalf("expected close count %d, got %d", tt.wantClose, got)
			}
		})
	}
}

func TestSQLState_ExecStatementsReturnsCopy(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returned slice mutation does not affect state"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &SQLState{}
			state.recordExec("SELECT 1")

			got := state.ExecStatements()
			got[0] = "MUTATED"

			reloaded := state.ExecStatements()
			if reloaded[0] != "SELECT 1" {
				t.Fatalf("expected original query to remain intact, got %q", reloaded[0])
			}
		})
	}
}
