# Pattern: ORDER BY + LIMIT 仍可能慢

## 症状
- SQL 带 `ORDER BY ... LIMIT N`，业务认为「只取 N 行应该很快」
- `Query_time` 仍高，`Rows_examined` 很大
- EXPLAIN 可能出现 `Using filesort`

## 原因
- 优化器需先 **找到并排序** 满足 WHERE 的行，再截断 LIMIT
- 若 ORDER BY 列 **不在索引** 或索引顺序与排序不一致，会在内存/磁盘做 filesort
- LIMIT 只减少 **返回行数**，不减少 **扫描与排序成本**

## 与演示慢日志的对应
- `ORDER BY created_at DESC LIMIT 20` 且 WHERE 未走索引时，仍可能扫数万行再排序

## 优化方向
- 让索引同时覆盖 **WHERE 过滤列 + ORDER BY 列**（如 `(price, created_at)`）
- 确认 EXPLAIN 中 `Extra` 是否出现 `Using index` / 避免 `Using filesort`

## 关联反模式
- 见 anti-pattern「有 LIMIT 就一定快」
