package testkit

import (
	"errors"
	"strings"
	"testing"
)

func TestBrokenFS_Open(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		wantErrSubstr string
	}{
		{name: "empty path", path: "", wantErrSubstr: "cannot open"},
		{name: "regular path", path: "migrations/001.sql", wantErrSubstr: "cannot open"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := (BrokenFS{}).Open(tt.path)
			if err == nil || !strings.Contains(err.Error(), tt.wantErrSubstr) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErrSubstr, err)
			}
		})
	}
}

func TestApplyWithPanicCapture(t *testing.T) {
	baseErr := errors.New("run failed")
	tests := []struct {
		name       string
		panicValue any
		err        error
		wantPanic  bool
		wantErr    error
	}{
		{
			name:    "returns error and no panic",
			err:     baseErr,
			wantErr: baseErr,
		},
		{
			name:       "captures panic",
			panicValue: "boom",
			wantPanic:  true,
		},
		{
			name:    "returns nil when no panic and no error",
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			panicValue, err := ApplyWithPanicCapture(t, func() error {
				if tt.panicValue != nil {
					panic(tt.panicValue)
				}
				return tt.err
			})

			if tt.wantPanic {
				if panicValue == nil {
					t.Fatalf("expected panic to be captured")
				}
				return
			}
			if panicValue != nil {
				t.Fatalf("did not expect panic value, got %v", panicValue)
			}

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}
