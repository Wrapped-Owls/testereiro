package mongochecker

import (
	"fmt"
	"maps"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/wrapped-owls/testereiro/puppetest"
)

type QueryBuilder interface {
	Build(ctx puppetest.Context) (Query, error)
}

type filterFromContext func(ctx puppetest.Context) (bson.M, error)

type BsonQueryBuilder struct {
	collection   string
	operation    Operation
	filters      []filterFromContext
	queryOptions any
}

func NewBsonQuery(collection string, filter bson.M) *BsonQueryBuilder {
	builder := &BsonQueryBuilder{collection: collection, operation: OpFind}
	if len(filter) > 0 {
		builder.AddFilter(func(_ puppetest.Context) (bson.M, error) {
			return filter, nil
		})
	}
	return builder
}

func (b *BsonQueryBuilder) AddFilter(filter filterFromContext) {
	b.filters = append(b.filters, filter)
}

func (b *BsonQueryBuilder) SetOperation(op Operation) {
	b.operation = op
}

func (b *BsonQueryBuilder) SetOptions(opts any) {
	b.queryOptions = opts
}

func (b *BsonQueryBuilder) Build(ctx puppetest.Context) (Query, error) {
	if b.collection == "" {
		return Query{}, fmt.Errorf("collection is required")
	}

	finalFilter := bson.M{}
	for _, filter := range b.filters {
		resolved, err := filter(ctx)
		if err != nil {
			return Query{}, fmt.Errorf("failed to resolve query filter: %w", err)
		}
		maps.Copy(finalFilter, resolved)
	}

	return Query{
		Collection: b.collection,
		Operation:  b.operation,
		Filter:     finalFilter,
		Options:    b.queryOptions,
	}, nil
}

type pipelineFromContext func(ctx puppetest.Context) (bson.A, error)

type AggregateQueryBuilder struct {
	collection   string
	pipelines    []pipelineFromContext
	queryOptions any
}

func NewAggregateQuery(collection string, pipeline bson.A) *AggregateQueryBuilder {
	builder := &AggregateQueryBuilder{
		collection: collection,
	}
	if len(pipeline) > 0 {
		builder.AddPipeline(func(_ puppetest.Context) (bson.A, error) {
			return pipeline, nil
		})
	}
	return builder
}

func (a *AggregateQueryBuilder) AddPipeline(pipeline pipelineFromContext) {
	a.pipelines = append(a.pipelines, pipeline)
}

func (a *AggregateQueryBuilder) SetOptions(opts any) {
	a.queryOptions = opts
}

func (a *AggregateQueryBuilder) Build(ctx puppetest.Context) (Query, error) {
	if a.collection == "" {
		return Query{}, fmt.Errorf("collection is required")
	}

	finalPipeline := bson.A{}
	for _, pipeline := range a.pipelines {
		resolved, err := pipeline(ctx)
		if err != nil {
			return Query{}, fmt.Errorf("failed to resolve query pipeline: %w", err)
		}
		finalPipeline = append(finalPipeline, resolved...)
	}

	return Query{
		Collection: a.collection,
		Operation:  OpAggregate,
		Pipeline:   finalPipeline,
		Options:    a.queryOptions,
	}, nil
}
