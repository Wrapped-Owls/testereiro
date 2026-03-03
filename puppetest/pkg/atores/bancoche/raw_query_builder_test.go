package bancoche

import (
	"reflect"
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest/internal/stgctx"
)

func TestRawQueryBuilder_Build(t *testing.T) {
	runnerCtx := stgctx.NewRunnerContext(t.Context())

	tests := []struct {
		name  string
		query string
		args  []any
	}{
		{
			name:  "returns query without args",
			query: "SELECT 1",
			args:  nil,
		},
		{
			name:  "returns query with positional args preserving order",
			query: "SELECT * FROM users WHERE id = ? AND active = ?",
			args:  []any{7, true},
		},
		{
			name:  "supports nil arg values",
			query: "INSERT INTO users(name, note) VALUES(?, ?)",
			args:  []any{"prime", nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewRawQuery(tt.query, tt.args...)

			gotQuery, gotArgs, err := builder.Build(runnerCtx)
			if err != nil {
				t.Fatalf("unexpected build error: %v", err)
			}

			if gotQuery != tt.query {
				t.Fatalf("expected query %q, got %q", tt.query, gotQuery)
			}
			if !reflect.DeepEqual(gotArgs, tt.args) {
				t.Fatalf("expected args %#v, got %#v", tt.args, gotArgs)
			}
		})
	}
}
