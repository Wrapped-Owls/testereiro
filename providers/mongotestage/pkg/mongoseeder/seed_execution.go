package mongoseeder

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/wrapped-owls/testereiro/providers/mongotestage"
	"github.com/wrapped-owls/testereiro/puppetest"
)

// ExecuteSeed resolves the mongo database from engine providers and applies configured seed plans.
func (r *SeedRunner) ExecuteSeed(engine *puppetest.Engine) error {
	if engine == nil {
		return fmt.Errorf("engine is nil")
	}

	database, err := mongotestage.DatabaseFromEngine(engine)
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
		docs = append(docs, plan.Documents...)

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
