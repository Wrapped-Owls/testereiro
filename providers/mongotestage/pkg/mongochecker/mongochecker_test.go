package mongochecker

import (
	"fmt"
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/wrapped-owls/testereiro/puppetest"
)

func TestBsonQueryBuilder_Build(t *testing.T) {
	builder := NewBsonQuery("users", bson.M{"active": true})
	builder.AddFilter(func(_ puppetest.Context) (bson.M, error) {
		return bson.M{"tenant": "acme"}, nil
	})

	query, err := builder.Build(nil)
	if err != nil {
		t.Fatalf("expected build to succeed, got %v", err)
	}
	if query.Collection != "users" {
		t.Fatalf("unexpected collection: %s", query.Collection)
	}

	expectedFilter := bson.M{"active": true, "tenant": "acme"}
	if !reflect.DeepEqual(expectedFilter, query.Filter) {
		t.Fatalf("unexpected filter: expected %+v, got %+v", expectedFilter, query.Filter)
	}
}

func TestWithCustomValidation_NilValidation(t *testing.T) {
	// The original test was checking if the validation function itself was nil.
	// The new structure means a non-nil validation function is always provided.

	runner := New(
		nil,
		WithCustomValidation(func(t testing.TB, ctx puppetest.Context, res *Cursor) error {
			if res == nil {
				return fmt.Errorf("expected cursor")
			}
			return nil
		}),
	)

	err := runner.validators[0].Validate(t, nil, nil)
	if err == nil {
		t.Fatal("expected nil validation function to fail")
	}
}
