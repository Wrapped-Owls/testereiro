package sqlite_test

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/wrapped-owls/examples/sqlite"
	"github.com/wrapped-owls/examples/sqlite/migrations"
	"github.com/wrapped-owls/testereiro/puppetest"
	"github.com/wrapped-owls/testereiro/puppetest/pkg/runners/reqrunner"
)

var NewEngine func(t testing.TB) *puppetest.Engine

func SQLitePerformer(_ context.Context, conf puppetest.DBConnectionConfig) (*sql.DB, error) {
	dsn := ":memory:"
	if conf.DBName != "" {
		dsn = fmt.Sprintf("file:%s?mode=memory&cache=shared", conf.DBName)
	}
	return sql.Open("sqlite", dsn)
}

func TestMain(m *testing.M) {
	engineFactory, err := puppetest.NewEngineFactory(
		puppetest.WithConnectionFactory(SQLitePerformer, false),
		puppetest.WithExtensions(
			puppetest.WithMigrationRunner(migrations.MigrationFS()),
			puppetest.WithTestServerFromEngine(func(e *puppetest.Engine) (http.Handler, error) {
				return sqlite.NewHandler(e.DB()), nil
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

func TestSQLiteIndieGames(t *testing.T) {
	engine := NewEngine(t)

	type Game sqlite.IndieGame // Create a new type to allow the engine to use the correct table name
	seedObject := Game{
		ID:          1,
		Title:       "Hollow Knight",
		Developer:   "Team Cherry",
		ReleaseYear: 2017,
	}
	if err := engine.Seed(seedObject); err != nil {
		t.Fatal(err)
	}

	// Use reqrunner to verify the API
	mr := reqrunner.NewHttpRunner(
		engine.BaseURL(),
		reqrunner.WithRequest(http.MethodGet, "/games", struct{}{}),
		reqrunner.ExpectStatus(http.StatusOK),
		reqrunner.ExpectBody([]sqlite.IndieGame{
			{
				ID:          seedObject.ID,
				Title:       seedObject.Title,
				Developer:   seedObject.Developer,
				ReleaseYear: seedObject.ReleaseYear,
			},
		}),
	)

	if err := engine.Execute(t, mr); err != nil {
		t.Fatalf("HttpRunner failed: %v", err)
	}
}
