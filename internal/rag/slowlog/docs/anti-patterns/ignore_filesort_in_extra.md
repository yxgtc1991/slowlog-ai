# Anti-Pattern: 忽视 EXPLAIN 的 Using filesort

## 错误结论
「有索引、type 不是 ALL，所以不会有排序开销。」

## 实际情况
- `Extra` 出现 **Using filesort** 表示仍需排序
- range/ref 访问后仍可能对大量行做 filesort
- ORDER BY 列顺序与索引不一致时常见

## 纠正方式
- 对照 ORDER BY 与索引列顺序
- 考虑 `(filter_cols..., sort_col)` 组合索引

## 关联
- 见 pattern「ORDER BY + LIMIT 仍可能慢」
