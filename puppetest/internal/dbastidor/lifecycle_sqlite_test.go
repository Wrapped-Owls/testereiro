package dbastidor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSQLiteLifecycle_Create(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		pathFor DBFilePathResolver
		wantDir bool
	}{
		{
			name:    "creates the directory holding the database file",
			pathFor: func(dbName string) string { return filepath.Join("nested", "deep", dbName+".db") },
			wantDir: true,
		},
		{
			name:    "memory-backed database needs no directory",
			pathFor: func(string) string { return "" },
		},
		{
			name:    "nil resolver is treated as memory-backed",
			pathFor: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			pathFor := rootedResolver(root, testCase.pathFor)

			if err := NewSQLiteLifecycle(pathFor).Create(t.Context(), "games"); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			entries, err := os.ReadDir(root)
			if err != nil {
				t.Fatalf("read temp dir: %v", err)
			}
			if hasEntries := len(entries) > 0; hasEntries != testCase.wantDir {
				t.Fatalf("expected directory created %v, got entries %v", testCase.wantDir, entries)
			}
		})
	}
}

func TestSQLiteLifecycle_Drop(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		existing  []string
		wantLeft  []string
		wantError bool
	}{
		{
			name:     "removes the database file and its wal and shm sidecars",
			existing: []string{"games.db", "games.db-wal", "games.db-shm", "other.db"},
			wantLeft: []string{"other.db"},
		},
		{
			name:     "missing files are not an error",
			existing: []string{"other.db"},
			wantLeft: []string{"other.db"},
		},
		{
			name:     "removes the database file when no sidecar exists",
			existing: []string{"games.db"},
			wantLeft: []string{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			for _, name := range testCase.existing {
				writeEmptyFile(t, filepath.Join(root, name))
			}

			pathFor := func(dbName string) string { return filepath.Join(root, dbName+".db") }
			if err := NewSQLiteLifecycle(pathFor).Drop(t.Context(), "games"); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			assertDirEntries(t, root, testCase.wantLeft)
		})
	}
}

func TestSQLiteLifecycle_DropReportsRemovalFailure(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	// A non-empty directory standing where the database file belongs cannot be removed, which is
	// the portable way to force os.Remove to fail.
	blocked := filepath.Join(root, "games.db")
	if err := os.MkdirAll(blocked, 0o750); err != nil {
		t.Fatalf("create blocking directory: %v", err)
	}
	writeEmptyFile(t, filepath.Join(blocked, "occupant"))

	pathFor := func(dbName string) string { return filepath.Join(root, dbName+".db") }
	if err := NewSQLiteLifecycle(pathFor).Drop(t.Context(), "games"); err == nil {
		t.Fatal("expected a removal error, got nil")
	}
}

func TestSQLiteLifecycle_DropSkipsMemoryBackedDatabase(t *testing.T) {
	t.Parallel()

	if err := NewSQLiteLifecycle(nil).Drop(t.Context(), "games"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func rootedResolver(root string, pathFor DBFilePathResolver) DBFilePathResolver {
	if pathFor == nil {
		return nil
	}
	return func(dbName string) string {
		relative := pathFor(dbName)
		if relative == "" {
			return ""
		}
		return filepath.Join(root, relative)
	}
}

func writeEmptyFile(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatalf("write %q: %v", path, err)
	}
}

func assertDirEntries(t *testing.T, root string, want []string) {
	t.Helper()

	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("read temp dir: %v", err)
	}
	if len(entries) != len(want) {
		t.Fatalf("expected entries %v, got %v", want, entries)
	}
	for index, entry := range entries {
		if entry.Name() != want[index] {
			t.Fatalf("expected entry %q, got %q", want[index], entry.Name())
		}
	}
}
