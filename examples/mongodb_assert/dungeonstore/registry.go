package dungeonstore

import (
	"fmt"
	"reflect"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/wrapped-owls/testereiro/examples/mongodb_assert/fixtures"
)

// decoderFunc creates a new zero-value of a registered type and unmarshals BSON into it.
type decoderFunc func(raw bson.Raw) (fixtures.Dungeonformer, error)

// registry maps _identity strings to decoder functions.
var registry = make(map[string]decoderFunc)

// Register adds a concrete Dungeonformer type to the registry.
// The identity key is derived from reflect.TypeFor[T]().Elem().Name().
func Register[T fixtures.Dungeonformer]() {
	var zero T
	typeName := reflect.TypeOf(zero).Elem().Name()
	registry[typeName] = func(raw bson.Raw) (fixtures.Dungeonformer, error) {
		target := reflect.New(reflect.TypeOf(zero).Elem()).Interface().(T)
		if err := bson.Unmarshal(raw, target); err != nil {
			return nil, fmt.Errorf("decode %s: %w", typeName, err)
		}
		return target, nil
	}
}

// Decode uses the registry to decode a raw BSON document into the correct concrete type.
func Decode(identity string, raw bson.Raw) (fixtures.Dungeonformer, error) {
	decoder, ok := registry[identity]
	if !ok {
		return nil, fmt.Errorf("unknown dungeonformer identity: %q", identity)
	}
	return decoder(raw)
}

// Identity returns the identity string for a value, using reflect.
func Identity(v any) string {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

func init() {
	Register[*fixtures.Bumblelf]()
	Register[*fixtures.OptimadinPrime]()
	Register[*fixtures.Ironknight]()
	Register[*fixtures.Ratcheric]()
	Register[*fixtures.Jazogue]()
	Register[*fixtures.Wheelificer]()
	Register[*fixtures.MegadwarfTron]()
	Register[*fixtures.Sorcerscream]()
	Register[*fixtures.Soundbard]()
	Register[*fixtures.Shocklock]()
}
