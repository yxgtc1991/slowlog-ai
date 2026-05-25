package rag

import (
	"path/filepath"
	"testing"
)

func TestSaveLoadTFIDFIndex_roundtrip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	idx, err := buildTFIDFIndex()
	if err != nil {
		t.Fatal(err)
	}
	if err := SaveTFIDFIndex(dir, idx); err != nil {
		t.Fatal(err)
	}
	r, ok, err := LoadTFIDFIndex(dir, 3)
	if err != nil || !ok || r == nil {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	chunks, err := r.Retrieve(t.Context(), "price 最左前缀")
	if err != nil || len(chunks) == 0 {
		t.Fatalf("retrieve: %v len=%d", err, len(chunks))
	}
}

func TestGetIndexStatus_afterBuild(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "rag-index")
	if err := BuildAllIndexes(dir, false); err != nil {
		t.Fatal(err)
	}
	st, err := GetIndexStatus(dir)
	if err != nil || !st.TFIDFOnDisk || st.Manifest == nil {
		t.Fatalf("st=%+v err=%v", st, err)
	}
}

func TestComputeEmbeddedDocsHash_stable(t *testing.T) {
	t.Parallel()
	h1, err := ComputeEmbeddedDocsHash()
	if err != nil {
		t.Fatal(err)
	}
	h2, err := ComputeEmbeddedDocsHash()
	if err != nil || h1 != h2 {
		t.Fatalf("%s vs %s", h1, h2)
	}
}
