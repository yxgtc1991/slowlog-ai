# RAG 检索怎么用

[文档索引](../INDEX.md) · [路线图](../agent/ROADMAP.md)

知识库：`internal/rag/slowlog/docs/**/*.md`（编译时 `embed` 进二进制），索引见 [slowlog/docs/README.md](../../internal/rag/slowlog/docs/README.md)。  
每篇 Markdown 按 **`##` 二级标题** 切成多个 chunk（工业级库 **14 篇 / 35+ chunk**），再检索 TopK（默认 3）。

流程图：[diagrams/rag-flow.md](../diagrams/rag-flow.md)

---

## 1. 三种检索模式（`SLOWLOG_RAG`）

| 值 | 何时用 | 需要 API Key | 说明 |
|----|--------|:------------:|------|
| **`tfidf`**（默认） | `make run` / `make agent-run` | 否 | 关键词 TF-IDF，与 query 相关 |
| **`embedding`** | 试语义检索、对比效果 | 否* | 内存向量 + 余弦 TopK；`*` 默认 `local` 不调 API |
| **`mock`** | `make agent-eval` | 否 | 固定 2 条，回归确定性 |

未设置 `SLOWLOG_RAG` 时等同于 **`tfidf`**。

---

## 2. 命令（不跑 LLM，秒级验证）

```bash
# 默认 TF-IDF，默认查询词
make rag-check

# 自定义查询（参数即查询词）
make rag-check "LIMIT ORDER BY filesort 索引"

# 检索 golden + 单元测试（改知识库后必跑）
make rag-test

# 并排对比 tfidf vs embedding（强制 local 向量，无需 Key）
make rag-check-compare

# 单独试向量检索（本地 embedder）
SLOWLOG_RAG=embedding make rag-check

# 单独试 TF-IDF（显式）
SLOWLOG_RAG=tfidf make rag-check

# 回归仍用 Mock（勿改默认 eval 脚本）
SLOWLOG_RAG=mock make agent-eval
```

---

## 3. 环境变量

写在 `.env` 或命令前 `export`（参考 `.env.example`）。

| 变量 | 默认 | 说明 |
|------|------|------|
| `SLOWLOG_RAG` | `tfidf` | `tfidf` \| `embedding` \| `mock` |
| `SLOWLOG_RAG_TOPK` | `3` | 返回条数上限 |
| `SLOWLOG_EMBEDDING_PROVIDER` | `local` | 仅 `embedding` 时：`local`（哈希向量）\| `http`（调 API） |
| `SLOWLOG_EMBEDDING_API_KEY` | — | `http` 时用；可复用 `DEEPSEEK_API_KEY` |
| `SLOWLOG_EMBEDDING_BASE_URL` | OpenAI 默认 | OpenAI **兼容** 根地址，如 `https://api.openai.com/v1` |
| `SLOWLOG_EMBEDDING_MODEL` | `text-embedding-3-small` | embeddings 模型名 |

### 用真 Embedding API（可选）

```bash
SLOWLOG_RAG=embedding \
SLOWLOG_EMBEDDING_PROVIDER=http \
SLOWLOG_EMBEDDING_API_KEY=sk-xxx \
SLOWLOG_EMBEDDING_BASE_URL=https://api.openai.com/v1 \
SLOWLOG_EMBEDDING_MODEL=text-embedding-3-small \
make rag-check
```

### 让 Agent 全程走向量检索

```bash
# .env 中写一行即可
SLOWLOG_RAG=embedding
SLOWLOG_EMBEDDING_PROVIDER=local   # 或 http + Key

make run          # 或 make agent-run
```

V6 里 LLM 每轮若选 `retrieve_rag`，会用当前 `SLOWLOG_RAG` 对应的后端；**接口不变**，只换实现。

---

## 4. 和 Agent / Eval 的关系

| 场景 | 建议配置 |
|------|----------|
| 日常演示、写报告 | 默认即可（`tfidf`） |
| 对比检索效果 | `make rag-check-compare` |
| 改 RAG 代码后 | `make rag-check` + `make agent-eval` |
| CI / 无 Token 回归 | `agent-eval` 内部或手动 `SLOWLOG_RAG=mock` |

---

## 5. 代码入口

| 文件 | 作用 |
|------|------|
| `internal/rag/chunks.go` | 按 `##` 切 chunk |
| `internal/rag/tfidf.go` | TF-IDF 检索 |
| `internal/rag/embedding.go` | 内存向量 TopK |
| `internal/rag/embedder.go` | `local` / `http` Embedder |
| `internal/rag/factory.go` | `NewDefaultRetriever()` |
| `cmd/rag-check/main.go` | 本地试跑 CLI |

扩展新知识：在 `internal/rag/slowlog/docs/` 下按目录加 `.md`（`patterns/`、`metrics/` 等），重新编译即可。
