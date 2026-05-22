// agent-eval 运行 V6 Agent 的 golden cases（脚本化 LLM，无 API Key）。
package main

import (
	"ai_slow_log/internal/eval"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	caseName := flag.String("case", "", "只跑指定 case 名称（如 guided_flow）")
	reportJSON := flag.String("report", "", "对已保存的 agent-run JSON 做断言（不跑 Agent）")
	verbose := flag.Bool("v", false, "打印每条 case 的轨迹与结论摘要")
	flag.Parse()

	if *reportJSON != "" {
		runReportAssert(*reportJSON, *verbose)
		return
	}

	ctx := context.Background()
	results := eval.RunAll(ctx, *caseName)
	if *caseName != "" && len(results) == 0 {
		fmt.Fprintf(os.Stderr, "unknown case: %s\n", *caseName)
		os.Exit(2)
	}

	pass, fail := 0, 0
	for _, r := range results {
		if r.Pass {
			pass++
		} else {
			fail++
		}
		printResult(r, *verbose)
	}

	fmt.Println()
	fmt.Printf("Agent Eval: %d passed, %d failed (total %d)\n", pass, fail, pass+fail)
	if fail > 0 {
		os.Exit(1)
	}
}

func runReportAssert(path string, verbose bool) {
	// 默认断言：至少 3 轮、含 RAG、结论含「扫描」或「索引」类关键词（宽松，适配真实 LLM 输出）
	errs := eval.AssertReportFile(path, eval.ReportExpect{
		MinRounds:     3,
		MustHaveTools: []string{"explain_mysql_query"},
		FinalContains: []string{"price"},
	})
	if verbose && len(errs) == 0 {
		fmt.Println("report:", path, "OK")
	}
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Println("FAIL:", e)
		}
		os.Exit(1)
	}
	fmt.Println("report assert OK:", path)
}

func printResult(r eval.Result, verbose bool) {
	status := "PASS"
	if !r.Pass {
		status = "FAIL"
	}
	fmt.Printf("[%s] %s  rounds=%d\n", status, r.CaseName, r.Iterations)
	if verbose || !r.Pass {
		fmt.Printf("  trajectory: %s\n", strings.Join(r.Actions, " → "))
		if r.Final != "" {
			fmt.Printf("  final: %s\n", r.Final)
		}
	}
	for _, e := range r.Errors {
		fmt.Printf("  ✗ %s\n", e)
	}
}
