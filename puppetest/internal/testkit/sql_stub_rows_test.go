package testkit

import (
	"database/sql/driver"
	"errors"
	"io"
	"reflect"
	"testing"
)

func TestStubSQLRows_Next(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		rows     [][]driver.Value
		wantRead [][]driver.Value
	}{
		{
			name:     "yields every row then reports EOF",
			rows:     [][]driver.Value{{true}, {false}},
			wantRead: [][]driver.Value{{true}, {false}},
		},
		{
			name:     "empty result reports EOF immediately",
			rows:     nil,
			wantRead: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			rows := &stubSQLRows{columns: []string{"value"}, rows: testCase.rows}

			var read [][]driver.Value
			for {
				dest := make([]driver.Value, 1)
				err := rows.Next(dest)
				if errors.Is(err, io.EOF) {
					break
				}
				if err != nil {
					t.Fatalf("unexpected next error: %v", err)
				}
				read = append(read, dest)
			}

			if !reflect.DeepEqual(read, testCase.wantRead) {
				t.Fatalf("expected rows %v, got %v", testCase.wantRead, read)
			}
		})
	}
}

func TestStubSQLRows_CloseStopsIteration(t *testing.T) {
	t.Parallel()

	rows := &stubSQLRows{columns: []string{"value"}, rows: [][]driver.Value{{true}}}
	if err := rows.Close(); err != nil {
		t.Fatalf("close error: %v", err)
	}

	if err := rows.Next(make([]driver.Value, 1)); !errors.Is(err, io.EOF) {
		t.Fatalf("expected EOF after close, got %v", err)
	}
}
