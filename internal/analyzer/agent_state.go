package analyzer

import (
	"encoding/json"
	"fmt"
	"strings"

	"ai_slow_log/internal/rag"
)

// AgentPhase V6 Agent 分析阶段（状态机），用于 Prompt 摘要与轨迹。
type AgentPhase string

const (
	PhaseInit         AgentPhase = "init"
	PhaseRAGDone      AgentPhase = "rag_done"
	PhaseDBReady      AgentPhase = "db_ready"
	PhaseExplained    AgentPhase = "explained"
	PhaseIndexPlanned AgentPhase = "index_planned"
	PhaseAnalyzed     AgentPhase = "analyzed"
	PhaseFinished     AgentPhase = "finished"
)

func (p AgentPhase) hint() string {
	switch p {
	case PhaseInit:
		return "刚开始，建议先 retrieve_rag 或阅读慢日志"
	case PhaseRAGDone:
		return "已检索知识库，可连库或 EXPLAIN"
	case PhaseDBReady:
		return "数据库可用，建议 explain_mysql_query"
	case PhaseExplained:
		return "已有执行计划，可 add_mysql_index(dry_run) 或 analyze"
	case PhaseIndexPlanned:
		return "已有索引建议，建议 analyze 归纳后 finish"
	case PhaseAnalyzed:
		return "已有中间分析，信息足够可 finish"
	case PhaseFinished:
		return "分析已结束"
	default:
		return ""
	}
}

// RAGStateEntry 单次 RAG 检索在状态中的记录。
type RAGStateEntry struct {
	Query      string   `json:"query"`
	ChunkCount int      `json:"chunk_count,omitempty"`
	Titles     []string `json:"titles,omitempty"`
	Error      string   `json:"error,omitempty"`
}

// ToolStateEntry 单次工具调用在状态中的记录。
type ToolStateEntry struct {
	OK      bool   `json:"ok"`
	Summary string `json:"summary,omitempty"`
	Error   string `json:"error,omitempty"`
}

// AgentState 类型化 Agent 上下文（替代 map[string]interface{}）。
type AgentState struct {
	Phase    AgentPhase                `json:"phase"`
	RAG      []RAGStateEntry           `json:"rag,omitempty"`
	Tools    map[string]ToolStateEntry `json:"tools,omitempty"`
	Analysis string                    `json:"analysis,omitempty"`
	Question string                    `json:"question,omitempty"`
}

func NewAgentState() *AgentState {
	return &AgentState{
		Phase: PhaseInit,
		Tools: make(map[string]ToolStateEntry),
	}
}

func (s *AgentState) RecordRAG(query string, chunks []rag.KnowledgeChunk, err error) {
	entry := RAGStateEntry{Query: query}
	if err != nil {
		entry.Error = err.Error()
	} else {
		entry.ChunkCount = len(chunks)
		for _, c := range chunks {
			if c.Title != "" {
				entry.Titles = append(entry.Titles, c.Title)
			}
		}
		if s.Phase == PhaseInit {
			s.Phase = PhaseRAGDone
		}
	}
	s.RAG = append(s.RAG, entry)
}

func (s *AgentState) RecordTool(name string, result interface{}, err error) {
	entry := ToolStateEntry{}
	if err != nil {
		entry.Error = err.Error()
	} else {
		entry.OK = true
		entry.Summary = summarizeToolResult(name, result, 320)
		s.advancePhaseForTool(name)
	}
	s.Tools[name] = entry
}

func (s *AgentState) advancePhaseForTool(name string) {
	switch name {
	case "connect_mysql_instance":
		if phaseAtMost(s.Phase, PhaseDBReady) {
			s.Phase = PhaseDBReady
		}
	case "explain_mysql_query":
		if phaseAtMost(s.Phase, PhaseExplained) {
			s.Phase = PhaseExplained
		}
	case "add_mysql_index":
		if phaseAtMost(s.Phase, PhaseIndexPlanned) {
			s.Phase = PhaseIndexPlanned
		}
	}
}

func phaseAtMost(current, max AgentPhase) bool {
	order := map[AgentPhase]int{
		PhaseInit: 0, PhaseRAGDone: 1, PhaseDBReady: 2,
		PhaseExplained: 3, PhaseIndexPlanned: 4, PhaseAnalyzed: 5, PhaseFinished: 6,
	}
	return order[current] <= order[max]
}

func (s *AgentState) RecordAnalyze(text string) {
	s.Analysis = truncateText(text, 800)
	if phaseAtMost(s.Phase, PhaseAnalyzed) {
		s.Phase = PhaseAnalyzed
	}
}

func (s *AgentState) RecordQuestion(q string) {
	s.Question = truncateText(q, 400)
}

func (s *AgentState) MarkFinished() {
	s.Phase = PhaseFinished
}

// PromptSummary 生成写入 Prompt 的紧凑摘要（控 Token，不含完整工具 JSON）。
func (s *AgentState) PromptSummary(maxField int) string {
	if maxField <= 0 {
		maxField = 400
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("阶段: %s", s.Phase))
	if h := s.Phase.hint(); h != "" {
		b.WriteString(fmt.Sprintf("（%s）", h))
	}
	b.WriteByte('\n')

	for _, r := range s.RAG {
		b.WriteString("- RAG ")
		b.WriteString(truncateText(r.Query, 80))
		if r.Error != "" {
			b.WriteString(fmt.Sprintf(": 失败 %s", truncateText(r.Error, 120)))
		} else {
			b.WriteString(fmt.Sprintf(": %d 条", r.ChunkCount))
			if len(r.Titles) > 0 {
				b.WriteString(" — ")
				b.WriteString(truncateText(strings.Join(r.Titles, "; "), maxField))
			}
		}
		b.WriteByte('\n')
	}
	for name, t := range s.Tools {
		b.WriteString("- 工具 ")
		b.WriteString(name)
		if t.Error != "" {
			b.WriteString(fmt.Sprintf(": 失败 %s", truncateText(t.Error, 120)))
		} else {
			b.WriteString(fmt.Sprintf(": OK — %s", truncateText(t.Summary, maxField)))
		}
		b.WriteByte('\n')
	}
	if s.Analysis != "" {
		b.WriteString("- 中间分析: ")
		b.WriteString(truncateText(s.Analysis, maxField))
		b.WriteByte('\n')
	}
	if s.Question != "" {
		b.WriteString("- 待澄清: ")
		b.WriteString(truncateText(s.Question, maxField))
		b.WriteByte('\n')
	}
	return strings.TrimSpace(b.String())
}

func summarizeToolResult(name string, result interface{}, maxLen int) string {
	m, ok := result.(map[string]interface{})
	if !ok {
		return truncateText(fmt.Sprint(result), maxLen)
	}
	switch name {
	case "connect_mysql_instance":
		if db, _ := m["database"].(string); db != "" {
			return "connected database=" + db
		}
		if ok, _ := m["connected"].(bool); ok {
			return "connected"
		}
	case "explain_mysql_query":
		if rows, _ := m["rows"].([]interface{}); len(rows) > 0 {
			if row0, _ := rows[0].(map[string]interface{}); row0 != nil {
				return fmt.Sprintf("plan type=%v key=%v rows=%v",
					row0["type"], row0["key"], row0["rows"])
			}
		}
	case "add_mysql_index":
		if ddl, _ := m["ddl"].(string); ddl != "" {
			return truncateText(ddl, maxLen)
		}
		if dry, _ := m["dry_run"].(bool); dry {
			return "dry_run index DDL generated"
		}
	}
	b, err := json.Marshal(result)
	if err != nil {
		return truncateText(fmt.Sprint(result), maxLen)
	}
	return truncateText(string(b), maxLen)
}

func truncateText(s string, max int) string {
	s = strings.TrimSpace(s)
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
