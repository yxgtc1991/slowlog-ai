package ops

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AuditEvent 操作审计（追加 JSONL）。
type AuditEvent struct {
	Time       time.Time `json:"time"`
	Action     string    `json:"action"`
	InstanceID string    `json:"instance_id,omitempty"`
	RequestID  string    `json:"request_id,omitempty"`
	ReportID   string    `json:"report_id,omitempty"`
	ClientIP   string    `json:"client_ip,omitempty"`
	Actor      string    `json:"actor,omitempty"`
	Detail     string    `json:"detail,omitempty"`
	Status     string    `json:"status,omitempty"`
}

// Auditor 追加写审计日志。
type Auditor struct {
	path string
	mu   sync.Mutex
}

func NewAuditor(path string) *Auditor {
	if path == "" {
		path = "data/audit.jsonl"
	}
	return &Auditor{path: path}
}

func (a *Auditor) Log(ev AuditEvent) error {
	if a == nil {
		return nil
	}
	if ev.Time.IsZero() {
		ev.Time = time.Now()
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := os.MkdirAll(filepath.Dir(a.path), 0o755); err != nil && filepath.Dir(a.path) != "." {
		return err
	}
	f, err := os.OpenFile(a.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	b, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	_, err = f.Write(append(b, '\n'))
	return err
}
