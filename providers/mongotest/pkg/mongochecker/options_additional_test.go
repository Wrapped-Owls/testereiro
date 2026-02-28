package mongochecker

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/wrapped-owls/testereiro/puppetest"
)

type modifierSpy struct {
	queryBuilder QueryBuilder
	queryOptions any
	validators   []validator
}

func (m *modifierSpy) SetQueryBuilder(builder QueryBuilder) { m.queryBuilder = builder }
func (m *modifierSpy) SetQueryOptions(opts any)             { m.queryOptions = opts }

func (m *modifierSpy) AddValidator(
	v validator,
) {
	m.validators = append(m.validators, v)
}

func TestOptions_ModifierIntegration(t *testing.T) {
	spy := &modifierSpy{}

	WithQueryBuilder(NewBsonQuery("users", bson.M{"id": 1}))(spy)
	if spy.queryBuilder == nil {
		t.Fatal("expected query builder to be set")
	}

	WithAggregateQuery("users", bson.A{bson.D{{Key: "$match", Value: bson.M{"active": true}}}})(spy)
	if spy.queryBuilder == nil {
		t.Fatal("expected aggregate query builder")
	}

	WithFindOneQuery("users", bson.M{"id": 1})(spy)
	WithCountQuery("users", bson.M{"active": true})(spy)
	if spy.queryBuilder == nil {
		t.Fatal("expected query builder from query options")
	}

	WithFindOptions(options.Find())(spy)
	if _, ok := spy.queryOptions.([]options.Lister[options.FindOptions]); !ok {
		t.Fatalf("unexpected find options type: %T", spy.queryOptions)
	}

	WithFindOneOptions(options.FindOne())(spy)
	if _, ok := spy.queryOptions.([]options.Lister[options.FindOneOptions]); !ok {
		t.Fatalf("unexpected findOne options type: %T", spy.queryOptions)
	}

	WithCountOptions(options.Count())(spy)
	if _, ok := spy.queryOptions.([]options.Lister[options.CountOptions]); !ok {
		t.Fatalf("unexpected count options type: %T", spy.queryOptions)
	}

	WithAggregateOptions(options.Aggregate())(spy)
	if _, ok := spy.queryOptions.([]options.Lister[options.AggregateOptions]); !ok {
		t.Fatalf("unexpected aggregate options type: %T", spy.queryOptions)
	}

	ExpectCount(3)(spy)
	ExpectDoc(map[string]int{"a": 1})(spy)
	ExpectDocs([]map[string]int{{"a": 1}})(spy)
	WithCustomValidation(func(testing.TB, puppetest.Context, *Cursor) error { return nil })(spy)
	if len(spy.validators) != 4 {
		t.Fatalf("expected 4 validators, got %d", len(spy.validators))
	}
}

func TestCtxBasedOptions_BuildersUseContextFunctions(t *testing.T) {
	ctx := testRunnerContext(t)
	puppetest.SaveOnCtx(ctx, "tenant-1")

	spy := &modifierSpy{}
	WithFindOneQueryFromCtx("users", func(ctx puppetest.Context) (bson.M, error) {
		return bson.M{"tenant": "tenant-1"}, nil
	})(spy)
	query, err := spy.queryBuilder.Build(ctx)
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}
	if !reflect.DeepEqual(query.Filter, bson.M{"tenant": "tenant-1"}) {
		t.Fatalf("unexpected filter: %#v", query.Filter)
	}

	WithCountQueryFromCtx("users", func(ctx puppetest.Context) (bson.M, error) {
		return bson.M{"active": true}, nil
	})(spy)
	query, err = spy.queryBuilder.Build(ctx)
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}
	if !reflect.DeepEqual(query.Filter, bson.M{"active": true}) {
		t.Fatalf("unexpected count filter: %#v", query.Filter)
	}

	WithAggregateQueryFromCtx("users", func(ctx puppetest.Context) (bson.A, error) {
		return bson.A{bson.D{{Key: "$match", Value: bson.M{"active": true}}}}, nil
	})(spy)
	query, err = spy.queryBuilder.Build(ctx)
	if err != nil {
		t.Fatalf("unexpected aggregate build error: %v", err)
	}
	if len(query.Pipeline) != 1 {
		t.Fatalf("unexpected pipeline length: %d", len(query.Pipeline))
	}
}

func TestWithCustomValidation_ForwardsError(t *testing.T) {
	spy := &modifierSpy{}
	WithCustomValidation(func(testing.TB, puppetest.Context, *Cursor) error {
		return errors.New("validation error")
	})(spy)

	if len(spy.validators) != 1 {
		t.Fatalf("expected validator to be added, got %d", len(spy.validators))
	}

	err := spy.validators[0].Validate(t, testRunnerContext(t), nil)
	if err == nil || !strings.Contains(err.Error(), "validation error") {
		t.Fatalf("unexpected validation error: %v", err)
	}
}
