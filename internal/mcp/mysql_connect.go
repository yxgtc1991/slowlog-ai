package mcp

import (
	"ai_slow_log/internal/mysql"
	"context"
	"errors"
	"fmt"
)

// ConnectMySQLCapability 连接（校验）配置中的本地 MySQL 实例，并返回版本与库列表。
type ConnectMySQLCapability struct {
	Client *mysql.Client
}

func (c *ConnectMySQLCapability) Name() string {
	return "connect_mysql_instance"
}

func (c *ConnectMySQLCapability) Description() string {
	return "连接环境变量中配置的 MySQL 实例，返回连接状态、服务端版本与可访问的数据库列表。可选指定 database 以切换当前库。"
}

func (c *ConnectMySQLCapability) InputSchema() map[string]string {
	return map[string]string{
		"database": "string // 可选，连接后 USE 的数据库名",
	}
}

func (c *ConnectMySQLCapability) Execute(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	if c.Client == nil {
		return nil, errors.New("mysql client is not configured")
	}

	if db, ok := input["database"].(string); ok && db != "" {
		if err := c.Client.UseDatabase(ctx, db); err != nil {
			return nil, fmt.Errorf("use database: %w", err)
		}
	}

	ping, err := c.Client.Ping(ctx)
	if err != nil {
		return map[string]interface{}{
			"connected": false,
			"error":     err.Error(),
			"instance":  c.Client.ConfigSnapshot(),
		}, nil
	}

	databases, err := c.Client.ListDatabases(ctx)
	if err != nil {
		ping["databases_error"] = err.Error()
	} else {
		ping["databases"] = databases
	}
	ping["current_database"] = c.Client.ConfigSnapshot()["database"]
	return ping, nil
}

// RegisterMySQLCapabilities 注册与 MySQL 实例相关的能力。
func RegisterMySQLCapabilities(server *Server, client *mysql.Client) {
	if client == nil {
		return
	}
	server.RegisterCapability(&ConnectMySQLCapability{Client: client})
	server.RegisterCapability(&ExplainMySQLQueryCapability{Client: client})
	server.RegisterCapability(&AddMySQLIndexCapability{Client: client})
}
