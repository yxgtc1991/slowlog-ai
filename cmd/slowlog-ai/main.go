package main

import (
	"ai_slow_log/internal/agentmode"
	"ai_slow_log/internal/analyzer"
	"ai_slow_log/internal/bootstrap"
	"ai_slow_log/internal/config"
	"ai_slow_log/internal/llm"
	"ai_slow_log/internal/mcp"
	"ai_slow_log/internal/rag"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	_ = config.LoadDotEnv(".env")

	agentTrace := agentTraceEnabled()
	if agentTrace {
		fmt.Fprintln(os.Stderr, "ℹ️  Agent 轨迹已开启（stderr 每轮输出，结束后再打汇总）")
	}

	mode, args, err := agentmode.ResolveFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	const defaultSlowLogPath = "testdata/slowlog-products.txt"
	slowLog := readSlowLogOrFatal(args, defaultSlowLogPath)

	ctx := context.Background()

	llmClient, err := llm.NewDeepSeekClient(os.Getenv("DEEPSEEK_API_KEY"), "")
	if err != nil {
		log.Fatalf("failed to create llm client: %v", err)
	}

	boot, err := bootstrap.SetupMCP(llmClient, func(format string, v ...any) {
		fmt.Printf(format+"\n", v...)
	})
	if err != nil {
		log.Fatalf("mcp: %v", err)
	}
	defer boot.Close()

	switch mode {
	case agentmode.V5:
		runV5(ctx, llmClient, boot.Server, slowLog)
	default:
		runV6(ctx, llmClient, boot.Server, slowLog, agentTrace)
	}
}

func runV5(ctx context.Context, llmClient *llm.DeepSeekClient, server *mcp.Server, slowLog string) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("⚡ V5 Tool Calling：API tool_calls 协议")
	fmt.Println(strings.Repeat("=", 60))

	v5Analyzer := analyzer.NewV5ToolCallingAnalyzer(
		llm.NewDeepSeekToolCallingAdapter(llmClient),
		analyzer.NewRAGRetrieverAdapter(rag.MustDefaultRetriever()),
		mcp.NewServerAsExecutor(server),
		server.GetCapabilitiesAsV4(),
	)

	v5Result, err := v5Analyzer.Analyze(ctx, slowLog)
	if err != nil {
		log.Fatalf("Failed to analyze with V5: %v", err)
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 慢日志分析结果（V5 Tool Calling）")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println(v5Result.Analysis)
	analyzer.PrintV5ToolCallingSummary(v5Result)
}

func runV6(ctx context.Context, llmClient *llm.DeepSeekClient, server *mcp.Server, slowLog string, agentTrace bool) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🤖 V6 Agent：LLM 自主决策下一步行动（NextAction）")
	fmt.Println(strings.Repeat("=", 60))

	v6Opts := []analyzer.V6AgentOption{}
	if agentTrace {
		v6Opts = append(v6Opts, analyzer.WithAgentVerbose(true))
	}
	if analyzer.HITLEnabledFromEnv() {
		v6Opts = append(v6Opts, analyzer.WithAgentHITL(true))
		fmt.Fprintln(os.Stderr, "ℹ️  HITL 已开启：ask_question 时将等待 stdin 输入")
	}
	v6Analyzer := analyzer.NewV6AgentAnalyzer(
		llmClient,
		analyzer.NewRAGRetrieverAdapter(rag.MustDefaultRetriever()),
		mcp.NewServerAsExecutor(server),
		server.GetCapabilitiesAsV4(),
		v6Opts...,
	)

	v6Result, err := v6Analyzer.Analyze(ctx, slowLog)
	if err != nil {
		log.Fatalf("Failed to analyze with V6: %v", err)
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 慢日志分析结果（V6 Agent）")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println(v6Result.FinalResult)
	analyzer.PrintV6AgentSummary(v6Result, agentTrace)
}

func agentTraceEnabled() bool {
	if os.Getenv("SLOWLOG_AGENT_TRACE") == "1" || strings.EqualFold(os.Getenv("SLOWLOG_AGENT_TRACE"), "true") {
		return true
	}
	for _, arg := range os.Args[1:] {
		if arg == "-agent-trace" || arg == "--agent-trace" {
			return true
		}
	}
	return false
}

// readSlowLogOrFatal 从剩余 args 取慢日志路径（已去掉模式参数）。
func readSlowLogOrFatal(args []string, defaultPath string) string {
	path := defaultPath
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			continue
		}
		path = arg
		break
	}
	content, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("read slow log %s: %v", path, err)
	}
	return string(content)
}
