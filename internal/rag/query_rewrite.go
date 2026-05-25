package rag

import (
	"regexp"
	"strings"
)

var (
	reFromTable = regexp.MustCompile(`(?i)(?:FROM|JOIN|INTO|UPDATE)\s+([a-zA-Z_][a-zA-Z0-9_]*)`)
	reMetric    = regexp.MustCompile(`(?i)(Query_time|Lock_time|Rows_sent|Rows_examined|Using filesort|type:\s*ALL|type\s+ALL)`)
)

// RewriteQueries 基于 LLM rag_query 与慢日志规则抽取，生成多路检索查询（去重，最多 maxQueries 条）。
func RewriteQueries(llmQuery, slowLog string) []string {
	primary := strings.TrimSpace(llmQuery)
	if primary == "" {
		primary = "慢 SQL 优化 索引"
	}
	seen := map[string]struct{}{}
	add := func(q string) []string {
		q = strings.TrimSpace(q)
		if q == "" {
			return nil
		}
		key := strings.ToLower(q)
		if _, ok := seen[key]; ok {
			return nil
		}
		seen[key] = struct{}{}
		return []string{q}
	}

	out := add(primary)
	hints := extractSlowLogHints(slowLog)
	if len(hints) > 0 {
		out = append(out, add(strings.Join(append([]string{primary}, hints...), " "))...)
		if len(hints) >= 2 {
			out = append(out, add(strings.Join(hints, " "))...)
		}
	}
	for _, h := range hints {
		if strings.Contains(strings.ToLower(primary), strings.ToLower(h)) {
			continue
		}
		out = append(out, add(primary+" "+h)...)
	}

	const maxQueries = 4
	if len(out) > maxQueries {
		out = out[:maxQueries]
	}
	return out
}

func extractSlowLogHints(slowLog string) []string {
	slowLog = strings.TrimSpace(slowLog)
	if slowLog == "" {
		return nil
	}
	seen := map[string]struct{}{}
	var hints []string
	add := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" {
			return
		}
		key := strings.ToLower(s)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		hints = append(hints, s)
	}

	for _, m := range reFromTable.FindAllStringSubmatch(slowLog, -1) {
		if len(m) > 1 {
			add(m[1])
		}
	}
	for _, m := range reMetric.FindAllStringSubmatch(slowLog, -1) {
		if len(m) > 1 {
			add(strings.ReplaceAll(m[1], ":", " "))
		}
	}
	lower := strings.ToLower(slowLog)
	switch {
	case strings.Contains(lower, "filesort"):
		add("filesort ORDER BY")
	case strings.Contains(lower, "order by"):
		add("ORDER BY filesort")
	}
	if strings.Contains(lower, "rows_examined") {
		add("Rows_examined 扫描行数")
	}
	if strings.Contains(lower, "lock_time") && strings.Contains(lower, "query_time") {
		add("Lock_time 锁等待")
	}
	if strings.Contains(lower, "join") {
		add("JOIN 驱动表")
	}
	if strings.Contains(lower, "offset") {
		add("深分页 OFFSET")
	}
	return hints
}
