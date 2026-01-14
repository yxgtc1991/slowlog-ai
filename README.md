# slowlog-ai

🚀 **基于 LLM 的 MySQL 慢日志智能分析工具（Golang）**

`slowlog-ai` 是一个使用 Go 编写的项目，旨在探索 **将大语言模型（LLM）引入云数据库 MySQL 慢日志分析场景**，自动完成 SQL 慢查询的原因分析与优化建议生成。

---

## ✨ 项目背景

在云数据库 MySQL 场景中，慢日志通常存在以下痛点：

- 慢 SQL 数量大，人工分析成本高
- SQL 优化依赖经验，质量不稳定
- 现有工具多聚焦统计，缺少"原因 + 建议"的语义分析

本项目尝试引入 **LLM（如 DeepSeek）**，通过 Prompt Engineering 和 RAG（检索增强生成），对 MySQL 慢日志进行**自动化智能分析**，作为 DBA / 平台侧的辅助决策工具。

---

## 🧠 核心能力

### 当前版本（v4）

- ✅ **模块化架构**：可插拔的 LLM 客户端、Prompt 构建器和 RAG 检索器
- ✅ **多版本 Prompt 支持**：
  - **v1 基础版本**：简单的提示和慢日志内容（可用于对比学习）
  - **v2 严格模式**：基于慢日志内容进行严格分析，输出结构化 JSON
  - **v3 RAG 增强**：结合专家知识库，提供更准确的推断和建议
- ✅ **RAG 检索**：支持从知识库中检索相关模式、反模式和指标说明
- ✅ **MCP 能力基础**：已实现 MCP（Model Context Protocol）能力接口，可扩展为 MCP 服务器

### Prompt 演进历程

- **v1 基础版本**：简单的提示和慢日志内容（已恢复，可用于对比学习）
- **v2 严格模式**：引入严格规则约束，要求 LLM 只基于日志内容分析，输出结构化 JSON
- **v3 RAG 增强**：在 v2 基础上增加 RAG 检索的知识块，用于辅助推断 suspected_issues

---

## 📁 项目结构

```
slowlog-ai/
├── cmd/
│   └── slowlog-ai/
│       └── main.go              # 程序入口
├── internal/
│   ├── analyzer/               # 分析器核心模块
│   │   ├── interfaces.go       # LLMClient, Retriever, PromptBuilder 接口
│   │   ├── options.go          # Option 函数和 PromptVersion
│   │   ├── slowlog.go          # SlowLogAnalyzer 实现
│   │   └── rag_adapter.go      # RAG 检索器适配器
│   ├── llm/
│   │   └── deepseek.go         # DeepSeek LLM 客户端封装
│   ├── prompt/
│   │   └── slowlog/
│   │       ├── v1_basic.go     # v1 基础版本 Prompt（用于对比学习）
│   │       ├── v2_strict.go    # v2 严格模式 Prompt
│   │       └── v3_rag.go       # v3 RAG 增强 Prompt
│   ├── rag/
│   │   ├── retriever.go        # RAG 检索器接口
│   │   ├── mock.go             # Mock 检索器实现
│   │   └── slowlog/
│   │       └── docs/           # 知识库文档（模式、反模式、指标等）
│   └── mcp/                    # MCP 能力模块
│       ├── capability.go       # MCP Capability 接口
│       └── slowlog.go          # AnalyzeSlowLogCapability 实现
├── go.mod
├── go.sum
└── README.md
```

---

## 🚀 快速开始

### 环境要求

- Go 1.23+
- DeepSeek API Key

### 安装

```bash
git clone <repository-url>
cd slowlog-ai
go mod download
```

### 配置

设置环境变量：

```bash
export DEEPSEEK_API_KEY="your-api-key-here"
```

### 使用示例

#### 1. 命令行使用

```bash
# 使用内置示例
go run cmd/slowlog-ai/main.go

# 从文件读取慢日志
go run cmd/slowlog-ai/main.go /path/to/slowlog.txt
```

#### 2. 代码示例

```go
package main

import (
    "context"
    "os"
    
    "ai_slow_log/internal/analyzer"
    "ai_slow_log/internal/llm"
    prompt "ai_slow_log/internal/prompt/slowlog"
    "ai_slow_log/internal/rag"
)

func main() {
    ctx := context.Background()
    
    // 创建 LLM 客户端
    llmClient, err := llm.NewDeepSeekClient(
        os.Getenv("DEEPSEEK_API_KEY"), 
        "", // 使用默认模型
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // 创建分析器（使用 v3 RAG 增强模式）
    analyzer := analyzer.NewAnalyzer(
        llmClient,
        &prompt.StrictV2Prompt{},  // v2 prompt
        &prompt.RagV3Prompt{},      // v3 prompt
        analyzer.WithRAGRetriever(
            analyzer.NewRAGRetrieverAdapter(
                rag.NewMockRetriever(),
            ),
        ),
        analyzer.WithPromptVersion(analyzer.PromptV3),
    )
    
    // 分析慢日志
    slowLog := `# Time: 2025-01-05T10:12:33.123456Z
# Query_time: 12.456  Lock_time: 0.001
# Rows_sent: 10  Rows_examined: 1250000
SELECT * FROM orders WHERE user_id = 123 ORDER BY created_at DESC LIMIT 10;`
    
    result, err := analyzer.Analyze(ctx, slowLog)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(result.RawOutput)
}
```

---

## 🏗️ 架构设计

### 核心接口

#### LLMClient
```go
type LLMClient interface {
    Chat(ctx context.Context, prompt string) (string, error)
}
```

#### Retriever
```go
type Retriever interface {
    Retrieve(ctx context.Context, query string) ([]rag.KnowledgeChunk, error)
}
```

#### PromptBuilder
```go
type PromptBuilder interface {
    Build(slowLog string, chunks []rag.KnowledgeChunk) string
}
```

### 函数式选项模式（Functional Options Pattern）

项目使用 **函数式选项模式** 来构建 `SlowLogAnalyzer`，这是 Go 语言中处理可选配置的优雅方式。

#### 设计优势

- ✅ **可扩展性**：新增配置项只需添加新的 `WithXxx` 函数，无需修改构造函数签名
- ✅ **灵活性**：可选参数，顺序无关，只设置需要的配置
- ✅ **可读性**：函数名清晰表达意图，自文档化
- ✅ **类型安全**：编译期检查，IDE 自动补全

#### 实现原理

```go
// 定义 Option 类型：一个函数，接收配置对象并修改它
type Option func(*SlowLogAnalyzer)

// 创建选项函数：返回一个 Option
func WithPromptBuilder(pb PromptBuilder) Option {
    return func(a *SlowLogAnalyzer) {
        a.promptBuilder = pb
    }
}

// 构造函数：接收可变参数 opts ...Option
func NewAnalyzer(llm LLMClient, opts ...Option) *SlowLogAnalyzer {
    a := &SlowLogAnalyzer{llm: llm}
    
    // 依次应用所有选项
    for _, opt := range opts {
        opt(a)  // 调用选项函数，修改配置对象
    }
    
    return a
}
```

#### 使用示例

```go
// 基础用法：只设置 Prompt
analyzer := analyzer.NewAnalyzer(
    llmClient,
    analyzer.WithPromptBuilder(&prompt.StrictV2Prompt{}),
)

// 完整配置：设置多个选项
analyzer := analyzer.NewAnalyzer(
    llmClient,
    analyzer.WithPromptBuilder(&prompt.RagV3Prompt{}),
    analyzer.WithRAGRetriever(analyzer.NewRAGRetrieverAdapter(rag.NewMockRetriever())),
)

// 选项顺序无关
analyzer := analyzer.NewAnalyzer(
    llmClient,
    analyzer.WithRAGRetriever(...),
    analyzer.WithPromptBuilder(&prompt.RagV3Prompt{}),
)
```

这个模式在 Go 标准库中广泛使用，如 `context.WithTimeout`、`grpc.Dial` 等。

### 分析流程

1. **输入慢日志** → `SlowLogAnalyzer.Analyze()`
2. **RAG 检索**（如果启用）→ 从知识库检索相关模式
3. **构建 Prompt** → 根据版本选择 v2 或 v3 Prompt
4. **LLM 分析** → 调用 LLM 生成分析结果
5. **返回结果** → 结构化的分析结果

### Prompt 版本说明

#### v2 严格模式
- 只基于慢日志内容进行分析
- 不允许假设表结构、索引或业务逻辑
- 输出结构化 JSON，包含：
  - `summary`: 问题总结
  - `metrics`: 指标提取
  - `confirmed_issues`: 可确认的问题
  - `suspected_issues`: 可能存在的问题（带置信度）
  - `required_information`: 需要补充的信息
  - `next_actions`: 下一步建议

#### v3 RAG 增强
- 在 v2 基础上增加专家知识库
- 知识库包含：
  - **模式（Patterns）**：常见性能问题模式
  - **反模式（Anti-patterns）**：常见误解和错误结论
  - **指标（Metrics）**：慢日志指标说明
  - **边界（Boundaries）**：分析边界和限制
- 知识仅用于推断 `suspected_issues`，不能作为 `confirmed_issues`

---

## 🔌 MCP 能力

项目已实现 MCP（Model Context Protocol）能力的基础结构，可以将慢日志分析能力暴露为 MCP 工具。

### MCP 接口

```go
type Capability interface {
    Name() string
    Description() string
    InputSchema() map[string]string
    Execute(ctx context.Context, input map[string]interface{}) (interface{}, error)
}
```

### 使用示例

```go
// 创建慢日志分析能力
capability := &mcp.AnalyzeSlowLogCapability{
    Analyzer: analyzer,
}

// 执行能力
result, err := capability.Execute(ctx, map[string]interface{}{
    "slow_log": slowLogText,
})
```

---

## 📚 知识库

RAG 知识库位于 `internal/rag/slowlog/docs/`，包含：

- **patterns/**: 性能问题模式（如 `rows_examined_high.md`）
- **anti-patterns/**: 常见误解（如 `limit_not_fast.md`）
- **metrics/**: 指标说明（如 `rows_examined.md`）
- **boundaries/**: 分析边界（如 `no_schema.md`）
- **actions/**: 优化建议（如 `add_index.md`）

---

## 🔧 扩展开发

### 添加新的 LLM 提供商

实现 `analyzer.LLMClient` 接口：

```go
type CustomLLMClient struct {
    // ...
}

func (c *CustomLLMClient) Chat(ctx context.Context, prompt string) (string, error) {
    // 实现 LLM 调用逻辑
}
```

### 添加新的 Prompt 版本

实现 `analyzer.PromptBuilder` 接口：

```go
type CustomPrompt struct{}

func (p *CustomPrompt) Build(slowLog string, chunks []rag.KnowledgeChunk) string {
    // 构建自定义 Prompt
}
```

### 实现真实的 RAG 检索器

替换 `rag.MockRetriever`，实现基于向量数据库的检索：

```go
type VectorRetriever struct {
    // 向量数据库客户端
}

func (r *VectorRetriever) Retrieve(ctx context.Context, query string) ([]rag.KnowledgeChunk, error) {
    // 向量检索逻辑
}
```

---

## 📝 输出格式

分析结果以 JSON 格式输出：

```json
{
  "summary": "扫描行数远大于返回行数，可能存在索引未命中问题",
  "metrics": {
    "query_time": "12.456",
    "rows_examined": "1250000",
    "rows_sent": "10",
    "lock_time": "0.001"
  },
  "confirmed_issues": [
    "Rows_examined (1250000) 远大于 Rows_sent (10)，扫描效率低"
  ],
  "suspected_issues": [
    {
      "issue": "可能缺少 user_id 或 created_at 的索引",
      "reason": "基于 Rows_examined >> Rows_sent 模式推断",
      "confidence": "medium"
    }
  ],
  "required_information": [
    "表结构信息",
    "索引信息",
    "EXPLAIN 执行计划"
  ],
  "next_actions": [
    {
      "action": "检查 user_id 和 created_at 字段的索引情况",
      "depends_on": "需要提供表结构和索引信息"
    }
  ]
}
```

---

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

---

## 📄 许可证

[待定]
