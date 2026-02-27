package dungeonstore

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/wrapped-owls/testereiro/examples/mongodb_assert/fixtures"
)

const CollectionName = "dungeonformers"

// identityField is the BSON field used as a type discriminator.
const identityField = "_identity"

// DungeonStore provides MongoDB operations for polymorphic Dungeonformers.
type DungeonStore struct {
	db *mongo.Database
}

// NewDungeonStore creates a new DungeonStore.
func NewDungeonStore(db *mongo.Database) *DungeonStore {
	return &DungeonStore{db: db}
}

func (s *DungeonStore) collection() *mongo.Collection {
	return s.db.Collection(CollectionName)
}

// WrapForSeed marshals a Dungeonformer into a bson.M with the _identity field injected.
// This is useful for seeding via mongoseeder (which accepts []any of bson-marshalable values).
func WrapForSeed(d any) (bson.M, error) {
	raw, err := bson.Marshal(d)
	if err != nil {
		return nil, fmt.Errorf("marshal dungeonformer: %w", err)
	}

	var doc bson.M
	if err = bson.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("unmarshal to bson.M: %w", err)
	}

	doc[identityField] = Identity(d)
	return doc, nil
}

// MustWrapForSeed is like WrapForSeed but panics on error. Use in test setup only.
func MustWrapForSeed(d any) bson.M {
	doc, err := WrapForSeed(d)
	if err != nil {
		panic(err)
	}
	return doc
}

// Save inserts a Dungeonformer into MongoDB with the _identity field injected.
func (s *DungeonStore) Save(ctx context.Context, d any) error {
	doc, err := WrapForSeed(d)
	if err != nil {
		return err
	}

	_, err = s.collection().InsertOne(ctx, doc)
	return err
}

// FindAll loads all documents from the collection, decoding each into its concrete type.
func (s *DungeonStore) FindAll(ctx context.Context) ([]fixtures.Dungeonformer, error) {
	return s.findByFilter(ctx, bson.M{})
}

// FindByClass loads documents matching the given class name.
func (s *DungeonStore) FindByClass(
	ctx context.Context,
	class string,
) ([]fixtures.Dungeonformer, error) {
	return s.findByFilter(ctx, bson.M{"class": class})
}

// FindByIdentity loads documents matching the given _identity (type name).
func (s *DungeonStore) FindByIdentity(
	ctx context.Context,
	identity string,
) ([]fixtures.Dungeonformer, error) {
	return s.findByFilter(ctx, bson.M{identityField: identity})
}

func (s *DungeonStore) findByFilter(
	ctx context.Context,
	filter bson.M,
) ([]fixtures.Dungeonformer, error) {
	cursor, err := s.collection().Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []fixtures.Dungeonformer
	for cursor.Next(ctx) {
		// Read the _identity field to determine the concrete type.
		identity, ok := cursor.Current.Lookup(identityField).StringValueOK()
		if !ok {
			return nil, fmt.Errorf("document missing %s field", identityField)
		}

		decoded, err := Decode(identity, cursor.Current)
		if err != nil {
			return nil, err
		}
		results = append(results, decoded)
	}

	if err = cursor.Err(); err != nil {
		return nil, err
	}
	return results, nil
}
