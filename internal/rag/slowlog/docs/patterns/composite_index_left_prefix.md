# Pattern: 复合索引最左前缀未命中

## 典型场景
- 表上已有复合索引，例如 `idx_code_price_time(code, price, created_at)`
- SQL 的 WHERE **未包含最左列** `code`，仅过滤 `price` 或 `created_at`
- 慢日志 `Rows_examined` 接近表行数，`Rows_sent` 很小

## 机制说明
- InnoDB 复合索引按 **(col1, col2, col3)** 排序存储
- 条件必须从 **最左列** 起连续匹配，优化器才能走索引范围扫描
- 跳过左列只查中间列时，索引往往 **用不上**，退化为全表扫描 + 过滤

## 与演示慢日志的对应
- `products` 表：`WHERE price >= 100`，索引以 `code` 打头
- 应怀疑 **左前缀断裂**，而非「完全没有索引」

## 建议动作
- 用 EXPLAIN 看 `type`、`key`、`rows`
- 评估改写 SQL（带上选择性好的左列）或 **新建以 price 打头的索引** `(price, created_at)`

## 常见误判
- 看到表上有复合索引就认为一定命中
- 未对照 WHERE 列顺序与索引列顺序
