package ops

import (
	"path/filepath"
	"testing"
)

func TestRegistry_resolveDefault(t *testing.T) {
	t.Parallel()
	r, err := LoadRegistry("")
	if err != nil {
		t.Fatal(err)
	}
	inst, err := r.Resolve("")
	if err != nil || inst.ID != "default" {
		t.Fatalf("%+v err=%v", inst, err)
	}
}

func TestRegistry_fromFile(t *testing.T) {
	t.Parallel()
	path := filepath.Join("..", "..", "config", "instances.example.json")
	r, err := LoadRegistry(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Resolve("readonly"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Resolve("no-such"); err == nil {
		t.Fatal("expected error")
	}
}
