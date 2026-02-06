package testcontainers_mysql_test

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/vinovest/sqlx"

	"github.com/wrapped-owls/testereiro/examples/balatro_mysql/db/daos"
	"github.com/wrapped-owls/testereiro/puppetest/pkg/runners/dbrunner"
)

func TestDatabaseOnlyAssertion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	engine := NewEngine(t)

	seedJokers := [...]daos.Joker{
		{
			Name:   "Gros Michel",
			Effect: "+15 Mult. 1 in 6 chance to destroy this card at end of round.",
			Rarity: "Common",
		},
		{
			Name:   "Cavendish",
			Effect: "X3 Mult. 1 in 1000 chance to destroy this card at end of round.",
			Rarity: "Common",
		},
		{Name: "Blueprint", Effect: "Copies ability of Joker to the right.", Rarity: "Rare"},
	}
	err := engine.Seed(seedJokers[0], seedJokers[1], seedJokers[2])
	assert.NoError(t, err)

	runner := dbrunner.NewDbRunner(
		engine.DB(),
		dbrunner.WithMapQuery("jokers", map[string]any{"rarity": "Common"}),
		// dbrunner.WithQuery(dbrunner.NewRawQuery("SELECT * FROM jokers WHERE rarity = ?", "Common")),
		dbrunner.WithCustomValidation(func(t testing.TB, rows *sql.Rows) error {
			var jokerList []daos.Joker
			if err = sqlx.StructScan(rows, &jokerList); err != nil {
				return fmt.Errorf("could not scan joker rows: %v", err)
			}

			expectedJokers := seedJokers[:2]
			assert.Equal(t, len(expectedJokers), len(jokerList))
			for index := range jokerList {
				// Clean up the ID to allow asserting it
				jokerList[index].ID = 0
			}
			assert.Equal(t, expectedJokers, jokerList)

			return nil
		}),
	)
	err = engine.Execute(t, runner)
	assert.NoError(t, err)
	t.Log("Tests executed successfully")
}
