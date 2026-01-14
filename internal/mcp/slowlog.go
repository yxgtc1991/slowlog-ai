package mcp

import (
	"ai_slow_log/internal/analyzer"
	"context"
	"errors"
)

type AnalyzeSlowLogCapability struct {
	Analyzer *analyzer.SlowLogAnalyzer
}

func (c *AnalyzeSlowLogCapability) Name() string {
	return "analyze_slow_log"
}

func (c *AnalyzeSlowLogCapability) Description() string {
	return "分析 MySQL 慢日志，输出结构化的性能问题与优化建议"
}

func (c *AnalyzeSlowLogCapability) InputSchema() map[string]string {
	return map[string]string{
		"slow_log": "string // 原始 MySQL 慢日志文本",
	}
}

func (c *AnalyzeSlowLogCapability) Execute(
	ctx context.Context,
	input map[string]interface{},
) (interface{}, error) {

	raw, ok := input["slow_log"].(string)
	if !ok || raw == "" {
		return nil, errors.New("slow_log is required")
	}

	return c.Analyzer.Analyze(ctx, raw)
}
