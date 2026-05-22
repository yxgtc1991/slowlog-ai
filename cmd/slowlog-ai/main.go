package main

import (
	"ai_slow_log/internal/analyzer"
	"ai_slow_log/internal/config"
	"ai_slow_log/internal/llm"
	"ai_slow_log/internal/mcp"
	"ai_slow_log/internal/mysql"
	prompt "ai_slow_log/internal/prompt/slowlog"
	"ai_slow_log/internal/rag"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
)

func slowLogFileArg() string {
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "-") {
			continue
		}
		return arg
	}
	return ""
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

func main() {
	agentTrace := agentTraceEnabled()
	if agentTrace {
		fmt.Fprintln(os.Stderr, "ℹ️  Agent 轨迹已开启（stderr 每轮输出，结束后再打汇总）")
	}

	const defaultSlowLogPath = "testdata/slowlog-products.txt"
	slowLog := readSlowLogOrFatal(slowLogFileArg(), defaultSlowLogPath)

	ctx := context.Background()

	llmClient, err := llm.NewDeepSeekClient(os.Getenv("DEEPSEEK_API_KEY"), "")
	if err != nil {
		log.Fatalf("failed to create llm client: %v", err)
	}

	// 创建 V4 分析器（用于 V4 能力感知演示）
	v4Analyzer := analyzer.NewAnalyzer(
		llmClient,
		analyzer.WithPromptBuilder(&prompt.RagV3Prompt{}),
		analyzer.WithRAGRetriever(analyzer.NewRAGRetrieverAdapter(rag.NewMockRetriever())),
	)

	// ===== V4 能力感知演示（已关闭，只保留 V5）=====
	// 创建 MCP 服务器并注册能力
	mcpServer := mcp.NewServer()
	capability := &mcp.AnalyzeSlowLogCapability{
		Analyzer: v4Analyzer,
	}
	mcpServer.RegisterCapability(capability)

	// 本地 MySQL：从 .env / 环境变量加载，注册 connect_mysql_instance
	mysqlCfg, err := config.MustLoadMySQL()
	if err != nil {
		log.Printf("mysql: %v (skip connect_mysql_instance)", err)
	} else {
		mysqlClient, err := mysql.NewClient(mysqlCfg)
		if err != nil {
			log.Printf("mysql: connect failed: %v", err)
		} else {
			defer func() { _ = mysqlClient.Close() }()
			mcp.RegisterMySQLCapabilities(mcpServer, mysqlClient)
			fmt.Printf("MySQL: connected to %s:%d as %s\n", mysqlCfg.Host, mysqlCfg.Port, mysqlCfg.User)
		}
	}

	// V4 演示代码已注释，如需查看 V4 功能，取消下面的注释
	/*
		// 1. 列出所有可用能力（JSON 格式）
		fmt.Println("=== 能力感知：列出所有可用能力 ===")
		capabilitiesJSON, err := mcpServer.ListCapabilities()
		if err != nil {
			log.Fatalf("Failed to list capabilities: %v", err)
		}
		fmt.Println(capabilitiesJSON)
		fmt.Println()

		// 2. 获取能力描述 prompt（给 LLM 看）
		fmt.Println("=== 能力感知：生成 LLM 可理解的能力描述 ===")
		capabilityPrompt := mcpServer.GetCapabilityPrompt()
		fmt.Println(capabilityPrompt)
		fmt.Println()

		// 3. 检查能力是否存在
		fmt.Println("=== 能力感知：检查能力是否存在 ===")
		fmt.Printf("是否支持 'analyze_slow_log'：%v\n", mcpServer.HasCapability("analyze_slow_log"))
		fmt.Printf("已注册能力数量：%d\n", mcpServer.CapabilityCount())
		fmt.Println()

		// 4. 使用能力感知执行分析
		fmt.Println("=== 使用能力感知执行慢日志分析（V4 方式）===")
		analyzeResult, err := v4Analyzer.Analyze(ctx, slowLog)
		if err != nil {
			log.Fatalf("Failed to analyze slow log: %v", err)
		}
		fmt.Println(analyzeResult.RawOutput)
	*/

	// ===== V5 Tool Calling 演示（已关闭，只保留 V6）=====
	/*
		// 1. 创建 Tool Calling 客户端适配器
		toolCallingClient := llm.NewDeepSeekToolCallingAdapter(llmClient)

		// 2. 将能力转换为工具定义
		caps := mcpServer.GetCapabilitiesAsV4()

		// 3. 创建 V5 Tool Calling 分析器
		v5Analyzer := analyzer.NewV5ToolCallingAnalyzer(
			toolCallingClient,
			analyzer.NewRAGRetrieverAdapter(rag.NewMockRetriever()),
			mcp.NewServerAsExecutor(mcpServer),
			caps,
		)

		// 4. 执行 V5 分析（LLM 会直接调用工具）
		v5Result, err := v5Analyzer.Analyze(ctx, slowLog)
		if err != nil {
			log.Fatalf("Failed to analyze with V5: %v", err)
		}

		// 输出最终分析结果（LLM 基于工具结果的总结）
		fmt.Println("\n" + strings.Repeat("=", 60))
		fmt.Println("📊 慢日志分析结果（V5 Tool Calling）")
		fmt.Println(strings.Repeat("=", 60))
		fmt.Println(v5Result.Analysis)

		// 只在需要时显示工具调用信息
		if len(v5Result.ToolCalls) > 0 {
			fmt.Println("\n" + strings.Repeat("-", 60))
			fmt.Printf("🔧 工具调用：%d 次 | 迭代：%d 轮\n", len(v5Result.ToolCalls), v5Result.Iterations)
			for i, tc := range v5Result.ToolCalls {
				fmt.Printf("  %d. %s\n", i+1, tc.Name)
			}
		}
	*/

	// ===== V6 Agent 演示 =====
	// 将能力转换为工具定义（V6 也需要）
	caps := mcpServer.GetCapabilitiesAsV4()
	fmt.Println("\n\n" + strings.Repeat("=", 60))
	fmt.Println("🤖 V6 Agent：LLM 自主决策下一步行动")
	fmt.Println(strings.Repeat("=", 60))

	// 1. 创建 V6 Agent 分析器（使用普通 LLM 客户端，不需要 Tool Calling）
	v6Opts := []analyzer.V6AgentOption{}
	if agentTrace {
		v6Opts = append(v6Opts, analyzer.WithAgentVerbose(true))
	}
	v6Analyzer := analyzer.NewV6AgentAnalyzer(
		llmClient, // 使用普通 LLM 客户端
		analyzer.NewRAGRetrieverAdapter(rag.NewMockRetriever()),
		mcp.NewServerAsExecutor(mcpServer),
		caps,
		v6Opts...,
	)

	// 2. 执行 V6 Agent 分析（LLM 自主决定每一步要做什么）
	v6Result, err := v6Analyzer.Analyze(ctx, slowLog)
	if err != nil {
		log.Fatalf("Failed to analyze with V6: %v", err)
	}

	// 输出最终分析结果
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 慢日志分析结果（V6 Agent）")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println(v6Result.FinalResult)

	analyzer.PrintV6AgentSummary(v6Result, agentTrace)
}

// readSlowLogOrFatal 优先读命令行路径，否则读默认 products 慢日志（与 agent-run 一致）。
func readSlowLogOrFatal(fileArg, defaultPath string) string {
	path := defaultPath
	if fileArg != "" {
		path = fileArg
	}
	content, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("read slow log %s: %v", path, err)
	}
	return string(content)
}
