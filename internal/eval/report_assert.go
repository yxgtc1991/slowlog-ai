package eval

import (
	"encoding/json"
	"fmt"
	"strings"

	"ai_slow_log/internal/analyzer"
)

// ReportExpect 对已保存的 agent-run JSON 做回归断言（无需重跑 LLM）。
type ReportExpect struct {
	MinRounds       int
	ActionTypes     []string // 按序匹配 actions[].type
	FinalContains   []string
	MustHaveTools   []string
}

// AssertReportFile 读取 reports/*.json 并断言。
func AssertReportFile(path string, exp ReportExpect) []string {
	data, err := readProjectFile(path)
	if err != nil {
		return []string{fmt.Sprintf("read report: %v", err)}
	}
	var report analyzer.V6AgentRunReport
	if err := json.Unmarshal(data, &report); err != nil {
		return []string{fmt.Sprintf("parse report: %v", err)}
	}
	return assertReport(&report, exp)
}

func assertReport(report *analyzer.V6AgentRunReport, exp ReportExpect) []string {
	var errs []string
	if report == nil {
		return []string{"report is nil"}
	}
	if exp.MinRounds > 0 && len(report.Actions) < exp.MinRounds {
		errs = append(errs, fmt.Sprintf("actions=%d want >= %d", len(report.Actions), exp.MinRounds))
	}
	if len(exp.ActionTypes) > 0 {
		if len(report.Actions) < len(exp.ActionTypes) {
			errs = append(errs, fmt.Sprintf("actions=%d want >= %d types", len(report.Actions), len(exp.ActionTypes)))
		} else {
			for i, want := range exp.ActionTypes {
				if string(report.Actions[i].Type) != want {
					errs = append(errs, fmt.Sprintf("action[%d] type=%q want %q", i, report.Actions[i].Type, want))
				}
			}
		}
	}
	for _, t := range exp.MustHaveTools {
		found := false
		for _, a := range report.Actions {
			if a.Type == "call_tool" && a.ToolName == t {
				found = true
				break
			}
		}
		if !found {
			errs = append(errs, fmt.Sprintf("report never called tool %q", t))
		}
	}
	lower := strings.ToLower(report.FinalResult)
	for _, sub := range exp.FinalContains {
		if !strings.Contains(lower, strings.ToLower(sub)) {
			errs = append(errs, fmt.Sprintf("final_result missing %q", sub))
		}
	}
	return errs
}
