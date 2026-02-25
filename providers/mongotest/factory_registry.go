package mongotest

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/wrapped-owls/testereiro/puppetest"
)

type factoryResource struct {
	client            *mongo.Client
	disconnectTimeout time.Duration
	ctx               context.Context
}

var factoryResourceKey = puppetest.NewTaggedProviderKey[factoryResource](
	"mongotest.factory.resource",
)

func storeFactoryResource(factory *puppetest.EngineFactory, resource *factoryResource) bool {
	registerErr := puppetest.RegisterFactoryProvider(
		factory, factoryResourceKey, resource,
		mongoDbBinder, teardownMongoDB,
	)

	return registerErr == nil
}

func mongoDbBinder(
	_ context.Context, engine *puppetest.Engine, stored *factoryResource,
) error {
	if engine == nil {
		return fmt.Errorf("engine is nil")
	}
	if stored == nil || stored.client == nil {
		return fmt.Errorf("mongo factory resource is nil")
	}

	databaseName := engine.DBName()
	database := stored.client.Database(databaseName)
	if err := puppetest.SetProvider(engine, databaseProviderKey, database, nil); err != nil {
		return fmt.Errorf("failed to bind mongo database into engine providers: %w", err)
	}
	return nil
}

func teardownMongoDB(parentCtx context.Context, stored *factoryResource) error {
	if stored == nil || stored.client == nil {
		return nil
	}

	baseCtx := parentCtx
	if baseCtx == nil {
		if baseCtx = stored.ctx; baseCtx == nil {
			baseCtx = context.Background()
		}
	}

	ctx, cancel := context.WithTimeout(baseCtx, stored.disconnectTimeout)
	defer cancel()
	return stored.client.Disconnect(ctx)
}

func loadFactoryResource(factory *puppetest.EngineFactory) (*factoryResource, bool) {
	return puppetest.FactoryProvider[factoryResource](factory, factoryResourceKey)
}
