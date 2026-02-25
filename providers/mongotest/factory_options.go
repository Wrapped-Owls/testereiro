package mongotest

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"

	"github.com/wrapped-owls/testereiro/puppetest"
)

func NewMongoClientOptions(cfg *ConnectionConfig) (clientOpts *options.ClientOptions, err error) {
	if cfg == nil {
		return nil, fmt.Errorf("connection config is nil")
	}
	if *cfg, err = cfg.normalize(); err != nil {
		return nil, err
	}

	normalizedCfg := *cfg // Create a temp alias
	clientOpts = options.Client()
	if normalizedCfg.URI == "" {
		normalizedCfg.URI = fmt.Sprintf("mongodb://%s:%d", normalizedCfg.Host, normalizedCfg.Port)
	}

	clientOpts.ApplyURI(normalizedCfg.URI)
	if normalizedCfg.Username != "" {
		clientOpts.SetAuth(options.Credential{
			Username:   normalizedCfg.Username,
			Password:   normalizedCfg.Password,
			AuthSource: normalizedCfg.AuthSource,
		})
	}
	clientOpts.SetServerSelectionTimeout(normalizedCfg.ConnectTimeout)

	return clientOpts, nil
}

func CreateMongoClient(clientOpts *options.ClientOptions) (*mongo.Client, error) {
	if clientOpts == nil {
		return nil, fmt.Errorf("mongo client options are nil")
	}
	client, err := mongo.Connect(clientOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect mongo client: %w", err)
	}
	return client, nil
}

func PingMongoClient(ctx context.Context, client *mongo.Client, timeout time.Duration) error {
	if client == nil {
		return fmt.Errorf("mongo client is nil")
	}

	if ctx == nil {
		ctx = context.Background()
	}
	pingCtx, pingCancel := context.WithTimeout(ctx, timeout)
	defer pingCancel()
	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		return fmt.Errorf("failed to ping mongo server: %w", err)
	}
	return nil
}

func ConnectAndPingMongoClient(
	cfg ConnectionConfig,
	clientOpts *options.ClientOptions,
) (*mongo.Client, error) {
	client, err := CreateMongoClient(clientOpts)
	if err != nil {
		return nil, err
	}

	if err = PingMongoClient(cfg.Context, client, cfg.PingTimeout); err != nil {
		disconnectCtx, cancel := context.WithTimeout(cfg.Context, cfg.DisconnectTimeout)
		defer cancel()
		_ = client.Disconnect(disconnectCtx)
		return nil, err
	}

	return client, nil
}

func WithMongoConnection(cfg ConnectionConfig) puppetest.EngineFactoryOption {
	return func(factory *puppetest.EngineFactory) error {
		clientOpts, err := NewMongoClientOptions(&cfg)
		if err != nil {
			return err
		}

		var client *mongo.Client
		if client, err = ConnectAndPingMongoClient(cfg, clientOpts); err != nil {
			return err
		}

		resource := &factoryResource{
			client:            client,
			disconnectTimeout: cfg.DisconnectTimeout,
			ctx:               cfg.Context,
		}
		if err = bindFactoryResource(factory, resource); err != nil {
			disconnectCtx, cancel := context.WithTimeout(cfg.Context, cfg.DisconnectTimeout)
			defer cancel()
			_ = client.Disconnect(disconnectCtx)
			return err
		}

		return nil
	}
}

func WithMongoClient(client *mongo.Client) puppetest.EngineFactoryOption {
	return func(factory *puppetest.EngineFactory) error {
		if client == nil {
			return fmt.Errorf("mongo client is nil")
		}

		resource := &factoryResource{
			client:            client,
			disconnectTimeout: defaultDisconnectTimeout,
			ctx:               context.Background(),
		}
		return bindFactoryResource(factory, resource)
	}
}

func WithMongoDb(cfg ConnectionConfig) puppetest.EngineFactoryOption {
	return WithMongoConnection(cfg)
}

func bindFactoryResource(factory *puppetest.EngineFactory, resource *factoryResource) error {
	if factory == nil {
		return fmt.Errorf("engine factory is nil")
	}
	if resource == nil || resource.client == nil {
		return fmt.Errorf("mongo factory resource is nil")
	}

	if _, exists := loadFactoryResource(factory); exists {
		return fmt.Errorf("mongo provider already configured for this factory")
	}

	if !storeFactoryResource(factory, resource) {
		return fmt.Errorf("mongo provider already configured for this factory")
	}

	return nil
}
