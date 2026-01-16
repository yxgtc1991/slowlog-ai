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
  - **v1 基础版本**：问模型 - 简单直接，无约束
  - **v2 严格模式**：约束模型 - 结构化 JSON 输出，避免过度假设
  - **v3 RAG 增强**：引入知识库 - 结合专家知识，提升分析准确性
  - **v4 能力感知**：能力抽象与感知 - 系统能力可发现、可描述
  - **v5 能力调用意图**：LLM 自主决策 - LLM 输出意图，系统自动执行
  - **v6 Agent**：真正的 AI Agent - LLM 自主决定下一步行动（调用工具、检索 RAG、分析、提问、完成）
- ✅ **RAG 检索**：支持从知识库中检索相关模式、反模式和指标说明
- ✅ **MCP 能力基础**：已实现 MCP（Model Context Protocol）能力接口，可扩展为 MCP 服务器

### Prompt 演进历程

本项目展示了 Prompt Engineering 的完整演进过程，从最基础的提示到能力感知系统。每个版本都解决了前一个版本的问题，逐步提升系统的可靠性和智能化水平。

#### 📌 V1 基础版本：问模型（Ask the Model）

**设计思路**：最直接的方式，将问题抛给 LLM，依赖模型的通用知识。

**核心特点**：
- 简单的角色设定："你是一个 MySQL 专家"
- 直接提问："请分析以下慢日志并给出优化建议"
- 无任何约束，完全依赖模型的自由发挥

**代码示例**：
```go
func BuildSlowLogPrompt(slowLog string) string {
    return fmt.Sprintf(`
你是一个 MySQL 专家，请分析以下慢日志并给出优化建议：

%s
`, slowLog)
}
```

**存在的问题**：
- ❌ 输出格式不统一，难以程序化处理
- ❌ 模型可能过度假设，给出无法验证的建议
- ❌ 缺少结构化输出，无法区分"确定的问题"和"推测的问题"
- ❌ 无法控制模型的推理边界

**适用场景**：快速原型、概念验证、人工审查场景

---

#### 🔒 V2 严格模式：约束模型（Constrain the Model）

**设计思路**：通过严格的规则约束，让模型只基于提供的信息进行分析，避免过度假设。

**核心改进**：
- ✅ **角色明确化**：从"专家"变为"分析组件"，强调工具属性
- ✅ **规则约束**：5 条严格规则，禁止假设、禁止无法推导的结论
- ✅ **结构化输出**：强制 JSON 格式，包含 `confirmed_issues` 和 `suspected_issues`
- ✅ **边界明确**：明确区分"可确认"和"需要更多信息"

**代码示例**：
```go
var StrictPromptTemplate = `
你是一个【MySQL 慢查询分析组件】，不是聊天助手。

你必须严格遵守以下规则：
1. 只能基于提供的慢日志内容进行分析
2. 不允许假设表结构、索引或业务逻辑
3. 不确定的信息必须明确标注为"需要更多信息"
4. 禁止给出无法从日志中直接推导的结论
5. 输出必须是 JSON，不允许任何额外说明文字
...
`
```

**解决的问题**：
- ✅ 输出格式统一，可程序化处理
- ✅ 避免模型过度假设
- ✅ 明确区分确定性和推测性结论
- ✅ 可追溯的分析依据

**技术亮点**：
- **Prompt 约束技术**：通过明确的规则限制模型行为
- **结构化输出**：JSON Schema 约束输出格式
- **置信度管理**：通过 `confidence` 字段管理不确定性

**仍存在的问题**：
- ⚠️ 模型缺乏专业知识，可能遗漏常见问题模式
- ⚠️ 无法利用历史经验和最佳实践
- ⚠️ 分析深度受限于模型的通用知识

---

#### 🧠 V3 RAG 增强：引入知识库（Retrieval-Augmented Generation）

**设计思路**：通过 RAG（检索增强生成）技术，将专家知识库注入到 Prompt 中，提升分析的准确性和深度。

**核心改进**：
- ✅ **知识检索**：从知识库中检索相关的模式、反模式、指标说明
- ✅ **知识注入**：将检索到的知识块注入到 Prompt 中
- ✅ **知识约束**：明确知识仅用于推断 `suspected_issues`，不能作为 `confirmed_issues`
- ✅ **降级机制**：如果没有检索到知识，自动降级到 v2

**代码示例**：
```go
func BuildSlowLogPromptV3(slowLog string, ragChunks []rag.KnowledgeChunk) string {
    // 构建知识块描述
    var sb strings.Builder
    for _, c := range ragChunks {
        sb.WriteString(fmt.Sprintf("- %s：%s\n", c.Title, c.Content))
    }
    
    return fmt.Sprintf(`
%s  // v2 的严格模板

【可参考的专家知识（仅用于推断 suspected_issues，不得作为 confirmed_issues）】
%s
`, BuildSlowLogPromptV2(slowLog), sb.String())
}
```

**解决的问题**：
- ✅ 利用专家知识，提升分析准确性
- ✅ 识别常见性能问题模式
- ✅ 避免常见误解和反模式
- ✅ 提供更专业的优化建议

**技术亮点**：
- **RAG 架构**：检索 + 生成，结合外部知识
- **知识分类**：模式（Patterns）、反模式（Anti-patterns）、指标（Metrics）
- **知识边界**：明确知识的使用范围和限制
- **优雅降级**：无知识时回退到 v2

**知识库结构**：
```
internal/rag/slowlog/docs/
├── patterns/          # 性能问题模式
├── anti-patterns/     # 常见误解
├── metrics/           # 指标说明
├── boundaries/        # 分析边界
└── actions/           # 优化建议
```

**仍存在的问题**：
- ⚠️ 知识库是静态的，需要人工维护
- ⚠️ 检索质量依赖查询策略
- ⚠️ 系统能力是固定的，无法动态扩展

---

#### 🎯 V4 能力感知：能力抽象与感知（Capability Awareness）

**设计思路**：将系统能力抽象为可发现、可描述的接口，让 LLM 能够感知和选择使用哪些能力。

**核心改进**：
- ✅ **能力抽象**：定义 `Capability` 接口，统一能力描述
- ✅ **能力发现**：系统可以自动列出所有可用能力
- ✅ **能力描述**：生成 LLM 可理解的能力说明
- ✅ **动态扩展**：新能力可以动态注册，无需修改核心代码

**代码示例**：
```go
// 能力接口定义
type Capability interface {
    Name() string                    // 能力名称
    Description() string             // 能力说明
    InputSchema() map[string]string  // 输入参数
    Execute(ctx, input) (interface{}, error)
}

// v4 Prompt：生成能力描述
func BuildCapabilityPromptV4(caps []CapabilityV4) string {
    // 为每个能力生成描述
    // 告诉 LLM 可以使用哪些工具
}
```

**解决的问题**：
- ✅ 系统能力可发现、可描述
- ✅ LLM 可以自主选择使用哪些能力
- ✅ 新能力可以动态添加
- ✅ 为 MCP（Model Context Protocol）奠定基础

**技术亮点**：
- **接口抽象**：通过接口统一能力定义
- **能力注册表**：Registry 模式管理能力
- **自描述系统**：系统可以描述自己的能力
- **MCP 兼容**：符合 Model Context Protocol 设计理念

**架构设计**：
```
Capability Interface
    ↓
Registry (能力注册表)
    ↓
Server (能力服务器)
    ↓
Prompt V4 (能力描述生成)
```

**演进价值**：
- 🎯 **可扩展性**：新能力只需实现接口即可
- 🎯 **可发现性**：系统自动发现和描述能力
- 🎯 **可组合性**：多个能力可以组合使用
- 🎯 **标准化**：为 MCP 协议做准备

---

### 📊 版本对比总结

| 版本 | 核心思想 | 输出格式 | 知识来源 | 扩展性 | 适用场景 |
|------|---------|---------|---------|--------|---------|
| **V1** | 问模型 | 自由文本 | 模型通用知识 | ❌ | 快速原型 |
| **V2** | 约束模型 | 结构化 JSON | 模型通用知识 | ❌ | 生产环境（基础） |
| **V3** | RAG 增强 | 结构化 JSON | 模型 + 知识库 | ⚠️ | 生产环境（推荐） |
| **V4** | 能力感知 | 能力描述 | 系统能力抽象 | ✅ | MCP 集成 |
| **V5** | 能力调用意图 | 分析 + 意图 | LLM 自主决策 | ✅ | 自动化执行 |

### 🎓 学习要点（面试准备）

1. **V1 → V2**：理解 Prompt 约束的重要性，如何通过规则限制模型行为
2. **V2 → V3**：掌握 RAG 技术，如何将外部知识注入到 LLM 推理过程
3. **V3 → V4**：理解能力抽象和系统自描述，如何设计可扩展的能力系统
4. **整体演进**：从"问模型"到"约束模型"到"增强模型"到"感知系统"的完整思路

### 💡 设计模式应用

- **V2**：使用了 **Prompt Template Pattern**（模板模式）
- **V3**：使用了 **RAG Pattern**（检索增强生成模式）
- **V4**：使用了 **Capability Pattern**（能力模式）和 **Registry Pattern**（注册表模式）

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
│   │   ├── rag_adapter.go      # RAG 检索器适配器
│   │   ├── v5_tool_calling.go  # V5 Tool Calling 分析器
│   │   └── v6_agent.go         # V6 Agent 分析器
│   ├── llm/
│   │   └── deepseek.go         # DeepSeek LLM 客户端封装
│   ├── prompt/
│   │   └── slowlog/
│   │       ├── v1_basic.go      # v1 基础版本：问模型
│   │       ├── v2_strict.go      # v2 严格模式：约束模型
│   │       ├── v3_rag.go         # v3 RAG 增强：引入知识库
│   │       ├── v4_capability.go  # v4 能力感知：能力抽象与感知
│   │       ├── v5_intent.go      # v5 能力调用意图：LLM 自主决策
│   │       └── v6_action.go       # v6 Agent：LLM 自主决定下一步行动
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

### Prompt 版本快速参考

详细的演进过程请参考上方的 [Prompt 演进历程](#prompt-演进历程) 章节。

#### V1 基础版本：问模型
- **文件**：`internal/prompt/slowlog/v1_basic.go`
- **特点**：简单直接，无约束
- **适用**：快速原型、概念验证

#### V2 严格模式：约束模型
- **文件**：`internal/prompt/slowlog/v2_strict.go`
- **特点**：5 条严格规则，结构化 JSON 输出
- **输出结构**：`summary`、`metrics`、`confirmed_issues`、`suspected_issues`、`required_information`、`next_actions`
- **适用**：生产环境（基础版本）

#### V3 RAG 增强：引入知识库
- **文件**：`internal/prompt/slowlog/v3_rag.go`
- **特点**：结合专家知识库，提升分析准确性
- **知识库**：patterns、anti-patterns、metrics、boundaries、actions
- **适用**：生产环境（推荐版本）

#### V4 能力感知：能力抽象与感知
- **文件**：`internal/prompt/slowlog/v4_capability.go`
- **特点**：系统能力抽象，自动发现和描述能力
- **适用**：MCP 集成、可扩展系统

#### V5 能力调用意图：LLM 自主决策
- **文件**：`internal/prompt/slowlog/v5_intent.go`
- **特点**：LLM 输出能力调用意图，系统自动执行
- **适用**：自动化执行、AI Agent

#### 🤖 V6 Agent：LLM 自主决策下一步行动
- **文件**：`internal/prompt/slowlog/v6_action.go`、`internal/analyzer/v6_agent.go`
- **特点**：LLM 自主决定下一步要做什么（调用工具、检索 RAG、继续分析、提问、完成）
- **适用**：真正的 AI Agent、自主规划任务

**设计思路**：从"LLM 决定是否调用工具"升级到"LLM 决定下一步做什么"，实现真正的 Agent 模式。

**核心改进**：
- ✅ **多行动类型**：支持 5 种行动类型（call_tool、retrieve_rag、analyze、ask_question、finish）
- ✅ **自主规划**：LLM 可以自主规划分析步骤，决定何时调用工具、何时检索知识库、何时完成
- ✅ **上下文感知**：每次决策都基于当前上下文（已执行的工具结果、RAG 检索结果等）
- ✅ **对话历史**：维护完整的对话历史，支持多轮交互

**行动类型说明**：
1. **call_tool** - 调用工具：当需要调用系统工具时使用
2. **retrieve_rag** - 检索知识库：当需要查询相关知识时使用
3. **analyze** - 继续分析：基于已有信息继续分析
4. **ask_question** - 提出问题：需要用户提供额外信息时使用
5. **finish** - 完成分析：收集足够信息，输出最终结果

**代码示例**：
```go
// 创建 V6 Agent 分析器
v6Analyzer := analyzer.NewV6AgentAnalyzer(
    llmClient,              // 使用普通 LLM 客户端（不需要 Tool Calling）
    ragRetriever,           // RAG 检索器
    capabilityExecutor,     // 能力执行器
    availableTools,          // 可用工具列表
)

// 执行分析（LLM 自主决定每一步要做什么）
result, err := v6Analyzer.Analyze(ctx, slowLog)

// 查看执行轨迹
fmt.Printf("迭代次数：%d\n", result.Iterations)
for i, action := range result.Actions {
    fmt.Printf("%d. [%s] %s\n", i+1, action.Type, action.Reasoning)
}
```

**与 V5 的区别**：
- **V5**：LLM 决定"是否调用工具"，系统负责执行工具调用
- **V6**：LLM 决定"下一步做什么"，可以是调用工具、检索 RAG、继续分析、提问或完成

---

## 🔌 MCP 能力与能力感知

项目已实现 MCP（Model Context Protocol）能力的基础结构，支持**能力感知**（Capability Awareness），可以让系统自动发现、描述和执行能力。

### 能力感知特性

- ✅ **能力发现**：自动列出所有已注册的能力
- ✅ **能力描述**：生成 LLM 可理解的能力描述
- ✅ **能力查询**：检查是否支持某个能力
- ✅ **统一执行**：通过统一接口执行能力

### MCP 接口

```go
type Capability interface {
    Name() string                    // 能力唯一标识
    Description() string             // 能力说明（告诉 LLM 什么时候用）
    InputSchema() map[string]string  // 输入参数结构（JSON Schema 简化版）
    Execute(ctx context.Context, input map[string]interface{}) (interface{}, error)
}
```

### 能力感知使用示例

```go
// 1. 创建 MCP 服务器
mcpServer := mcp.NewServer()

// 2. 注册能力
capability := &mcp.AnalyzeSlowLogCapability{
    Analyzer: analyzer,
}
mcpServer.RegisterCapability(capability)

// 3. 列出所有可用能力（JSON 格式）
capabilitiesJSON, err := mcpServer.ListCapabilities()
fmt.Println(capabilitiesJSON)

// 4. 获取能力描述 prompt（给 LLM 看）
capabilityPrompt := mcpServer.GetCapabilityPrompt()
fmt.Println(capabilityPrompt)

// 5. 检查能力是否存在
if mcpServer.HasCapability("analyze_slow_log") {
    fmt.Println("支持慢日志分析能力")
}

// 6. 执行能力
result, err := mcpServer.ExecuteCapability(ctx, "analyze_slow_log", map[string]interface{}{
    "slow_log": slowLogText,
})
```

### 能力描述格式

能力描述会自动生成，包含：
- **能力名称**：唯一标识符
- **能力说明**：告诉 LLM 什么时候使用这个能力
- **输入参数**：参数名称、类型和说明

示例输出：
```
你可以使用以下系统能力（Tools）：

【能力 1】analyze_slow_log
  说明：分析 MySQL 慢日志，输出结构化的性能问题与优化建议
  输入参数：
    - slow_log: string // 原始 MySQL 慢日志文本
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
