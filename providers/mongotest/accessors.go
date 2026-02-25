package mongotest

import (
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/wrapped-owls/testereiro/puppetest"
)

func DatabaseFromEngine(engine *puppetest.Engine) (*mongo.Database, error) {
	if engine == nil {
		return nil, fmt.Errorf("engine is nil")
	}

	database, ok := puppetest.Provider[mongo.Database](engine, databaseProviderKey)
	if !ok || database == nil {
		return nil, fmt.Errorf("mongo database not found in engine provider storage")
	}
	return database, nil
}

func ClientFromFactory(factory *puppetest.EngineFactory) (*mongo.Client, error) {
	resource, ok := loadFactoryResource(factory)
	if !ok || resource == nil || resource.client == nil {
		return nil, fmt.Errorf("mongo client not found in factory resource registry")
	}
	return resource.client, nil
}
