# Metric: Query_time 与 Lock_time

## Query_time
- 慢日志中的 **语句执行耗时**（秒），用户感知的主指标
- 包含解析、优化、执行、发送结果等阶段（不含应用层网络）

## Lock_time
- 等待 **表级/行级锁** 的时间
- Lock_time 高而 Rows_examined 不大时，优先怀疑 **锁竞争** 而非索引

## 联合解读
| 现象 | 可能方向 |
|------|----------|
| Query_time 高、Rows_examined 高 | 扫描/排序/索引问题 |
| Query_time 高、Lock_time 高 | 并发写、长事务、锁等待 |
| Query_time 高、Rows_sent 很小 | 过滤或 LIMIT 后结果少，但扫描多 |

## 注意事项
- 不要只盯 Query_time 忽视 Lock_time
- 优化索引无法解决纯锁等待问题
