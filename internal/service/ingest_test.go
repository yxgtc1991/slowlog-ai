package service

import (
	"testing"
)

func TestParseQueryTimeSeconds(t *testing.T) {
	t.Parallel()
	s := "# Query_time: 2.350  Lock_time: 0.000\nSELECT 1"
	if got := ParseQueryTimeSeconds(s); got < 2.34 || got > 2.36 {
		t.Fatalf("got %v", got)
	}
}

func TestPassesIngestThreshold(t *testing.T) {
	t.Parallel()
	slow := "# Query_time: 0.045\nSELECT 1"
	if !PassesIngestThreshold(slow, 0) {
		t.Fatal("zero threshold should pass")
	}
	if PassesIngestThreshold(slow, 1.0) {
		t.Fatal("should skip fast query")
	}
}

func TestParseIngestRequest(t *testing.T) {
	t.Parallel()
	req, err := ParseIngestRequest([]byte(`{"slow_log":"# Time\nSELECT 1","source":"fluent-bit","async":true}`))
	if err != nil || req.Source != "fluent-bit" || !req.IngestAsyncDefault(true) {
		t.Fatalf("%+v err=%v", req, err)
	}
}

func TestJobStore_createGet(t *testing.T) {
	t.Parallel()
	store := NewJobStore()
	store.Create("job-1", "test")
	j, ok := store.Get("job-1")
	if !ok || j.Status != JobPending {
		t.Fatalf("%+v", j)
	}
}
