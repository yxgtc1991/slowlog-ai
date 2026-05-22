package analyzer

import (
	"strings"
	"testing"

	"ai_slow_log/internal/rag"
)

func TestAgentState_phaseAndSummary(t *testing.T) {
	t.Parallel()
	s := NewAgentState()
	s.RecordRAG("全表扫描", []rag.KnowledgeChunk{{Title: "Rows_examined"}}, nil)
	if s.Phase != PhaseRAGDone {
		t.Fatalf("phase=%s want rag_done", s.Phase)
	}
	s.RecordTool("connect_mysql_instance", map[string]interface{}{"database": "test"}, nil)
	if s.Phase != PhaseDBReady {
		t.Fatalf("phase=%s want db_ready", s.Phase)
	}
	sum := s.PromptSummary(200)
	if !strings.Contains(sum, "db_ready") || !strings.Contains(sum, "connect_mysql") {
		t.Fatalf("summary=%q", sum)
	}
	if strings.Contains(sum, `"database"`) {
		t.Fatal("summary should not contain full JSON dump")
	}
}

func TestAgentState_explainAdvancesPhase(t *testing.T) {
	t.Parallel()
	s := NewAgentState()
	s.RecordTool("explain_mysql_query", map[string]interface{}{
		"rows": []interface{}{
			map[string]interface{}{"type": "ALL", "key": nil, "rows": 48000},
		},
	}, nil)
	if s.Phase != PhaseExplained {
		t.Fatalf("phase=%s want explained", s.Phase)
	}
	if sum := s.PromptSummary(100); !strings.Contains(sum, "ALL") {
		t.Fatalf("summary=%q", sum)
	}
}
