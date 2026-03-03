package mongochecker

import (
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/wrapped-owls/testereiro/providers/mongotestage/internal/mongoqueries"
	"github.com/wrapped-owls/testereiro/puppetest"
)

// WithQueryBuilder sets the query builder used by MongoChecker.
func WithQueryBuilder(builder QueryBuilder) Option {
	return func(modifier CheckerModifier) {
		modifier.SetQueryBuilder(builder)
	}
}

// ExpectDocs validates all decoded documents against expected values.
func ExpectDocs[T any](expected []T, sanitizer ...objSanitizer[T]) Option {
	return func(modifier CheckerModifier) {
		v := genericValidator[T, []T]{ExpectedList: expected}
		if len(sanitizer) > 0 {
			v.Sanitizer = sanitizer[0]
		}
		modifier.AddValidator(v)
	}
}

// ExpectDoc validates the first decoded document against expected value.
func ExpectDoc[T any](expected T, sanitizer ...objSanitizer[T]) Option {
	return func(modifier CheckerModifier) {
		v := genericValidator[T, []T]{Expected: expected}
		if len(sanitizer) > 0 {
			v.Sanitizer = sanitizer[0]
		}
		modifier.AddValidator(v)
	}
}

// WithAggregateQuery configures an aggregate query with a static pipeline.
func WithAggregateQuery(collection string, pipeline bson.A) Option {
	return func(modifier CheckerModifier) {
		modifier.SetQueryBuilder(NewAggregateQuery(collection, pipeline))
	}
}

// WithAggregateQueryFromCtx configures an aggregate query whose pipeline is resolved from context.
func WithAggregateQueryFromCtx(
	collection string, pipelineFn func(ctx puppetest.Context) (bson.A, error),
) Option {
	return func(modifier CheckerModifier) {
		builder := NewAggregateQuery(collection, nil)
		builder.AddPipeline(pipelineFn)
		modifier.SetQueryBuilder(builder)
	}
}

// WithAggregateOptions sets options for aggregate operations.
func WithAggregateOptions(opts ...options.Lister[options.AggregateOptions]) Option {
	return func(modifier CheckerModifier) {
		modifier.SetQueryOptions(opts)
	}
}

// WithFindOneQuery configures a findOne query with a static filter.
func WithFindOneQuery(collection string, filter bson.M) Option {
	return func(modifier CheckerModifier) {
		qb := NewBsonQuery(collection, filter)
		qb.SetOperation(OpFindOne)
		modifier.SetQueryBuilder(qb)
	}
}

// WithFindOneQueryFromCtx configures a findOne query whose filter is resolved from context.
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

// WithFindOptions sets options for find operations.
func WithFindOptions(opts ...options.Lister[options.FindOptions]) Option {
	return func(modifier CheckerModifier) {
		modifier.SetQueryOptions(opts)
	}
}

// WithFindOneOptions sets options for findOne operations.
func WithFindOneOptions(opts ...options.Lister[options.FindOneOptions]) Option {
	return func(modifier CheckerModifier) {
		modifier.SetQueryOptions(opts)
	}
}

// WithCountQuery configures a count query with a static filter.
func WithCountQuery(collection string, filter bson.M) Option {
	return func(modifier CheckerModifier) {
		qb := NewBsonQuery(collection, filter)
		qb.SetOperation(OpCount)
		modifier.SetQueryBuilder(qb)
	}
}

// WithCountQueryFromCtx configures a count query whose filter is resolved from context.
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

// WithCountOptions sets options for count operations.
func WithCountOptions(opts ...options.Lister[options.CountOptions]) Option {
	return func(modifier CheckerModifier) {
		modifier.SetQueryOptions(opts)
	}
}

// ExpectCount validates the resulting count value.
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

// WithCustomValidation appends a custom validation callback.
func WithCustomValidation(
	validate func(t testing.TB, ctx puppetest.Context, cursor *Cursor) error,
) Option {
	return func(modifier CheckerModifier) {
		modifier.AddValidator(&customValidator{
			validate: validate,
		})
	}
}
