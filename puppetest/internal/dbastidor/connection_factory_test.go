package dbastidor

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest/internal/testkit"
)

func stubPerformer(t *testing.T, rootState, subState *testkit.SQLState) ConnectionPerformer {
	t.Helper()

	return func(_ context.Context, conf ConnectionConfig) (*sql.DB, error) {
		if conf.DBName == "" {
			return testkit.OpenStubDB(t, t.Name()+"-root", rootState), nil
		}
		return testkit.OpenStubDB(t, t.Name()+"-sub-"+conf.DBName, subState), nil
	}
}

func TestNewConnectionFactory(t *testing.T) {
	t.Parallel()

	connectErr := errors.New("connect failed")
	testCases := []struct {
		name           string
		performer      ConnectionPerformer
		buildLifecycle LifecycleBuilder
		wantErrText    string
		wantLifecycle  DBLifecycle
	}{
		{
			name:          "defaults to the no-op lifecycle",
			wantLifecycle: NoOpLifecycle{},
		},
		{
			name:           "installs the built lifecycle",
			buildLifecycle: func(rootDB *sql.DB) DBLifecycle { return NewMySQLLifecycle(rootDB) },
		},
		{
			name:        "rejects a nil performer",
			performer:   nil,
			wantErrText: "nil connection performer",
		},
		{
			name: "propagates a root connection failure",
			performer: func(context.Context, ConnectionConfig) (*sql.DB, error) {
				return nil, connectErr
			},
			wantErrText: "connect failed",
		},
		{
			name:           "rejects a builder that returns no lifecycle",
			buildLifecycle: func(*sql.DB) DBLifecycle { return nil },
			wantErrText:    "database lifecycle builder returned nil",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			performer := testCase.performer
			if performer == nil && testCase.wantErrText != "nil connection performer" {
				performer = stubPerformer(t, &testkit.SQLState{}, &testkit.SQLState{})
			}

			factory, err := NewConnectionFactory(t.Context(), performer, testCase.buildLifecycle)
			if testCase.wantErrText != "" {
				if err == nil || !strings.Contains(err.Error(), testCase.wantErrText) {
					t.Fatalf("expected error containing %q, got %v", testCase.wantErrText, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			t.Cleanup(func() { _ = factory.Close() })

			if !factory.IsSetup() {
				t.Fatal("expected the factory to report itself set up")
			}
			if testCase.wantLifecycle != nil &&
				!reflect.DeepEqual(factory.lifecycle, testCase.wantLifecycle) {
				t.Fatalf("expected lifecycle %T, got %T", testCase.wantLifecycle, factory.lifecycle)
			}
		})
	}
}

func TestConnectionFactory_NewDatabase(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		dbName         string
		buildLifecycle LifecycleBuilder
		wantRootStmts  []string
		wantStmtShapes []string
	}{
		{
			name:          "no-op lifecycle runs no statement",
			dbName:        "TestThing/Sub-Case",
			wantRootStmts: []string{},
		},
		{
			name:           "mysql lifecycle creates then drops the normalized name",
			dbName:         "TestThing/Sub-Case",
			buildLifecycle: func(rootDB *sql.DB) DBLifecycle { return NewMySQLLifecycle(rootDB) },
			wantStmtShapes: []string{
				"CREATE DATABASE IF NOT EXISTS `%s`",
				"DROP DATABASE IF EXISTS `%s`",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			rootState, subState := &testkit.SQLState{}, &testkit.SQLState{}
			factory, err := NewConnectionFactory(
				t.Context(),
				stubPerformer(t, rootState, subState),
				testCase.buildLifecycle,
			)
			if err != nil {
				t.Fatalf("create factory: %v", err)
			}
			t.Cleanup(func() { _ = factory.Close() })

			newDb, err := factory.NewDatabase(t.Context(), testCase.dbName)
			if err != nil {
				t.Fatalf("new database: %v", err)
			}
			wantName := NormalizeDBName(testCase.dbName)
			if newDb.Name != wantName {
				t.Fatalf("expected database name %q, got %q", wantName, newDb.Name)
			}

			if err = newDb.Teardown(t.Context()); err != nil {
				t.Fatalf("teardown: %v", err)
			}
			if got := subState.CloseCount(); got != 1 {
				t.Fatalf("expected the test connection to be closed once, got %d", got)
			}
			want := testCase.wantRootStmts
			for _, shape := range testCase.wantStmtShapes {
				want = append(want, fmt.Sprintf(shape, wantName))
			}
			if got := rootState.ExecStatements(); !reflect.DeepEqual(got, want) {
				t.Fatalf("expected statements %v, got %v", want, got)
			}
		})
	}
}

func TestConnectionFactory_NewDatabaseFailsWhenCreateFails(t *testing.T) {
	t.Parallel()

	rootState := &testkit.SQLState{ExecErr: errors.New("exec failed")}
	factory, err := NewConnectionFactory(
		t.Context(),
		stubPerformer(t, rootState, &testkit.SQLState{}),
		func(rootDB *sql.DB) DBLifecycle { return NewMySQLLifecycle(rootDB) },
	)
	if err != nil {
		t.Fatalf("create factory: %v", err)
	}
	t.Cleanup(func() { _ = factory.Close() })

	if _, err = factory.NewDatabase(t.Context(), "games"); err == nil {
		t.Fatal("expected a create error, got nil")
	}
}

func TestConnectionFactory_CloseIsIdempotent(t *testing.T) {
	t.Parallel()

	rootState := &testkit.SQLState{}
	factory, err := NewConnectionFactory(
		t.Context(),
		stubPerformer(t, rootState, &testkit.SQLState{}),
		nil,
	)
	if err != nil {
		t.Fatalf("create factory: %v", err)
	}

	for range 2 {
		if closeErr := factory.Close(); closeErr != nil {
			t.Fatalf("close: %v", closeErr)
		}
	}
	if got := rootState.CloseCount(); got != 1 {
		t.Fatalf("expected the root connection to be closed once, got %d", got)
	}
	if factory.IsSetup() {
		t.Fatal("expected the factory to stop reporting itself set up after close")
	}
}

func TestNewConnectionFactory_ClosesRootWhenBuilderReturnsNil(t *testing.T) {
	t.Parallel()

	rootState := &testkit.SQLState{}
	_, err := NewConnectionFactory(
		t.Context(),
		stubPerformer(t, rootState, &testkit.SQLState{}),
		func(*sql.DB) DBLifecycle { return nil },
	)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if got := rootState.CloseCount(); got != 1 {
		t.Fatalf("expected the root connection to be closed once, got %d", got)
	}
}
