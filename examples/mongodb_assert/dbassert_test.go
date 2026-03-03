package mongodb_assert_test

import (
	"fmt"
	"net/http"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/wrapped-owls/testereiro/examples/mongodb_assert/dungeonstore"
	"github.com/wrapped-owls/testereiro/examples/mongodb_assert/fixtures"
	"github.com/wrapped-owls/testereiro/providers/mongotestage"
	"github.com/wrapped-owls/testereiro/providers/mongotestage/pkg/mongochecker"
	"github.com/wrapped-owls/testereiro/providers/mongotestage/pkg/mongoseeder"
	"github.com/wrapped-owls/testereiro/puppetest"
	"github.com/wrapped-owls/testereiro/puppetest/pkg/atores/netoche"
)

// seedDungeonformers wraps each struct with _identity before seeding.
func seedDungeonformers() []any {
	return []any{
		dungeonstore.MustWrapForSeed(fixtures.NewBumblelf()),
		dungeonstore.MustWrapForSeed(fixtures.NewOptimadinPrime()),
		dungeonstore.MustWrapForSeed(fixtures.NewIronknight()),
		dungeonstore.MustWrapForSeed(fixtures.NewRatcheric()),
		dungeonstore.MustWrapForSeed(fixtures.NewJazogue()),
		dungeonstore.MustWrapForSeed(fixtures.NewWheelificer()),
		dungeonstore.MustWrapForSeed(fixtures.NewMegadwarfTron()),
		dungeonstore.MustWrapForSeed(fixtures.NewSorcerscream()),
		dungeonstore.MustWrapForSeed(fixtures.NewSoundbard()),
		dungeonstore.MustWrapForSeed(fixtures.NewShocklock()),
	}
}

func TestDungeonformers_ListAll(t *testing.T) {
	engine := NewEngine(t)

	err := engine.SeedWithProvider(
		mongoseeder.WithClearAndSeed(dungeonstore.CollectionName, seedDungeonformers()...),
	)
	if err != nil {
		t.Fatal(err)
	}

	mr := netoche.New(
		engine.BaseURL(),
		netoche.WithRequest(http.MethodGet, "/dungeonformers", netoche.NoBody{}),
		netoche.ExpectStatus(http.StatusOK),
		netoche.ExpectBodyWithComparator(
			[]bson.M{},
			func(t testing.TB, _ []bson.M, actual []bson.M) bool {
				if len(actual) != 10 {
					t.Errorf("expected 10 dungeonformers, got %d", len(actual))
					return false
				}
				return true
			},
		),
	)

	if err = engine.Execute(t, mr); err != nil {
		t.Fatalf("HttpRunner failed: %v", err)
	}
}

func TestDungeonformers_FilterByClass(t *testing.T) {
	engine := NewEngine(t)

	err := engine.SeedWithProvider(
		mongoseeder.WithClearAndSeed(dungeonstore.CollectionName, seedDungeonformers()...),
	)
	if err != nil {
		t.Fatal(err)
	}

	mr := netoche.New(
		engine.BaseURL(),
		netoche.WithRequest(http.MethodGet, "/dungeonformers/class/{class}", netoche.NoBody{}),
		netoche.WithPathParam("class", "Paladin Commander"),
		netoche.ExpectStatus(http.StatusOK),
		netoche.ExpectBodyWithComparator(
			[]bson.M{},
			func(t testing.TB, _ []bson.M, actual []bson.M) bool {
				if len(actual) != 1 {
					t.Errorf("expected 1 Paladin Commander, got %d", len(actual))
					return false
				}
				name, _ := actual[0]["name"].(string)
				if name != "OptimadinPrime" {
					t.Errorf("expected name OptimadinPrime, got %s", name)
					return false
				}
				return true
			},
		),
	)

	if err = engine.Execute(t, mr); err != nil {
		t.Fatalf("HttpRunner failed: %v", err)
	}
}

func TestDungeonformers_MongoDirectQuery_Identity(t *testing.T) {
	engine := NewEngine(t)

	err := engine.SeedWithProvider(
		mongoseeder.WithClearAndSeed(dungeonstore.CollectionName, seedDungeonformers()...),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Query by _identity to verify the discriminator was stored
	queryRunner, err := mongotestage.NewMongoRunnerFromEngine(
		engine,
		mongochecker.WithFindOneQuery(
			dungeonstore.CollectionName,
			bson.M{"_identity": "MegadwarfTron"},
		),
		mongochecker.ExpectCount(1),
		mongochecker.WithCustomValidation(
			func(_ testing.TB, ctx puppetest.Context, cursor *mongochecker.Cursor) error {
				docs, err := mongochecker.DecodeAll[bson.M](ctx, cursor)
				if err != nil {
					return err
				}
				if len(docs) != 1 {
					return fmt.Errorf("expected 1 doc, got %d", len(docs))
				}
				doc := docs[0]

				name, _ := doc["name"].(string)
				if name != "MegadwarfTron" {
					return fmt.Errorf("expected name MegadwarfTron, got %s", name)
				}

				class, _ := doc["class"].(string)
				if class != "Dwarven Warlord" {
					return fmt.Errorf("expected class Dwarven Warlord, got %s", class)
				}

				// Verify unique fields are stored
				battleCry, _ := doc["battle_cry"].(string)
				if battleCry == "" {
					return fmt.Errorf("expected battle_cry to be set")
				}

				return nil
			},
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	if err = engine.Execute(t, queryRunner); err != nil {
		t.Fatal(err)
	}
}

func TestDungeonformers_UniqueFieldsPreserved(t *testing.T) {
	engine := NewEngine(t)

	err := engine.SeedWithProvider(
		mongoseeder.WithClearAndSeed(dungeonstore.CollectionName, seedDungeonformers()...),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Check that Wheelificer's inventions array was stored correctly
	queryRunner, err := mongotestage.NewMongoRunnerFromEngine(
		engine,
		mongochecker.WithFindOneQuery(
			dungeonstore.CollectionName,
			bson.M{"_identity": "Wheelificer"},
		),
		mongochecker.ExpectCount(1),
		mongochecker.WithCustomValidation(
			func(_ testing.TB, ctx puppetest.Context, cursor *mongochecker.Cursor) error {
				docs, err := mongochecker.DecodeAll[bson.M](ctx, cursor)
				if err != nil {
					return err
				}
				if len(docs) == 0 {
					return fmt.Errorf("expected Wheelificer doc, got none")
				}
				doc := docs[0]

				workshopName, _ := doc["workshop_name"].(string)
				if workshopName != "The Spark Forge" {
					return fmt.Errorf(
						"expected workshop_name 'The Spark Forge', got %q",
						workshopName,
					)
				}

				inventionsRaw, ok := doc["inventions"].(bson.A)
				if !ok {
					return fmt.Errorf(
						"expected inventions to be an array, got %T",
						doc["inventions"],
					)
				}
				if len(inventionsRaw) != 3 {
					return fmt.Errorf("expected 3 inventions, got %d", len(inventionsRaw))
				}

				return nil
			},
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	if err = engine.Execute(t, queryRunner); err != nil {
		t.Fatal(err)
	}
}
