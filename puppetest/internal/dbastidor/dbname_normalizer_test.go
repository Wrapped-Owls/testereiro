package dbastidor

import "testing"

func TestNormalizeDBName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "lowercases letters",
			in:   "MyDatabase",
			want: "mydatabase",
		},
		{
			name: "replaces separators and punctuation",
			in:   "my-db name.v1",
			want: "my_db_name_v1",
		},
		{
			name: "keeps unicode letters and numbers",
			in:   "Árvore４2",
			want: "árvore４2",
		},
		{
			name: "replaces emoji with underscore",
			in:   "db🔥name",
			want: "db_name",
		},
		{
			name: "empty string remains empty",
			in:   "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeDBName(tt.in); got != tt.want {
				t.Fatalf("expected normalized name %q, got %q", tt.want, got)
			}
		})
	}
}
