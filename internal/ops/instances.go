package ops

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Instance 注册的 MySQL/业务实例（PoC：元数据 + 工具白名单）。
type Instance struct {
	ID            string   `json:"id"`
	Label         string   `json:"label,omitempty"`
	MySQLDatabase string   `json:"mysql_database,omitempty"`
	AllowedTools  []string `json:"allowed_tools,omitempty"` // 空=不限制
	Disabled      bool     `json:"disabled,omitempty"`
}

type instancesFile struct {
	Instances []Instance `json:"instances"`
}

// Registry 实例目录。
type Registry struct {
	byID map[string]Instance
}

// LoadRegistry 从 JSON 文件加载；文件缺失时仅允许 instance_id=default。
func LoadRegistry(path string) (*Registry, error) {
	r := &Registry{byID: map[string]Instance{
		"default": {ID: "default", Label: "default (.env MySQL)"},
	}}
	if strings.TrimSpace(path) == "" {
		return r, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return r, nil
		}
		return nil, err
	}
	var f instancesFile
	if err := json.Unmarshal(b, &f); err != nil {
		return nil, fmt.Errorf("parse instances: %w", err)
	}
	for _, inst := range f.Instances {
		id := strings.TrimSpace(inst.ID)
		if id == "" {
			continue
		}
		inst.ID = id
		r.byID[id] = inst
	}
	return r, nil
}

func (r *Registry) Resolve(instanceID string) (Instance, error) {
	id := strings.TrimSpace(instanceID)
	if id == "" {
		id = "default"
	}
	inst, ok := r.byID[id]
	if !ok {
		return Instance{}, fmt.Errorf("unknown instance_id %q", id)
	}
	if inst.Disabled {
		return Instance{}, fmt.Errorf("instance %q is disabled", id)
	}
	return inst, nil
}

func (r *Registry) List() []Instance {
	out := make([]Instance, 0, len(r.byID))
	for _, v := range r.byID {
		out = append(out, v)
	}
	return out
}
