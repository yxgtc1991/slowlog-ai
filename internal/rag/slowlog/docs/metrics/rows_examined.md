# Metric: Rows_examined

## 含义
MySQL 执行查询过程中扫描的行数。

## 常见解读
- 数值越大，CPU / IO 开销越高
- 与 Rows_sent 差距大时，通常存在过滤或排序开销

## 注意事项
- InnoDB 的 Rows_examined 是估算值
- 不能单独作为是否缺索引的唯一依据