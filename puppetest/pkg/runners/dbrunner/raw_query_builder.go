package dbrunner

import "github.com/wrapped-owls/testereiro/puppetest/internal/stgctx"

// RawQueryBuilder holds a raw SQL query and its arguments.
type RawQueryBuilder struct {
	query string
	args  []any
}

func NewRawQuery(query string, args ...any) *RawQueryBuilder {
	return &RawQueryBuilder{
		query: query,
		args:  args,
	}
}

func (b *RawQueryBuilder) Build(_ stgctx.RunnerContext) (string, []any, error) {
	return b.query, b.args, nil
}
