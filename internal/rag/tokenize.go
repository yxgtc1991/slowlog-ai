package rag

import (
	"regexp"
	"strings"
	"unicode"
)

var wordRe = regexp.MustCompile(`[a-z0-9_]+`)

// 慢日志领域常用短语（优先整词匹配，提升 TF-IDF 命中率）。
var domainPhrases = []string{
	"最左前缀", "复合索引", "全表扫描", "左前缀", "filesort",
	"rows_examined", "query_time", "lock_time", "dry_run",
	"created_at", "order_by",
}

// tokenize 中英文混合分词（英文按词；中文短语优先，其余按单字）。
func tokenize(text string) []string {
	text = strings.ToLower(text)
	var tokens []string
	remaining := text
	for _, p := range domainPhrases {
		pLower := strings.ToLower(p)
		for strings.Contains(remaining, pLower) {
			tokens = append(tokens, pLower)
			remaining = strings.Replace(remaining, pLower, " ", 1)
		}
	}
	for _, w := range wordRe.FindAllString(remaining, -1) {
		if len(w) > 1 {
			tokens = append(tokens, w)
		}
	}
	for _, r := range remaining {
		if unicode.Is(unicode.Han, r) {
			tokens = append(tokens, string(r))
		}
	}
	return tokens
}

func termFreq(tokens []string) map[string]float64 {
	tf := make(map[string]float64)
	for _, t := range tokens {
		tf[t]++
	}
	n := float64(len(tokens))
	if n == 0 {
		return tf
	}
	for k, v := range tf {
		tf[k] = v / n
	}
	return tf
}
