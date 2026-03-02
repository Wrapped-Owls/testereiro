package mongoseeder

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/wrapped-owls/testereiro/providers/mongotest"
	"github.com/wrapped-owls/testereiro/puppetest"
)

// SeedOperationMode selects how documents are inserted during seeding.
type SeedOperationMode uint8

const (
	// SeedModeInsertMany inserts documents per collection with InsertMany.
	SeedModeInsertMany SeedOperationMode = iota
	// SeedModeClientBulkWrite inserts documents using client-level bulk writes.
	SeedModeClientBulkWrite
)

// SeedPlan describes one collection and its documents to seed.
type SeedPlan struct {
	Collection string
	Documents  []any
}

// SeedRunner is a configurable mongo seed provider for puppetest engines.
type SeedRunner struct {
	clearBefore bool
	ordered     bool
	mode        SeedOperationMode
	plans       []SeedPlan
}

var _ puppetest.SeedProvider = (*SeedRunner)(nil)

// New creates a SeedRunner with default behavior.
func New() *SeedRunner {
	return &SeedRunner{
		clearBefore: true,
		ordered:     true,
		mode:        SeedModeInsertMany,
	}
}

// WithSeedDocuments creates a SeedRunner and appends a seed plan.
func WithSeedDocuments(collection string, docs ...any) *SeedRunner {
	return New().WithSeedDocuments(collection, docs...)
}

// WithClearAndSeed creates a SeedRunner that clears DB before seeding and appends a plan.
func WithClearAndSeed(collection string, docs ...any) *SeedRunner {
	return New().WithClearAndSeed(collection, docs...)
}

// WithSeedDocuments appends a seed plan for one collection.
func (r *SeedRunner) WithSeedDocuments(collection string, docs ...any) *SeedRunner {
	r.plans = append(r.plans, SeedPlan{
		Collection: collection,
		Documents:  docs,
	})
	return r
}

// WithClearAndSeed enables clear-before mode and appends a seed plan.
func (r *SeedRunner) WithClearAndSeed(collection string, docs ...any) *SeedRunner {
	r.clearBefore = true
	return r.WithSeedDocuments(collection, docs...)
}

// WithClearBeforeSeed controls whether the database is dropped before inserts.
func (r *SeedRunner) WithClearBeforeSeed(clearBefore bool) *SeedRunner {
	r.clearBefore = clearBefore
	return r
}

// WithOrderedInsert controls ordered semantics for insert operations.
func (r *SeedRunner) WithOrderedInsert(ordered bool) *SeedRunner {
	r.ordered = ordered
	return r
}

// WithSeedOperationMode sets the insertion strategy used by ExecuteSeed.
func (r *SeedRunner) WithSeedOperationMode(mode SeedOperationMode) *SeedRunner {
	r.mode = mode
	return r
}

// WithInsertManySeedMode selects SeedModeInsertMany.
func (r *SeedRunner) WithInsertManySeedMode() *SeedRunner {
	return r.WithSeedOperationMode(SeedModeInsertMany)
}

// WithClientBulkWriteSeedMode selects SeedModeClientBulkWrite.
func (r *SeedRunner) WithClientBulkWriteSeedMode() *SeedRunner {
	return r.WithSeedOperationMode(SeedModeClientBulkWrite)
}

// ExecuteSeed resolves the mongo database from engine providers and applies configured seed plans.
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
