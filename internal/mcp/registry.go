package mcp

// CapabilityMeta 是给 LLM 看的能力描述
type CapabilityMeta struct {
	Name        string
	Description string
	InputSchema map[string]string
}

// DescribeCapability 把 Capability 转成 Meta
func DescribeCapability(c Capability) CapabilityMeta {
	return CapabilityMeta{
		Name:        c.Name(),
		Description: c.Description(),
		InputSchema: c.InputSchema(),
	}
}

// Registry 管理所有 MCP 能力
// 用于能力发现和统一管理
type Registry struct {
	capabilities map[string]Capability
}

// NewRegistry 创建新的能力注册表
func NewRegistry() *Registry {
	return &Registry{
		capabilities: make(map[string]Capability),
	}
}

// Register 注册一个能力
func (r *Registry) Register(c Capability) {
	r.capabilities[c.Name()] = c
}

// Get 根据名称获取能力
func (r *Registry) Get(name string) (Capability, bool) {
	c, ok := r.capabilities[name]
	return c, ok
}

// List 列出所有已注册的能力
func (r *Registry) List() []Capability {
	list := make([]Capability, 0, len(r.capabilities))
	for _, c := range r.capabilities {
		list = append(list, c)
	}
	return list
}

// ListMeta 列出所有能力的元数据（用于能力感知）
func (r *Registry) ListMeta() []CapabilityMeta {
	metas := make([]CapabilityMeta, 0, len(r.capabilities))
	for _, c := range r.capabilities {
		metas = append(metas, DescribeCapability(c))
	}
	return metas
}

// Count 返回已注册能力的数量
func (r *Registry) Count() int {
	return len(r.capabilities)
}

// Has 检查是否已注册某个能力
func (r *Registry) Has(name string) bool {
	_, ok := r.capabilities[name]
	return ok
}
