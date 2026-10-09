package dbastidor

import (
	"strings"
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest/internal/testkit"
)

type seedJoker struct {
	Name   string `db:"name"`
	Rarity string `db:"rarity"`
	Hidden string `db:"-"`
	NoTag  string
}

type seedEmpty struct {
	NoTag string
}

type seedInbox struct {
	Card string `db:"card"`
}

type seedCategory struct {
	Card string `db:"card"`
}

type seedKey struct {
	Card string `db:"card"`
}

type seedDeck struct {
	Card string `db:"card"`
}

func (seedDeck) TableName() string {
	return "joker_deck"
}

func TestExecuteSeedStruct(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		item        any
		placeholder PlaceholderStyle
		wantStmt    string
		wantErrText string
	}{
		{
			name:     "nil style defaults to question marks",
			item:     seedJoker{Name: "Jimbo", Rarity: "Common"},
			wantStmt: "INSERT INTO seed_jokers (name, rarity) VALUES (?, ?)",
		},
		{
			name:        "question marks stay positional-free",
			item:        seedJoker{Name: "Jimbo", Rarity: "Common"},
			placeholder: QuestionPlaceholder,
			wantStmt:    "INSERT INTO seed_jokers (name, rarity) VALUES (?, ?)",
		},
		{
			name:        "ordinals number each column from one",
			item:        seedJoker{Name: "Jimbo", Rarity: "Common"},
			placeholder: OrdinalPlaceholder,
			wantStmt:    "INSERT INTO seed_jokers (name, rarity) VALUES ($1, $2)",
		},
		{
			name:        "pointer to struct is dereferenced",
			item:        &seedJoker{Name: "Jimbo", Rarity: "Common"},
			placeholder: OrdinalPlaceholder,
			wantStmt:    "INSERT INTO seed_jokers (name, rarity) VALUES ($1, $2)",
		},
		{
			name:        "a name ending in x is already plural",
			item:        seedInbox{Card: "Joker"},
			placeholder: OrdinalPlaceholder,
			wantStmt:    "INSERT INTO seed_inbox (card) VALUES ($1)",
		},
		{
			name:        "a consonant before the final y takes ies",
			item:        seedCategory{Card: "Joker"},
			placeholder: OrdinalPlaceholder,
			wantStmt:    "INSERT INTO seed_categories (card) VALUES ($1)",
		},
		{
			name:        "a vowel before the final y takes s",
			item:        seedKey{Card: "Joker"},
			placeholder: OrdinalPlaceholder,
			wantStmt:    "INSERT INTO seed_keys (card) VALUES ($1)",
		},
		{
			name:        "an explicit table name wins over the type name",
			item:        seedDeck{Card: "Joker"},
			placeholder: OrdinalPlaceholder,
			wantStmt:    "INSERT INTO joker_deck (card) VALUES ($1)",
		},
		{
			name:        "an explicit table name is read through a pointer",
			item:        &seedDeck{Card: "Joker"},
			placeholder: OrdinalPlaceholder,
			wantStmt:    "INSERT INTO joker_deck (card) VALUES ($1)",
		},
		{
			name:        "non-struct is rejected",
			item:        "not a struct",
			wantErrText: "expected struct or pointer to struct",
		},
		{
			name:        "struct without db tags is rejected",
			item:        seedEmpty{},
			wantErrText: "no database fields found",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			state := &testkit.SQLState{}
			db := testkit.OpenStubDB(t, t.Name(), state)
			t.Cleanup(func() { _ = db.Close() })

			err := ExecuteSeedStruct(db, testCase.item, testCase.placeholder)
			if testCase.wantErrText != "" {
				if err == nil || !strings.Contains(err.Error(), testCase.wantErrText) {
					t.Fatalf("expected error containing %q, got %v", testCase.wantErrText, err)
				}
				if got := state.ExecStatements(); len(got) != 0 {
					t.Fatalf("expected no statement to run, got %v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			got := state.ExecStatements()
			if len(got) != 1 || got[0] != testCase.wantStmt {
				t.Fatalf("expected statement %q, got %v", testCase.wantStmt, got)
			}
		})
	}
}
