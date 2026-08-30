package dbastidor

import "testing"

func TestPlaceholderStyles(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		style PlaceholderStyle
		want  []string
	}{
		{
			name:  "question marks ignore the index",
			style: QuestionPlaceholder,
			want:  []string{"?", "?", "?"},
		},
		{
			name:  "ordinals count from one",
			style: OrdinalPlaceholder,
			want:  []string{"$1", "$2", "$3"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			for index, want := range testCase.want {
				if got := testCase.style(index + 1); got != want {
					t.Fatalf("index %d: expected %q, got %q", index+1, want, got)
				}
			}
		})
	}
}

func TestOrdinalPlaceholderPastNine(t *testing.T) {
	t.Parallel()

	// Guards against a byte-arithmetic shortcut that would render 10 as ":".
	if got := OrdinalPlaceholder(10); got != "$10" {
		t.Fatalf("expected %q, got %q", "$10", got)
	}
}
