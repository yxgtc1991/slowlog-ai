package rag

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

// loadKnowledgeChunks 从嵌入的 slowlog/docs 加载知识块（按 ## 切分）。
func loadKnowledgeChunks() ([]KnowledgeChunk, error) {
	var out []KnowledgeChunk
	err := fs.WalkDir(slowlogDocsFS, "slowlog/docs", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		b, err := fs.ReadFile(slowlogDocsFS, path)
		if err != nil {
			return err
		}
		docTitle, body := parseMarkdownDoc(string(b))
		source := sourceFromPath(path)
		chunks := splitMarkdownBySections(docTitle, body, source)
		out = append(out, chunks...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no knowledge chunks found")
	}
	return out, nil
}

// splitMarkdownBySections 按二级标题 ## 切分；无 ## 时整篇为一 chunk。
func splitMarkdownBySections(docTitle, body, source string) []KnowledgeChunk {
	body = strings.TrimSpace(body)
	if body == "" {
		return []KnowledgeChunk{{
			Title:   docTitle,
			Content: "",
			Source:  source,
		}}
	}

	lines := strings.Split(body, "\n")
	var chunks []KnowledgeChunk
	var sectionTitle string
	var sectionLines []string
	var preamble []string

	flushSection := func() {
		content := strings.TrimSpace(strings.Join(sectionLines, "\n"))
		if sectionTitle == "" && content == "" {
			return
		}
		title := docTitle
		if sectionTitle != "" {
			title = docTitle + " · " + sectionTitle
		}
		chunks = append(chunks, KnowledgeChunk{
			Title:   title,
			Content: truncateContent(content, 600),
			Source:  source,
		})
		sectionTitle = ""
		sectionLines = nil
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") {
			if sectionTitle == "" && len(sectionLines) == 0 && len(preamble) > 0 {
				chunks = append(chunks, KnowledgeChunk{
					Title:   docTitle,
					Content: truncateContent(strings.Join(preamble, "\n"), 600),
					Source:  source,
				})
				preamble = nil
			} else {
				flushSection()
			}
			sectionTitle = strings.TrimSpace(strings.TrimPrefix(trimmed, "## "))
			sectionLines = nil
			continue
		}
		if sectionTitle == "" && len(preamble) == 0 && len(chunks) == 0 {
			preamble = append(preamble, line)
		} else if sectionTitle == "" {
			preamble = append(preamble, line)
		} else {
			sectionLines = append(sectionLines, line)
		}
	}

	if sectionTitle != "" || len(sectionLines) > 0 {
		flushSection()
	} else if len(preamble) > 0 {
		chunks = append(chunks, KnowledgeChunk{
			Title:   docTitle,
			Content: truncateContent(strings.Join(preamble, "\n"), 600),
			Source:  source,
		})
	}

	if len(chunks) == 0 {
		return []KnowledgeChunk{{
			Title:   docTitle,
			Content: truncateContent(body, 600),
			Source:  source,
		}}
	}
	return chunks
}

func parseMarkdownDoc(raw string) (title, body string) {
	lines := strings.Split(raw, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			title = strings.TrimPrefix(line, "# ")
			body = strings.TrimSpace(strings.Join(lines[i+1:], "\n"))
			return title, body
		}
	}
	return "", strings.TrimSpace(raw)
}

func sourceFromPath(path string) string {
	dir := filepath.Dir(path)
	base := filepath.Base(dir)
	switch base {
	case "patterns":
		return "pattern"
	case "anti-patterns":
		return "anti-pattern"
	case "metrics":
		return "metric"
	case "actions":
		return "action"
	case "boundaries":
		return "boundary"
	default:
		return base
	}
}

func truncateContent(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
