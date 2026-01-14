package analyzer

// PromptVersion 表示 Prompt 版本
type PromptVersion string

const (
	PromptV2 PromptVersion = "v2"
	PromptV3 PromptVersion = "v3"
)

// Option 是函数式选项模式的类型定义
// 这是一个函数类型，接收 SlowLogAnalyzer 指针并修改其配置
// 使用函数式选项模式的优势：
// 1. 可扩展：新增配置项只需添加新的 WithXxx 函数，无需修改构造函数签名
// 2. 灵活：可选参数，顺序无关，只设置需要的配置
// 3. 可读：函数名清晰表达意图，自文档化
// 4. 类型安全：编译期检查，IDE 自动补全
type Option func(*SlowLogAnalyzer)

// WithRAGRetriever 设置 RAG 检索器选项
// 这是函数式选项模式的实现：返回一个 Option 函数，该函数会修改 SlowLogAnalyzer 的配置
func WithRAGRetriever(r Retriever) Option {
	return func(a *SlowLogAnalyzer) {
		a.retriever = r
	}
}

// WithPromptBuilder 设置 Prompt 构建器选项
func WithPromptBuilder(pb PromptBuilder) Option {
	return func(a *SlowLogAnalyzer) {
		a.promptBuilder = pb
	}
}

// WithPromptVersion 设置 Prompt 版本选项（可选，主要用于日志记录）
func WithPromptVersion(v PromptVersion) Option {
	return func(a *SlowLogAnalyzer) {
		a.promptVersion = v
	}
}
