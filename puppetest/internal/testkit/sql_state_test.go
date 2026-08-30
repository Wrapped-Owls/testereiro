package testkit

import (
	"database/sql/driver"
	"reflect"
	"testing"
)

func TestSQLState_Recorders(t *testing.T) {
	t.Parallel()

	testCases := []struct {
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

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			state := &SQLState{}
			for _, query := range testCase.queries {
				state.recordExec(query, nil)
			}
			for range testCase.closeCalls {
				state.recordClose()
			}

			if got := state.ExecStatements(); !reflect.DeepEqual(got, testCase.wantQueries) {
				t.Fatalf("expected queries %v, got %v", testCase.wantQueries, got)
			}
			if got := state.CloseCount(); got != testCase.wantClose {
				t.Fatalf("expected close count %d, got %d", testCase.wantClose, got)
			}
		})
	}
}

func TestSQLState_ExecStatementsReturnsCopy(t *testing.T) {
	t.Parallel()

	state := &SQLState{}
	state.recordExec("SELECT 1", nil)

	mutated := state.ExecStatements()
	mutated[0] = "MUTATED"

	reloaded := state.ExecStatements()
	if reloaded[0] != "SELECT 1" {
		t.Fatalf("expected original query to remain intact, got %q", reloaded[0])
	}
}

func TestSQLState_CallRecorders(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		record     func(state *SQLState, statement string, args []driver.Value)
		read       func(state *SQLState) []RecordedQuery
		statement  string
		args       []driver.Value
		wantNoArgs bool
	}{
		{
			name:      "exec calls keep their bound arguments",
			record:    (*SQLState).recordExec,
			read:      (*SQLState).ExecCalls,
			statement: "DROP DATABASE IF EXISTS x",
			args:      []driver.Value{"games"},
		},
		{
			name:      "query calls keep their bound arguments",
			record:    (*SQLState).recordQuery,
			read:      (*SQLState).QueryCalls,
			statement: "SELECT 1 FROM pg_database WHERE datname = $1",
			args:      []driver.Value{"games"},
		},
		{
			name:       "argument-less call records an empty argument list",
			record:     (*SQLState).recordExec,
			read:       (*SQLState).ExecCalls,
			statement:  "SELECT 1",
			args:       nil,
			wantNoArgs: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			state := &SQLState{}
			testCase.record(state, testCase.statement, testCase.args)

			calls := testCase.read(state)
			if len(calls) != 1 {
				t.Fatalf("expected 1 recorded call, got %d", len(calls))
			}
			if calls[0].Statement != testCase.statement {
				t.Fatalf("expected statement %q, got %q", testCase.statement, calls[0].Statement)
			}
			if testCase.wantNoArgs {
				if len(calls[0].Args) != 0 {
					t.Fatalf("expected no arguments, got %v", calls[0].Args)
				}
				return
			}
			if !reflect.DeepEqual(calls[0].Args, testCase.args) {
				t.Fatalf("expected args %v, got %v", testCase.args, calls[0].Args)
			}
		})
	}
}

func TestSQLState_RecordedArgsAreCloned(t *testing.T) {
	t.Parallel()

	args := []driver.Value{"games"}
	state := &SQLState{}
	state.recordExec("SELECT 1", args)

	// The driver reuses its argument slice between calls, so the recorder must not alias it.
	args[0] = "mutated"

	calls := state.ExecCalls()
	if calls[0].Args[0] != "games" {
		t.Fatalf("expected recorded arg to stay %q, got %v", "games", calls[0].Args[0])
	}
}
