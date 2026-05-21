package mysql

import (
	"context"
	"fmt"
	"strings"
)

// AddIndexOptions 创建索引的参数。
type AddIndexOptions struct {
	Database  string
	Table     string
	IndexName string
	Columns   []string
	Unique    bool
}

// BuildAddIndexDDL 生成 ALTER TABLE ... ADD [UNIQUE] INDEX DDL（不执行）。
func (c *Client) BuildAddIndexDDL(opts AddIndexOptions) (string, error) {
	db := strings.TrimSpace(opts.Database)
	if db == "" {
		db = c.cfg.Database
	}
	if err := validateIdent(opts.Table, "table"); err != nil {
		return "", err
	}
	if err := validateIdent(opts.IndexName, "index_name"); err != nil {
		return "", err
	}
	if len(opts.Columns) == 0 {
		return "", fmt.Errorf("columns is required")
	}
	if err := validateIdents("column", opts.Columns...); err != nil {
		return "", err
	}
	if db != "" {
		if err := validateIdent(db, "database"); err != nil {
			return "", err
		}
	}

	quotedCols := make([]string, len(opts.Columns))
	for i, col := range opts.Columns {
		quotedCols[i] = "`" + col + "`"
	}

	kind := "INDEX"
	if opts.Unique {
		kind = "UNIQUE INDEX"
	}

	tableRef := "`" + opts.Table + "`"
	if db != "" {
		tableRef = fmt.Sprintf("`%s`.`%s`", db, opts.Table)
	}

	return fmt.Sprintf(
		"ALTER TABLE %s ADD %s `%s` (%s)",
		tableRef,
		kind,
		opts.IndexName,
		strings.Join(quotedCols, ", "),
	), nil
}

// AddIndex 执行建索引 DDL。
func (c *Client) AddIndex(ctx context.Context, opts AddIndexOptions) (map[string]interface{}, error) {
	ddl, err := c.BuildAddIndexDDL(opts)
	if err != nil {
		return nil, err
	}
	if _, err := c.db.ExecContext(ctx, ddl); err != nil {
		return nil, fmt.Errorf("add index: %w", err)
	}
	return map[string]interface{}{
		"executed":   true,
		"ddl":        ddl,
		"database":   opts.Database,
		"table":      opts.Table,
		"index_name": opts.IndexName,
		"columns":    opts.Columns,
		"unique":     opts.Unique,
	}, nil
}
