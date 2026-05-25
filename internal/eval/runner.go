package eval

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ai_slow_log/internal/analyzer"
	promptv6 "ai_slow_log/internal/prompt/slowlog"
	"ai_slow_log/internal/rag"
)

// RunCase 执行单条 golden case。
func RunCase(ctx context.Context, c Case) Result {
	res := Result{CaseName: c.Name}

	slowLog, err := readProjectFile(c.SlowLogPath)
	if err != nil {
		res.Errors = []string{fmt.Sprintf("read slow log: %v", err)}
		return res
	}

	llm := NewScriptLLM(c.Script...)
	exec := c.Executor
	if exec == nil {
		exec = NewStubExecutor()
	}

	opts := []analyzer.V6AgentOption{
		analyzer.WithAgentRecordRounds(true),
	}
	if c.Guide {
		opts = append(opts, analyzer.WithAgentGuide(promptv6.GuidedSlowLogPreamble))
	}
	if len(c.HITLReplies) > 0 {
		opts = append(opts, analyzer.WithAgentUserInput(analyzer.NewStubUserInput(c.HITLReplies...)))
	}

	v6 := analyzer.NewV6AgentAnalyzer(
		llm,
		analyzer.NewRAGRetrieverAdapter(rag.NewMockRetriever()),
		exec,
		nil,
		opts...,
	)

	out, err := v6.Analyze(ctx, string(slowLog))
	if err != nil {
		res.Errors = []string{fmt.Sprintf("analyze: %v", err)}
		return res
	}

	res.Iterations = out.Iterations
	res.Actions = formatActions(out.Actions)
	res.Final = truncate(out.FinalResult, 200)
	res.Errors = assertResult(out, c.Expect)
	res.Pass = len(res.Errors) == 0
	return res
}

// RunAll 运行全部（或按名称过滤）cases。
func RunAll(ctx context.Context, filter string) []Result {
	cases := AllCases()
	if filter != "" {
		filtered := make([]Case, 0)
		for _, c := range cases {
			if c.Name == filter {
				filtered = append(filtered, c)
			}
		}
		cases = filtered
	}
	results := make([]Result, 0, len(cases))
	for _, c := range cases {
		results = append(results, RunCase(ctx, c))
	}
	return results
}

// readProjectFile 从仓库根或 go test 的包目录解析相对路径。
func readProjectFile(rel string) ([]byte, error) {
	candidates := []string{rel}
	dir := "."
	for i := 0; i < 4; i++ {
		dir = filepath.Join(dir, "..")
		candidates = append(candidates, filepath.Join(dir, rel))
	}
	for _, p := range candidates {
		b, err := os.ReadFile(p)
		if err == nil {
			return b, nil
		}
	}
	return nil, fmt.Errorf("cannot read %q (tried from cwd %s)", rel, mustGetwd())
}

func mustGetwd() string {
	wd, _ := os.Getwd()
	return wd
}

func truncate(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}

// FailedCount 失败条数。
func FailedCount(results []Result) int {
	n := 0
	for _, r := range results {
		if !r.Pass {
			n++
		}
	}
	return n
}
