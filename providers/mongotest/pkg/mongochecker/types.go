package mongochecker

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Operation uint8

const (
	OpFind Operation = iota
	OpFindOne
	OpAggregate
	OpCount
)

type Query struct {
	Collection string
	Operation  Operation
	Filter     bson.M
	Pipeline   bson.A
	Options    any // holds []options.Lister[options.XxxOptions] for the specific operation
}
