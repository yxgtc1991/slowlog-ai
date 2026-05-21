package mcp

import (
	"ai_slow_log/internal/mysql"
	"context"
	"errors"
	"fmt"
)

// ExplainMySQLQueryCapability 对 SELECT 执行 EXPLAIN（如 test.products 上的查询）。
type ExplainMySQLQueryCapability struct {
	Client *mysql.Client
}

func (c *ExplainMySQLQueryCapability) Name() string {
	return "explain_mysql_query"
}

func (c *ExplainMySQLQueryCapability) Description() string {
	return "对单条 SELECT 语句执行 EXPLAIN，返回执行计划（仅允许 SELECT，禁止多语句）。可先 connect_mysql_instance 切到 test 库，或传 database=test。示例表：test.products。"
}

func (c *ExplainMySQLQueryCapability) InputSchema() map[string]string {
	return map[string]string{
		"sql":      "string // 必填，单条 SELECT，如 SELECT * FROM products WHERE id = 1",
		"database": "string // 可选，执行前 USE 的数据库，如 test",
		"format":   "string // 可选，traditional（默认）或 json",
	}
}

func (c *ExplainMySQLQueryCapability) Execute(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	if c.Client == nil {
		return nil, errors.New("mysql client is not configured")
	}

	sql, ok := stringInput(input, "sql")
	if !ok {
		return nil, errors.New("sql is required")
	}

	if db, ok := stringInput(input, "database"); ok {
		if err := c.Client.UseDatabase(ctx, db); err != nil {
			return nil, fmt.Errorf("use database: %w", err)
		}
	}

	format := mysql.ExplainFormatTraditional
	if f, ok := stringInput(input, "format"); ok {
		format = mysql.ExplainFormat(f)
	}

	out, err := c.Client.Explain(ctx, sql, format)
	if err != nil {
		return nil, err
	}
	out["current_database"] = c.Client.ConfigSnapshot()["database"]
	return out, nil
}
