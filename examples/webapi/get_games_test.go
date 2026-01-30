package webapi

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest"
	"github.com/wrapped-owls/testereiro/puppetest/pkg/runners/reqrunner"
)

var NewEngine func(t testing.TB) *puppetest.Engine

func TestMain(m *testing.M) {
	engineFactory, err := puppetest.NewEngineFactory(
		func(ctx context.Context, conf puppetest.DBConnectionConfig) (*sql.DB, error) {
			// WebAPI example doesn't use DB, but puppetest requires a performer for now
			return nil, nil
		},
		puppetest.WithTestServer(NewHandler()),
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

func TestIndieGames(t *testing.T) {
	engine := NewEngine(t)

	// Use reqrunner to verify the API
	mr := reqrunner.NewHttpRunner(
		engine.BaseURL(),
		reqrunner.WithRequest(http.MethodGet, "/games", struct{}{}),
		reqrunner.ExpectStatus(http.StatusOK),
		reqrunner.ExpectBody(Games),
	)

	if err := mr.Run(t, nil); err != nil {
		t.Fatalf("HttpRunner failed: %v", err)
	}
}
