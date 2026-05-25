# 架构、MCP 与扩展

[路线图](../agent/ROADMAP.md) · [文档索引](../INDEX.md) · [版本详解](VERSIONS.md)

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

### analysis-flow

**分析流程**（V1–V3 `SlowLogAnalyzer` 一次调用路径）：

1. **输入慢日志** → `SlowLogAnalyzer.Analyze()`
2. **RAG 检索**（如果启用）→ 从知识库检索相关模式
3. **构建 Prompt** → 根据版本选择 v2 或 v3 Prompt
4. **LLM 分析** → 调用 LLM 生成分析结果
5. **返回结果** → 结构化的分析结果

各版本 Prompt 与 Agent 行为详见 [VERSIONS.md](VERSIONS.md)。

**V6 `AgentState`**（`internal/analyzer/agent_state.go`）：用阶段 `init → rag_done → db_ready → explained → index_planned → analyzed → finished` 替代 `map[string]interface{}`；每轮 Prompt 只注入 `PromptSummary()`（工具要点 + RAG 标题），完整工具 JSON 仍写入报告 `tool_results`，不重复灌进 LLM。

**工具错误码**（`internal/toolerr`）：MCP 失败经 `toolerr.From` 映射为 `code`（如 `mysql_table_not_found`）与 `retryable`；Agent 摘要中标注「可重试 / 勿重试」，报告 `action_outcome` 含 `{ok:false, code, message, retryable}`。

**结构化 Trace**（`internal/analyzer/trace.go`）：`make agent-run`（开启 round 记录）时写入 `trace.total_duration_ms` 与每轮 `rounds[].trace[]`（`llm.chat`、`llm.parse`、`tool.*` 等 span 及 `duration_ms`）；`*.brief.html` 逐轮表含耗时列。

### v1-v3-example

**V1–V3 集成示例**（V3 + RAG）：

```go
llmClient, _ := llm.NewDeepSeekClient(os.Getenv("DEEPSEEK_API_KEY"), "")
a := analyzer.NewAnalyzer(
    llmClient,
    analyzer.WithPromptBuilder(&prompt.RagV3Prompt{}),
    analyzer.WithRAGRetriever(analyzer.NewRAGRetrieverAdapter(rag.NewMockRetriever())),
    analyzer.WithPromptVersion(analyzer.PromptV3),
)
result, err := a.Analyze(ctx, slowLogText)
```

## mcp

**MCP 能力与能力感知**

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

### 本地 MySQL 实例连接（`connect_mysql_instance`）

凭证通过 **`.env` / 环境变量** 配置（见 `.env.example`），不写入代码：

```bash
MYSQL_HOST=127.0.0.1
MYSQL_PORT=3306
MYSQL_USER=root
MYSQL_PASSWORD=your_password
MYSQL_DATABASE=          # 可选，默认库
```

架构分层：

- `internal/config`：加载 `.env` 与 `MYSQL_*` 环境变量
- `internal/mysql`：连接池与 Ping / `SHOW DATABASES`（后续 EXPLAIN 复用 `Client`）
- `internal/mcp`：`ConnectMySQLCapability` → `connect_mysql_instance`
- `ExplainMySQLQueryCapability` → `explain_mysql_query`（仅单条 SELECT）
- `AddMySQLIndexCapability` → `add_mysql_index`（默认 `dry_run=true`，不直接改表）

**示例（test.products）**：

```json
// explain_mysql_query
{ "database": "test", "sql": "SELECT * FROM products WHERE sku = 'A001'" }

// add_mysql_index（先预览 DDL）
{ "database": "test", "table": "products", "index_name": "idx_sku", "columns": "sku", "dry_run": true }

// 确认后再执行
{ "database": "test", "table": "products", "index_name": "idx_sku", "columns": "sku", "dry_run": false }
```

仅校验连接（不跑 V6 / LLM）：

```bash
make mysql-check
# 或：go run ./cmd/mysql-check
```

若 `go mod tidy` 因公司 GOPROXY 超时，请用：

```bash
make deps    # 默认 GOPROXY=https://goproxy.cn,direct
make vendor  # 同步 vendor/ 后可直接 go build
```

启动 `slowlog-ai` 时若 MySQL 配置有效，会自动注册该能力，V6 Agent 可通过 `call_tool` 调用。

> 版本与能力对照见 [版本演进速查 · MCP 能力](../../README.md#mcp-能力)。

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

RAG 知识库位于 `internal/rag/slowlog/docs/`（按 **`##` 切 chunk**；**默认 `TFIDFRetriever`**，`SLOWLOG_RAG=embedding` 为内存向量 TopK，`mock` 用于 eval）。**用法**：[guides/RAG.md](../guides/RAG.md) · 流程图：[diagrams/rag-flow.md](../diagrams/rag-flow.md)。包含：

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
