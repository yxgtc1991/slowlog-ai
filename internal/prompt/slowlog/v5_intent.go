package prompt

import (
	"encoding/json"
	"fmt"
	"strings"
)

// CapabilityIntentV5 是 v5 版本的能力调用意图
// LLM 输出这个结构，系统负责执行
type CapabilityIntentV5 struct {
	CapabilityName string                 `json:"capability_name"` // 要调用的能力名称
	Input          map[string]interface{} `json:"input"`           // 能力输入参数
	Reason         string                 `json:"reason"`          // 为什么调用这个能力
}

// AnalysisResultV5 是 v5 版本的分析结果，包含分析内容和能力调用意图
type AnalysisResultV5 struct {
	Analysis      interface{}          `json:"analysis"`       // 分析结果（可以是字符串或对象）
	CapabilityUse []CapabilityIntentV5 `json:"capability_use"` // 能力调用意图列表
	NextStep      string               `json:"next_step"`      // 下一步建议
}

// BuildSlowLogPromptV5 构建 v5 版本的慢日志分析 prompt
// v5 版本引入了能力调用意图（Capability Intent），让 LLM 自主决定使用哪些能力
// 这是 Prompt 演进的第五个版本：从能力感知 → 能力调用意图
func BuildSlowLogPromptV5(slowLog string, availableCapabilities []CapabilityV4) string {
	var sb strings.Builder

	// 1. 基础分析要求（类似 v2）
	sb.WriteString(`你是一个【MySQL 慢查询分析组件】，不是聊天助手。

你必须严格遵守以下规则：
1. 只能基于提供的慢日志内容进行分析
2. 不允许假设表结构、索引或业务逻辑
3. 不确定的信息必须明确标注为"需要更多信息"
4. 禁止给出无法从日志中直接推导的结论
5. 输出必须是 JSON，不允许任何额外说明文字

你的目标是：
- 识别"可以从慢日志中确定的问题"
- 指出"需要补充哪些信息才能进一步分析"
- 给出"下一步分析建议"，但必须说明依赖条件

`)

	// 2. 可用能力列表（类似 v4）
	if len(availableCapabilities) > 0 {
		sb.WriteString("【可用系统能力】\n")
		sb.WriteString("你可以通过输出能力调用意图来使用以下系统能力：\n\n")

		for i, c := range availableCapabilities {
			meta := DescribeCapabilityV4(c)
			sb.WriteString(fmt.Sprintf("能力 %d：%s\n", i+1, meta.Name))
			sb.WriteString(fmt.Sprintf("  说明：%s\n", meta.Description))
			sb.WriteString("  输入参数：\n")
			for k, v := range meta.InputSchema {
				sb.WriteString(fmt.Sprintf("    - %s: %s\n", k, v))
			}
			sb.WriteString("\n")
		}

		sb.WriteString(`【能力调用规则】
- 如果你认为某个能力可以帮助进一步分析，请在输出中包含 "capability_use" 字段
- 每个能力调用意图必须包含：capability_name（能力名称）、input（输入参数）、reason（调用原因）
- 可以同时调用多个能力
- 如果不需要调用任何能力，capability_use 可以为空数组

`)
	}

	// 3. 输出格式要求
	sb.WriteString(`【输出格式】
你必须输出严格的 JSON 格式，结构如下：
{
  "analysis": {
    "summary": "一句话总结慢查询主要问题",
    "metrics": {
      "query_time": "",
      "rows_examined": "",
      "rows_sent": "",
      "lock_time": ""
    },
    "confirmed_issues": ["可以从日志中100%确认的问题"],
    "suspected_issues": [
      {
        "issue": "可能存在的问题",
        "reason": "基于哪些日志现象推断",
        "confidence": "low | medium"
      }
    ],
    "required_information": ["进一步分析必须提供的信息"],
    "next_actions": [
      {
        "action": "下一步建议",
        "depends_on": "依赖哪些补充信息"
      }
    ]
  },
  "capability_use": [
    {
      "capability_name": "能力名称",
      "input": {
        "参数名": "参数值"
      },
      "reason": "为什么调用这个能力"
    }
  ],
  "next_step": "基于当前分析和能力调用结果的下一步建议"
}

`)

	// 4. 慢日志内容
	sb.WriteString("【慢日志内容】\n")
	sb.WriteString(slowLog)
	sb.WriteString("\n\n")

	// 5. 重要提示
	sb.WriteString(`【重要提示】
- 分析部分（analysis）必须基于慢日志内容，不能依赖能力调用结果
- 能力调用意图（capability_use）是可选的，只有在需要进一步分析时才添加
- 如果调用能力，reason 字段必须说明为什么需要这个能力
- 输出必须是有效的 JSON，可以直接被程序解析

`)

	return sb.String()
}

// ParseAnalysisResultV5 解析 v5 版本的 LLM 输出
// 尝试从 LLM 输出中提取分析结果和能力调用意图
func ParseAnalysisResultV5(llmOutput string) (*AnalysisResultV5, error) {
	// 尝试解析 JSON
	var result AnalysisResultV5
	if err := json.Unmarshal([]byte(llmOutput), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &result, nil
}

// FormatCapabilityIntentJSON 将能力调用意图格式化为 JSON（用于执行）
func FormatCapabilityIntentJSON(intent CapabilityIntentV5) (string, error) {
	data, err := json.Marshal(intent)
	if err != nil {
		return "", fmt.Errorf("failed to marshal intent: %w", err)
	}
	return string(data), nil
}
