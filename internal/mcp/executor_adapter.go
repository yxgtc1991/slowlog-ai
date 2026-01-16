package mcp

import (
	"ai_slow_log/internal/analyzer"
	"context"
)

// ServerAsExecutor 将 Server 适配为 analyzer.CapabilityExecutor
// 这个适配器允许 V5Analyzer 使用 MCP Server 来执行能力调用
type ServerAsExecutor struct {
	server *Server
}

// NewServerAsExecutor 创建适配器
func NewServerAsExecutor(server *Server) analyzer.CapabilityExecutor {
	return &ServerAsExecutor{server: server}
}

// ExecuteCapability 执行能力
func (a *ServerAsExecutor) ExecuteCapability(ctx context.Context, name string, input map[string]interface{}) (interface{}, error) {
	return a.server.ExecuteCapability(ctx, name, input)
}

// HasCapability 检查能力是否存在
func (a *ServerAsExecutor) HasCapability(name string) bool {
	return a.server.HasCapability(name)
}
