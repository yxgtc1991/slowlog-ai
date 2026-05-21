package mysql

import (
	"ai_slow_log/internal/config"
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Client 封装本地 MySQL 连接（供 MCP Capability 使用）。
type Client struct {
	cfg config.MySQLConfig
	db  *sql.DB
}

// NewClient 使用配置建立连接池并 Ping。
func NewClient(cfg config.MySQLConfig) (*Client, error) {
	dsn := buildDSN(cfg)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping mysql: %w", err)
	}
	return &Client{cfg: cfg, db: db}, nil
}

func buildDSN(cfg config.MySQLConfig) string {
	host := cfg.Host
	if cfg.Port > 0 {
		host = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	}
	// database 可为空，连接后再 USE db
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&charset=utf8mb4&loc=Local&timeout=10s",
		cfg.User, cfg.Password, host, cfg.Database)
}

// Close 关闭连接池。
func (c *Client) Close() error {
	if c.db == nil {
		return nil
	}
	return c.db.Close()
}

// ConfigSnapshot 返回可展示的配置（不含密码）。
func (c *Client) ConfigSnapshot() map[string]interface{} {
	return map[string]interface{}{
		"host":     c.cfg.Host,
		"port":     c.cfg.Port,
		"user":     c.cfg.User,
		"database": c.cfg.Database,
	}
}

// Ping 检查连接并返回版本等信息。
func (c *Client) Ping(ctx context.Context) (map[string]interface{}, error) {
	if err := c.db.PingContext(ctx); err != nil {
		return nil, err
	}
	var version string
	if err := c.db.QueryRowContext(ctx, "SELECT VERSION()").Scan(&version); err != nil {
		return nil, err
	}
	out := map[string]interface{}{
		"connected": true,
		"version":   version,
		"instance":  c.ConfigSnapshot(),
	}
	return out, nil
}

// ListDatabases 列出可访问的数据库（过滤系统库可选保留 information_schema 等供诊断）。
func (c *Client) ListDatabases(ctx context.Context) ([]string, error) {
	rows, err := c.db.QueryContext(ctx, "SHOW DATABASES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dbs []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		dbs = append(dbs, name)
	}
	return dbs, rows.Err()
}

// UseDatabase 切换默认库（用于后续 EXPLAIN 等）。
func (c *Client) UseDatabase(ctx context.Context, dbName string) error {
	dbName = strings.TrimSpace(dbName)
	if dbName == "" {
		return fmt.Errorf("database name is empty")
	}
	// 简单校验，防止注入
	for _, r := range dbName {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
			return fmt.Errorf("invalid database name: %s", dbName)
		}
	}
	_, err := c.db.ExecContext(ctx, "USE `"+dbName+"`")
	if err != nil {
		return err
	}
	c.cfg.Database = dbName
	return nil
}

// DB 暴露底层连接（供后续 EXPLAIN Capability 使用）。
func (c *Client) DB() *sql.DB {
	return c.db
}
