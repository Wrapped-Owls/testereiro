package mongoseeder

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/wrapped-owls/testereiro/providers/mongotest"
	"github.com/wrapped-owls/testereiro/puppetest"
)

type SeedOperationMode uint8

const (
	SeedModeInsertMany SeedOperationMode = iota
	SeedModeClientBulkWrite
)

type SeedPlan struct {
	Collection string
	Documents  []any
}

type SeedRunner struct {
	clearBefore bool
	ordered     bool
	mode        SeedOperationMode
	plans       []SeedPlan
}

var _ puppetest.SeedProvider = (*SeedRunner)(nil)

func New() *SeedRunner {
	return &SeedRunner{
		clearBefore: true,
		ordered:     true,
		mode:        SeedModeInsertMany,
	}
}

func WithSeedDocuments(collection string, docs ...any) *SeedRunner {
	return New().WithSeedDocuments(collection, docs...)
}

func WithClearAndSeed(collection string, docs ...any) *SeedRunner {
	return New().WithClearAndSeed(collection, docs...)
}

func (r *SeedRunner) WithSeedDocuments(collection string, docs ...any) *SeedRunner {
	r.plans = append(r.plans, SeedPlan{
		Collection: collection,
		Documents:  docs,
	})
	return r
}

func (r *SeedRunner) WithClearAndSeed(collection string, docs ...any) *SeedRunner {
	r.clearBefore = true
	return r.WithSeedDocuments(collection, docs...)
}

func (r *SeedRunner) WithClearBeforeSeed(clearBefore bool) *SeedRunner {
	r.clearBefore = clearBefore
	return r
}

func (r *SeedRunner) WithOrderedInsert(ordered bool) *SeedRunner {
	r.ordered = ordered
	return r
}

func (r *SeedRunner) WithSeedOperationMode(mode SeedOperationMode) *SeedRunner {
	r.mode = mode
	return r
}

func (r *SeedRunner) WithInsertManySeedMode() *SeedRunner {
	return r.WithSeedOperationMode(SeedModeInsertMany)
}

func (r *SeedRunner) WithClientBulkWriteSeedMode() *SeedRunner {
	return r.WithSeedOperationMode(SeedModeClientBulkWrite)
}

func (r *SeedRunner) ExecuteSeed(engine *puppetest.Engine) error {
	if engine == nil {
		return fmt.Errorf("engine is nil")
	}

	database, err := mongotest.DatabaseFromEngine(engine)
	if err != nil {
		return err
	}
	if database == nil {
		return fmt.Errorf("mongo database is nil")
	}
	return r.seed(engine.Context(), database)
}

func (r *SeedRunner) seed(ctx context.Context, db *mongo.Database) error {
	if db == nil {
		return fmt.Errorf("mongo database is nil")
	}
	if r.clearBefore {
		if err := db.Drop(ctx); err != nil {
			return fmt.Errorf("failed to drop database %s: %w", db.Name(), err)
		}
	}

	switch r.mode {
	case SeedModeClientBulkWrite:
		return r.runClientBulkWriteMode(ctx, db)
	default:
		return r.runInsertManyMode(ctx, db)
	}
}

func (r *SeedRunner) runInsertManyMode(ctx context.Context, db *mongo.Database) error {
	for _, plan := range r.plans {
		if plan.Collection == "" {
			return fmt.Errorf("seed collection name is required")
		}

		if len(plan.Documents) == 0 {
			continue
		}

		collection := db.Collection(plan.Collection)
		docs := make([]any, 0, len(plan.Documents))
		for _, doc := range plan.Documents {
			docs = append(docs, doc)
		}

		_, err := collection.InsertMany(ctx, docs, options.InsertMany().SetOrdered(r.ordered))
		if err != nil {
			return fmt.Errorf("failed to insert documents into %s: %w", plan.Collection, err)
		}
	}
	return nil
}

func (r *SeedRunner) runClientBulkWriteMode(ctx context.Context, db *mongo.Database) error {
	var writes []mongo.ClientBulkWrite
	for _, plan := range r.plans {
		if plan.Collection == "" {
			return fmt.Errorf("seed collection name is required")
		}
		for _, document := range plan.Documents {
			writes = append(writes, mongo.ClientBulkWrite{
				Database:   db.Name(),
				Collection: plan.Collection,
				Model:      mongo.NewClientInsertOneModel().SetDocument(document),
			})
		}
	}

	if len(writes) == 0 {
		return nil
	}

	_, err := db.Client().BulkWrite(ctx, writes, options.ClientBulkWrite().SetOrdered(r.ordered))
	if err != nil {
		return fmt.Errorf("failed to execute mongo client bulk write seed: %w", err)
	}
	return nil
}
