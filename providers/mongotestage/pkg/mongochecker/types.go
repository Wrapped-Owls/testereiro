package mongochecker

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Operation identifies which mongo query operation should be executed.
type Operation uint8

const (
	// OpFind executes a collection find query.
	OpFind Operation = iota
	// OpFindOne executes a collection findOne query.
	OpFindOne
	// OpAggregate executes a collection aggregate query.
	OpAggregate
	// OpCount executes a collection count query.
	OpCount
)

// Query contains the resolved mongo query data passed to execution.
type Query struct {
	Collection string
	Operation  Operation
	Filter     bson.M
	Pipeline   bson.A
	Options    any // holds []options.Lister[options.XxxOptions] for the specific operation
}
