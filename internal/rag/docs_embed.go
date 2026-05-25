package rag

import "embed"

//go:embed slowlog/docs/**/*.md
var slowlogDocsFS embed.FS
