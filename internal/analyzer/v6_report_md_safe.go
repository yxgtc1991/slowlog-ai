package analyzer

import (
	"fmt"
	"strings"
)

// mdSafeText 转义 <，避免 CommonMark / GoLand 将 <nil> 等当成 HTML 导致预览空白。
func mdSafeText(s string) string {
	s = sanitizeUTF8(s)
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	return s
}

// mdTableCell 表格单元格：转义 HTML 与竖线，去掉换行。
func mdTableCell(s string) string {
	s = mdSafeText(s)
	s = strings.ReplaceAll(s, "|", "｜")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.Join(strings.Fields(s), " ")
	return s
}

func mdCodeFence(lang, body string) string {
	body = strings.TrimSuffix(mdSafeText(body), "\n")
	fence := "```"
	for strings.Contains(body, fence) {
		fence += "`"
	}
	if lang != "" {
		return fence + lang + "\n" + body + "\n" + fence + "\n"
	}
	return fence + "\n" + body + "\n" + fence + "\n"
}

// formatGoValueForMD 避免 %v 把 nil 打成 <nil>。
func formatGoValueForMD(v interface{}) string {
	if v == nil {
		return "null"
	}
	return mdSafeText(fmt.Sprint(v))
}

// limitLineLengthOutsideFences 仅折行普通段落，不破坏 ``` 围栏块。
func limitLineLengthOutsideFences(s string, maxRunes int) string {
	if maxRunes <= 0 {
		return s
	}
	var out strings.Builder
	inFence := false
	for _, line := range strings.Split(s, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "```") {
			inFence = !inFence
			out.WriteString(line)
			out.WriteByte('\n')
			continue
		}
		if inFence || strings.HasPrefix(line, "    ") || strings.HasPrefix(strings.TrimSpace(line), "|") {
			out.WriteString(line)
			out.WriteByte('\n')
			continue
		}
		out.WriteString(wrapRunes(line, maxRunes))
		out.WriteByte('\n')
	}
	return strings.TrimSuffix(out.String(), "\n")
}

func wrapRunes(line string, maxRunes int) string {
	runes := []rune(line)
	if len(runes) <= maxRunes {
		return line
	}
	var parts []string
	for len(runes) > maxRunes {
		parts = append(parts, string(runes[:maxRunes]))
		runes = runes[maxRunes:]
	}
	if len(runes) > 0 {
		parts = append(parts, string(runes))
	}
	return strings.Join(parts, "\n")
}
