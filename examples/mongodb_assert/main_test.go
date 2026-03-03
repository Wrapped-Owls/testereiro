package mongodb_assert_test

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/wrapped-owls/testereiro/examples/mongodb_assert"
	"github.com/wrapped-owls/testereiro/providers/mongotestage"
	"github.com/wrapped-owls/testereiro/puppetest"
)

var NewEngine func(t testing.TB) *puppetest.Engine

func TestMain(m *testing.M) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "mongo:8",
			ExposedPorts: []string{"27017/tcp"},
			WaitingFor: wait.ForLog("Waiting for connections").
				WithStartupTimeout(2 * time.Minute),
		},
		Started: true,
	}

	container, err := testcontainers.GenericContainer(context.Background(), req)
	if err != nil {
		slog.Error("failed to start container", slog.String("error", err.Error()))
		os.Exit(1)
	}

	connCfg, err := mongoConnectionFromContainer(container)
	if err != nil {
		slog.Error("failed to build mongo config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	engineFactory, err := puppetest.NewEngineFactory(
		puppetest.WithAfterFactoryClose(func(_ *puppetest.FactoryCloseEvent) error {
			return containerHook{container: container}.Close()
		}),
		mongotestage.WithMongoConnection(connCfg),
		puppetest.WithExtensions(
			puppetest.WithTestServerFromEngine(func(e *puppetest.Engine) (http.Handler, error) {
				db, err := mongotestage.DatabaseFromEngine(e)
				if err != nil {
					return nil, err
				}
				return mongodb_assert.NewHandler(db), nil
			}),
		),
	)
	if err != nil {
		slog.Error("failed to setup engine factory", slog.String("error", err.Error()))
		os.Exit(1)
	}

	NewEngine = engineFactory.NewEngine
	code := m.Run()

	if err = engineFactory.Close(); err != nil {
		slog.Error("failed to close engine factory", slog.String("error", err.Error()))
		os.Exit(1)
	}

	os.Exit(code)
}

type containerHook struct {
	container testcontainers.Container
}

func (o containerHook) Close() error {
	if o.container == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return o.container.Terminate(ctx)
}

func mongoConnectionFromContainer(
	container testcontainers.Container,
) (mongotestage.ConnectionConfig, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	port, err := container.MappedPort(ctx, "27017")
	if err != nil {
		return mongotestage.ConnectionConfig{}, fmt.Errorf(
			"failed to read mapped mongo port: %w",
			err,
		)
	}
	host, err := container.Host(ctx)
	if err != nil {
		return mongotestage.ConnectionConfig{}, fmt.Errorf(
			"failed to read mongo container host: %w",
			err,
		)
	}
	portNum, err := strconv.Atoi(port.Port())
	if err != nil {
		return mongotestage.ConnectionConfig{}, fmt.Errorf(
			"failed to parse mongo mapped port: %w",
			err,
		)
	}

	return mongotestage.ConnectionConfig{
		Host: host,
		Port: portNum,
	}, nil
}
