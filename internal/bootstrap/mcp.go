package bootstrap

import (
	"ai_slow_log/internal/analyzer"
	"ai_slow_log/internal/config"
	"ai_slow_log/internal/llm"
	"ai_slow_log/internal/mcp"
	"ai_slow_log/internal/mysql"
	prompt "ai_slow_log/internal/prompt/slowlog"
	"ai_slow_log/internal/rag"
)

// MCP 启动 MCP Server、注册能力与 MySQL（可选）。
type MCP struct {
	Server *mcp.Server
	Caps   []prompt.CapabilityV4
	close  func()
}

// Close 释放 MySQL 等资源。
func (m *MCP) Close() {
	if m != nil && m.close != nil {
		m.close()
	}
}

// SetupMCP 与 make run / agent-run 共用：V4 分析器 + MySQL 工具注册。
func SetupMCP(llmClient *llm.DeepSeekClient, logf func(string, ...any)) (*MCP, error) {
	if logf == nil {
		logf = func(string, ...any) {}
	}
	server := mcp.NewServer()
	server.RegisterCapability(&mcp.AnalyzeSlowLogCapability{
		Analyzer: analyzer.NewAnalyzer(
			llmClient,
			analyzer.WithPromptBuilder(&prompt.RagV3Prompt{}),
			analyzer.WithRAGRetriever(analyzer.NewRAGRetrieverAdapter(rag.MustDefaultRetriever())),
		),
	})

	var closeFn func()
	if mysqlCfg, err := config.MustLoadMySQL(); err != nil {
		logf("mysql: %v (skip connect_mysql_instance)", err)
	} else {
		client, err := mysql.NewClient(mysqlCfg)
		if err != nil {
			logf("mysql: connect failed: %v", err)
		} else {
			closeFn = func() { _ = client.Close() }
			mcp.RegisterMySQLCapabilities(server, client)
			logf("MySQL: connected to %s:%d as %s", mysqlCfg.Host, mysqlCfg.Port, mysqlCfg.User)
		}
	}

	return &MCP{
		Server: server,
		Caps:   server.GetCapabilitiesAsV4(),
		close:  closeFn,
	}, nil
}
