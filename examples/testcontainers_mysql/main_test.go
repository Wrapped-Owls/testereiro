package testcontainers_mysql_test

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/wrapped-owls/testereiro/examples/balatro_mysql/db/migrations"
	"github.com/wrapped-owls/testereiro/puppetest"
)

var NewEngine func(t testing.TB) *puppetest.Engine

func TestMain(m *testing.M) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "mysql:8",
			Env: map[string]string{
				"MYSQL_ROOT_PASSWORD": "jimbo",
				"MYSQL_DATABASE":      "balatro_db",
			},
			ExposedPorts: []string{"3306/tcp"},
			WaitingFor: wait.ForLog("port: 3306  MySQL Community Server").
				WithStartupTimeout(2 * time.Minute),
		},
		Started: true,
	}

	container, err := testcontainers.GenericContainer(context.Background(), req)
	if err != nil {
		slog.Error("failed to start container", slog.String("error", err.Error()))
		os.Exit(1)
	}

	engineFactory, err := puppetest.NewEngineFactory(
		puppetest.WithHook(containerHook{container: container}),
		puppetest.WithConnectionFactory(mysqlPerformerFromContainer(container), true),
		puppetest.WithExtensions(
			puppetest.WithMigrationRunner(migrations.FS()),
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

func mysqlPerformerFromContainer(container testcontainers.Container) puppetest.ConnectionPerformer {
	return func(ctx context.Context, conf puppetest.DBConnectionConfig) (*sql.DB, error) {
		port, err := container.MappedPort(ctx, "3306")
		if err != nil {
			return nil, err
		}
		host, err := container.Host(ctx)
		if err != nil {
			return nil, err
		}

		dbName := conf.DBName
		if dbName == "" {
			dbName = "balatro_db"
		}
		dsn := fmt.Sprintf("root:jimbo@tcp(%s:%s)/%s?parseTime=true", host, port.Port(), dbName)
		if conf.AllowMultiStatements {
			dsn += "&multiStatements=true"
		}
		return sql.Open("mysql", dsn)
	}
}
