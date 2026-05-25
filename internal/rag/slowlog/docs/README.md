# slowlog 知识库索引

编译时嵌入 `internal/rag`（`//go:embed`），供 TF-IDF / embedding 检索。

## 目录

| 目录 | 用途 |
|------|------|
| **patterns/** | 可识别的性能模式 |
| **anti-patterns/** | 常见误判 |
| **metrics/** | 慢日志指标含义 |
| **actions/** | 建议动作与风险 |
| **boundaries/** | Agent 能力边界 |

## 与演示数据对齐

默认慢日志：`testdata/slowlog-products.txt`（`test.products`，`price` 过滤 + `ORDER BY created_at`）。

优先检索命中：

- `patterns/composite_index_left_prefix.md`
- `patterns/order_by_with_limit.md`
- `anti-patterns/filter_missing_leftmost_prefix.md`
- `anti-patterns/ignore_filesort_in_extra.md`
- `metrics/rows_sent.md`
- `boundaries/lock_contention.md`
- `actions/index_for_products_price_filter.md`
- `patterns/covering_index_basics.md`
- `patterns/join_driving_table.md`
- `patterns/lock_wait_vs_scan.md`
- `metrics/slowlog_header_fields.md`
- `anti-patterns/deep_offset.md`
- `anti-patterns/over_indexing.md`
- `boundaries/schema_change.md`

测试慢日志（多场景）：仓库根目录 `testdata/README.md`。

## 验证

```bash
make rag-check "price 复合索引 最左前缀"
go test ./internal/rag/... -run Golden
```

扩展：新增 `.md` 后重新编译；保持每篇多个 `##` 标题便于 chunk 切分。
