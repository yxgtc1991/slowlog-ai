# Action: products 表 price 过滤的索引建议

## 场景
- 慢 SQL：`SELECT * FROM products WHERE price >= ? ORDER BY created_at DESC LIMIT 20`
- 现有索引以 `code` 为最左列，无法服务仅按 `price` 过滤

## 推荐索引（示例）
- `(price, created_at)`：同时服务 **范围过滤** 与 **降序排序**
- 是否包含 `code` 取决于是否常按 code 等值 + price 范围查询

## 执行策略
- 先 `EXPLAIN` 对比加索引前后 `rows`、`Extra`
- 使用 MCP **add_mysql_index** 时默认 **dry_run=true**，确认 DDL 再上线

## 风险
- 写放大、缓冲池压力、在线 DDL 时间窗
- 需与 DBA 确认表量级与业务窗口

## 边界
- 不能替代对 **锁等待、缓存命中率** 的排查
- 一次性报表任务可能不值得加索引
