package runners

import (
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest/internal/stgctx"
)

type Runner interface {
	Run(t testing.TB, ctx stgctx.RunnerContext) error
}

type MultiRunner struct {
	Runners []Runner
}

func (mr MultiRunner) Run(t testing.TB) error {
	ctx := stgctx.NewRunnerContext(t.Context())
	for _, runner := range mr.Runners {
		if err := runner.Run(t, ctx); err != nil {
			return err
		}
	}
	return nil
}
