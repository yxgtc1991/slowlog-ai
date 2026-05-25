package service

import (
	"ai_slow_log/internal/llm"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// JobStatus 异步 ingest 任务状态。
type JobStatus string

const (
	JobPending   JobStatus = "pending"
	JobRunning   JobStatus = "running"
	JobCompleted JobStatus = "completed"
	JobFailed    JobStatus = "failed"
	JobSkipped   JobStatus = "skipped"
)

// Job 一次 ingest 异步任务。
type Job struct {
	ID          string    `json:"id"`
	Status      JobStatus `json:"status"`
	Source      string    `json:"source,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
	ReportID    string    `json:"report_id,omitempty"`
	FinalResult string    `json:"final_result,omitempty"`
	Error       string    `json:"error,omitempty"`
	SkipReason  string    `json:"skip_reason,omitempty"`
}

// JobStore 内存任务表（PoC；重启丢失）。
type JobStore struct {
	mu   sync.RWMutex
	jobs map[string]*Job
}

func NewJobStore() *JobStore {
	return &JobStore{jobs: make(map[string]*Job)}
}

func (s *JobStore) Create(id, source string) *Job {
	s.mu.Lock()
	defer s.mu.Unlock()
	j := &Job{
		ID:        id,
		Status:    JobPending,
		Source:    source,
		CreatedAt: time.Now(),
	}
	s.jobs[id] = j
	return j
}

func (s *JobStore) Get(id string) (*Job, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	j, ok := s.jobs[id]
	return j, ok
}

func (s *JobStore) update(id string, fn func(*Job)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if j, ok := s.jobs[id]; ok {
		fn(j)
	}
}

// RunIngestAsync 后台执行分析并可选 callback。
func RunIngestAsync(
	parent context.Context,
	store *JobStore,
	jobID string,
	client *llm.DeepSeekClient,
	cfg RunV6Config,
	callbackURL string,
) {
	go func() {
		store.update(jobID, func(j *Job) { j.Status = JobRunning })

		ctx, cancel := context.WithTimeout(parent, cfg.timeoutDuration())
		defer cancel()

		result, err := RunV6(ctx, client, cfg)
		if err != nil {
			store.update(jobID, func(j *Job) {
				j.Status = JobFailed
				j.Error = err.Error()
				j.CompletedAt = time.Now()
			})
			postCallback(callbackURL, jobID, store)
			return
		}

		store.update(jobID, func(j *Job) {
			j.Status = JobCompleted
			j.ReportID = result.ReportID
			j.FinalResult = truncate(result.FinalResult, 2000)
			j.CompletedAt = time.Now()
		})
		postCallback(callbackURL, jobID, store)
	}()
}

func postCallback(url, jobID string, store *JobStore) {
	if url == "" {
		return
	}
	j, ok := store.Get(jobID)
	if !ok {
		return
	}
	body, _ := json.Marshal(j)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	_ = resp.Body.Close()
}

func truncate(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "…"
}

// NewJobID 生成任务 ID。
func NewJobID() string {
	return "job-" + time.Now().Format("20060102-150405.000")
}
