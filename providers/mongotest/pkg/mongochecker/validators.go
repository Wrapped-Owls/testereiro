package mongochecker

import (
	"fmt"
	"testing"

	"github.com/wrapped-owls/testereiro/providers/mongotest/internal/mongoqueries"
	"github.com/wrapped-owls/testereiro/puppetest"
)

type (
	validator interface {
		Validate(t testing.TB, ctx puppetest.Context, cursor *Cursor) error
	}

	objSanitizer[O any]  func(expected, actual *O) error
	objComparator[O any] func(t testing.TB, expected, actual O) bool

	// Cursor wraps different MongoDB operation results.
	Cursor = mongoqueries.Cursor
)

type customValidator struct {
	validate func(testing.TB, puppetest.Context, *Cursor) error
}

func (v *customValidator) Validate(t testing.TB, ctx puppetest.Context, cursor *Cursor) error {
	if v.validate == nil {
		return fmt.Errorf("custom validation function is nil")
	}
	return v.validate(t, ctx, cursor)
}

type countValidator struct {
	expected int
}

func (v *countValidator) Validate(_ testing.TB, ctx puppetest.Context, cursor *Cursor) error {
	count := cursor.Count()
	puppetest.SaveOnCtx(ctx, count)

	if int(count) != v.expected {
		return fmt.Errorf("expected %d documents, got %d", v.expected, count)
	}
	return nil
}

type genericValidator[T any, L ~[]T] struct {
	Expected     T
	ExpectedList L
	Sanitizer    objSanitizer[T]
	Comparator   objComparator[T]
}

func (v genericValidator[T, L]) Validate(
	t testing.TB, ctx puppetest.Context, cursor *Cursor,
) error {
	actualList, err := v.extractResult(ctx, cursor)
	if err != nil {
		return err
	}

	expected := v.ExpectedList
	if len(expected) == 0 {
		expected = L{v.Expected}
	}

	if len(actualList) != len(expected) {
		t.Fatalf("got %d documents, expected %d", len(actualList), len(expected))
	}

	for index, expectedItem := range expected {
		actualItem := actualList[index]
		if v.Sanitizer != nil {
			if err = v.Sanitizer(&expectedItem, &actualItem); err != nil {
				return fmt.Errorf("failed to sanitize result: %w", err)
			}
		}

		if v.Comparator != nil {
			if !v.Comparator(t, expectedItem, actualItem) {
				return fmt.Errorf("mongo result mismatch")
			}
			return nil
		}
	}

	return nil
}

func (v genericValidator[T, L]) extractResult(ctx puppetest.Context, cursor *Cursor) (L, error) {
	var actualList L
	var err error

	// If we expect multiple docs, result is a slice
	if len(v.ExpectedList) > 0 {
		if actualList, err = mongoqueries.All[T](ctx, cursor); err != nil {
			return nil, err
		}
		puppetest.SaveOnCtx(ctx, actualList)
		return actualList, nil
	}

	var actual T
	if actual, err = mongoqueries.First[T](ctx, cursor); err != nil {
		return nil, err
	}

	puppetest.SaveOnCtx(ctx, actual)
	actualList = L{actual}
	return actualList, nil
}
