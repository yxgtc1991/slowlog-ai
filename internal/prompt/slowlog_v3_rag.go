package prompt

import (
	"ai_slow_log/internal/rag"
	"fmt"
	"strings"
)

// BuildSlowLogPromptV3 构建 v3 版本的慢日志分析 prompt（带 RAG）
// v3 版本在 v2 的基础上增加了 RAG 检索的知识块，用于辅助分析
// ragChunks: 从 RAG 系统检索到的相关知识块
func BuildSlowLogPromptV3(slowLog string, ragChunks []rag.KnowledgeChunk) string {
	if len(ragChunks) == 0 {
		// 如果没有 RAG 知识块，降级到 v2
		return BuildSlowLogPromptV2(slowLog)
	}

	var sb strings.Builder
	for _, c := range ragChunks {
		sb.WriteString(fmt.Sprintf(
			"- %s：%s\n",
			c.Title,
			c.Content,
		))
	}

	return fmt.Sprintf(`
%s

【可参考的专家知识（仅用于推断 suspected_issues，不得作为 confirmed_issues）】
%s
`,
		BuildSlowLogPromptV2(slowLog),
		sb.String(),
	)
}
