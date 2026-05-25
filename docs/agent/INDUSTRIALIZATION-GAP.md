# 距离「工业化 Agent」还有多远

[生产化缺口](PRODUCTION-GAPS.md) · [路线图](ROADMAP.md) · [调研问答](RESEARCH-QA.md)

> **本仓定位**：慢 SQL 领域的 **工程化 PoC**（V6 Agent + RAG + MCP + 双层 golden + HTTP 雏形）。  
> **G01～G16 已全部完成** → 可封存演示与工位复盘。  
> **工业化 Agent** 指：可 7×24 托管、可观测、可回归、可扩租户、可接真实日志与变更流程，而不是「能跑通一条 Demo」。

---

## 一句话距离感

| 维度 | 粗估 | 说明 |
|------|:----:|------|
| **演示级 Agent 工程化** | **~85%** | 多轮协议、Eval、Trace、RAG 真检索、报告存档、文档闭环 |
| **可对内试点的「准生产」** | **~45%** | 有 HTTP/ingest/审计/限流 PoC，缺统一鉴权、分布式队列、SLO |
| **对外商业级工业化平台** | **~25%** | 缺多租户隔离、向量库运维、成本治理、变更审批与合规 |

不必用单一百分比「完工」——下面按 **能力域** 拆开看缺口与建议下一步。

---

## 能力域对照（0～5 级）

| 能力域 | 本仓现状（约） | 工业化常见目标 | 差距要点 |
|--------|:--------------:|----------------|----------|
| **Agent 协议与状态** | 4 | 5 | V6 NextAction + AgentState + 阶段工具披露；缺 Plan-and-Execute、Reflection、长期 Memory |
| **工具 / MCP** | 3 | 5 | 真连库 + EXPLAIN + dry_run；缺连接池、超时重试策略、工具级 RBAC 强制执行 |
| **RAG** | 3 | 5 | 21 篇 + TF-IDF/embedding + 磁盘索引 + G16 改写/RRF；缺 Milvus 运维、HyDE、在线 A/B、租户隔离索引 |
| **可观测** | 3 | 5 | Trace span + 报告 JSON；缺 Prometheus、分布式 trace、告警与 SLO 看板 |
| **质量与回归** | 4 | 5 | 6 条 Agent eval + 15 条检索 golden + CI；缺 LLM 评判器、线上 badcase 回流 |
| **服务化** | 2 | 5 | agent-api + ingest job；缺 OAuth/mTLS、水平扩展、削峰队列 |
| **安全与合规** | 2 | 5 | dry_run、凭证不进 Git、审计 JSONL；缺 SSO、密钥轮转、操作审批流 |
| **成本与容量** | 2 | 5 | G15 进程内限流；缺按租户配额、Token 计量、模型路由与熔断 |

**综合（加权心算）**：在「慢 SQL 单域 Agent」前提下，约 **3.0 / 5** —— 已是**像样的工程原型**，距**平台化工业产品**通常还有 **1～2 个里程碑**（各 2～4 周 PoC 级，或一个季度产品级）。

---

## 本仓已具备的「工业化要素」（可写进汇报）

- **可回归**：`make check` = 单测 + Agent eval + RAG golden + 文档链；CI 已接。
- **可复盘**：报告 JSON / MD / HTML / brief + Trace span；不必重跑 LLM。
- **可演进**：V1～V6 同仓；`Retriever` / MCP 接口可换实现而不改 V6 循环。
- **可演示多场景**：4 条慢日志 + 21 篇知识库 + 真实 LLM 抽检清单。
- **可接下游雏形**：HTTP analyze、ingest、实例注册、审计、限流、RAG 热重建。

这些在 **实验室 / 工位 / 内部分享** 场景已够用；**缺的是「上线运营」那一圈**。

---

## 仍缺什么（按投入产出排序）

### 第一梯队 — 想「内网试点」优先做

| 项 | 价值 | 工作量（粗估） |
|----|------|----------------|
| API **统一鉴权**（`SLOWLOG_API_KEY`） | 堵住裸奔接口 | **PoC 已做**（D1） |
| **Prometheus `/metrics`** | 请求/analyze/限流计数 | **PoC 已做**（D1） |
| **异步队列**（ingest → worker，Redis/DB job 表已有可加深） | 削峰、可重启 | 3～5 天 |
| JWT / SSO、ingest 与 analyze 统一身份 | 企业登录 | 1 周+ |
| **allowed_tools 真正约束 Agent**（G14 元数据落地） | 多实例叙事可信 | 2～4 天 |
| 定期 **real-run 报告基线** 进文档（不进 Git） | 防 Prompt/模型漂移 | 0.5 天/次 |

### 第二梯队 — 想「像平台」再做

| 项 | 价值 | 工作量（粗估） |
|----|------|----------------|
| **Milvus / pgvector** 替换 JSON 索引 + 按实例 partition | RAG 规模化 | 1～2 周 |
| **HyDE 或 LLM Query 改写**（与 G16 规则并存） | 检索召回上限 | 3～5 天 |
| **Reflection 轮**（finish 前核对 EXPLAIN 证据） | 降幻觉 | 3～5 天 |
| **Plan-and-Execute**（先 plan 再工具） | 复杂案可讲 | 1 周 |
| 历史 case **向量记忆**（schema+sql 去重） | 重复慢 SQL 加速 | 1～2 周 |

### 第三梯队 — 商业 / 合规级

多区域部署、SSO、变更审批对接、按租户计费、模型网关、数据驻留与审计归档 —— **新里程碑或新仓库**，不宜在本 PoC 上无限堆。

---

## 和 G01～G16 的关系

| 阶段 | 含义 |
|------|------|
| **G01～G16 完成** | 本仓库 **PoC 路线图结案** |
| **工业化** | **下一张路线图**（本文 + [PRODUCTION-GAPS](PRODUCTION-GAPS.md) 未列的 P4） |

G 表解决的是「演示与叙事不脱节」；工业化解决的是「7×24 有人敢用」。

---

## 建议的「还能继续做什么」

按你剩余时间选 **一条主线**，避免散弹：

1. **封存收工（0 代码）**  
   跑通 `make check` + [REAL-RUN-CHECKLIST](REAL-RUN-CHECKLIST.md) + 更新 [REVIEW-CHECKLIST](REVIEW-CHECKLIST.md) 演示顺序 → **对外只说 PoC 已封存**。

2. **内网试点线（2～3 周）**  
   鉴权 → 队列 worker → `/metrics` → Fluent Bit 联调一条真实慢日志源。

3. **Agent 研究线（1～2 周）**  
   Reflection 或 Plan-and-Execute 二选一 + 1 条 eval golden。

4. **RAG 研究线（1 周）**  
   Milvus PoC 或 HyDE + 对比 `make rag-test` 命中率。

5. **新仓库产品线**  
   本仓 tag `poc-sealed`；生产化（租户、计费、审批）fork 新 repo，避免 PoC 叙事被污染。

---

## 对外怎么说（避免过度承诺）

| 可以说 | 不要说 |
|--------|--------|
| 具备工业化 **演进基础** 的慢 SQL Agent 原型 | 已工业级 / 已全量上线 |
| V6 + MCP + RAG + 回归 + HTTP 雏形 | 替代 DBA / 自动加索引上线 |
| G16 规则改写 + RRF 提升召回 | 已达商业检索引擎效果 |

---

## 自检

```bash
make check && make api-test
make rag-check "JOIN 驱动表"
SLOWLOG_RAG_MULTI=1 SLOWLOG_RAG_SLOWLOG_FILE=testdata/slowlog-products.txt make rag-check "全表扫描"
```

文档链：`make doc-links`
