package dbastidor

import (
	"database/sql/driver"
	"errors"
	"reflect"
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest/internal/testkit"
)

// pgErrorStub stands in for pgx's *pgconn.PgError, which puppetest only ever sees through its
// SQLState method.
type pgErrorStub struct {
	state string
}

func (e pgErrorStub) Error() string {
	return "postgres error " + e.state
}

func (e pgErrorStub) SQLState() string {
	return e.state
}

func databaseFoundResult() *testkit.SQLState {
	return &testkit.SQLState{
		QueryCols: []string{"exists"},
		QueryRows: [][]driver.Value{{true}},
	}
}

func databaseMissingResult() *testkit.SQLState {
	return &testkit.SQLState{
		QueryCols: []string{"exists"},
		QueryRows: [][]driver.Value{{false}},
	}
}

func TestPostgresLifecycle_Create(t *testing.T) {
	t.Parallel()

	execErr := errors.New("exec failed")
	queryErr := errors.New("query failed")
	testCases := []struct {
		name       string
		dbName     string
		state      *testkit.SQLState
		wantErr    error
		wantQuery  []testkit.RecordedQuery
		wantExec   []string
		wantNoExec bool
	}{
		{
			name:   "creates the database when pg_database has no row for it",
			dbName: "games_db",
			state:  databaseMissingResult(),
			wantQuery: []testkit.RecordedQuery{
				{Statement: postgresExistsStmt, Args: []driver.Value{"games_db"}},
			},
			wantExec: []string{`CREATE DATABASE "games_db"`},
		},
		{
			name:   "skips creation when the database already exists",
			dbName: "games_db",
			state:  databaseFoundResult(),
			wantQuery: []testkit.RecordedQuery{
				{Statement: postgresExistsStmt, Args: []driver.Value{"games_db"}},
			},
			wantNoExec: true,
		},
		{
			name:   "escapes an embedded double quote instead of interpolating it",
			dbName: `games"; DROP DATABASE other; --`,
			state:  databaseMissingResult(),
			wantQuery: []testkit.RecordedQuery{
				{Statement: postgresExistsStmt, Args: []driver.Value{`games"; DROP DATABASE other; --`}},
			},
			wantExec: []string{`CREATE DATABASE "games""; DROP DATABASE other; --"`},
		},
		{
			name:       "rejects an unquotable name before touching the connection",
			dbName:     "",
			state:      databaseMissingResult(),
			wantErr:    ErrInvalidIdentifier,
			wantNoExec: true,
		},
		{
			name:       "wraps a failing existence check",
			dbName:     "games_db",
			state:      &testkit.SQLState{QueryErr: queryErr},
			wantErr:    queryErr,
			wantNoExec: true,
		},
		{
			name:   "treats a lost create race as success",
			dbName: "games_db",
			state: &testkit.SQLState{
				QueryCols: []string{"exists"},
				QueryRows: [][]driver.Value{{false}},
				ExecErr:   pgErrorStub{state: duplicateDatabaseSQLState},
			},
			wantExec: []string{`CREATE DATABASE "games_db"`},
		},
		{
			name:   "propagates a create failure that is not a duplicate database",
			dbName: "games_db",
			state: &testkit.SQLState{
				QueryCols: []string{"exists"},
				QueryRows: [][]driver.Value{{false}},
				ExecErr:   execErr,
			},
			wantErr:  execErr,
			wantExec: []string{`CREATE DATABASE "games_db"`},
		},
		{
			name:   "propagates a create failure carrying another sqlstate",
			dbName: "games_db",
			state: &testkit.SQLState{
				QueryCols: []string{"exists"},
				QueryRows: [][]driver.Value{{false}},
				ExecErr:   pgErrorStub{state: "42501"},
			},
			wantErr:  pgErrorStub{state: "42501"},
			wantExec: []string{`CREATE DATABASE "games_db"`},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			rootDB := testkit.OpenStubDB(t, t.Name(), testCase.state)
			t.Cleanup(func() { _ = rootDB.Close() })

			err := NewPostgresLifecycle(rootDB).Create(t.Context(), testCase.dbName)
			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("expected error %v, got %v", testCase.wantErr, err)
			}

			if testCase.wantQuery != nil {
				if got := testCase.state.QueryCalls(); !reflect.DeepEqual(got, testCase.wantQuery) {
					t.Fatalf("expected queries %v, got %v", testCase.wantQuery, got)
				}
			}

			got := testCase.state.ExecStatements()
			if testCase.wantNoExec {
				if len(got) != 0 {
					t.Fatalf("expected no statement to run, got %v", got)
				}
				return
			}
			if !reflect.DeepEqual(got, testCase.wantExec) {
				t.Fatalf("expected statements %v, got %v", testCase.wantExec, got)
			}
		})
	}
}

func TestPostgresLifecycle_Drop(t *testing.T) {
	t.Parallel()

	execErr := errors.New("exec failed")
	testCases := []struct {
		name       string
		dbName     string
		execErr    error
		wantErr    error
		wantExec   []testkit.RecordedQuery
		wantNoExec bool
	}{
		{
			name:   "terminates live backends before dropping",
			dbName: "games_db",
			wantExec: []testkit.RecordedQuery{
				{Statement: postgresTerminateStmt, Args: []driver.Value{"games_db"}},
				{Statement: `DROP DATABASE IF EXISTS "games_db"`, Args: []driver.Value{}},
			},
		},
		{
			name:   "escapes an embedded double quote instead of interpolating it",
			dbName: `games"; DROP DATABASE other; --`,
			wantExec: []testkit.RecordedQuery{
				{
					Statement: postgresTerminateStmt,
					Args:      []driver.Value{`games"; DROP DATABASE other; --`},
				},
				{
					Statement: `DROP DATABASE IF EXISTS "games""; DROP DATABASE other; --"`,
					Args:      []driver.Value{},
				},
			},
		},
		{
			name:       "rejects an unquotable name before touching the connection",
			dbName:     "",
			wantErr:    ErrInvalidIdentifier,
			wantNoExec: true,
		},
		{
			name:    "wraps a failing terminate so a stuck drop is not silent",
			dbName:  "games_db",
			execErr: execErr,
			wantErr: execErr,
			wantExec: []testkit.RecordedQuery{
				{Statement: postgresTerminateStmt, Args: []driver.Value{"games_db"}},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			state := &testkit.SQLState{ExecErr: testCase.execErr}
			rootDB := testkit.OpenStubDB(t, t.Name(), state)
			t.Cleanup(func() { _ = rootDB.Close() })

			err := NewPostgresLifecycle(rootDB).Drop(t.Context(), testCase.dbName)
			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("expected error %v, got %v", testCase.wantErr, err)
			}

			got := state.ExecCalls()
			if testCase.wantNoExec {
				if len(got) != 0 {
					t.Fatalf("expected no statement to run, got %v", got)
				}
				return
			}
			if !reflect.DeepEqual(got, testCase.wantExec) {
				t.Fatalf("expected statements %v, got %v", testCase.wantExec, got)
			}
		})
	}
}
