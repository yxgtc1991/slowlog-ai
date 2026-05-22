package eval

import "encoding/json"

func decisionJSON(state string, na nextActionPayload) string {
	b, _ := json.Marshal(map[string]interface{}{
		"current_state": state,
		"next_action":   na,
	})
	return string(b)
}

type nextActionPayload map[string]interface{}

func retrieveRAG(query string) nextActionPayload {
	return nextActionPayload{
		"type":       "retrieve_rag",
		"reasoning":  "需要专家知识",
		"rag_query":  query,
	}
}

func callTool(name string, args map[string]interface{}) nextActionPayload {
	if args == nil {
		args = map[string]interface{}{}
	}
	return nextActionPayload{
		"type":       "call_tool",
		"reasoning":  "调用 MCP 工具",
		"tool_name":  name,
		"tool_args":  args,
	}
}

func analyze(text string) nextActionPayload {
	return nextActionPayload{
		"type":      "analyze",
		"reasoning": "综合已有信息",
		"analysis":  text,
	}
}

func finish(result string) nextActionPayload {
	return nextActionPayload{
		"type":       "finish",
		"reasoning":  "信息足够",
		"result":     result,
	}
}
