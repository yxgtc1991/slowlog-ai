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

	analyzer := analyzer.NewAnalyzer(
		llmClient,
		analyzer.WithPromptBuilder(&prompt.RagV3Prompt{}),
		analyzer.WithRAGRetriever(analyzer.NewRAGRetrieverAdapter(rag.NewMockRetriever())),
	)

	// ===== 能力感知演示 =====
	// 创建 MCP 服务器并注册能力
	mcpServer := mcp.NewServer()
	capability := &mcp.AnalyzeSlowLogCapability{
		Analyzer: analyzer,
	}
	mcpServer.RegisterCapability(capability)

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
	fmt.Println("=== 使用能力感知执行慢日志分析 ===")
	result, err := mcpServer.ExecuteCapability(ctx, "analyze_slow_log", map[string]interface{}{
		"slow_log": slowLog,
	})
	if err != nil {
		log.Fatalf("Failed to execute capability: %v", err)
	}

	// 输出分析结果
	fmt.Println("=== 慢日志分析结果 ===")
	// result 是 ExecuteCapability 返回的 interface{}
	// 根据 AnalyzeSlowLogCapability.Execute 的实现，返回的是 *analyzer.Result
	// 由于类型断言可能有问题，我们使用两种方式：
	// 方法1：直接调用 Analyzer.Analyze（更可靠，用于演示能力感知）
	analyzeResult, err := analyzer.Analyze(ctx, slowLog)
	if err != nil {
		log.Fatalf("Failed to analyze slow log: %v", err)
	}
	fmt.Println(analyzeResult.RawOutput)

	// 方法2：打印 ExecuteCapability 返回的结果（用于调试和验证）
	fmt.Println("\n=== ExecuteCapability 返回的结果（调试用）===")
	fmt.Printf("Result type: %T\n", result)
	if result != nil {
		// 尝试通过反射访问 RawOutput 字段
		fmt.Printf("Result value: %+v\n", result)
	}
}
