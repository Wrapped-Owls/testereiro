package dbastidor

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest/internal/testkit"
)

func TestMySQLLifecycle(t *testing.T) {
	t.Parallel()

	execErr := errors.New("exec failed")
	testCases := []struct {
		name     string
		dbName   string
		execErr  error
		run      func(lifecycle *MySQLLifecycle, ctx context.Context, dbName string) error
		want     []string
		wantErr  error
		wantNone bool
	}{
		{
			name:   "create quotes the name with backticks",
			dbName: "games_db",
			run:    (*MySQLLifecycle).Create,
			want:   []string{"CREATE DATABASE IF NOT EXISTS `games_db`"},
		},
		{
			name:   "drop quotes the name with backticks",
			dbName: "games_db",
			run:    (*MySQLLifecycle).Drop,
			want:   []string{"DROP DATABASE IF EXISTS `games_db`"},
		},
		{
			name:   "create escapes an embedded backtick instead of interpolating it",
			dbName: "games`; DROP DATABASE other; --",
			run:    (*MySQLLifecycle).Create,
			want:   []string{"CREATE DATABASE IF NOT EXISTS `games``; DROP DATABASE other; --`"},
		},
		{
			name:     "create rejects an unquotable name before touching the connection",
			dbName:   "",
			run:      (*MySQLLifecycle).Create,
			wantErr:  ErrInvalidIdentifier,
			wantNone: true,
		},
		{
			name:     "drop rejects an unquotable name before touching the connection",
			dbName:   "",
			run:      (*MySQLLifecycle).Drop,
			wantErr:  ErrInvalidIdentifier,
			wantNone: true,
		},
		{
			name:    "create wraps a failing statement",
			dbName:  "games_db",
			execErr: execErr,
			run:     (*MySQLLifecycle).Create,
			want:    []string{"CREATE DATABASE IF NOT EXISTS `games_db`"},
			wantErr: execErr,
		},
		{
			name:    "drop wraps a failing statement",
			dbName:  "games_db",
			execErr: execErr,
			run:     (*MySQLLifecycle).Drop,
			want:    []string{"DROP DATABASE IF EXISTS `games_db`"},
			wantErr: execErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			state := &testkit.SQLState{ExecErr: testCase.execErr}
			rootDB := testkit.OpenStubDB(t, t.Name(), state)
			t.Cleanup(func() { _ = rootDB.Close() })

			err := testCase.run(NewMySQLLifecycle(rootDB), t.Context(), testCase.dbName)
			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("expected error %v, got %v", testCase.wantErr, err)
			}

			got := state.ExecStatements()
			if testCase.wantNone {
				if len(got) != 0 {
					t.Fatalf("expected no statement to run, got %v", got)
				}
				return
			}
			if !reflect.DeepEqual(got, testCase.want) {
				t.Fatalf("expected statements %v, got %v", testCase.want, got)
			}
		})
	}
}
