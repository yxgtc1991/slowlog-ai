package analyzer

import (
	"strings"
	"testing"
)

func TestHITLEnabledFromEnv(t *testing.T) {
	t.Setenv("SLOWLOG_AGENT_HITL", "true")
	if !HITLEnabledFromEnv() {
		t.Fatal("expected true")
	}
}

func TestStubUserInput(t *testing.T) {
	t.Parallel()
	s := NewStubUserInput("answer-one", "answer-two")
	a, err := s.ReadLine("prompt")
	if err != nil || a != "answer-one" {
		t.Fatalf("got %q err=%v", a, err)
	}
	a2, _ := s.ReadLine("")
	if a2 != "answer-two" {
		t.Fatalf("got %q", a2)
	}
}

func TestAgentState_userReplyInSummary(t *testing.T) {
	t.Parallel()
	s := NewAgentState()
	s.RecordQuestion("表名？")
	s.RecordUserReply("products")
	sum := s.PromptSummary(200)
	if !strings.Contains(sum, "用户已答") || !strings.Contains(sum, "products") {
		t.Fatalf("summary=%q", sum)
	}
}
