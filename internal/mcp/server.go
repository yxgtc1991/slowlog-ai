package mcp

import (
	promptv4 "ai_slow_log/internal/prompt/slowlog"
	"context"
	"fmt"
)

// Server MCP 服务器，提供能力感知和执行功能
type Server struct {
	registry *Registry
}

// NewServer 创建新的 MCP 服务器
func NewServer() *Server {
	return &Server{
		registry: NewRegistry(),
	}
}

// RegisterCapability 注册一个能力
func (s *Server) RegisterCapability(c Capability) {
	s.registry.Register(c)
}

// ListCapabilities 列出所有能力（用于能力感知）
// 返回 JSON 格式的能力列表
// 使用 prompt/slowlog v4 版本的格式化函数
func (s *Server) ListCapabilities() (string, error) {
	metas := s.registry.ListMeta()
	// 转换为 prompt/slowlog 的 CapabilityMetaV4 格式
	v4Metas := make([]promptv4.CapabilityMetaV4, 0, len(metas))
	for _, meta := range metas {
		v4Metas = append(v4Metas, promptv4.ConvertToCapabilityMetaV4(
			meta.Name,
			meta.Description,
			meta.InputSchema,
		))
	}
	return promptv4.FormatCapabilitiesJSON(v4Metas)
}

// GetCapabilityPrompt 获取能力描述 prompt（给 LLM 看）
// 使用 v4 版本的能力感知 prompt
func (s *Server) GetCapabilityPrompt() string {
	caps := s.registry.List()
	// 将 mcp.Capability 转换为 prompt/slowlog.CapabilityV4
	// 由于 Capability 接口与 CapabilityV4 接口方法相同，可以直接转换
	v4Caps := make([]promptv4.CapabilityV4, len(caps))
	for i, c := range caps {
		v4Caps[i] = c // Capability 接口与 CapabilityV4 接口兼容
	}
	return promptv4.BuildCapabilityPromptV4(v4Caps)
}

// ExecuteCapability 执行指定的能力
func (s *Server) ExecuteCapability(ctx context.Context, name string, input map[string]interface{}) (interface{}, error) {
	capability, ok := s.registry.Get(name)
	if !ok {
		return nil, fmt.Errorf("capability '%s' not found", name)
	}

	return capability.Execute(ctx, input)
}

// GetCapability 获取能力实例（用于直接调用）
func (s *Server) GetCapability(name string) (Capability, error) {
	capability, ok := s.registry.Get(name)
	if !ok {
		return nil, fmt.Errorf("capability '%s' not found", name)
	}
	return capability, nil
}

// HasCapability 检查是否支持某个能力
func (s *Server) HasCapability(name string) bool {
	return s.registry.Has(name)
}

// CapabilityCount 返回已注册能力的数量
func (s *Server) CapabilityCount() int {
	return s.registry.Count()
}
