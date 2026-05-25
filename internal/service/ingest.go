package service

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// IngestRequest 日志平台 / Fluent Bit webhook 入参。
type IngestRequest struct {
	SlowLog     string            `json:"slow_log"`
	Source      string            `json:"source,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Guided      *bool             `json:"guided,omitempty"`
	Async       *bool             `json:"async,omitempty"`
	CallbackURL string            `json:"callback_url,omitempty"`
}

var queryTimeRe = regexp.MustCompile(`(?m)^#\s*Query_time:\s*([\d.]+)`)

// ParseIngestRequest 解析 JSON 入参；slow_log 必填。
func ParseIngestRequest(body []byte) (*IngestRequest, error) {
	var req IngestRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("invalid json: %w", err)
	}
	if strings.TrimSpace(req.SlowLog) == "" {
		return nil, fmt.Errorf("slow_log is required")
	}
	return &req, nil
}

// IngestAsyncDefault 未指定 async 时是否异步（webhook 默认 true）。
func (r *IngestRequest) IngestAsyncDefault(defaultAsync bool) bool {
	if r.Async == nil {
		return defaultAsync
	}
	return *r.Async
}

func (r *IngestRequest) GuidedEnabled() bool {
	if r.Guided == nil {
		return false
	}
	return *r.Guided
}

// ParseQueryTimeSeconds 从慢日志头解析 Query_time（秒）；解析失败返回 0。
func ParseQueryTimeSeconds(slowLog string) float64 {
	m := queryTimeRe.FindStringSubmatch(slowLog)
	if len(m) < 2 {
		return 0
	}
	v, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return 0
	}
	return v
}

// PassesIngestThreshold 是否满足最小 Query_time 阈值（0 表示不过滤）。
func PassesIngestThreshold(slowLog string, minQueryTimeSec float64) bool {
	if minQueryTimeSec <= 0 {
		return true
	}
	return ParseQueryTimeSeconds(slowLog) >= minQueryTimeSec
}
