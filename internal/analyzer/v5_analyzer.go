package analyzer

import (
	promptv5 "ai_slow_log/internal/prompt/slowlog"
	"context"
	"encoding/json"
	"fmt"
)

// V5Analyzer v5 版本的分析器
// 支持 LLM 输出能力调用意图，系统自动执行
type V5Analyzer struct {
	llm              LLMClient
	retriever        Retriever
	intentExecutor   IntentExecutor
	availableCaps    []promptv5.CapabilityV4
	maxIterations    int // 最大迭代次数，防止无限循环
	currentIteration int
}

// V5AnalyzerOption V5 分析器的配置选项
type V5AnalyzerOption func(*V5Analyzer)

// WithIntentExecutor 设置意图执行器
func WithIntentExecutor(executor IntentExecutor) V5AnalyzerOption {
	return func(a *V5Analyzer) {
		a.intentExecutor = executor
	}
}

// WithAvailableCapabilities 设置可用能力列表
func WithAvailableCapabilities(caps []promptv5.CapabilityV4) V5AnalyzerOption {
	return func(a *V5Analyzer) {
		a.availableCaps = caps
	}
}

// WithMaxIterations 设置最大迭代次数
func WithMaxIterations(max int) V5AnalyzerOption {
	return func(a *V5Analyzer) {
		a.maxIterations = max
	}
}

// NewV5Analyzer 创建 v5 版本的分析器
func NewV5Analyzer(
	llm LLMClient,
	retriever Retriever,
	opts ...V5AnalyzerOption,
) *V5Analyzer {
	analyzer := &V5Analyzer{
		llm:           llm,
		retriever:     retriever,
		maxIterations: 3, // 默认最多 3 轮交互
	}

	for _, opt := range opts {
		opt(analyzer)
	}

	return analyzer
}

// V5Result v5 版本的分析结果
type V5Result struct {
	Analysis        string                 // 最终分析结果
	CapabilityCalls map[string]interface{} // 能力调用结果
	Iterations      int                    // 实际迭代次数
	RawOutput       string                 // 原始 LLM 输出
}

// Analyze 执行 v5 版本的分析
// 支持多轮交互：LLM 输出意图 → 系统执行 → LLM 继续分析
func (a *V5Analyzer) Analyze(ctx context.Context, slowLog string) (*V5Result, error) {
	if slowLog == "" {
		return nil, fmt.Errorf("slow log is empty")
	}

	a.currentIteration = 0
	capabilityResults := make(map[string]interface{})

	// 第一轮：初始分析
	prompt := promptv5.BuildSlowLogPromptV5(slowLog, a.availableCaps)
	llmOutput, err := a.llm.Chat(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to get LLM response: %w", err)
	}

	// 解析 LLM 输出
	result, err := promptv5.ParseAnalysisResultV5(llmOutput)
	if err != nil {
		// 如果解析失败，返回原始输出
		return &V5Result{
			Analysis:        llmOutput,
			CapabilityCalls: capabilityResults,
			Iterations:      1,
			RawOutput:       llmOutput,
		}, nil
	}

	// 提取分析结果
	var finalAnalysis string
	if result.Analysis != nil {
		analysisBytes, err := json.MarshalIndent(result.Analysis, "", "  ")
		if err == nil {
			finalAnalysis = string(analysisBytes)
		} else {
			// 如果是字符串，直接使用
			if str, ok := result.Analysis.(string); ok {
				finalAnalysis = str
			} else {
				finalAnalysis = llmOutput
			}
		}
	} else {
		finalAnalysis = llmOutput
	}

	// 执行能力调用意图
	if len(result.CapabilityUse) > 0 && a.intentExecutor != nil {
		a.currentIteration++
		if a.currentIteration <= a.maxIterations {
			// 转换为执行器需要的格式
			intents := make([]CapabilityIntent, len(result.CapabilityUse))
			for i, cu := range result.CapabilityUse {
				intents[i] = CapabilityIntent{
					CapabilityName: cu.CapabilityName,
					Input:          cu.Input,
					Reason:         cu.Reason,
				}
			}

			// 执行能力调用
			results, err := a.intentExecutor.ExecuteIntents(ctx, intents)
			if err != nil {
				// 记录错误但继续
				capabilityResults["error"] = err.Error()
			} else {
				capabilityResults = results
			}

			// 如果有能力调用结果，可以继续让 LLM 基于结果进行进一步分析
			// 这里简化处理，直接返回结果
		}
	}

	return &V5Result{
		Analysis:        finalAnalysis,
		CapabilityCalls: capabilityResults,
		Iterations:      a.currentIteration + 1,
		RawOutput:       llmOutput,
	}, nil
}
