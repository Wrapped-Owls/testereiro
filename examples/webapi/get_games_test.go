package webapi

import (
	"net/http"
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest/pkg/runners/reqrunner"
)

func TestIndieGames(t *testing.T) {
	engine := NewEngine(t)

	// Use reqrunner to verify the API
	mr := reqrunner.NewHttpRunner(
		engine.BaseURL(),
		reqrunner.WithRequest(http.MethodGet, "/games", struct{}{}),
		reqrunner.ExpectStatus(http.StatusOK),
		reqrunner.ExpectBody(Games),
	)

	if err := engine.Execute(t, mr); err != nil {
		t.Fatalf("HttpRunner failed: %v", err)
	}
}
