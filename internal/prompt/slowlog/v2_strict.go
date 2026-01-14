package prompt

import (
	"ai_slow_log/internal/rag"
	"fmt"
)

type StrictV2Prompt struct{}

func (p *StrictV2Prompt) Build(slowLog string, _ []rag.KnowledgeChunk) string {
	return BuildSlowLogPromptV2(slowLog)
}

// StrictPromptTemplate 是 v2 版本的严格 prompt 模板
// 该模板要求 LLM 严格遵守规则，只基于慢日志内容进行分析，输出 JSON 格式
var StrictPromptTemplate = `
你是一个【MySQL 慢查询分析组件】，不是聊天助手。

你必须严格遵守以下规则：
1. 只能基于提供的慢日志内容进行分析
2. 不允许假设表结构、索引或业务逻辑
3. 不确定的信息必须明确标注为“需要更多信息”
4. 禁止给出无法从日志中直接推导的结论
5. 输出必须是 JSON，不允许任何额外说明文字

你的目标是：
- 识别“可以从慢日志中确定的问题”
- 指出“需要补充哪些信息才能进一步分析”
- 给出“下一步分析建议”，但必须说明依赖条件

输出 JSON 结构如下：
{
  "summary": "一句话总结慢查询主要问题",
  "metrics": {
    "query_time": "",
    "rows_examined": "",
    "rows_sent": "",
    "lock_time": ""
  },
  "confirmed_issues": [
    "可以从日志中100%确认的问题"
  ],
  "suspected_issues": [
    {
      "issue": "可能存在的问题",
      "reason": "基于哪些日志现象推断",
      "confidence": "low | medium"
    }
  ],
  "required_information": [
    "进一步分析必须提供的信息"
  ],
  "next_actions": [
    {
      "action": "下一步建议",
      "depends_on": "依赖哪些补充信息"
    }
  ]
}
`

// BuildSlowLogPromptV2 构建 v2 版本的慢日志分析 prompt
// v2 版本引入了严格的规则约束，要求 LLM 只基于日志内容分析，输出结构化 JSON
func BuildSlowLogPromptV2(slowLog string) string {
	return fmt.Sprintf(`
%s

【慢日志内容】
%s
`, StrictPromptTemplate, slowLog)
}
