package dbrunner

import (
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest/internal/stgctx"
)

func TestMapQueryBuilder_Build(t *testing.T) {
	ctx := stgctx.NewRunnerContext(t.Context())

	tests := []struct {
		name          string
		table         string
		filters       map[string]any
		expectedQuery string
		expectedArgs  []any
	}{
		{
			name:          "basic_query",
			table:         "users",
			filters:       map[string]any{"id": 1},
			expectedQuery: "SELECT * FROM users WHERE id = ?",
			expectedArgs:  []any{1},
		},
		{
			name:          "single_filter_name",
			table:         "users",
			filters:       map[string]any{"name": "john"},
			expectedQuery: "SELECT * FROM users WHERE name = ?",
			expectedArgs:  []any{"john"},
		},
		{
			name:          "no_filters",
			table:         "users",
			filters:       nil,
			expectedQuery: "SELECT * FROM users",
			expectedArgs:  nil,
		},
	}

	for _, tCase := range tests {
		t.Run(tCase.name, func(t *testing.T) {
			qb := NewMapQuery(tCase.table, tCase.filters)

			query, args, err := qb.Build(ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if query != tCase.expectedQuery {
				t.Errorf("expected query %q, got %q", tCase.expectedQuery, query)
			}

			if len(args) != len(tCase.expectedArgs) {
				t.Fatalf("expected %d args, got %d (%v)", len(tCase.expectedArgs), len(args), args)
			}

			for index := range args {
				if args[index] != tCase.expectedArgs[index] {
					t.Errorf(
						"expected arg[%d] = %v, got %v",
						index,
						tCase.expectedArgs[index],
						args[index],
					)
				}
			}
		})
	}
}
