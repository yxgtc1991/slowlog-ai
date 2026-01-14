package mcp

import "context"

// Capability 是一个 MCP 能力的最小抽象
type Capability interface {
	// Name 能力唯一标识（给 LLM 用）
	Name() string

	// Description 能力说明（告诉 LLM 什么时候用）
	Description() string

	// InputSchema 描述输入参数结构（JSON Schema 简化版）
	InputSchema() map[string]string

	// Execute 真正执行能力
	Execute(ctx context.Context, input map[string]interface{}) (interface{}, error)
}
