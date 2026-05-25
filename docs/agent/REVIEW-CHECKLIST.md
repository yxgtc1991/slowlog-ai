# 封存后复习清单（基建 / AI 应用向）

项目实现已告一段落；本文档用于 **复习与口述**，不再驱动新功能开发。

← [INDEX](../INDEX.md) · [AI-APPLICATION-BRIEF](AI-APPLICATION-BRIEF.md)

---

## 1. 项目 30 秒版（背这个就够开场）

Go 做的 **MySQL 慢日志多轮 Agent**：RAG 查知识、MCP 做 EXPLAIN/索引建议、V6 自主决策、报告可复盘；**agent-eval + rag-test + CI** 保证可回归。

---

## 2. 演示命令（交流前一天跑一遍）

```bash
make check
# 或拆开：make agent-eval && make rag-test && make test
make rag-check "price 最左前缀"
# 有 Key：make agent-run → 打开 reports/*brief.html
```

---

## 3. 必会深挖（按重要性）

1. V6 循环：`retrieve_rag` / `call_tool` / `analyze` / `finish`  
2. AgentState 为何摘要进 Prompt（控 Token）  
3. 双层 golden：Agent vs RAG 分工  
4. V5 vs V6 协议差异（`make run-v5`）  
5. EXPLAIN 与慢日志 SQL 对齐、索引 dry_run  

代码：`v6_agent.go`、`agent_state.go`、`internal/eval/`、`internal/rag/`。

---

## 4. 结合 9 年后端 + Fluent Bit（岗位叙事）

| 你的主业 | 本项目 |
|----------|--------|
| K8s 日志采集（Fluent Bit） | 日志 **到了之后** 怎么智能诊断 |
| 稳定性 / 排障 | 慢 SQL **证据链**（日志→EXPLAIN→建议） |
| 0→1 基建 | slowlog **0→1 Agent 链路** |

连接句：**采集层我熟悉；智能层我用 Agent 验证过工程化闭环。**

---

## 5. 技术调研问答（可选深读）

- [RESEARCH-QA.md](RESEARCH-QA.md)：Agent / RAG / MySQL / MCP 等调研题，带 **slowlog-ai 对照** 与可深化方向。

---

## 6. 复习周计划（建议 5～7 天）

| 天 | 内容 |
|----|------|
| D1 | 本项目：口述 2 遍 + brief 报告走读 |
| D2 | RAG/Agent 概念：检索、Tool Calling、ReAct、评测 |
| D3 | Go + MySQL：索引、EXPLAIN、慢日志字段 |
| D4 | 系统设计：「设计一个 SQL 诊断 Agent」（用本仓库当答案骨架） |
| D5 | K8s/可观测：Fluent Bit 链路复习 + 1 个排障故事 |
| D6 | 基础算法/并发：中等题 2～3 道（巩固基本功） |
| D7 | 模拟 30 分钟问答（录音自检） |

---

## 7. 刻意不做（封存边界）

- 向量库 / HITL / Plan-and-Execute（P2）  
- 扩游戏/推荐业务 Demo  
- 无休止加知识库（现有 16 篇 + golden 已够讲）

---

## 8. 文档索引

| 文档 | 用途 |
|------|------|
| [AI-APPLICATION-BRIEF](AI-APPLICATION-BRIEF.md) | 2 分钟稿 + 深挖答法 |
| [ROADMAP](ROADMAP.md) | 路线与变更 |
| [EVAL](EVAL.md) | Agent golden |
| [guides/RAG](../guides/RAG.md) | RAG 环境变量 |
| [RESEARCH-QA](RESEARCH-QA.md) | 调研问答与题单对照 |
