package agentmode

import "testing"

func TestResolve_cliOverridesEnv(t *testing.T) {
	t.Parallel()
	m, rest, err := Resolve("v6", []string{"-agent-mode=v5", "slow.txt"})
	if err != nil {
		t.Fatal(err)
	}
	if m != V5 {
		t.Fatalf("mode=%s", m)
	}
	if len(rest) != 1 || rest[0] != "slow.txt" {
		t.Fatalf("rest=%v", rest)
	}
}

func TestResolve_defaultV6(t *testing.T) {
	t.Parallel()
	m, _, err := Resolve("", nil)
	if err != nil || m != V6 {
		t.Fatalf("mode=%s err=%v", m, err)
	}
}

func TestResolve_unknown(t *testing.T) {
	t.Parallel()
	_, _, err := Resolve("v99", nil)
	if err == nil {
		t.Fatal("expected error")
	}
}
