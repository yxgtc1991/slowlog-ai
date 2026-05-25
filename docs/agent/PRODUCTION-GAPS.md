# 生产化缺口与优化路线

[文档索引](../INDEX.md) · [调研问答](RESEARCH-QA.md) · [测试指南](TESTING.md) · [路线图](ROADMAP.md)

> **定位**：slowlog-ai 当前是 **工程化 PoC**（CLI + Agent + RAG + MCP + 回归）。本文列出相对「全量生产上线」的缺口，按 **优先级 + 状态** 跟踪；逐项消化，避免对外表述与实现脱节。

**推荐对外说法**：具备上线演进基础的慢 SQL Agent 诊断原型 — 采集与审批接入待建设。

---

## 优先级说明

| 级别 | 含义 | 目标 |
|:----:|------|------|
| **P0** | 安全与叙事底线 | 不夸大已上线；风险可控 |
| **P1** | 演示与回归加固 | 多场景可讲、可测，不绑单表 |
| **P2** | 工程质量提升 | 结论可追溯、Token/误调更稳 |
| **P3** | 生产形态 | 服务化、接入日志链路、规模化 RAG |

状态：`待做` · `进行中` · `已完成`

---

## 缺口总表

| ID | 优先级 | 缺口 | 现状 | 目标 | 状态 |
|:--:|:------:|------|------|------|:----:|
| G01 | P0 | 对外表述与真实边界 | 文档已写 PoC/封存 | 统一口径：非全量线上服务 | 已完成 |
| G02 | P0 | 索引/DDL 误操作 | `add_mysql_index` 默认 dry_run | 保持默认；文档强调审批 | 已完成 |
| G03 | P0 | 凭证不进 Git | `.env` + example | CI/文档禁止提交密钥 | 已完成 |
| G04 | P1 | 测试数据单场景 | 仅 `slowlog-products.txt` | +锁等待 / JOIN / 索引命中 等 | 已完成 |
| G05 | P1 | 知识库偏单表 price | 16 篇，products 对齐好 | +JOIN / 深分页 / 过度索引 等 | 已完成 |
| G06 | P1 | 检索 golden 覆盖不足 | 10 条 | 与新 md、新慢日志对齐 | 已完成 |
| G07 | P2 | 结论缺少强制证据字段 | finish 自由文本 | Prompt 要求 result 含「证据：」行 | 已完成 |
| G08 | P2 | 工具列表一次性注入 | 全量 MCP Meta 进 Prompt | 按 `AgentPhase` 渐进披露 | 已完成 |
| G09 | P2 | LLM 非确定性 | 仅脚本 eval 确定性 | 真实 run 抽检 + 报告 `-report` 基线 | 已完成 |
| G10 | P2 | `ask_question` 不阻塞 | 只写状态 | HITL 暂停/恢复（stdin 即可） | 已完成 |
| G11 | P3 | 无 HTTP/API 服务 | CLI only | REST/内部 RPC + 异步任务 | 已完成 |
| G12 | P3 | 未接日志链路 | 与 Fluent Bit 叙事分离 | Bit/平台 → 对象存储 → 触发诊断 | 已完成 |
| G13 | P3 | RAG 内存索引 | embed + 进程内 TF-IDF | 向量库版本、热更新、按租户隔离 | 已完成 |
| G14 | P3 | 无多租户与审计 | 单 `.env` 连库 | 实例注册、操作审计、RBAC | 已完成 |
| G15 | P3 | 无成本与限流 | 每 run 直调 LLM | 配额、并发、超时、熔断 | 已完成 |
| G16 | P3 | Query 改写 / 多路召回 | 单路 `rag_query` | 规则抽取 + 可选 RRF | 待做 |

---

## 分步执行计划

### 阶段 A — P1 演示加固 ✅

1. **G04** `testdata/slowlog-{lock-wait,join-large,index-hit}.txt` + `testdata/README.md`  
2. **G05** 知识库 +5（JOIN、锁等待、深分页、过度索引、schema 变更边界）→ **21 篇**  
3. **G06** 检索 golden +5  
4. `make check` 通过  

### 阶段 B — P2 质量（改 Agent 行为，仍 CLI）← 当前

1. ~~**G07**~~ Prompt 已要求 finish 含「证据：」  
2. ~~**G08**~~ `ToolsForPhase` 过滤 Prompt 内工具列表  
3. ~~**G09**~~ 2026-05-25 四条慢日志真实跑通，见 [REAL-RUN-CHECKLIST](REAL-RUN-CHECKLIST.md)  
4. ~~**G10**~~ `SLOWLOG_AGENT_HITL=1` 时 ask_question 读 stdin；eval `ask_question_hitl`  
5. ~~**G11**~~ `cmd/agent-api`：`POST /v1/analyze`、GET 报告（见 [API.md](../design/API.md)）  

### 阶段 C — P3 生产形态（新模块或新仓库）← 当前

1. ~~**G11**~~ 最小 HTTP 已实现；生产需鉴权、异步队列  
2. ~~**G12**~~ `POST /v1/ingest` + Job 轮询 + [LOG-INGESTION](../design/LOG-INGESTION.md)  
3. ~~**G13**~~ 磁盘索引 JSON + `make rag-index-build` + `/v1/rag/rebuild`（见 [RAG-INDEX](../design/RAG-INDEX.md)）  
4. ~~**G14**~~ 实例注册 + JSONL 审计 + admin token（见 [OPS.md](../design/OPS.md)）  
5. ~~**G15**~~ 进程内限流 / 日配额 / 并发（`internal/ops/limits`）  
6. **G16** 多路召回按业务需要排期  

---

## 变更记录

| 日期 | 说明 |
|------|------|
| 2026-05-21 | 初版：缺口表 + P0～P3 分阶段 |
| 2026-05-21 | 阶段 A 完成：+3 慢日志、+5 知识库、+5 golden；G07 Prompt 证据行 |
| 2026-05-21 | G12 设计草案：LOG-INGESTION.md |
| 2026-05-25 | G09 完成：4 场景真实 LLM 抽检 + products 报告 `-report` 基线 |
| 2026-05-25 | G10 完成：SLOWLOG_AGENT_HITL + eval `ask_question_hitl` |
| 2026-05-25 | G11 完成：`agent-api` + [API.md](../design/API.md) + `make api-test` |
| 2026-05-25 | G12 完成：`/v1/ingest` 异步任务 + Fluent Bit 示例配置 |
| 2026-05-25 | G13 完成：RAG 磁盘索引 + rag-index CLI + API rebuild |
| 2026-05-25 | G14 完成：实例注册、JSONL 审计、admin token（[OPS.md](../design/OPS.md)） |
| 2026-05-25 | G15 完成：API 限流、日配额、并发（[OPS.md](../design/OPS.md)） |

---

## 完成度自检

```bash
make check
make api-test
make rag-check "JOIN 驱动表"
ls testdata/slowlog-*.txt
```

对外交流前对照：**G01～G03 必能口述**；**G04～G06 有样例可演示**；**G11～G15 为 PoC 已实现**；**G16 仍为路线图**。
