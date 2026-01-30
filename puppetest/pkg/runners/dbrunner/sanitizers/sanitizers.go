package sanitizers

// Placeholder for reusable sanitizers.
// These could be functions that clean up database results before comparison.

type DbSanitizer[O any] func(expected, actual *O) error

func NoOpSanitizer[O any](expected, actual *O) error {
	return nil
}
