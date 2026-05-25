package rag

import (
	"strings"
	"testing"
)

func TestSplitMarkdownBySections(t *testing.T) {
	t.Parallel()
	chunks := splitMarkdownBySections("Demo Title", "## 症状\n行1\n\n## 原因\n行2", "pattern")
	if len(chunks) != 2 {
		t.Fatalf("len=%d", len(chunks))
	}
	if !strings.Contains(chunks[0].Title, "症状") || chunks[0].Source != "pattern" {
		t.Fatalf("first=%+v", chunks[0])
	}
	if !strings.Contains(chunks[1].Title, "原因") {
		t.Fatalf("second=%+v", chunks[1])
	}
}

func TestSplitMarkdownBySections_noHeading(t *testing.T) {
	t.Parallel()
	chunks := splitMarkdownBySections("Only", "plain body", "metric")
	if len(chunks) != 1 || !strings.Contains(chunks[0].Content, "plain") {
		t.Fatalf("chunks=%+v", chunks)
	}
}
