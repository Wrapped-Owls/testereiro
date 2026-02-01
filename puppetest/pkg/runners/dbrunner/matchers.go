package dbrunner

import (
	"fmt"

	"github.com/wrapped-owls/testereiro/puppetest/internal/stgctx"
)

// WithQuery sets the query builder for the DB runner.
func WithQuery(qb QueryBuilder) Option {
	return func(r RunnerModifier) {
		r.SetQueryBuilder(qb)
	}
}

// WithMapQuery sets the from and a late-bound filter builder.
func WithMapQuery(table string, filter map[string]any) Option {
	return func(r RunnerModifier) {
		qb := NewMapQuery(table, filter)
		r.SetQueryBuilder(qb)
	}
}

// WithMapQueryFromCtx sets the from and a late-bound filter builder.
func WithMapQueryFromCtx[V any](table string, fn func(val V) map[string]any) Option {
	return func(r RunnerModifier) {
		lateFilter := func(ctx stgctx.RunnerContext) (map[string]any, error) {
			val, ok := stgctx.LoadFromCtx[V](ctx)
			if !ok {
				var zero V
				return nil, fmt.Errorf("failed to load value of type %T from storage", zero)
			}
			return fn(val), nil
		}

		qb := NewMapQuery(table, nil)
		qb.AddFilter(lateFilter)
		r.SetQueryBuilder(qb)
	}
}
