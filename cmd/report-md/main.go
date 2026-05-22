// 从已保存的 agent-run JSON 重新生成全部报告（无需重跑 Agent）。
package main

import (
	"ai_slow_log/internal/analyzer"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: go run ./cmd/report-md reports/agent-run-YYYYMMDD-HHMMSS.json")
	}
	jsonPath := os.Args[1]
	jb, err := os.ReadFile(jsonPath)
	if err != nil {
		log.Fatalf("read: %v", err)
	}
	var report analyzer.V6AgentRunReport
	if err := json.Unmarshal(jb, &report); err != nil {
		log.Fatalf("parse: %v", err)
	}
	base := strings.TrimSuffix(jsonPath, filepath.Ext(jsonPath))
	jsonBase := filepath.Base(jsonPath)
	write := func(path, content string) {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			log.Fatalf("write %s: %v", path, err)
		}
	}
	write(base+".md", analyzer.FormatV6ReportMarkdownForFile(&report, jsonBase))
	write(base+".html", analyzer.FormatV6ReportHTML(&report, jsonBase))
	write(base+".brief.md", analyzer.FormatV6ReportBriefMarkdown(&report))
	write(base+".brief.html", analyzer.FormatV6ReportBriefHTML(&report))
	fmt.Println("已写入:")
	fmt.Println("  精简 HTML:", base+".brief.html")
	fmt.Println("  精简 MD:  ", base+".brief.md")
	fmt.Println("  完整 HTML:", base+".html")
	fmt.Println("  完整 MD:  ", base+".md")
}
