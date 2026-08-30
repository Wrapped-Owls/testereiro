package puppetest

import "testing"

func TestReExportedPlaceholderStyles(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		style PlaceholderStyle
		index int
		want  string
	}{
		{name: "question mark", style: QuestionPlaceholder, index: 1, want: "?"},
		{name: "ordinal first", style: OrdinalPlaceholder, index: 1, want: "$1"},
		{name: "ordinal past nine", style: OrdinalPlaceholder, index: 12, want: "$12"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got := testCase.style(testCase.index); got != testCase.want {
				t.Fatalf("expected %q, got %q", testCase.want, got)
			}
		})
	}
}
