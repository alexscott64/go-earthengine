package earthengine

import (
	"sync"
	"time"
)

// QuotaTracker tracks API usage for quota management.
type QuotaTracker struct {
	mu            sync.RWMutex
	dailyRequests map[string]int // date -> request count
	totalRequests int64
	startTime     time.Time
	requestLimit  int // daily request limit (0 = unlimited)
}

// NewQuotaTracker creates a new quota tracker.
func NewQuotaTracker(dailyLimit int) *QuotaTracker {
	return &QuotaTracker{
		dailyRequests: make(map[string]int),
		startTime:     time.Now(),
		requestLimit:  dailyLimit,
	}
}

// RecordRequest records an API request for quota tracking.
func (qt *QuotaTracker) RecordRequest() {
	qt.mu.Lock()
	defer qt.mu.Unlock()

	today := time.Now().Format("2006-01-02")
	qt.dailyRequests[today]++
	qt.totalRequests++
}

// GetDailyUsage returns the number of requests made today.
func (qt *QuotaTracker) GetDailyUsage() int {
	qt.mu.RLock()
	defer qt.mu.RUnlock()

	today := time.Now().Format("2006-01-02")
	return qt.dailyRequests[today]
}

// GetTotalUsage returns the total number of requests since tracker creation.
func (qt *QuotaTracker) GetTotalUsage() int64 {
	qt.mu.RLock()
	defer qt.mu.RUnlock()

	return qt.totalRequests
}

// IsQuotaExceeded checks if daily quota has been exceeded.
func (qt *QuotaTracker) IsQuotaExceeded() bool {
	if qt.requestLimit == 0 {
		return false // unlimited
	}

	qt.mu.RLock()
	defer qt.mu.RUnlock()

	today := time.Now().Format("2006-01-02")
	return qt.dailyRequests[today] >= qt.requestLimit
}

// GetRemainingQuota returns the number of remaining requests for today.
func (qt *QuotaTracker) GetRemainingQuota() int {
	if qt.requestLimit == 0 {
		return -1 // unlimited
	}

	qt.mu.RLock()
	defer qt.mu.RUnlock()

	today := time.Now().Format("2006-01-02")
	used := qt.dailyRequests[today]
	remaining := qt.requestLimit - used
	if remaining < 0 {
		return 0
	}
	return remaining
}

// GetUsageStats returns usage statistics.
func (qt *QuotaTracker) GetUsageStats() QuotaStats {
	qt.mu.RLock()
	defer qt.mu.RUnlock()

	today := time.Now().Format("2006-01-02")
	uptime := time.Since(qt.startTime)

	return QuotaStats{
		TodayRequests:   qt.dailyRequests[today],
		TotalRequests:   qt.totalRequests,
		DailyLimit:      qt.requestLimit,
		RemainingQuota:  qt.GetRemainingQuota(),
		Uptime:          uptime,
		RequestsPerHour: float64(qt.totalRequests) / uptime.Hours(),
	}
}

// CleanupOldData removes usage data older than the specified number of days.
func (qt *QuotaTracker) CleanupOldData(daysToKeep int) {
	qt.mu.Lock()
	defer qt.mu.Unlock()

	cutoff := time.Now().AddDate(0, 0, -daysToKeep)
	cutoffStr := cutoff.Format("2006-01-02")

	for date := range qt.dailyRequests {
		if date < cutoffStr {
			delete(qt.dailyRequests, date)
		}
	}
}

// QuotaStats represents quota usage statistics.
type QuotaStats struct {
	TodayRequests   int           // Requests made today
	TotalRequests   int64         // Total requests since start
	DailyLimit      int           // Daily request limit (0 = unlimited)
	RemainingQuota  int           // Remaining requests for today (-1 = unlimited)
	Uptime          time.Duration // Time since tracker creation
	RequestsPerHour float64       // Average requests per hour
}

// QuotaError is returned when quota is exceeded.
type QuotaError struct {
	Message        string
	DailyLimit     int
	CurrentUsage   int
	ResetTime      time.Time
}

func (e *QuotaError) Error() string {
	return e.Message
}

// NewQuotaError creates a new quota error.
func NewQuotaError(limit, usage int) *QuotaError {
	tomorrow := time.Now().AddDate(0, 0, 1)
	resetTime := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, tomorrow.Location())

	return &QuotaError{
		Message:      "daily quota exceeded",
		DailyLimit:   limit,
		CurrentUsage: usage,
		ResetTime:    resetTime,
	}
}
