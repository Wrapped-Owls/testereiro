package mongochecker

import (
	"errors"
	"fmt"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/wrapped-owls/testereiro/providers/mongotest/internal/mongoqueries"
	"github.com/wrapped-owls/testereiro/puppetest"
)

// CheckerModifier receives option-driven mutations for MongoChecker.
type CheckerModifier interface {
	// SetQueryBuilder sets the query builder used by Run.
	SetQueryBuilder(QueryBuilder)
	// SetQueryOptions sets driver options for the selected operation.
	SetQueryOptions(opts any)
	// AddValidator appends a validation step executed after the query.
	AddValidator(v validator)
}

// Option configures a MongoChecker.
type Option func(modifier CheckerModifier)

// MongoChecker runs a mongo query and validates its results.
type MongoChecker struct {
	db           *mongo.Database
	query        QueryBuilder
	queryOptions any
	validators   []validator
}

// New creates a MongoChecker and applies options in order.
func New(db *mongo.Database, opts ...Option) *MongoChecker {
	r := &MongoChecker{db: db}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// AddValidator appends a result validator.
func (r *MongoChecker) AddValidator(v validator) {
	r.validators = append(r.validators, v)
}

// SetQueryBuilder sets the query builder used at execution time.
func (r *MongoChecker) SetQueryBuilder(builder QueryBuilder) {
	r.query = builder
}

// SetQueryOptions stores operation-specific mongo driver options.
func (r *MongoChecker) SetQueryOptions(opts any) {
	r.queryOptions = opts
}

// Run builds and executes the query, then runs validators over the cursor.
func (r *MongoChecker) Run(t testing.TB, ctx puppetest.Context) error {
	if r.db == nil {
		return fmt.Errorf("mongo database is nil")
	}
	if r.query == nil {
		return fmt.Errorf("query builder is required")
	}

	builtQuery, err := r.query.Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to build mongo query: %w", err)
	}
	if builtQuery.Collection == "" {
		return fmt.Errorf("query collection is required")
	}

	builtQuery.Options = r.queryOptions
	curWrap, queryErr := executeQuery(r.db, ctx, builtQuery)
	if queryErr != nil && errors.Is(queryErr, mongo.ErrNoDocuments) {
		return queryErr
	}

	defer func() { _ = curWrap.Close(ctx) }()

	for _, v := range r.validators {
		if err = v.Validate(t, ctx, curWrap); err != nil {
			return err
		}
	}
	return nil
}

func executeQuery(
	db *mongo.Database, ctx puppetest.Context, builtQuery Query,
) (*mongoqueries.Cursor, error) {
	collection := db.Collection(builtQuery.Collection)
	var curWrap *mongoqueries.Cursor
	var queryErr error

	switch builtQuery.Operation {
	case OpFindOne:
		var raw bson.Raw
		findOneOpts := castOptions[options.FindOneOptions](builtQuery.Options)
		raw, queryErr = collection.FindOne(ctx, builtQuery.Filter, findOneOpts...).Raw()
		curWrap = mongoqueries.NewCursorResult(raw)
	case OpCount:
		countOpts := castOptions[options.CountOptions](builtQuery.Options)
		var count int64
		count, queryErr = collection.CountDocuments(ctx, builtQuery.Filter, countOpts...)
		curWrap = mongoqueries.NewCursorCount(count)
	case OpAggregate:
		aggOpts := castOptions[options.AggregateOptions](builtQuery.Options)
		var cursor *mongo.Cursor
		cursor, queryErr = collection.Aggregate(ctx, builtQuery.Pipeline, aggOpts...)
		curWrap = mongoqueries.NewCursor(cursor)
	case OpFind:
		findOpts := castOptions[options.FindOptions](builtQuery.Options)
		var cursor *mongo.Cursor
		cursor, queryErr = collection.Find(ctx, builtQuery.Filter, findOpts...)
		curWrap = mongoqueries.NewCursor(cursor)
	default:
		return nil, fmt.Errorf("unsupported mongo operation: %d", builtQuery.Operation)
	}
	return curWrap, queryErr
}

func castOptions[T any](inputOpt any) []options.Lister[T] {
	if opt, ok := inputOpt.([]options.Lister[T]); ok {
		return opt
	}
	return nil
}
