package eval

import (
	"context"
	"fmt"
)

// StubExecutor 为 Eval 提供确定性工具返回值（不连真实 MySQL）。
type StubExecutor struct {
	tools map[string]bool
	reply map[string]interface{}
	errs  map[string]error
}

func NewStubExecutor() *StubExecutor {
	return &StubExecutor{
		tools: map[string]bool{},
		reply: map[string]interface{}{},
		errs:  map[string]error{},
	}
}

func (s *StubExecutor) WithTool(name string, result interface{}) *StubExecutor {
	s.tools[name] = true
	s.reply[name] = result
	return s
}

func (s *StubExecutor) WithToolError(name string, err error) *StubExecutor {
	s.tools[name] = true
	s.errs[name] = err
	return s
}

func (s *StubExecutor) ExecuteCapability(_ context.Context, name string, _ map[string]interface{}) (interface{}, error) {
	if err, ok := s.errs[name]; ok {
		return nil, err
	}
	if r, ok := s.reply[name]; ok {
		return r, nil
	}
	if s.tools[name] {
		return map[string]interface{}{"ok": true, "tool": name}, nil
	}
	return nil, fmt.Errorf("capability %q not stubbed", name)
}

func (s *StubExecutor) HasCapability(name string) bool {
	return s.tools[name]
}
