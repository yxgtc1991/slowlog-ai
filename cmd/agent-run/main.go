// 完整跑一遍 V6 Agent，并将每轮 LLM / RAG / MCP 输出保存到 reports/（便于复盘，无需重复消耗 Token）。
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
	"path/filepath"
	"strings"
)

func main() {
	_ = config.LoadDotEnv(".env")

	reportDir := "reports"
	slowLogPath := "testdata/slowlog-products.txt"
	guided := true
	trace := true

	for _, arg := range os.Args[1:] {
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

	mcpServer := mcp.NewServer()
	mcpServer.RegisterCapability(&mcp.AnalyzeSlowLogCapability{
		Analyzer: analyzer.NewAnalyzer(
			llmClient,
			analyzer.WithPromptBuilder(&prompt.RagV3Prompt{}),
			analyzer.WithRAGRetriever(analyzer.NewRAGRetrieverAdapter(rag.NewMockRetriever())),
		),
	})

	if mysqlCfg, err := config.MustLoadMySQL(); err != nil {
		log.Printf("mysql: %v (Agent 将跳过 MySQL 相关工具)", err)
	} else {
		client, err := mysql.NewClient(mysqlCfg)
		if err != nil {
			log.Printf("mysql: connect: %v", err)
		} else {
			defer client.Close()
			mcp.RegisterMySQLCapabilities(mcpServer, client)
			fmt.Fprintf(os.Stderr, "MySQL: %s:%d db=%s\n", mysqlCfg.Host, mysqlCfg.Port, mysqlCfg.Database)
		}
	}

	caps := mcpServer.GetCapabilitiesAsV4()
	toolNames := make([]string, 0, len(caps))
	for _, c := range caps {
		toolNames = append(toolNames, c.Name())
	}

	opts := []analyzer.V6AgentOption{
		analyzer.WithAgentRecordRounds(true),
		analyzer.WithAgentVerbose(trace),
	}
	if guided {
		opts = append(opts, analyzer.WithAgentGuide(prompt.GuidedSlowLogPreamble))
	}
	v6 := analyzer.NewV6AgentAnalyzer(
		llmClient,
		analyzer.NewRAGRetrieverAdapter(rag.NewMockRetriever()),
		mcp.NewServerAsExecutor(mcpServer),
		caps,
		opts...,
	)

	if guided {
		fmt.Fprintln(os.Stderr, "ℹ️  已启用 guided 推荐流程（RAG → MySQL → EXPLAIN → 索引 dry_run → finish）")
	}

	fmt.Fprintln(os.Stderr, "▶ 开始 V6 Agent 分析（结果将写入 reports/）...")
	result, err := v6.Analyze(ctx, string(slowLog))
	if err != nil {
		log.Fatalf("agent: %v", err)
	}

	report := analyzer.BuildV6RunReport(string(slowLog), result, toolNames)
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
