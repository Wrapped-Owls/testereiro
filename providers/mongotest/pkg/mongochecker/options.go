package mongochecker

import (
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/wrapped-owls/testereiro/providers/mongotest/internal/mongoqueries"
	"github.com/wrapped-owls/testereiro/puppetest"
)

func WithQueryBuilder(builder QueryBuilder) Option {
	return func(modifier CheckerModifier) {
		modifier.SetQueryBuilder(builder)
	}
}

func ExpectDocs[T any](expected []T, sanitizer ...objSanitizer[T]) Option {
	return func(modifier CheckerModifier) {
		v := genericValidator[T, []T]{ExpectedList: expected}
		if len(sanitizer) > 0 {
			v.Sanitizer = sanitizer[0]
		}
		modifier.AddValidator(v)
	}
}

func ExpectDoc[T any](expected T, sanitizer ...objSanitizer[T]) Option {
	return func(modifier CheckerModifier) {
		v := genericValidator[T, []T]{Expected: expected}
		if len(sanitizer) > 0 {
			v.Sanitizer = sanitizer[0]
		}
		modifier.AddValidator(v)
	}
}

func WithAggregateQuery(collection string, pipeline bson.A) Option {
	return func(modifier CheckerModifier) {
		modifier.SetQueryBuilder(NewAggregateQuery(collection, pipeline))
	}
}

func WithAggregateQueryFromCtx(
	collection string, pipelineFn func(ctx puppetest.Context) (bson.A, error),
) Option {
	return func(modifier CheckerModifier) {
		builder := NewAggregateQuery(collection, nil)
		builder.AddPipeline(pipelineFn)
		modifier.SetQueryBuilder(builder)
	}
}

func WithAggregateOptions(opts ...options.Lister[options.AggregateOptions]) Option {
	return func(modifier CheckerModifier) {
		modifier.SetQueryOptions(opts)
	}
}

func WithFindOneQuery(collection string, filter bson.M) Option {
	return func(modifier CheckerModifier) {
		qb := NewBsonQuery(collection, filter)
		qb.SetOperation(OpFindOne)
		modifier.SetQueryBuilder(qb)
	}
}

func WithFindOneQueryFromCtx(
	collection string, filterFn func(ctx puppetest.Context) (bson.M, error),
) Option {
	return func(modifier CheckerModifier) {
		builder := NewBsonQuery(collection, nil)
		builder.AddFilter(filterFn)
		builder.SetOperation(OpFindOne)
		modifier.SetQueryBuilder(builder)
	}
}

func WithFindOptions(opts ...options.Lister[options.FindOptions]) Option {
	return func(modifier CheckerModifier) {
		modifier.SetQueryOptions(opts)
	}
}

func WithFindOneOptions(opts ...options.Lister[options.FindOneOptions]) Option {
	return func(modifier CheckerModifier) {
		modifier.SetQueryOptions(opts)
	}
}

func WithCountQuery(collection string, filter bson.M) Option {
	return func(modifier CheckerModifier) {
		qb := NewBsonQuery(collection, filter)
		qb.SetOperation(OpCount)
		modifier.SetQueryBuilder(qb)
	}
}

func WithCountQueryFromCtx(
	collection string, filterFn func(ctx puppetest.Context) (bson.M, error),
) Option {
	return func(modifier CheckerModifier) {
		builder := NewBsonQuery(collection, nil)
		builder.AddFilter(filterFn)
		builder.SetOperation(OpCount)
		modifier.SetQueryBuilder(builder)
	}
}

func WithCountOptions(opts ...options.Lister[options.CountOptions]) Option {
	return func(modifier CheckerModifier) {
		modifier.SetQueryOptions(opts)
	}
}

func ExpectCount(expected int) Option {
	return func(modifier CheckerModifier) {
		modifier.AddValidator(&countValidator{expected: expected})
	}
}

// DecodeAll retrieves all results from a Cursor.
func DecodeAll[O any](ctx puppetest.Context, cursor *Cursor) ([]O, error) {
	return mongoqueries.All[O](ctx, cursor)
}

// DecodeFirst retrieves the first result from a Cursor.
func DecodeFirst[O any](ctx puppetest.Context, cursor *Cursor) (O, error) {
	return mongoqueries.First[O](ctx, cursor)
}

func WithCustomValidation(
	validate func(t testing.TB, ctx puppetest.Context, cursor *Cursor) error,
) Option {
	return func(modifier CheckerModifier) {
		modifier.AddValidator(&customValidator{
			validate: validate,
		})
	}
}
