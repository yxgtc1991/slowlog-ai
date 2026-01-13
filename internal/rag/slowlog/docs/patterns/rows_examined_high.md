# Pattern: Rows_examined 显著大于 Rows_sent

## 症状
- Rows_examined 远大于 Rows_sent（例如 > 100x）
- Query_time 较高

## 可能原因
- 未命中合适索引
- 使用范围查询但索引选择性低
- 扫描后排序（filesort）

## 判断依据
- 慢日志中的 Rows_examined 指标
- 无法从慢日志确认具体索引情况

## 常见误判
- ❌ 一定是缺少索引（不一定）
- ❌ 一定是全表扫描（需要 EXPLAIN）

## 下一步需要的信息
- 表结构
- 索引信息
- EXPLAIN 执行计划