package prompt

import (
	"encoding/json"
	"fmt"
	"strings"
)

// CapabilityV4 是 v4 版本的能力接口（避免循环依赖）
// 这个接口与 mcp.Capability 相同，但定义在这里以避免循环导入
type CapabilityV4 interface {
	Name() string
	Description() string
	InputSchema() map[string]string
}

// CapabilityMetaV4 是给 LLM 看的能力描述（v4 版本）
type CapabilityMetaV4 struct {
	Name        string
	Description string
	InputSchema map[string]string
}

// DescribeCapabilityV4 把 CapabilityV4 转成 Meta
func DescribeCapabilityV4(c CapabilityV4) CapabilityMetaV4 {
	return CapabilityMetaV4{
		Name:        c.Name(),
		Description: c.Description(),
		InputSchema: c.InputSchema(),
	}
}

// BuildCapabilityPromptV4 构建 v4 版本的能力描述 prompt（给 LLM 看）
// v4 版本引入了能力感知（Capability Awareness），可以自动发现和描述系统能力
// 这是 Prompt 演进的第四个版本，从简单的提示 → 严格模式 → RAG 增强 → 能力感知
func BuildCapabilityPromptV4(caps []CapabilityV4) string {
	var sb strings.Builder

	sb.WriteString("你可以使用以下系统能力（Tools）：\n\n")

	for i, c := range caps {
		meta := DescribeCapabilityV4(c)

		sb.WriteString(fmt.Sprintf("【能力 %d】%s\n", i+1, meta.Name))
		sb.WriteString(fmt.Sprintf("  说明：%s\n", meta.Description))
		sb.WriteString("  输入参数：\n")

		for k, v := range meta.InputSchema {
			sb.WriteString(fmt.Sprintf("    - %s: %s\n", k, v))
		}

		sb.WriteString("\n")
	}

	sb.WriteString("当你认为某个能力可以帮助解决问题时，请说明你想使用哪个能力，以及对应的输入参数。\n")

	return sb.String()
}

// BuildCapabilityPromptV4FromList 从能力列表生成能力描述 prompt（v4 版本）
// 这个函数接收实现了 CapabilityV4 接口的切片
func BuildCapabilityPromptV4FromList(caps []CapabilityV4) string {
	return BuildCapabilityPromptV4(caps)
}

// FormatCapabilitiesJSON 将能力列表格式化为 JSON（用于 API 返回）
func FormatCapabilitiesJSON(metas []CapabilityMetaV4) (string, error) {
	data, err := json.MarshalIndent(metas, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal capabilities: %w", err)
	}
	return string(data), nil
}

// FormatCapabilitiesMarkdown 将能力列表格式化为 Markdown（用于文档）
func FormatCapabilitiesMarkdown(metas []CapabilityMetaV4) string {
	var sb strings.Builder

	sb.WriteString("# 可用能力列表\n\n")

	for i, meta := range metas {
		sb.WriteString(fmt.Sprintf("## %d. %s\n\n", i+1, meta.Name))
		sb.WriteString(fmt.Sprintf("**说明**：%s\n\n", meta.Description))
		sb.WriteString("**输入参数**：\n\n")

		if len(meta.InputSchema) == 0 {
			sb.WriteString("无\n\n")
		} else {
			sb.WriteString("| 参数名 | 类型 | 说明 |\n")
			sb.WriteString("|--------|------|------|\n")
			for k, v := range meta.InputSchema {
				// 解析类型和说明（格式：type // description）
				parts := strings.SplitN(v, "//", 2)
				paramType := strings.TrimSpace(parts[0])
				description := ""
				if len(parts) > 1 {
					description = strings.TrimSpace(parts[1])
				}
				sb.WriteString(fmt.Sprintf("| `%s` | %s | %s |\n", k, paramType, description))
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// ConvertToCapabilityMetaV4 将 mcp.CapabilityMeta 转换为 CapabilityMetaV4
// 这个函数用于在 mcp 包中使用，避免循环依赖
func ConvertToCapabilityMetaV4(name, description string, inputSchema map[string]string) CapabilityMetaV4 {
	return CapabilityMetaV4{
		Name:        name,
		Description: description,
		InputSchema: inputSchema,
	}
}
