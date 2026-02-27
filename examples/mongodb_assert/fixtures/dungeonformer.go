package fixtures

// Dungeonformer is the common interface for all Dungeon Transformers.
type Dungeonformer interface {
	Name() string
	Class() string
}

// DnDBase is the embedded base for all Dungeonformers.
type DnDBase struct {
	EntityName string `bson:"name"  json:"name"`
	ClassName  string `bson:"class" json:"class"`
}

func (b *DnDBase) Name() string {
	return b.EntityName
}

func (b *DnDBase) Class() string {
	return b.ClassName
}

// GenericDungeonformer is a generic Dungeonformer with a name and height.
// It does NOT implement the Dungeonformer interface (Name is a field, not a method).
type GenericDungeonformer struct {
	DnDBase `       bson:",inline"`
	Name    string `bson:"name"    json:"name"`
	Height  uint32 `bson:"height"  json:"height"`
}

func NewGenericDungeonformer(name string, height uint32) GenericDungeonformer {
	return GenericDungeonformer{Name: name, Height: height}
}
