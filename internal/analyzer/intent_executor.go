package analyzer

import (
	"context"
	"fmt"
)

// IntentExecutor 能力调用意图执行器
// 负责解析和执行 LLM 输出的能力调用意图
type IntentExecutor interface {
	// ExecuteIntent 执行单个能力调用意图
	ExecuteIntent(ctx context.Context, intent CapabilityIntent) (interface{}, error)

	// ExecuteIntents 批量执行能力调用意图
	ExecuteIntents(ctx context.Context, intents []CapabilityIntent) (map[string]interface{}, error)
}

// CapabilityIntent 能力调用意图
type CapabilityIntent struct {
	CapabilityName string                 `json:"capability_name"`
	Input          map[string]interface{} `json:"input"`
	Reason         string                 `json:"reason"`
}

// DefaultIntentExecutor 默认的意图执行器实现
type DefaultIntentExecutor struct {
	capabilityExecutor CapabilityExecutor
}

// CapabilityExecutor 能力执行器接口
// 用于执行具体的能力调用
type CapabilityExecutor interface {
	ExecuteCapability(ctx context.Context, name string, input map[string]interface{}) (interface{}, error)
	HasCapability(name string) bool
}

// NewIntentExecutor 创建意图执行器
func NewIntentExecutor(executor CapabilityExecutor) IntentExecutor {
	return &DefaultIntentExecutor{
		capabilityExecutor: executor,
	}
}

// ExecuteIntent 执行单个能力调用意图
func (e *DefaultIntentExecutor) ExecuteIntent(ctx context.Context, intent CapabilityIntent) (interface{}, error) {
	if intent.CapabilityName == "" {
		return nil, fmt.Errorf("capability_name is required")
	}

	if !e.capabilityExecutor.HasCapability(intent.CapabilityName) {
		return nil, fmt.Errorf("capability '%s' not found", intent.CapabilityName)
	}

	return e.capabilityExecutor.ExecuteCapability(ctx, intent.CapabilityName, intent.Input)
}

// ExecuteIntents 批量执行能力调用意图
// 返回一个 map，key 是能力名称，value 是执行结果
func (e *DefaultIntentExecutor) ExecuteIntents(ctx context.Context, intents []CapabilityIntent) (map[string]interface{}, error) {
	results := make(map[string]interface{})

	for _, intent := range intents {
		result, err := e.ExecuteIntent(ctx, intent)
		if err != nil {
			// 记录错误但继续执行其他意图
			results[intent.CapabilityName] = map[string]interface{}{
				"error": err.Error(),
			}
			continue
		}
		results[intent.CapabilityName] = result
	}

	return results, nil
}
