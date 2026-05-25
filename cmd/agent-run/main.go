// 完整跑一遍 Agent（V6 默认 / V5 Tool Calling），并将输出保存到 reports/。
package main

import (
	"ai_slow_log/internal/agentmode"
	"ai_slow_log/internal/analyzer"
	"ai_slow_log/internal/bootstrap"
	"ai_slow_log/internal/config"
	"ai_slow_log/internal/llm"
	"ai_slow_log/internal/mcp"
	"ai_slow_log/internal/rag"
	prompt "ai_slow_log/internal/prompt/slowlog"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	_ = config.LoadDotEnv(".env")

	mode, args, err := agentmode.ResolveFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	reportDir := "reports"
	slowLogPath := "testdata/slowlog-products.txt"
	guided := true
	trace := true

	for _, arg := range args {
		switch {
		case arg == "-guided=false", arg == "--guided=false":
			guided = false
		case arg == "-trace=false":
			trace = false
		case strings.HasPrefix(arg, "-report="):
			reportDir = strings.TrimPrefix(arg, "-report=")
		case !strings.HasPrefix(arg, "-"):
			slowLogPath = arg
		}
	}

	slowLog, err := os.ReadFile(slowLogPath)
	if err != nil {
		log.Fatalf("read slow log: %v", err)
	}

	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		log.Fatal("DEEPSEEK_API_KEY is required in .env or environment")
	}

	ctx := context.Background()
	llmClient, err := llm.NewDeepSeekClient(apiKey, "")
	if err != nil {
		log.Fatalf("llm: %v", err)
	}

	boot, err := bootstrap.SetupMCP(llmClient, func(format string, v ...any) {
		fmt.Fprintf(os.Stderr, format+"\n", v...)
	})
	if err != nil {
		log.Fatalf("mcp: %v", err)
	}
	defer boot.Close()

	caps := boot.Caps
	toolNames := make([]string, 0, len(caps))
	for _, c := range caps {
		toolNames = append(toolNames, c.Name())
	}

	switch mode {
	case agentmode.V5:
		runV5Agent(ctx, llmClient, boot.Server, string(slowLog), reportDir)
	default:
		runV6Agent(ctx, llmClient, boot.Server, string(slowLog), reportDir, toolNames, guided, trace)
	}
}

func runV5Agent(ctx context.Context, llmClient *llm.DeepSeekClient, server *mcp.Server, slowLog, reportDir string) {
	fmt.Fprintln(os.Stderr, "▶ 开始 V5 Tool Calling 分析（结果将写入 reports/）...")
	v5 := analyzer.NewV5ToolCallingAnalyzer(
		llm.NewDeepSeekToolCallingAdapter(llmClient),
		analyzer.NewRAGRetrieverAdapter(rag.MustDefaultRetriever()),
		mcp.NewServerAsExecutor(server),
		server.GetCapabilitiesAsV4(),
	)
	result, err := v5.Analyze(ctx, slowLog)
	if err != nil {
		log.Fatalf("v5 agent: %v", err)
	}

	report := analyzer.BuildV5RunReport(slowLog, result)
	jsonPath, mdPath, err := analyzer.SaveV5RunReport(reportDir, report)
	if err != nil {
		log.Fatalf("save report: %v", err)
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("最终结论（V5）")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println(result.Analysis)
	analyzer.PrintV5ToolCallingSummary(result)

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("报告已保存")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("JSON:", jsonPath)
	fmt.Println("MD:  ", mdPath)
}

func runV6Agent(ctx context.Context, llmClient *llm.DeepSeekClient, server *mcp.Server, slowLog, reportDir string, toolNames []string, guided, trace bool) {
	opts := []analyzer.V6AgentOption{
		analyzer.WithAgentRecordRounds(true),
		analyzer.WithAgentVerbose(trace),
	}
	if guided {
		opts = append(opts, analyzer.WithAgentGuide(prompt.GuidedSlowLogPreamble))
	}
	if analyzer.HITLEnabledFromEnv() {
		opts = append(opts, analyzer.WithAgentHITL(true))
		fmt.Fprintln(os.Stderr, "ℹ️  HITL 已开启：ask_question 时将等待 stdin 输入（SLOWLOG_AGENT_HITL）")
	}
	v6 := analyzer.NewV6AgentAnalyzer(
		llmClient,
		analyzer.NewRAGRetrieverAdapter(rag.MustDefaultRetriever()),
		mcp.NewServerAsExecutor(server),
		server.GetCapabilitiesAsV4(),
		opts...,
	)

	if guided {
		fmt.Fprintln(os.Stderr, "ℹ️  已启用 guided 推荐流程（RAG → MySQL → EXPLAIN → 索引 dry_run → finish）")
	}
	fmt.Fprintln(os.Stderr, "▶ 开始 V6 Agent 分析（结果将写入 reports/）...")
	result, err := v6.Analyze(ctx, slowLog)
	if err != nil {
		log.Fatalf("agent: %v", err)
	}

	report := analyzer.BuildV6RunReport(slowLog, result, toolNames)
	paths, err := analyzer.SaveV6RunReport(reportDir, report)
	if err != nil {
		log.Fatalf("save report: %v", err)
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("最终结论")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println(result.FinalResult)

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("报告已保存（无需重跑即可复盘）")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("精简 HTML:", paths.BriefHTML, "  ← 推荐给客户/快速看每轮做了什么")
	fmt.Println("精简 MD:  ", paths.BriefMD)
	fmt.Println("完整 HTML:", paths.FullHTML)
	fmt.Println("完整 MD:  ", paths.FullMD)
	fmt.Println("JSON:     ", paths.JSON)
	fmt.Println("\n查阅建议：打开", filepath.Base(paths.BriefHTML), "一眼看逐轮过程；", filepath.Base(paths.FullHTML), "看完整细节。")
}
