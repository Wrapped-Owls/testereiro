package bancoche

import "github.com/wrapped-owls/testereiro/puppetest/internal/stgctx"

// RawQueryBuilder holds a raw SQL query and its arguments.
type RawQueryBuilder struct {
	query string
	args  []any
}

// NewRawQuery creates a RawQueryBuilder from a static query and arguments.
func NewRawQuery(query string, args ...any) *RawQueryBuilder {
	return &RawQueryBuilder{
		query: query,
		args:  args,
	}
}

// Build returns the raw SQL query and arguments unchanged.
func (b *RawQueryBuilder) Build(_ stgctx.RunnerContext) (string, []any, error) {
	return b.query, b.args, nil
}
