package ops

import (
	"errors"
	"testing"
)

func TestLimits_rateLimit(t *testing.T) {
	t.Parallel()
	l := &Limits{perMinute: 1}
	if err := l.Acquire(); err != nil {
		t.Fatal(err)
	}
	if err := l.Acquire(); !errors.Is(err, ErrRateLimited) {
		t.Fatalf("got %v", err)
	}
}

func TestLimits_enabled(t *testing.T) {
	t.Parallel()
	if (&Limits{}).Enabled() {
		t.Fatal("empty limits should not be enabled")
	}
	if !(&Limits{perMinute: 1}).Enabled() {
		t.Fatal("expected enabled")
	}
}
