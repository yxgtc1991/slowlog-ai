package eval

import (
	"context"
	"fmt"
)

// ScriptLLM 按预定顺序返回 LLM 输出，用于 Agent Eval（不调用真实 API）。
type ScriptLLM struct {
	script []string
	step   int
}

func NewScriptLLM(script ...string) *ScriptLLM {
	return &ScriptLLM{script: script}
}

func (s *ScriptLLM) Chat(_ context.Context, _ string) (string, error) {
	if s.step >= len(s.script) {
		return "", fmt.Errorf("script LLM exhausted after %d responses (add more script steps)", s.step)
	}
	out := s.script[s.step]
	s.step++
	return out, nil
}

func (s *ScriptLLM) StepsUsed() int { return s.step }
