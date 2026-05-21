// 校验仓库内 Markdown 相对路径与标题锚点（对齐 GoLand / GitHub 标题 slug）。
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

var (
	linkRe    = regexp.MustCompile(`\[[^\]]*\]\(([^)]+)\)`)
	customID  = regexp.MustCompile(`\s+\{#([a-zA-Z0-9_-]+)\}\s*$`)
	headingRe = regexp.MustCompile(`^(#{1,6})\s+(.+?)\s*$`)
)

func main() {
	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cwd: %v\n", err)
		os.Exit(1)
	}
	// 从仓库根运行：go run ./scripts/validate-md-links
	if _, err := os.Stat(filepath.Join(root, "README.md")); err != nil {
		root = filepath.Join(root, "..", "..")
	}

	var mdFiles []string
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".md") && !strings.Contains(path, "/vendor/") {
			mdFiles = append(mdFiles, path)
		}
		return nil
	})

	headings := make(map[string]map[string]bool) // file -> anchor -> true
	for _, f := range mdFiles {
		ids, err := extractAnchors(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "parse %s: %v\n", f, err)
			os.Exit(1)
		}
		rel, _ := filepath.Rel(root, f)
		headings[rel] = ids
	}

	var failed int
	for _, src := range mdFiles {
		relSrc, _ := filepath.Rel(root, src)
		links, err := extractLinks(src)
		if err != nil {
			fmt.Fprintf(os.Stderr, "links %s: %v\n", relSrc, err)
			os.Exit(1)
		}
		for _, lk := range links {
			if strings.HasPrefix(lk, "http://") || strings.HasPrefix(lk, "https://") {
				continue
			}
			target, anchor := lk, ""
			if i := strings.Index(lk, "#"); i >= 0 {
				target = lk[:i]
				anchor = lk[i+1:]
			}
			if target == "" {
				// 同文件锚点
				if anchor == "" {
					continue
				}
				if !headings[relSrc][anchor] {
					fmt.Printf("FAIL %s: #%s not found in %s\n", relSrc, anchor, relSrc)
					failed++
				}
				continue
			}
			resolved := filepath.Clean(filepath.Join(filepath.Dir(src), target))
			relTarget, _ := filepath.Rel(root, resolved)
			if _, ok := headings[relTarget]; !ok {
				if _, err := os.Stat(resolved); err != nil {
					fmt.Printf("FAIL %s: file not found %s (from %s)\n", relSrc, relTarget, lk)
					failed++
				}
				continue
			}
			if anchor != "" && !headings[relTarget][anchor] {
				fmt.Printf("FAIL %s: #%s not found in %s (available: %s)\n",
					relSrc, anchor, relTarget, joinKeys(headings[relTarget]))
				failed++
			}
		}
	}

	if failed > 0 {
		fmt.Printf("\n%d broken link(s)\n", failed)
		os.Exit(1)
	}
	fmt.Printf("OK: validated links in %d markdown file(s)\n", len(mdFiles))
}

func extractLinks(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, m := range linkRe.FindAllStringSubmatch(string(data), -1) {
		out = append(out, m[1])
	}
	return out, nil
}

func extractAnchors(path string) (map[string]bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	ids := make(map[string]bool)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		m := headingRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		title := m[2]
		if cid := customID.FindStringSubmatch(title); len(cid) == 2 {
			ids[cid[1]] = true
			title = customID.ReplaceAllString(title, "")
		}
		ids[githubSlug(title)] = true
	}
	return ids, sc.Err()
}

// githubSlug 近似 GoLand / GitHub 标题锚点（ASCII 小写 + 连字符；保留中文）。
func githubSlug(title string) string {
	title = strings.TrimSpace(title)
	var b strings.Builder
	prevDash := false
	for _, r := range title {
		if r == '{' || r == '#' {
			break
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if r < 128 {
				b.WriteRune(unicode.ToLower(r))
			} else {
				b.WriteRune(r)
			}
			prevDash = false
			continue
		}
		if r == '-' || r == '_' {
			if !prevDash && b.Len() > 0 {
				b.WriteByte('-')
				prevDash = true
			}
			continue
		}
		if unicode.IsSpace(r) {
			if !prevDash && b.Len() > 0 {
				b.WriteByte('-')
				prevDash = true
			}
			continue
		}
		// 跳过标点；中文等字母数字已处理
		if unicode.Is(unicode.Han, r) {
			b.WriteRune(r)
			prevDash = false
		}
	}
	return strings.Trim(b.String(), "-")
}

func joinKeys(m map[string]bool) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return strings.Join(keys, ", ")
}
