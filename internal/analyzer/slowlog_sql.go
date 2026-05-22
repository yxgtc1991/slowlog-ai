package analyzer

import (
	"strings"
)

// ExtractSelectSQL 从慢日志文本中提取第一条 SELECT … ; 语句（去注释行后拼接）。
func ExtractSelectSQL(slowLog string) string {
	lines := strings.Split(slowLog, "\n")
	var stmt []string
	inSelect := false
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if trim == "" || strings.HasPrefix(trim, "#") {
			continue
		}
		upper := strings.ToUpper(trim)
		if !inSelect {
			if strings.HasPrefix(upper, "SELECT") {
				inSelect = true
				stmt = append(stmt, trim)
				if strings.HasSuffix(trim, ";") {
					break
				}
			}
			continue
		}
		stmt = append(stmt, trim)
		if strings.HasSuffix(trim, ";") {
			break
		}
	}
	if len(stmt) == 0 {
		return ""
	}
	sql := strings.Join(stmt, " ")
	return strings.TrimSuffix(sql, ";")
}

// NormalizeExplainArgs 若 LLM 给出的 sql 与慢日志不一致，改用慢日志中的 SELECT。
func NormalizeExplainArgs(toolArgs map[string]interface{}, slowLog string) map[string]interface{} {
	want := ExtractSelectSQL(slowLog)
	if want == "" {
		return toolArgs
	}
	got, _ := toolArgs["sql"].(string)
	got = strings.TrimSpace(got)
	if got == "" || !sqlSemanticallyMatches(got, want) {
		if toolArgs == nil {
			toolArgs = make(map[string]interface{})
		}
		toolArgs["sql"] = want
	}
	return toolArgs
}

func sqlSemanticallyMatches(got, want string) bool {
	g := normalizeSQL(got)
	w := normalizeSQL(want)
	return g == w
}

func normalizeSQL(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, ";")
	return strings.Join(strings.Fields(strings.ToLower(s)), " ")
}
