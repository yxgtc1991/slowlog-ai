package mcp

import (
	"ai_slow_log/internal/mysql"
	"context"
	"errors"
	"fmt"
)

// AddMySQLIndexCapability 在指定表上创建索引（默认 dry_run 只返回 DDL，不执行）。
type AddMySQLIndexCapability struct {
	Client *mysql.Client
}

func (c *AddMySQLIndexCapability) Name() string {
	return "add_mysql_index"
}

func (c *AddMySQLIndexCapability) Description() string {
	return "为 MySQL 表添加索引。默认 dry_run=true 仅返回 ALTER TABLE DDL；设 dry_run=false 才会真正执行。示例：database=test, table=products, index_name=idx_sku, columns=sku。"
}

func (c *AddMySQLIndexCapability) InputSchema() map[string]string {
	return map[string]string{
		"database":   "string // 可选，如 test；默认用当前库",
		"table":      "string // 必填，如 products",
		"index_name": "string // 必填，新索引名",
		"columns":    "string | array // 必填，列名，如 sku 或 [\"category_id\",\"created_at\"]",
		"unique":     "bool // 可选，是否 UNIQUE INDEX，默认 false",
		"dry_run":    "bool // 可选，默认 true；false 时执行 DDL",
	}
}

func (c *AddMySQLIndexCapability) Execute(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	if c.Client == nil {
		return nil, errors.New("mysql client is not configured")
	}

	table, ok := stringInput(input, "table")
	if !ok {
		return nil, errors.New("table is required")
	}
	indexName, ok := stringInput(input, "index_name")
	if !ok {
		return nil, errors.New("index_name is required")
	}
	columns, err := parseColumns(input)
	if err != nil {
		return nil, err
	}

	db, _ := stringInput(input, "database")
	if db != "" {
		if err := c.Client.UseDatabase(ctx, db); err != nil {
			return nil, fmt.Errorf("use database: %w", err)
		}
	} else {
		db = ""
		if snap, ok := c.Client.ConfigSnapshot()["database"].(string); ok {
			db = snap
		}
	}

	opts := mysql.AddIndexOptions{
		Database:  db,
		Table:     table,
		IndexName: indexName,
		Columns:   columns,
		Unique:    boolInput(input, "unique", false),
	}

	ddl, err := c.Client.BuildAddIndexDDL(opts)
	if err != nil {
		return nil, err
	}

	dryRun := boolInput(input, "dry_run", true)
	result := map[string]interface{}{
		"dry_run":        dryRun,
		"ddl":            ddl,
		"table":          table,
		"index_name":     indexName,
		"columns":        columns,
		"unique":         opts.Unique,
		"current_database": c.Client.ConfigSnapshot()["database"],
	}

	if dryRun {
		result["executed"] = false
		result["message"] = "dry_run: DDL not executed; set dry_run=false to apply"
		return result, nil
	}

	execOut, err := c.Client.AddIndex(ctx, opts)
	if err != nil {
		return nil, err
	}
	for k, v := range execOut {
		result[k] = v
	}
	result["dry_run"] = false
	result["message"] = "index created"
	return result, nil
}
