package ops

import (
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	ErrRateLimited   = errors.New("rate limit exceeded")
	ErrQuotaExceeded = errors.New("daily analyze quota exceeded")
	ErrTooManyConcurrent = errors.New("too many concurrent analyzes")
)

// Limits 全局限流、并发与配额（进程内 PoC）。
type Limits struct {
	mu              sync.Mutex
	perMinute       int
	minuteWindow    time.Time
	minuteCount     int
	dailyQuota      int
	dailyDate       string
	dailyCount      int
	maxConcurrent   int
	sem             chan struct{}
	OnRateLimited   func()
}

// NewLimitsFromEnv 解析环境变量。
func NewLimitsFromEnv() *Limits {
	l := &Limits{
		perMinute:     envInt("SLOWLOG_RATE_LIMIT_PER_MIN", 0),
		dailyQuota:    envInt("SLOWLOG_DAILY_ANALYZE_QUOTA", 0),
		maxConcurrent: envInt("SLOWLOG_MAX_CONCURRENT", 0),
	}
	if l.maxConcurrent > 0 {
		l.sem = make(chan struct{}, l.maxConcurrent)
	}
	return l
}

func envInt(key string, def int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return def
	}
	return n
}

// Status 当前限流状态（可对外展示）。
type LimitStatus struct {
	RateLimitPerMin   int    `json:"rate_limit_per_min"`
	RateUsedThisMin   int    `json:"rate_used_this_min"`
	DailyQuota        int    `json:"daily_quota"`
	DailyUsed         int    `json:"daily_used"`
	DailyDate         string `json:"daily_date"`
	MaxConcurrent     int    `json:"max_concurrent"`
	ConcurrentInUse   int    `json:"concurrent_in_use"`
}

// Enabled 是否配置了任一限流项。
func (l *Limits) Enabled() bool {
	if l == nil {
		return false
	}
	return l.perMinute > 0 || l.dailyQuota > 0 || l.maxConcurrent > 0
}

func (l *Limits) Status() LimitStatus {
	l.mu.Lock()
	defer l.mu.Unlock()
	st := LimitStatus{
		RateLimitPerMin: l.perMinute,
		RateUsedThisMin: l.minuteCount,
		DailyQuota:      l.dailyQuota,
		DailyUsed:       l.dailyCount,
		DailyDate:       l.dailyDate,
		MaxConcurrent:   l.maxConcurrent,
	}
	if l.sem != nil {
		st.ConcurrentInUse = len(l.sem)
	}
	return st
}

// Acquire 分析前占用配额；分析结束须 Release。
func (l *Limits) Acquire() error {
	l.mu.Lock()
	now := time.Now()
	if l.perMinute > 0 {
		if now.Sub(l.minuteWindow) >= time.Minute {
			l.minuteWindow = now
			l.minuteCount = 0
		}
		if l.minuteCount >= l.perMinute {
			l.mu.Unlock()
			return ErrRateLimited
		}
		l.minuteCount++
	}
	if l.dailyQuota > 0 {
		day := now.Format("2006-01-02")
		if l.dailyDate != day {
			l.dailyDate = day
			l.dailyCount = 0
		}
		if l.dailyCount >= l.dailyQuota {
			l.mu.Unlock()
			return ErrQuotaExceeded
		}
		l.dailyCount++
	}
	l.mu.Unlock()

	if l.sem == nil {
		return nil
	}
	select {
	case l.sem <- struct{}{}:
		return nil
	default:
		return ErrTooManyConcurrent
	}
}

func (l *Limits) Release() {
	if l.sem == nil {
		return
	}
	select {
	case <-l.sem:
	default:
	}
}

// Middleware 对昂贵路由限流（health/status 除外由调用方决定）。
func (l *Limits) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if l == nil || !isExpensivePath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		if err := l.Acquire(); err != nil {
			if l.OnRateLimited != nil {
				l.OnRateLimited()
			}
			status := http.StatusTooManyRequests
			if errors.Is(err, ErrTooManyConcurrent) {
				status = http.StatusServiceUnavailable
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			_, _ = w.Write([]byte(`{"error":"` + err.Error() + `"}`))
			return
		}
		defer l.Release()
		next.ServeHTTP(w, r)
	})
}

func isExpensivePath(path string) bool {
	switch {
	case path == "/v1/analyze", path == "/v1/ingest":
		return true
	case strings.HasPrefix(path, "/v1/rag/rebuild"):
		return true
	default:
		return false
	}
}
