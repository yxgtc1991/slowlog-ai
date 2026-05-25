package analyzer

import promptv6 "ai_slow_log/internal/prompt/slowlog"

// ToolsForPhase 按阶段渐进披露 MCP 工具（减 Prompt Token、降低误调）。
// 返回 nil 表示不过滤，使用全部 availableTools。
func ToolsForPhase(phase AgentPhase, all []promptv6.CapabilityV4) []promptv6.CapabilityV4 {
	names := toolNamesForPhase(phase)
	if names == nil {
		return all
	}
	if len(names) == 0 {
		return nil
	}
	allow := make(map[string]struct{}, len(names))
	for _, n := range names {
		allow[n] = struct{}{}
	}
	var out []promptv6.CapabilityV4
	for _, t := range all {
		meta := promptv6.DescribeCapabilityV4(t)
		if _, ok := allow[meta.Name]; ok {
			out = append(out, t)
		}
	}
	return out
}

func toolNamesForPhase(phase AgentPhase) []string {
	switch phase {
	case PhaseInit:
		return nil
	case PhaseRAGDone:
		return []string{
			"connect_mysql_instance",
			"explain_mysql_query",
			"add_mysql_index",
			"analyze_slow_log",
		}
	case PhaseDBReady:
		return []string{"explain_mysql_query", "add_mysql_index"}
	case PhaseExplained:
		return []string{"add_mysql_index"}
	case PhaseIndexPlanned, PhaseAnalyzed:
		return []string{}
	default:
		return nil
	}
}
