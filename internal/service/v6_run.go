package service

import (
	"ai_slow_log/internal/analyzer"
	"ai_slow_log/internal/bootstrap"
	"ai_slow_log/internal/llm"
	"ai_slow_log/internal/mcp"
	"ai_slow_log/internal/rag"
	prompt "ai_slow_log/internal/prompt/slowlog"
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// RunV6Config 一次 V6 分析请求（HTTP / 内部任务共用）。
type RunV6Config struct {
	SlowLog         string
	ReportDir       string
	Guided          bool
	HITL            bool
	AnalyzeTimeout  time.Duration
}

// RunV6Result 分析结果与报告路径。
type RunV6Result struct {
	ReportID    string
	FinalResult string
	Iterations  int
	Paths       *analyzer.SavedReportPaths
}

// ReportIDFromJSONPath 从 agent-run-*.json 路径得到 report_id。
func ReportIDFromJSONPath(jsonPath string) string {
	base := filepath.Base(jsonPath)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

func (c RunV6Config) timeoutDuration() time.Duration {
	if c.AnalyzeTimeout > 0 {
		return c.AnalyzeTimeout
	}
	return 15 * time.Minute
}

// RunV6 执行 V6 Agent 并写入 reports/。
func RunV6(ctx context.Context, client *llm.DeepSeekClient, cfg RunV6Config) (*RunV6Result, error) {
	if strings.TrimSpace(cfg.SlowLog) == "" {
		return nil, fmt.Errorf("slow_log is empty")
	}
	dir := cfg.ReportDir
	if dir == "" {
		dir = "reports"
	}

	boot, err := bootstrap.SetupMCP(client, nil)
	if err != nil {
		return nil, fmt.Errorf("mcp setup: %w", err)
	}
	defer boot.Close()

	toolNames := make([]string, 0, len(boot.Caps))
	for _, c := range boot.Caps {
		toolNames = append(toolNames, c.Name())
	}

	opts := []analyzer.V6AgentOption{
		analyzer.WithAgentRecordRounds(true),
		analyzer.WithAgentVerbose(false),
	}
	if cfg.Guided {
		opts = append(opts, analyzer.WithAgentGuide(prompt.GuidedSlowLogPreamble))
	}
	if cfg.HITL {
		opts = append(opts, analyzer.WithAgentHITL(true))
	}

	v6 := analyzer.NewV6AgentAnalyzer(
		client,
		analyzer.NewRAGRetrieverAdapter(rag.MustDefaultRetriever()),
		mcp.NewServerAsExecutor(boot.Server),
		boot.Server.GetCapabilitiesAsV4(),
		opts...,
	)

	result, err := v6.Analyze(ctx, cfg.SlowLog)
	if err != nil {
		return nil, fmt.Errorf("analyze: %w", err)
	}

	report := analyzer.BuildV6RunReport(cfg.SlowLog, result, toolNames)
	paths, err := analyzer.SaveV6RunReport(dir, report)
	if err != nil {
		return nil, fmt.Errorf("save report: %w", err)
	}

	return &RunV6Result{
		ReportID:    ReportIDFromJSONPath(paths.JSON),
		FinalResult: result.FinalResult,
		Iterations:  result.Iterations,
		Paths:       paths,
	}, nil
}
