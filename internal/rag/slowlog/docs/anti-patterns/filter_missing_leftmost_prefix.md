# Anti-Pattern: 有复合索引但 WHERE 漏最左列

## 错误结论
「表上已经有 `(code, price, created_at)` 索引，所以这条 SQL 会用索引。」

## 实际情况
- WHERE 只有 `price >= ?` 时，**最左列 code 未参与条件**
- 复合索引 **左前缀原则** 不满足，常见 `type=ALL`、key 为 NULL
- 表现为全表扫描或大量回表，与「无索引」症状相似

## 纠正方式
- EXPLAIN 核对 `key` 是否真命中、是否预期索引名
- 改写：在业务允许时增加 **高选择性左列** 条件
- 或新增 **以过滤列打头** 的索引（如 `(price, created_at)`）

## 汇报话术
- 强调「索引存在 ≠ 索引被使用」
- 用 Rows_examined / EXPLAIN 证据链说明
