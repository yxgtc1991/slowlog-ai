package eval

import (
	promptv6 "ai_slow_log/internal/prompt/slowlog"
)

const defaultSlowLogPath = "testdata/slowlog-products.txt"

// AllCases 内置 golden cases（脚本化 LLM，无 API 费用）。
func AllCases() []Case {
	executor := NewStubExecutor().
		WithTool("connect_mysql_instance", map[string]interface{}{
			"connected": true,
			"database":  "test",
		}).
		WithTool("explain_mysql_query", map[string]interface{}{
			"rows": []map[string]interface{}{
				{"type": "ALL", "key": nil, "rows": 48000},
			},
		}).
		WithTool("add_mysql_index", map[string]interface{}{
			"dry_run": true,
			"ddl":     "ALTER TABLE products ADD INDEX idx_price_created (price, created_at)",
		})

	return []Case{
		{
			Name:        "guided_flow",
			SlowLogPath: defaultSlowLogPath,
			Guide:       true,
			Executor:    executor,
			Script: []string{
				decisionJSON("已读慢日志，先查 RAG", retrieveRAG("rows_examined 高 全表扫描 复合索引")),
				decisionJSON("确认库可用", callTool("connect_mysql_instance", map[string]interface{}{})),
				decisionJSON("对慢 SQL 做 EXPLAIN", callTool("explain_mysql_query", map[string]interface{}{
					"database": "test",
					"sql":      "SELECT * FROM products WHERE price >= 100 ORDER BY created_at DESC LIMIT 20",
				})),
				decisionJSON("dry_run 建索引", callTool("add_mysql_index", map[string]interface{}{
					"table":      "products",
					"index_name": "idx_price_created",
					"columns":    []string{"price", "created_at"},
					"dry_run":    true,
				})),
				decisionJSON("归纳根因", analyze("WHERE 仅 price，复合索引左前缀用不上，Rows_examined 高")),
				decisionJSON("输出结论", finish(
					"诊断：products 表对 price 过滤导致全表扫描（Rows_examined 约 48000）。"+
						"建议索引：(price, created_at)。预期减少扫描行数。风险：写入略增。",
				)),
			},
			Expect: Expect{
				Trajectory: []TrajectoryStep{
					{Type: string(promptv6.ActionRetrieveRAG)},
					{Type: string(promptv6.ActionCallTool), ToolName: "connect_mysql_instance"},
					{Type: string(promptv6.ActionCallTool), ToolName: "explain_mysql_query"},
					{Type: string(promptv6.ActionCallTool), ToolName: "add_mysql_index"},
					{Type: string(promptv6.ActionAnalyze)},
					{Type: string(promptv6.ActionFinish)},
				},
				ToolsMustCall: []string{
					"connect_mysql_instance",
					"explain_mysql_query",
					"add_mysql_index",
				},
				FinalContains:    []string{"全表扫描", "price", "索引"},
				MinIterations:    6,
				MaxIterations:    6,
				NoActionErrors:   true,
				FinalPhase:       "finished",
				MinSpansPerRound: 2,
			},
		},
		{
			Name:        "tool_name_as_type",
			SlowLogPath: defaultSlowLogPath,
			Guide:       false,
			Executor: NewStubExecutor().WithTool("analyze_slow_log", map[string]interface{}{
				"summary": "stub slow log analysis",
			}),
			Script: []string{
				`{
  "current_state": "模型误把工具名写在 type",
  "next_action": {
    "type": "analyze_slow_log",
    "reasoning": "先跑慢日志工具",
    "tool_args": {"slow_log": "stub"}
  }
}`,
				decisionJSON("完成", finish("已通过 normalizeNextAction 纠正 type，并完成分析。")),
			},
			Expect: Expect{
				Trajectory: []TrajectoryStep{
					{Type: string(promptv6.ActionCallTool), ToolName: "analyze_slow_log"},
					{Type: string(promptv6.ActionFinish)},
				},
				FinalContains: []string{"normalize"},
				MaxIterations: 2,
			},
		},
		{
			Name:        "tool_error_then_finish",
			SlowLogPath: defaultSlowLogPath,
			Guide:       false,
			Executor: NewStubExecutor().
				WithTool("connect_mysql_instance", nil).
				WithToolError("connect_mysql_instance", errStub("connection refused")),
			Script: []string{
				decisionJSON("连库", callTool("connect_mysql_instance", nil)),
				decisionJSON("库不可用仍给建议", finish("无法连接 MySQL，基于慢日志推断：全表扫描，建议检查 price 索引。")),
			},
			Expect: Expect{
				Trajectory: []TrajectoryStep{
					{Type: string(promptv6.ActionCallTool), ToolName: "connect_mysql_instance"},
					{Type: string(promptv6.ActionFinish)},
				},
				FinalContains: []string{"全表扫描"},
				ToolFailures: map[string]ToolFailureExpect{
					"connect_mysql_instance": {Code: "mysql_connection", Retryable: true},
				},
				NoActionErrors: false,
			},
		},
	}
}

func errStub(msg string) error {
	return &stubError{msg: msg}
}

type stubError struct{ msg string }

func (e *stubError) Error() string { return e.msg }
