package mongotest

import (
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/wrapped-owls/testereiro/providers/mongotest/pkg/runners/mongorunner"
	"github.com/wrapped-owls/testereiro/puppetest"
)

var databaseProviderKey = puppetest.NewTaggedProviderKey[mongo.Database]("mongo.database.resource")

func NewMongoRunnerFromEngine(
	engine *puppetest.Engine,
	opts ...mongorunner.Option,
) (*mongorunner.MongoRunner, error) {
	database, err := DatabaseFromEngine(engine)
	if err != nil {
		return nil, err
	}
	if database == nil {
		return nil, fmt.Errorf("mongo database is nil")
	}

	return mongorunner.NewMongoRunner(database, opts...), nil
}
