package webapi

import (
	"log/slog"
	"os"
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest"
)

var NewEngine func(t testing.TB) *puppetest.Engine

func TestMain(m *testing.M) {
	engineFactory, err := puppetest.NewEngineFactory(
		puppetest.WithExtensions(
			puppetest.WithTestServer(NewHandler()),
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
