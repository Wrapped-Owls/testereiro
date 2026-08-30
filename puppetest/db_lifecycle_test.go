package puppetest

import (
	"context"
	"database/sql/driver"
	"errors"
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest/internal/testkit"
)

func TestQuoteIdentifier(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		input   string
		quote   rune
		want    string
		wantErr error
	}{
		{name: "mysql name", input: "games", quote: '`', want: "`games`"},
		{name: "postgres name", input: "games", quote: '"', want: `"games"`},
		{
			name:  "doubles the embedded quote rune",
			input: `games"; DROP DATABASE other; --`,
			quote: '"',
			want:  `"games""; DROP DATABASE other; --"`,
		},
		{name: "empty name", input: "", quote: '"', wantErr: ErrInvalidIdentifier},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := QuoteIdentifier(testCase.input, testCase.quote)
			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("expected error %v, got %v", testCase.wantErr, err)
			}
			if testCase.wantErr == nil && got != testCase.want {
				t.Fatalf("expected %q, got %q", testCase.want, got)
			}
		})
	}
}

func TestShippedLifecycleBuilders(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		build   DBLifecycleBuilder
		wantDDL string
	}{
		{
			name:    "mysql builder quotes with backticks",
			build:   MySQLLifecycle,
			wantDDL: "CREATE DATABASE IF NOT EXISTS `games`",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			state := &testkit.SQLState{}
			rootDB := testkit.OpenStubDB(t, t.Name(), state)
			t.Cleanup(func() { _ = rootDB.Close() })

			lifecycle := testCase.build(rootDB)
			if lifecycle == nil {
				t.Fatal("expected a lifecycle, got nil")
			}
			if err := lifecycle.Create(t.Context(), "games"); err != nil {
				t.Fatalf("create: %v", err)
			}

			statements := state.ExecStatements()
			if testCase.wantDDL == "" {
				if len(statements) != 0 {
					t.Fatalf("expected no statement, got %v", statements)
				}
				return
			}
			if len(statements) != 1 || statements[0] != testCase.wantDDL {
				t.Fatalf("expected statement %q, got %v", testCase.wantDDL, statements)
			}
		})
	}
}

func TestPostgresLifecycleBuilder(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		exists   bool
		wantDDL  []string
		wantSkip bool
	}{
		{
			name:    "creates the database when pg_database has no row for it",
			exists:  false,
			wantDDL: []string{`CREATE DATABASE "games"`},
		},
		{
			name:     "skips creation when the database already exists",
			exists:   true,
			wantSkip: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			state := &testkit.SQLState{
				QueryCols: []string{"exists"},
				QueryRows: [][]driver.Value{{testCase.exists}},
			}
			rootDB := testkit.OpenStubDB(t, t.Name(), state)
			t.Cleanup(func() { _ = rootDB.Close() })

			if err := PostgresLifecycle(rootDB).Create(t.Context(), "games"); err != nil {
				t.Fatalf("create: %v", err)
			}

			statements := state.ExecStatements()
			if testCase.wantSkip {
				if len(statements) != 0 {
					t.Fatalf("expected no statement, got %v", statements)
				}
				return
			}
			if len(statements) != 1 || statements[0] != testCase.wantDDL[0] {
				t.Fatalf("expected statement %v, got %v", testCase.wantDDL, statements)
			}
		})
	}
}

func TestDBLifecycleIsImplementableOutsideTheLibrary(t *testing.T) {
	t.Parallel()

	// The alias is what lets a consumer name the type at all; internal/ would forbid it.
	var custom DBLifecycle = countingLifecycle{}
	if err := custom.Create(t.Context(), "games"); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := custom.Drop(t.Context(), "games"); err != nil {
		t.Fatalf("drop: %v", err)
	}
}

type countingLifecycle struct{}

var _ DBLifecycle = countingLifecycle{}

func (countingLifecycle) Create(context.Context, string) error { return nil }
func (countingLifecycle) Drop(context.Context, string) error   { return nil }
