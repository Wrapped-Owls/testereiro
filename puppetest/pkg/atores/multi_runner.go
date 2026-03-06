package atores

import (
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest/internal/stgctx"
)

// Runner defines a unit of work executable in a shared runner context.
type Runner interface {
	Run(t testing.TB, ctx stgctx.RunnerContext) error
}

var _ Runner = (*MultiRunner)(nil)

// MultiRunner executes multiple runners sequentially with a shared context.
type MultiRunner struct {
	Runners []Runner
}

// Run executes each runner and stops on the first error.
func (mr MultiRunner) Run(t testing.TB, ctx stgctx.RunnerContext) error {
	for _, runner := range mr.Runners {
		if err := runner.Run(t, ctx); err != nil {
			return err
		}
	}
	return nil
}
