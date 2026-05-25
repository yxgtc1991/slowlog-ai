package analyzer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// V5AgentRunReport V5 Tool Calling 一次运行的可持久化报告（无逐轮 NextAction）。
type V5AgentRunReport struct {
	GeneratedAt string                 `json:"generated_at"`
	Mode        string                 `json:"mode"`
	SlowLog     string                 `json:"slow_log"`
	Analysis    string                 `json:"analysis"`
	Iterations  int                    `json:"iterations"`
	ToolCalls   []ToolCall             `json:"tool_calls"`
	ToolResults map[string]interface{} `json:"tool_results"`
	RawOutput   string                 `json:"raw_output,omitempty"`
}

// BuildV5RunReport 从 V5 Analyze 结果组装报告。
func BuildV5RunReport(slowLog string, result *V5ToolCallingResult) *V5AgentRunReport {
	if result == nil {
		return nil
	}
	return &V5AgentRunReport{
		GeneratedAt: time.Now().Format(time.RFC3339),
		Mode:        "v5",
		SlowLog:     slowLog,
		Analysis:    result.Analysis,
		Iterations:  result.Iterations,
		ToolCalls:   result.ToolCalls,
		ToolResults: result.ToolResults,
		RawOutput:   result.RawOutput,
	}
}

// SaveV5RunReport 写入 JSON 与简要 Markdown。
func SaveV5RunReport(dir string, report *V5AgentRunReport) (jsonPath, mdPath string, err error) {
	if report == nil {
		return "", "", fmt.Errorf("report is nil")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", err
	}
	stamp := time.Now().Format("20060102-150405")
	base := "v5-run-" + stamp
	jsonPath = filepath.Join(dir, base+".json")
	mdPath = filepath.Join(dir, base+".md")

	raw, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", "", err
	}
	if err := os.WriteFile(jsonPath, raw, 0o644); err != nil {
		return "", "", err
	}
	md := fmt.Sprintf("# V5 Tool Calling 报告\n\n- 生成时间：%s\n- 迭代：%d\n- 工具调用：%d 次\n\n## 结论\n\n%s\n\n## 工具\n\n",
		report.GeneratedAt, report.Iterations, len(report.ToolCalls), report.Analysis)
	for i, tc := range report.ToolCalls {
		md += fmt.Sprintf("%d. **%s**\n", i+1, tc.Name)
	}
	if err := os.WriteFile(mdPath, []byte(md), 0o644); err != nil {
		return "", "", err
	}
	return jsonPath, mdPath, nil
}

// PrintV5ToolCallingSummary 分析结束后输出工具调用摘要。
func PrintV5ToolCallingSummary(result *V5ToolCallingResult) {
	if result == nil {
		return
	}
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("🔧 V5 Tool Calling：%d 次工具 | %d 轮迭代\n", len(result.ToolCalls), result.Iterations)
	for i, tc := range result.ToolCalls {
		fmt.Printf("  %d. %s\n", i+1, tc.Name)
	}
}
