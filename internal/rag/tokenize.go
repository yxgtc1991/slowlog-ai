package rag

import (
	"regexp"
	"strings"
	"unicode"
)

var wordRe = regexp.MustCompile(`[a-z0-9_]+`)

// tokenize 中英文混合分词（英文按词，中文按单字，便于 TF-IDF PoC）。
func tokenize(text string) []string {
	text = strings.ToLower(text)
	var tokens []string
	for _, w := range wordRe.FindAllString(text, -1) {
		if len(w) > 1 {
			tokens = append(tokens, w)
		}
	}
	for _, r := range text {
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
