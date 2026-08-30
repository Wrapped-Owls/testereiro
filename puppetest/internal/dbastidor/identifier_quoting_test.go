package dbastidor

import (
	"errors"
	"testing"
)

func TestQuoteIdentifier(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		input   string
		quote   rune
		want    string
		wantErr error
	}{
		{name: "mysql name", input: "games_db", quote: '`', want: "`games_db`"},
		{name: "postgres name", input: "games_db", quote: '"', want: `"games_db"`},
		{
			name:  "doubles an embedded backtick so it cannot close the identifier",
			input: "games`; DROP DATABASE other; --",
			quote: '`',
			want:  "`games``; DROP DATABASE other; --`",
		},
		{
			name:  "doubles an embedded double quote",
			input: `games"; DROP DATABASE other; --`,
			quote: '"',
			want:  `"games""; DROP DATABASE other; --"`,
		},
		{
			name:  "leaves the other engine's quote rune untouched",
			input: "games`db",
			quote: '"',
			want:  "\"games`db\"",
		},
		{name: "empty name", input: "", quote: '"', wantErr: ErrInvalidIdentifier},
		{name: "name with a null byte", input: "games\x00db", quote: '"', wantErr: ErrInvalidIdentifier},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := QuoteIdentifier(testCase.input, testCase.quote)
			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("expected error %v, got %v", testCase.wantErr, err)
			}
			if testCase.wantErr != nil {
				return
			}
			if got != testCase.want {
				t.Fatalf("expected %q, got %q", testCase.want, got)
			}
		})
	}
}
