package main

import (
	"ai_slow_log/internal/analyzer"
	"ai_slow_log/internal/llm"
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

	result, err := analyzer.Analyze(ctx, slowLog)
	if err != nil {
		log.Fatalf("Failed to analyze slow log: %v", err)
	}

	// 输出分析结果
	fmt.Println("=== 慢日志分析结果 ===")
	fmt.Println(result.RawOutput)
}
