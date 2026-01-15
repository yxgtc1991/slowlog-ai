package analyzer

import (
	"ai_slow_log/internal/rag"
	"context"
	"fmt"
)

// SlowLogAnalyzer 慢日志分析器
type SlowLogAnalyzer struct {
	llm           LLMClient
	retriever     Retriever
	promptBuilder PromptBuilder
	promptVersion PromptVersion
}

// NewAnalyzer 创建新的慢日志分析器
// 使用函数式选项模式（Functional Options Pattern）来处理可选配置
//
// 参数：
//   - llm: LLM 客户端（必需）
//   - opts: 可变参数，使用 WithXxx 函数设置可选配置
//
// 示例：
//
//	analyzer := NewAnalyzer(
//	    llmClient,
//	    WithPromptBuilder(&prompt.RagV3Prompt{}),
//	    WithRAGRetriever(retriever),
//	)
func NewAnalyzer(
	llm LLMClient,
	opts ...Option,
) *SlowLogAnalyzer {

	a := &SlowLogAnalyzer{
		llm:           llm,
		promptVersion: PromptV2,
	}

	// 应用所有选项函数，依次修改配置
	for _, opt := range opts {
		opt(a)
	}

	return a
}

// Result 分析结果
type Result struct {
	RawOutput string
}

func (a *SlowLogAnalyzer) Analyze(ctx context.Context, slowLog string) (*Result, error) {
	if slowLog == "" {
		return nil, fmt.Errorf("slow log is empty")
	}

	if a.promptBuilder == nil {
		return nil, fmt.Errorf("prompt builder is not set, use WithPromptBuilder option")
	}

	var chunks []rag.KnowledgeChunk
	if a.retriever != nil {
		var err error
		chunks, err = a.retriever.Retrieve(ctx, slowLog)
		if err != nil {
			return nil, err
		}
	}

	prompt := a.promptBuilder.Build(slowLog, chunks)

	out, err := a.llm.Chat(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return &Result{RawOutput: out}, nil
}
