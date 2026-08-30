package mongoseeder

import (
	"github.com/wrapped-owls/testereiro/puppetest"
)

// SeedOperationMode selects how documents are inserted during seeding.
type SeedOperationMode uint8

const (
	// SeedModeInsertMany inserts documents per collection with InsertMany.
	SeedModeInsertMany SeedOperationMode = iota
	// SeedModeClientBulkWrite inserts documents using client-level bulk writes.
	SeedModeClientBulkWrite
)

// SeedPlan describes one collection and its documents to seed.
type SeedPlan struct {
	Collection string
	Documents  []any
}

// SeedRunner is a configurable mongo seed provider for puppetest engines.
type SeedRunner struct {
	clearBefore bool
	ordered     bool
	mode        SeedOperationMode
	plans       []SeedPlan
}

var _ puppetest.SeedProvider = (*SeedRunner)(nil)

// New creates a SeedRunner with default behavior.
func New() *SeedRunner {
	return &SeedRunner{
		clearBefore: true,
		ordered:     true,
		mode:        SeedModeInsertMany,
	}
}

// WithSeedDocuments creates a SeedRunner and appends a seed plan.
func WithSeedDocuments(collection string, docs ...any) *SeedRunner {
	return New().WithSeedDocuments(collection, docs...)
}

// WithClearAndSeed creates a SeedRunner that clears DB before seeding and appends a plan.
func WithClearAndSeed(collection string, docs ...any) *SeedRunner {
	return New().WithClearAndSeed(collection, docs...)
}

// WithSeedDocuments appends a seed plan for one collection.
func (r *SeedRunner) WithSeedDocuments(collection string, docs ...any) *SeedRunner {
	r.plans = append(r.plans, SeedPlan{
		Collection: collection,
		Documents:  docs,
	})
	return r
}

// WithClearAndSeed enables clear-before mode and appends a seed plan.
func (r *SeedRunner) WithClearAndSeed(collection string, docs ...any) *SeedRunner {
	r.clearBefore = true
	return r.WithSeedDocuments(collection, docs...)
}

// WithClearBeforeSeed controls whether the database is dropped before inserts.
func (r *SeedRunner) WithClearBeforeSeed(clearBefore bool) *SeedRunner {
	r.clearBefore = clearBefore
	return r
}

// WithOrderedInsert controls ordered semantics for insert operations.
func (r *SeedRunner) WithOrderedInsert(ordered bool) *SeedRunner {
	r.ordered = ordered
	return r
}

// WithSeedOperationMode sets the insertion strategy used by ExecuteSeed.
func (r *SeedRunner) WithSeedOperationMode(mode SeedOperationMode) *SeedRunner {
	r.mode = mode
	return r
}

// WithInsertManySeedMode selects SeedModeInsertMany.
func (r *SeedRunner) WithInsertManySeedMode() *SeedRunner {
	return r.WithSeedOperationMode(SeedModeInsertMany)
}

// WithClientBulkWriteSeedMode selects SeedModeClientBulkWrite.
func (r *SeedRunner) WithClientBulkWriteSeedMode() *SeedRunner {
	return r.WithSeedOperationMode(SeedModeClientBulkWrite)
}
