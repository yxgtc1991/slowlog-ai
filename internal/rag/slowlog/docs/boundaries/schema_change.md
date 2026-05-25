# Boundary:  schema 变更与上线流程

## Agent 能力边界
- 本工具可 **dry_run** 生成 `ALTER TABLE ... ADD INDEX`，**不替代** DBA 变更评审
- 不自动执行线上 DDL，不处理 **pt-osc/gh-ost** 排期与回滚预案

## 生产建议流程
1. 慢日志 + EXPLAIN 证据归档  
2. dry_run DDL 评审（表大小、锁、复制延迟）  
3. 低峰执行或在线变更工具  
4. 对比变更前后慢日志指标  

## 风险说明
- 大表加索引可能长时间 metadata lock  
- 主从延迟、磁盘空间需提前检查  

## 与 MCP 的对应
- `add_mysql_index` 默认 `dry_run=true`；仅人工确认后才可 `dry_run=false`
