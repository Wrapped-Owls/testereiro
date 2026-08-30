package puppetest

import (
	"context"
	"database/sql"
	"reflect"
	"strings"
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest/internal/dbastidor"
	"github.com/wrapped-owls/testereiro/puppetest/internal/testkit"
)

func stubConnectionPerformer(
	t *testing.T, rootState, subState *testkit.SQLState,
) ConnectionPerformer {
	t.Helper()

	return func(_ context.Context, conf DBConnectionConfig) (*sql.DB, error) {
		if conf.DBName == "" {
			return testkit.OpenStubDB(t, t.Name()+"-root", rootState), nil
		}
		return testkit.OpenStubDB(t, t.Name()+"-sub-"+conf.DBName, subState), nil
	}
}

func TestWithDatabaseLifecycle_OrderIndependent(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		options func(performer ConnectionPerformer) []EngineFactoryOption
	}{
		{
			name: "lifecycle declared after the connection factory",
			options: func(performer ConnectionPerformer) []EngineFactoryOption {
				return []EngineFactoryOption{
					WithConnectionFactory(performer),
					WithDatabaseLifecycle(MySQLLifecycle),
				}
			},
		},
		{
			name: "lifecycle declared before the connection factory",
			options: func(performer ConnectionPerformer) []EngineFactoryOption {
				return []EngineFactoryOption{
					WithDatabaseLifecycle(MySQLLifecycle),
					WithConnectionFactory(performer),
				}
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			rootState := &testkit.SQLState{}
			performer := stubConnectionPerformer(t, rootState, &testkit.SQLState{})

			fac, err := NewEngineFactory(testCase.options(performer)...)
			if err != nil {
				t.Fatalf("create factory: %v", err)
			}
			t.Cleanup(func() { _ = fac.Close() })

			newDb, err := fac.dbFactory.NewDatabase(t.Context(), "Games")
			if err != nil {
				t.Fatalf("new database: %v", err)
			}
			if err = newDb.Teardown(t.Context()); err != nil {
				t.Fatalf("teardown: %v", err)
			}

			dbName := dbastidor.NormalizeDBName("Games")
			want := []string{
				"CREATE DATABASE IF NOT EXISTS `" + dbName + "`",
				"DROP DATABASE IF EXISTS `" + dbName + "`",
			}
			if got := rootState.ExecStatements(); !reflect.DeepEqual(got, want) {
				t.Fatalf("expected statements %v, got %v", want, got)
			}
		})
	}
}

func TestEngineFactoryOptionErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		options     []EngineFactoryOption
		wantErrText string
	}{
		{
			name:        "nil connection performer",
			options:     []EngineFactoryOption{WithConnectionFactory(nil)},
			wantErrText: "nil connection performer",
		},
		{
			name:        "nil lifecycle builder",
			options:     []EngineFactoryOption{WithDatabaseLifecycle(nil)},
			wantErrText: "nil database lifecycle builder",
		},
		{
			name:        "lifecycle without a connection factory",
			options:     []EngineFactoryOption{WithDatabaseLifecycle(MySQLLifecycle)},
			wantErrText: "database lifecycle set without a connection factory",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			fac, err := NewEngineFactory(testCase.options...)
			if err == nil {
				_ = fac.Close()
				t.Fatalf("expected error containing %q, got nil", testCase.wantErrText)
			}
			if !strings.Contains(err.Error(), testCase.wantErrText) {
				t.Fatalf("expected error containing %q, got %v", testCase.wantErrText, err)
			}
		})
	}
}

func TestWithConnectionFactory_NoLifecycleRunsNoDDL(t *testing.T) {
	t.Parallel()

	rootState := &testkit.SQLState{}
	fac, err := NewEngineFactory(
		WithConnectionFactory(stubConnectionPerformer(t, rootState, &testkit.SQLState{})),
	)
	if err != nil {
		t.Fatalf("create factory: %v", err)
	}
	t.Cleanup(func() { _ = fac.Close() })

	newDb, err := fac.dbFactory.NewDatabase(t.Context(), "Games")
	if err != nil {
		t.Fatalf("new database: %v", err)
	}
	if err = newDb.Teardown(t.Context()); err != nil {
		t.Fatalf("teardown: %v", err)
	}

	if got := rootState.ExecStatements(); len(got) != 0 {
		t.Fatalf("expected no statement without a lifecycle, got %v", got)
	}
}
