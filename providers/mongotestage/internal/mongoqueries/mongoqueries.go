package mongoqueries

import (
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/wrapped-owls/testereiro/puppetest"
)

type Cursor struct {
	cursor *mongo.Cursor
	all    []bson.Raw
	first  bson.Raw
	count  int64
}

func NewCursor(mc *mongo.Cursor) *Cursor {
	return &Cursor{cursor: mc}
}

func NewCursorCount(cnt int64) *Cursor {
	return &Cursor{count: cnt}
}

func NewCursorResult(elements ...bson.Raw) *Cursor {
	cur := new(Cursor)
	if cur.count = int64(len(elements)); cur.count >= 1 {
		cur.all = elements
		cur.first = elements[0]
	}

	return cur
}

func (c *Cursor) Count() int64 { return c.count }

func (c *Cursor) Close(ctx puppetest.Context) error {
	if c.cursor != nil {
		return c.cursor.Close(ctx)
	}
	return nil
}

func All[O any](ctx puppetest.Context, cursor *Cursor) ([]O, error) {
	if cursor.all != nil {
		results := make([]O, 0, len(cursor.all))
		for _, rawItem := range cursor.all {
			var obj O
			if err := bson.Unmarshal(rawItem, &obj); err != nil {
				return nil, fmt.Errorf("failed to unmarshal cached item into %T: %w", obj, err)
			}
			results = append(results, obj)
		}
		return results, nil
	}

	if mc := cursor.cursor; mc != nil {
		var results []O
		if err := mc.All(ctx, &results); err != nil {
			return nil, fmt.Errorf("failed to decode mongo cursor into []%T: %w", results, err)
		}
		cursor.count = int64(len(results))
		return results, nil
	}

	return nil, fmt.Errorf("no data available in cursor")
}

func First[O any](ctx puppetest.Context, cursor *Cursor) (O, error) {
	var obj O

	if first := cursor.first; first != nil {
		if err := bson.Unmarshal(first, &obj); err != nil {
			return obj, fmt.Errorf("failed to unmarshal first result into %T: %w", obj, err)
		}
		return obj, nil
	}

	if len(cursor.all) > 0 {
		if err := bson.Unmarshal(cursor.all[0], &obj); err != nil {
			return obj, fmt.Errorf("failed to unmarshal first cached item into %T: %w", obj, err)
		}
		return obj, nil
	}

	if mc := cursor.cursor; mc != nil {
		var results []O
		if err := mc.All(ctx, &results); err != nil {
			return obj, err
		}
		if len(results) > 0 {
			return results[0], nil
		}
	}

	return obj, fmt.Errorf("no documents found")
}
