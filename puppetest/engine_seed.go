package puppetest

import (
	"errors"
	"fmt"
	"slices"

	"github.com/wrapped-owls/testereiro/puppetest/internal/dbastidor"
)

// Seed executes SQL-based seed structs against the engine database.
func (e *Engine) Seed(seeds ...any) error {
	seedEvent := &EngineSeedEvent{
		Engine: e,
		Seeds:  append([]any(nil), seeds...),
	}
	if beforeHookErr := runHooks(seedEvent, e.hooks.beforeSeedHooks); beforeHookErr != nil {
		return beforeHookErr
	}

	if e.db == nil || e.db.IsZero() {
		return errors.New("database not initialized")
	}
	for _, s := range seeds {
		if err := dbastidor.ExecuteSeedStruct(e.db.Connection(), s, e.db.PlaceholderStyle()); err != nil {
			return fmt.Errorf("failed to seed data: %w", err)
		}
	}
	return nil
}

// SeedWithProvider executes provider-backed seeding strategies.
func (e *Engine) SeedWithProvider(providers ...SeedProvider) error {
	seedEvent := &EngineSeedEvent{
		Engine:        e,
		ProviderSeeds: slices.Clone(providers),
	}
	if beforeHookErr := runHooks(seedEvent, e.hooks.beforeSeedHooks); beforeHookErr != nil {
		return beforeHookErr
	}

	var seedErrs []error
	for index, provider := range providers {
		if provider == nil {
			seedErrs = append(seedErrs, fmt.Errorf("seed provider at index %d is nil", index))
			continue
		}
		if err := provider.ExecuteSeed(e); err != nil {
			seedErrs = append(
				seedErrs,
				fmt.Errorf("seed provider at index %d failed: %w", index, err),
			)
		}
	}
	return errors.Join(seedErrs...)
}
