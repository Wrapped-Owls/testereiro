package puppetest

import (
	"testing"
)

func TestEngineFactory_NoDB(t *testing.T) {
	fac, err := NewEngineFactory()
	if err != nil {
		t.Fatalf("failed to create factory: %v", err)
	}

	engine := fac.NewEngine(t)
	if engine.DB() != nil {
		t.Error("expected nil DB when no factory is set")
	}

	if err = engine.Teardown(); err != nil {
		t.Errorf("teardown failed: %v", err)
	}
}

func TestEngineFactory_WithExtension(t *testing.T) {
	extensionCalled := false
	ext := func(e *Engine) error {
		extensionCalled = true
		return nil
	}

	fac, err := NewEngineFactory(WithExtensions(ext))
	if err != nil {
		t.Fatalf("failed to create factory: %v", err)
	}

	fac.NewEngine(t)
	if !extensionCalled {
		t.Error("extension was not called")
	}
}
