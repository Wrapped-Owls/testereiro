package mongotestage

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/wrapped-owls/testereiro/puppetest"
)

func TestWithMongoClient_RegistersFactoryResourceAndBindsDatabase(t *testing.T) {
	factory, err := puppetest.NewEngineFactory(
		WithMongoClient(new(mongo.Client)),
	)
	if err != nil {
		t.Fatalf("failed to create engine factory: %v", err)
	}

	engine := factory.NewEngine(t)
	db, err := DatabaseFromEngine(engine)
	if err != nil {
		t.Fatalf("expected mongo database to be present on engine provider storage, got %v", err)
	}
	if db == nil {
		t.Fatal("expected mongo database to be non-nil")
	}
	if db.Name() != "testwithmongoclient_registersfactoryresourceandbindsdatabase_puppetest" {
		t.Fatalf(
			"expected database name `testwithmongoclient_registersfactoryresourceandbindsdatabase_puppetest`, got `%s`",
			db.Name(),
		)
	}
}

func TestWithMongoClient_RejectsDuplicateRegistration(t *testing.T) {
	_, err := puppetest.NewEngineFactory(
		WithMongoClient(new(mongo.Client)),
		WithMongoClient(new(mongo.Client)),
	)
	if err == nil {
		t.Fatal("expected duplicate mongo provider registration to fail")
	}
}

func TestWithMongoClient_ValidatesInputs(t *testing.T) {
	if _, err := puppetest.NewEngineFactory(WithMongoClient(nil)); err == nil {
		t.Fatal("expected nil mongo client to fail")
	}

	if _, err := puppetest.NewEngineFactory(WithMongoClient(new(mongo.Client))); err != nil {
		t.Fatalf("expected provided mongo.Client to perform correctly, got %v", err.Error())
	}
}

func TestWithMongoConnection_RequiresDatabase(t *testing.T) {
	_, err := puppetest.NewEngineFactory(WithMongoConnection(ConnectionConfig{}))
	if err == nil {
		t.Fatal("expected missing database to fail")
	}
}

func TestDatabaseFromEngine_NotFoundForEmptyEngine(t *testing.T) {
	engine := new(puppetest.Engine)
	_, err := DatabaseFromEngine(engine)
	if err == nil {
		t.Fatal("expected mongo database to be absent")
	}
}

func TestClientFromFactory_RequiresRegisteredResource(t *testing.T) {
	factory, err := puppetest.NewEngineFactory()
	if err != nil {
		t.Fatalf("failed to create engine factory: %v", err)
	}
	client, err := ClientFromFactory(factory)
	if err == nil || client != nil {
		t.Fatal("expected no client to be registered")
	}
}

func TestDatabaseFromEngine_ReturnsError(t *testing.T) {
	_, err := DatabaseFromEngine(new(puppetest.Engine))
	if err == nil {
		t.Fatal("expected missing database error")
	}
}

func TestClientFromFactory_ReturnsError(t *testing.T) {
	factory, err := puppetest.NewEngineFactory()
	if err != nil {
		t.Fatalf("failed to create engine factory: %v", err)
	}

	_, err = ClientFromFactory(factory)
	if err == nil {
		t.Fatal("expected missing client error")
	}
}

func TestBindFactoryResource_RejectsNilInputs(t *testing.T) {
	err := bindFactoryResource(nil, &factoryResource{client: new(mongo.Client)})
	if err == nil {
		t.Fatal("expected nil factory error")
	}

	factory, facErr := puppetest.NewEngineFactory()
	if facErr != nil {
		t.Fatalf("failed to create engine factory: %v", facErr)
	}
	err = bindFactoryResource(factory, nil)
	if err == nil {
		t.Fatal("expected nil resource error")
	}
}

func TestRegistryStoreLoad(t *testing.T) {
	factory, err := puppetest.NewEngineFactory()
	if err != nil {
		t.Fatalf("failed to create engine factory: %v", err)
	}

	resource := &factoryResource{client: new(mongo.Client)}
	if ok := storeFactoryResource(factory, resource); !ok {
		t.Fatal("expected first store to succeed")
	}
	if ok := storeFactoryResource(factory, resource); ok {
		t.Fatal("expected duplicate store to fail")
	}

	loaded, ok := loadFactoryResource(factory)
	if !ok || loaded != resource {
		t.Fatal("expected load to return stored resource")
	}
}

func TestNormalizeConnectionConfig_Defaults(t *testing.T) {
	cfg, err := (ConnectionConfig{}).normalize()
	if err != nil {
		t.Fatalf("unexpected normalize error: %v", err)
	}
	if cfg.Host != defaultMongoHost {
		t.Fatalf("expected host %s, got %s", defaultMongoHost, cfg.Host)
	}
	if cfg.Port != defaultMongoPort {
		t.Fatalf("expected port %d, got %d", defaultMongoPort, cfg.Port)
	}
	if cfg.Context == nil {
		t.Fatal("expected default context to be set")
	}
}

func TestPingMongoClient_ValidatesInputs(t *testing.T) {
	err := PingMongoClient(context.Background(), nil, time.Second)
	if err == nil {
		t.Fatal("expected nil client validation error")
	}

	err = PingMongoClient(context.Background(), new(mongo.Client), time.Millisecond)
	if err == nil {
		t.Fatal("expected ping failure for zero-value client")
	}
}
