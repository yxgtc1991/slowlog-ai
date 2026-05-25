# RAG 磁盘索引（G13）

[guides/RAG.md](../guides/RAG.md) · [生产化缺口](../agent/PRODUCTION-GAPS.md)

> 将 TF-IDF / embedding 索引 **持久化到磁盘**，知识库变更时按 `docs_hash` 自动失效重建。PoC 阶段用 JSON 文件，接口与 Milvus 等向量库可后续替换 `Retriever` 实现。

---

## 目录与文件

默认目录：`data/rag-index/`（或 `SLOWLOG_RAG_INDEX_DIR`）

| 文件 | 内容 |
|------|------|
| `tfidf-index.json` | 全部 chunk + TF-IDF 权重 + manifest |
| `embedding-index.json` | chunk + 向量 + embed 模型名（可选） |

`manifest.docs_hash` 与当前 `//go:embed` 知识库一致时才加载；否则回退内存重建。

---

## 环境变量

| 变量 | 说明 |
|------|------|
| `SLOWLOG_RAG_PERSIST=1` | 启动时优先读盘；缺失则构建并写盘 |
| `SLOWLOG_RAG_INDEX_DIR` | 索引目录 |
| `SLOWLOG_RAG` | `tfidf` / `embedding`（加载逻辑同上） |

未设 `SLOWLOG_RAG_PERSIST` 时：若目录下已有未过期 `tfidf-index.json`，仍会 **自动加载**（加速冷启动）。

---

## 命令

```bash
make rag-index-build          # 仅 TF-IDF
go run ./cmd/rag-index -embedding   # 含向量（local/http embedder）
make rag-index-status         # 查看 docs_hash 是否一致
```

---

## HTTP 管理（agent-api）

需 `X-Webhook-Secret`（与 ingest 相同，若配置了 `SLOWLOG_WEBHOOK_SECRET`）：

```bash
curl -s http://127.0.0.1:8080/v1/rag/status -H 'X-Webhook-Secret: ...'
curl -s -X POST 'http://127.0.0.1:8080/v1/rag/rebuild?embedding=true' -H 'X-Webhook-Secret: ...'
```

改知识库 md 后：执行 **rebuild** 或 `make rag-index-build`，无需改 Agent 代码。

---

## 与 Milvus 的关系

| 本阶段 | 后续 |
|--------|------|
| JSON 文件 + 单进程 | Milvus Collection + partition（按 `source` 分类） |
| `Retriever` 接口不变 | 新增 `MilvusRetriever` 实现 |

分区策略调研结论：单 Collection 多 **partition**（按 pattern/metric 等）通常比多 Collection 易运维；见 [RESEARCH-QA.md](../agent/RESEARCH-QA.md) Q17。

---

## .gitignore

`data/rag-index/` 不提交仓库；CI 仍用内存索引 + `make rag-test`。
