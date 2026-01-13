package slowlog

import (
	"regexp"
	"strconv"
	"strings"
)

// SlowLogMetrics 慢日志指标
type SlowLogMetrics struct {
	QueryTime    float64
	LockTime     float64
	RowsSent     int
	RowsExamined int
	SQL          string
}

// ParseSlowLog 解析慢日志文本，提取关键指标
func ParseSlowLog(slowLog string) (*SlowLogMetrics, error) {
	metrics := &SlowLogMetrics{}

	// 提取 Query_time
	if matches := regexp.MustCompile(`Query_time:\s+([\d.]+)`).FindStringSubmatch(slowLog); len(matches) > 1 {
		if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
			metrics.QueryTime = val
		}
	}

	// 提取 Lock_time
	if matches := regexp.MustCompile(`Lock_time:\s+([\d.]+)`).FindStringSubmatch(slowLog); len(matches) > 1 {
		if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
			metrics.LockTime = val
		}
	}

	// 提取 Rows_sent
	if matches := regexp.MustCompile(`Rows_sent:\s+(\d+)`).FindStringSubmatch(slowLog); len(matches) > 1 {
		if val, err := strconv.Atoi(matches[1]); err == nil {
			metrics.RowsSent = val
		}
	}

	// 提取 Rows_examined
	if matches := regexp.MustCompile(`Rows_examined:\s+(\d+)`).FindStringSubmatch(slowLog); len(matches) > 1 {
		if val, err := strconv.Atoi(matches[1]); err == nil {
			metrics.RowsExamined = val
		}
	}

	// 提取 SQL 语句（SET timestamp 之后的内容）
	lines := strings.Split(slowLog, "\n")
	var sqlLines []string
	inSQL := false
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "SET timestamp") {
			inSQL = true
			continue
		}
		if inSQL {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" {
				sqlLines = append(sqlLines, trimmed)
			}
		}
	}
	metrics.SQL = strings.Join(sqlLines, " ")

	return metrics, nil
}
