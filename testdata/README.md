# 测试数据

与 Agent eval、RAG 知识库、`make rag-check` 演示对齐。

| 文件 | 场景 | 要点 |
|------|------|------|
| `slowlog-products.txt` | 单表 price 过滤 + ORDER BY | 左前缀断裂、Rows_examined 高（**默认 eval**） |
| `slowlog-lock-wait.txt` | UPDATE 锁等待 | Lock_time ≈ Query_time，Rows_examined 小 |
| `slowlog-join-large.txt` | 多表 JOIN | 驱动表、JOIN 索引、百万级 examined |
| `slowlog-index-hit.txt` | 索引已命中 | examined≈sent，避免误建议加索引 |
| `eval/minimal-report.json` | 报告断言 fixture | `AssertReportFile` |

```bash
go run ./cmd/slowlog-ai testdata/slowlog-lock-wait.txt
make agent-run   # 默认仍用 products；可改入口或传路径
```
