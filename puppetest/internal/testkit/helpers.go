package testkit

import (
	"errors"
	"io/fs"
	"testing"
)

type BrokenFS struct{}

func (BrokenFS) Open(string) (fs.File, error) {
	return nil, errors.New("cannot open")
}

func ApplyWithPanicCapture(t testing.TB, apply func() error) (panicValue any, err error) {
	t.Helper()
	defer func() {
		panicValue = recover()
	}()
	err = apply()
	return
}
