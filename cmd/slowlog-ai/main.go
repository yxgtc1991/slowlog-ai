package main

import (
	"ai_slow_log/internal/analyzer"
	"ai_slow_log/internal/llm"
	"ai_slow_log/internal/mcp"
	prompt "ai_slow_log/internal/prompt/slowlog"
	"ai_slow_log/internal/rag"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	// 示例慢日志
	slowLog := `
# Time: 2025-01-05T10:12:33.123456Z
# User@Host: app_user[app_user] @ 10.0.0.12 []
# Query_time: 12.456  Lock_time: 0.001
# Rows_sent: 10  Rows_examined: 1250000
SET timestamp=1736071953;
SELECT *
FROM orders
WHERE user_id = 123
ORDER BY created_at DESC
LIMIT 10;
`

	// 支持从命令行参数读取慢日志文件
	if len(os.Args) > 1 {
		filePath := os.Args[1]
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Fatalf("Failed to read slow log file: %v", err)
		}
		slowLog = string(content)
	}

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

	// ===== V5 Tool Calling 演示 =====
	fmt.Println("\n\n=== V5 Tool Calling：真正的 MCP/Agent 模式 ===")

	// 1. 创建 Tool Calling 客户端适配器
	toolCallingClient := llm.NewDeepSeekToolCallingAdapter(llmClient)

	// 2. 将能力转换为工具定义
	caps := mcpServer.GetCapabilitiesAsV4()

	// 3. 创建 V5 Tool Calling 分析器
	// 注意：需要使用包名 analyzer，而不是变量 analyzer
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
}
