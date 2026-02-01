package sqlite_test

import (
	"net/http"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/wrapped-owls/examples/sqlite"
	"github.com/wrapped-owls/testereiro/puppetest/pkg/runners/reqrunner"
)

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
