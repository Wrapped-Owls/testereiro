package webapi

import (
	"net/http"
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest/pkg/atores/netoche"
)

func TestIndieGames(t *testing.T) {
	engine := NewEngine(t)

	// Use reqrunner to verify the API
	mr := netoche.New(
		engine.BaseURL(),
		netoche.WithRequest(http.MethodGet, "/games", struct{}{}),
		netoche.ExpectStatus(http.StatusOK),
		netoche.ExpectBody(Games),
	)

	if err := engine.Execute(t, mr); err != nil {
		t.Fatalf("HttpRunner failed: %v", err)
	}
}
