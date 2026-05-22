package eval

import (
	"fmt"
	"strings"

	"ai_slow_log/internal/analyzer"
	promptv6 "ai_slow_log/internal/prompt/slowlog"
)

func assertResult(result *analyzer.V6AgentResult, exp Expect) []string {
	var errs []string
	if result == nil {
		return []string{"result is nil"}
	}

	actions := result.Actions
	if exp.MinIterations > 0 && result.Iterations < exp.MinIterations {
		errs = append(errs, fmt.Sprintf("iterations=%d want >= %d", result.Iterations, exp.MinIterations))
	}
	if exp.MaxIterations > 0 && result.Iterations > exp.MaxIterations {
		errs = append(errs, fmt.Sprintf("iterations=%d want <= %d", result.Iterations, exp.MaxIterations))
	}

	if len(exp.Trajectory) > 0 {
		if len(actions) < len(exp.Trajectory) {
			errs = append(errs, fmt.Sprintf("actions=%d want at least %d for trajectory", len(actions), len(exp.Trajectory)))
		} else {
			for i, want := range exp.Trajectory {
				got := actions[i]
				if string(got.Type) != want.Type {
					errs = append(errs, fmt.Sprintf("step %d type=%q want %q", i+1, got.Type, want.Type))
					continue
				}
				if want.ToolName != "" && got.ToolName != want.ToolName {
					errs = append(errs, fmt.Sprintf("step %d tool=%q want %q", i+1, got.ToolName, want.ToolName))
				}
			}
		}
	}

	for _, tool := range exp.ToolsMustCall {
		if !actionCallsTool(actions, tool) {
			errs = append(errs, fmt.Sprintf("tool %q never called", tool))
		}
	}

	final := strings.TrimSpace(result.FinalResult)
	lower := strings.ToLower(final)
	for _, sub := range exp.FinalContains {
		if !strings.Contains(lower, strings.ToLower(sub)) {
			errs = append(errs, fmt.Sprintf("final_result missing %q", sub))
		}
	}
	for _, sub := range exp.FinalNotContains {
		if strings.Contains(lower, strings.ToLower(sub)) {
			errs = append(errs, fmt.Sprintf("final_result must not contain %q", sub))
		}
	}

	if exp.NoActionErrors && len(result.Rounds) > 0 {
		for _, r := range result.Rounds {
			if r.ActionError != "" {
				errs = append(errs, fmt.Sprintf("round %d action_error: %s", r.Round, r.ActionError))
			}
		}
	}

	return errs
}

func actionCallsTool(actions []promptv6.NextAction, name string) bool {
	for _, a := range actions {
		if a.Type == promptv6.ActionCallTool && a.ToolName == name {
			return true
		}
	}
	return false
}

func formatActions(actions []promptv6.NextAction) []string {
	out := make([]string, 0, len(actions))
	for _, a := range actions {
		if a.Type == promptv6.ActionCallTool && a.ToolName != "" {
			out = append(out, fmt.Sprintf("call_tool:%s", a.ToolName))
			continue
		}
		out = append(out, string(a.Type))
	}
	return out
}
