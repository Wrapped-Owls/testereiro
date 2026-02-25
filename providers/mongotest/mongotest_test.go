package mongotest

import (
	"testing"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/wrapped-owls/testereiro/puppetest"
)

func TestNewMongoRunnerFromEngine(t *testing.T) {
	engine := new(puppetest.Engine)
	database := new(mongo.Database)
	if err := puppetest.SetProvider(engine, databaseProviderKey, database, nil); err != nil {
		t.Fatalf("failed to set test database provider: %v", err)
	}

	runner, err := NewMongoRunnerFromEngine(engine)
	if err != nil {
		t.Fatalf("expected mongo runner helper to succeed, got %v", err)
	}
	if runner == nil {
		t.Fatal("expected runner to be created")
	}
}

func TestNewMongoRunnerFromEngine_MissingDatabase(t *testing.T) {
	_, err := NewMongoRunnerFromEngine(new(puppetest.Engine))
	if err == nil {
		t.Fatal("expected missing database to fail")
	}
}
