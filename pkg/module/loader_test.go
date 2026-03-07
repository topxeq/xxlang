package module

import (
	"testing"
)

func TestLoaderCache(t *testing.T) {
	loader := NewLoader()

	// Create and cache a module
	m := NewModule("./math")
	loader.Set("./math", m)

	// Get should return cached module
	m2, err := loader.Get("./math")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m2 != m {
		t.Error("expected same module instance from cache")
	}
}

func TestLoaderLoadingState(t *testing.T) {
	loader := NewLoader()

	// Mark as loading
	loader.MarkLoading("./math")
	if !loader.IsLoading("./math") {
		t.Error("expected IsLoading to be true")
	}

	// Mark as done
	loader.MarkDone("./math")
	if loader.IsLoading("./math") {
		t.Error("expected IsLoading to be false after MarkDone")
	}
}

func TestLoaderHasModule(t *testing.T) {
	loader := NewLoader()

	if loader.HasModule("./math") {
		t.Error("expected HasModule to be false for non-existent module")
	}

	m := NewModule("./math")
	loader.Set("./math", m)

	if !loader.HasModule("./math") {
		t.Error("expected HasModule to be true after Set")
	}
}
