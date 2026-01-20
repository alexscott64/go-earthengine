package earthengine

import (
	"testing"
	"time"
)

func TestNewQuotaTracker(t *testing.T) {
	qt := NewQuotaTracker(1000)

	if qt.requestLimit != 1000 {
		t.Errorf("requestLimit = %d, want 1000", qt.requestLimit)
	}

	if qt.totalRequests != 0 {
		t.Errorf("totalRequests = %d, want 0", qt.totalRequests)
	}
}

func TestRecordRequest(t *testing.T) {
	qt := NewQuotaTracker(1000)

	qt.RecordRequest()
	qt.RecordRequest()
	qt.RecordRequest()

	if qt.GetTotalUsage() != 3 {
		t.Errorf("totalRequests = %d, want 3", qt.GetTotalUsage())
	}

	if qt.GetDailyUsage() != 3 {
		t.Errorf("dailyUsage = %d, want 3", qt.GetDailyUsage())
	}
}

func TestGetDailyUsage(t *testing.T) {
	qt := NewQuotaTracker(1000)

	if qt.GetDailyUsage() != 0 {
		t.Errorf("dailyUsage = %d, want 0", qt.GetDailyUsage())
	}

	qt.RecordRequest()
	if qt.GetDailyUsage() != 1 {
		t.Errorf("dailyUsage = %d, want 1", qt.GetDailyUsage())
	}
}

func TestIsQuotaExceeded(t *testing.T) {
	qt := NewQuotaTracker(5)

	if qt.IsQuotaExceeded() {
		t.Error("quota should not be exceeded initially")
	}

	for i := 0; i < 5; i++ {
		qt.RecordRequest()
	}

	if !qt.IsQuotaExceeded() {
		t.Error("quota should be exceeded after 5 requests")
	}
}

func TestUnlimitedQuota(t *testing.T) {
	qt := NewQuotaTracker(0) // unlimited

	for i := 0; i < 10000; i++ {
		qt.RecordRequest()
	}

	if qt.IsQuotaExceeded() {
		t.Error("unlimited quota should never be exceeded")
	}

	if qt.GetRemainingQuota() != -1 {
		t.Errorf("remainingQuota = %d, want -1 for unlimited", qt.GetRemainingQuota())
	}
}

func TestGetRemainingQuota(t *testing.T) {
	qt := NewQuotaTracker(100)

	if qt.GetRemainingQuota() != 100 {
		t.Errorf("remainingQuota = %d, want 100", qt.GetRemainingQuota())
	}

	qt.RecordRequest()
	qt.RecordRequest()

	if qt.GetRemainingQuota() != 98 {
		t.Errorf("remainingQuota = %d, want 98", qt.GetRemainingQuota())
	}
}

func TestGetUsageStats(t *testing.T) {
	qt := NewQuotaTracker(1000)

	qt.RecordRequest()
	qt.RecordRequest()

	stats := qt.GetUsageStats()

	if stats.TodayRequests != 2 {
		t.Errorf("todayRequests = %d, want 2", stats.TodayRequests)
	}

	if stats.TotalRequests != 2 {
		t.Errorf("totalRequests = %d, want 2", stats.TotalRequests)
	}

	if stats.DailyLimit != 1000 {
		t.Errorf("dailyLimit = %d, want 1000", stats.DailyLimit)
	}

	if stats.RemainingQuota != 998 {
		t.Errorf("remainingQuota = %d, want 998", stats.RemainingQuota)
	}
}

func TestCleanupOldData(t *testing.T) {
	qt := NewQuotaTracker(1000)

	// Record some requests
	qt.RecordRequest()

	// Add old data manually
	qt.mu.Lock()
	qt.dailyRequests["2020-01-01"] = 100
	qt.dailyRequests["2020-01-02"] = 200
	qt.mu.Unlock()

	// Cleanup data older than 1 day
	qt.CleanupOldData(1)

	// Old data should be removed
	qt.mu.RLock()
	_, exists := qt.dailyRequests["2020-01-01"]
	qt.mu.RUnlock()

	if exists {
		t.Error("old data should be removed")
	}

	// Today's data should still exist
	if qt.GetDailyUsage() != 1 {
		t.Error("today's data should still exist")
	}
}

func TestQuotaError(t *testing.T) {
	err := NewQuotaError(1000, 1500)

	if err.DailyLimit != 1000 {
		t.Errorf("dailyLimit = %d, want 1000", err.DailyLimit)
	}

	if err.CurrentUsage != 1500 {
		t.Errorf("currentUsage = %d, want 1500", err.CurrentUsage)
	}

	if err.Error() != "daily quota exceeded" {
		t.Errorf("error message = %s, want 'daily quota exceeded'", err.Error())
	}

	// Reset time should be tomorrow at midnight
	now := time.Now()
	tomorrow := now.AddDate(0, 0, 1)
	expectedReset := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, tomorrow.Location())

	if !err.ResetTime.Equal(expectedReset) {
		t.Errorf("resetTime = %v, want %v", err.ResetTime, expectedReset)
	}
}

func TestConcurrentAccess(t *testing.T) {
	qt := NewQuotaTracker(10000)

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				qt.RecordRequest()
				qt.GetDailyUsage()
				qt.IsQuotaExceeded()
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	if qt.GetTotalUsage() != 1000 {
		t.Errorf("totalRequests = %d, want 1000", qt.GetTotalUsage())
	}
}
