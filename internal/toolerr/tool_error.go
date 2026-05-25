package toolerr

import (
	"errors"
	"strings"
)

// 工具错误码（MCP / Agent / 报告统一消费）。
const (
	CodeCapabilityNotFound = "capability_not_found"
	CodeNotConfigured      = "not_configured"
	CodeInvalidInput       = "invalid_input"
	CodeMySQLConnection    = "mysql_connection"
	CodeMySQLTableNotFound = "mysql_table_not_found"
	CodeMySQLQuery         = "mysql_query"
	CodeInternal           = "internal"
)

// ToolError 结构化工具失败。
type ToolError struct {
	Tool      string `json:"tool"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
}

// ToMap 写入报告 action_outcome。
func (e ToolError) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"ok":        false,
		"tool":      e.Tool,
		"code":      e.Code,
		"message":   e.Message,
		"retryable": e.Retryable,
	}
}

// Classify 将 Go error 映射为统一错误码与是否可重试。
func Classify(toolName string, err error) ToolError {
	if err == nil {
		return ToolError{Tool: toolName, Code: CodeInternal, Message: "nil error", Retryable: false}
	}
	msg := err.Error()
	lower := strings.ToLower(msg)

	te := ToolError{Tool: toolName, Message: msg}

	switch {
	case strings.Contains(lower, "capability") && strings.Contains(lower, "not found"):
		te.Code = CodeCapabilityNotFound
	case strings.Contains(lower, "not configured"):
		te.Code = CodeNotConfigured
	case strings.Contains(lower, "required"), strings.Contains(lower, "invalid"):
		te.Code = CodeInvalidInput
	case strings.Contains(lower, "doesn't exist"), strings.Contains(lower, "42s02"):
		te.Code = CodeMySQLTableNotFound
	case strings.Contains(lower, "connection refused"),
		strings.Contains(lower, "timeout"),
		strings.Contains(lower, "dial tcp"),
		strings.Contains(lower, "can't connect"):
		te.Code = CodeMySQLConnection
		te.Retryable = true
		return te
	case strings.Contains(lower, "explain:"), strings.Contains(lower, "mysql"):
		te.Code = CodeMySQLQuery
	default:
		te.Code = CodeInternal
	}
	return te
}

// ErrCoded 供 MCP 能力返回带码错误（可选）。
type ErrCoded struct {
	Code      string
	Message   string
	Retryable bool
	Tool      string
}

func (e *ErrCoded) Error() string { return e.Message }

// New 构造带码错误。
func New(tool, code, message string, retryable bool) error {
	return &ErrCoded{Tool: tool, Code: code, Message: message, Retryable: retryable}
}

// From 若 err 为 ErrCoded 则直接转换，否则 Classify。
func From(toolName string, err error) ToolError {
	var ec *ErrCoded
	if errors.As(err, &ec) {
		return ToolError{
			Tool:      firstNonEmpty(ec.Tool, toolName),
			Code:      ec.Code,
			Message:   ec.Message,
			Retryable: ec.Retryable,
		}
	}
	return Classify(toolName, err)
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
