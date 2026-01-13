package main

import (
	"ai_slow_log/internal/prompt"
	"ai_slow_log/internal/rag"
	analyzer "ai_slow_log/internal/slowlog"
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

	// 使用 Mock Retriever 进行 RAG 检索
	// 在实际应用中，可以替换为真实的向量数据库检索器
	retriever := rag.NewMockRetriever()
	ragQuery := "Rows_examined high, Rows_sent low, ORDER BY + LIMIT"
	chunks, err := retriever.Retrieve(ctx, ragQuery)
	if err != nil {
		log.Fatalf("Failed to retrieve RAG chunks: %v", err)
	}

	// 构建 prompt（使用 v3 版本，包含 RAG 知识）
	prompt := prompt.BuildSlowLogPromptV3(slowLog, chunks)

	// 调用 LLM 进行分析
	result, err := analyzer.AnalyzeSlowLog(ctx, prompt)
	if err != nil {
		log.Fatalf("Failed to analyze slow log: %v", err)
	}

	// 输出分析结果
	fmt.Println("=== 慢日志分析结果 ===")
	fmt.Println(result)
}
