# Agent Eval（P0）

[路线图](ROADMAP.md) · [完整跑通](RUN.md) · [文档索引](../INDEX.md)

用 **golden case + 轨迹/结论断言** 证明 V6 Agent 可回归，改 Prompt / 解析 / 执行逻辑后不必每次手跑 `make agent-run`。

## 做了什么

| 组件 | 作用 |
|------|------|
| `ScriptLLM` | 按脚本顺序返回「假 LLM」JSON，**不消耗 API** |
| `StubExecutor` | 工具返回固定 JSON，**不连 MySQL** |
| `internal/eval/cases.go` | 4 条内置 golden：guided 全流程、RAG 后 finish、type 归一化、工具失败仍 finish |
| `AssertReportFile` | 对 `reports/*.json` 做结构/结论断言（真实跑完后的存档） |

## 为什么做

- **之前**：只有 `v6_action_test` 测解析片段；完整 Agent 只能靠 `make agent-run` 肉眼对比。
- **现在**：CI / 本地 `make agent-eval` 秒级验证「轨迹对不对、结论有没有关键信息」；检索另见 `make rag-test`。
- **解决的问题**：改 `normalizeNextAction`、循环逻辑、报告字段时，**有自动化护栏**，避免回归。

## 如何运行

```bash
# 跑全部 golden（推荐，无需 DEEPSEEK_API_KEY）
make agent-eval

# 只看一条
go run ./cmd/agent-eval -case=guided_flow -v

# 单元测试（同上 cases）
go test ./internal/eval/... -v

# 对真实 agent-run 报告做断言（需先 make agent-run 生成 JSON）
go run ./cmd/agent-eval -report=reports/agent-run-YYYYMMDD-HHMMSS.json
```

## 如何对比效果

| 方式 | 何时用 | 对比什么 |
|------|--------|----------|
| **`make agent-eval`** | 改代码后每次提交前 | 确定性：轨迹 6 步、工具名、结论含「全表扫描/price/索引」 |
| **`make agent-run`** | 改 Prompt / 模型行为后 | 非确定性：看 `*.brief.html` 逐轮是否合理 |
| **`-report=...json`** | 有一份「满意」的存档后 | 把该 JSON 当基线：改代码后同命令仍应 PASS |

**对比示例**

1. 改 `v6_action.go` 前：`make agent-eval` → 3 passed  
2. 改坏 `validateNextAction` → `make agent-eval` → FAIL，立刻看到哪一步 type 不对  
3. 修好后再 `make agent-run`，用 `-report` 确认真实 LLM 报告仍含 `explain_mysql_query` 与 `price`

## 内置 case 说明

| Case | 验证点 |
|------|--------|
| `guided_flow` | RAG → 三工具 → analyze → finish；与 guided 推荐流程一致 |
| `tool_name_as_type` | 模型把 `analyze_slow_log` 写在 `type` 时仍能 `call_tool` 执行 |
| `tool_error_then_finish` | `connect_mysql` 失败仍有 `action_error` 且能 finish |

## 扩展

在 `internal/eval/cases.go` 的 `AllCases()` 增加 `Script` + `Expect` 即可；复杂场景可把 `reports/xxx.json` 复制到 `testdata/eval/` 用 `AssertReportFile` 固定断言。
