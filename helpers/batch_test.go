package helpers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yourusername/go-earthengine"
)

// mockQuery is a query that returns a predefined value or error
type mockQuery struct {
	value interface{}
	err   error
	delay time.Duration
}

func (m *mockQuery) Execute(ctx context.Context, client *earthengine.Client) (interface{}, error) {
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return m.value, m.err
}

func TestNewBatch(t *testing.T) {
	batch := NewBatch(nil, 10)

	if batch == nil {
		t.Fatal("NewBatch returned nil")
	}
	if batch.concurrency != 10 {
		t.Errorf("concurrency = %d, want 10", batch.concurrency)
	}
	if len(batch.queries) != 0 {
		t.Errorf("queries length = %d, want 0", len(batch.queries))
	}
}

func TestNewBatchDefaultConcurrency(t *testing.T) {
	batch := NewBatch(nil, 0) // Invalid concurrency

	if batch.concurrency != 10 {
		t.Errorf("default concurrency = %d, want 10", batch.concurrency)
	}
}

func TestBatchAdd(t *testing.T) {
	batch := NewBatch(nil, 10)
	query := &mockQuery{value: 42}

	batch.Add(query)

	if batch.Size() != 1 {
		t.Errorf("Size() = %d, want 1", batch.Size())
	}
}

func TestBatchAddChaining(t *testing.T) {
	batch := NewBatch(nil, 10)

	batch.Add(&mockQuery{value: 1}).
		Add(&mockQuery{value: 2}).
		Add(&mockQuery{value: 3})

	if batch.Size() != 3 {
		t.Errorf("Size() = %d, want 3", batch.Size())
	}
}

func TestBatchExecuteEmpty(t *testing.T) {
	batch := NewBatch(nil, 10)
	ctx := context.Background()

	results, err := batch.Execute(ctx)

	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}
	if len(results) != 0 {
		t.Errorf("results length = %d, want 0", len(results))
	}
}

func TestBatchExecuteSuccess(t *testing.T) {
	batch := NewBatch(nil, 10)
	batch.Add(&mockQuery{value: 42})
	batch.Add(&mockQuery{value: "hello"})
	batch.Add(&mockQuery{value: 3.14})

	ctx := context.Background()
	results, err := batch.Execute(ctx)

	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}
	if len(results) != 3 {
		t.Fatalf("results length = %d, want 3", len(results))
	}

	// Check results are in order
	if results[0].Value != 42 {
		t.Errorf("results[0].Value = %v, want 42", results[0].Value)
	}
	if results[1].Value != "hello" {
		t.Errorf("results[1].Value = %v, want hello", results[1].Value)
	}
	if results[2].Value != 3.14 {
		t.Errorf("results[2].Value = %v, want 3.14", results[2].Value)
	}

	// Check no errors
	for i, r := range results {
		if r.Error != nil {
			t.Errorf("results[%d].Error = %v, want nil", i, r.Error)
		}
	}
}

func TestBatchExecuteWithErrors(t *testing.T) {
	batch := NewBatch(nil, 10)
	batch.Add(&mockQuery{value: 42})
	batch.Add(&mockQuery{err: errors.New("query failed")})
	batch.Add(&mockQuery{value: 3.14})

	ctx := context.Background()
	results, err := batch.Execute(ctx)

	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}
	if len(results) != 3 {
		t.Fatalf("results length = %d, want 3", len(results))
	}

	// Check first query succeeded
	if results[0].Error != nil {
		t.Errorf("results[0].Error = %v, want nil", results[0].Error)
	}

	// Check second query failed
	if results[1].Error == nil {
		t.Error("results[1].Error = nil, want error")
	}

	// Check third query succeeded
	if results[2].Error != nil {
		t.Errorf("results[2].Error = %v, want nil", results[2].Error)
	}
}

func TestBatchExecuteCancellation(t *testing.T) {
	batch := NewBatch(nil, 1)
	batch.Add(&mockQuery{value: 1, delay: 100 * time.Millisecond})
	batch.Add(&mockQuery{value: 2, delay: 100 * time.Millisecond})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	results, err := batch.Execute(ctx)

	if err != context.Canceled {
		t.Errorf("Execute() error = %v, want context.Canceled", err)
	}

	// At least some queries should have context.Canceled error
	hasContextError := false
	for _, r := range results {
		if r.Error == context.Canceled {
			hasContextError = true
			break
		}
	}
	if !hasContextError {
		t.Error("Expected at least one query to have context.Canceled error")
	}
}

func TestBatchExecuteWithProgressEmpty(t *testing.T) {
	batch := NewBatch(nil, 10)
	ctx := context.Background()

	progressCalled := false
	results, err := batch.ExecuteWithProgress(ctx, func(completed, total int) {
		progressCalled = true
	})

	if err != nil {
		t.Errorf("ExecuteWithProgress() error = %v, want nil", err)
	}
	if len(results) != 0 {
		t.Errorf("results length = %d, want 0", len(results))
	}
	if progressCalled {
		t.Error("progress callback should not be called for empty batch")
	}
}

func TestBatchExecuteWithProgress(t *testing.T) {
	batch := NewBatch(nil, 10)
	batch.Add(&mockQuery{value: 1})
	batch.Add(&mockQuery{value: 2})
	batch.Add(&mockQuery{value: 3})

	ctx := context.Background()
	progressCalls := 0
	var lastCompleted, lastTotal int

	results, err := batch.ExecuteWithProgress(ctx, func(completed, total int) {
		progressCalls++
		lastCompleted = completed
		lastTotal = total
	})

	if err != nil {
		t.Errorf("ExecuteWithProgress() error = %v, want nil", err)
	}
	if len(results) != 3 {
		t.Errorf("results length = %d, want 3", len(results))
	}
	if progressCalls == 0 {
		t.Error("progress callback was never called")
	}
	if lastTotal != 3 {
		t.Errorf("lastTotal = %d, want 3", lastTotal)
	}
	if lastCompleted != 3 {
		t.Errorf("lastCompleted = %d, want 3", lastCompleted)
	}
}

func TestSummarize(t *testing.T) {
	results := []Result{
		{Value: 1, Error: nil},
		{Value: 2, Error: errors.New("failed")},
		{Value: 3, Error: nil},
		{Value: 4, Error: errors.New("failed")},
		{Value: 5, Error: nil},
	}

	summary := Summarize(results)

	if summary.Total != 5 {
		t.Errorf("Total = %d, want 5", summary.Total)
	}
	if summary.Succeeded != 3 {
		t.Errorf("Succeeded = %d, want 3", summary.Succeeded)
	}
	if summary.Failed != 2 {
		t.Errorf("Failed = %d, want 2", summary.Failed)
	}
	if summary.SuccessRate != 0.6 {
		t.Errorf("SuccessRate = %f, want 0.6", summary.SuccessRate)
	}
}

func TestSummarizeEmpty(t *testing.T) {
	results := []Result{}
	summary := Summarize(results)

	if summary.Total != 0 {
		t.Errorf("Total = %d, want 0", summary.Total)
	}
	if summary.SuccessRate != 0 {
		t.Errorf("SuccessRate = %f, want 0", summary.SuccessRate)
	}
}

func TestFilterSuccessful(t *testing.T) {
	results := []Result{
		{Value: 1, Error: nil},
		{Value: 2, Error: errors.New("failed")},
		{Value: 3, Error: nil},
	}

	successful := FilterSuccessful(results)

	if len(successful) != 2 {
		t.Errorf("FilterSuccessful() length = %d, want 2", len(successful))
	}
	if successful[0].Value != 1 {
		t.Errorf("successful[0].Value = %v, want 1", successful[0].Value)
	}
	if successful[1].Value != 3 {
		t.Errorf("successful[1].Value = %v, want 3", successful[1].Value)
	}
}

func TestFilterFailed(t *testing.T) {
	results := []Result{
		{Value: 1, Error: nil, Index: 0},
		{Value: 2, Error: errors.New("failed"), Index: 1},
		{Value: 3, Error: nil, Index: 2},
		{Value: 4, Error: errors.New("failed"), Index: 3},
	}

	failed := FilterFailed(results)

	if len(failed) != 2 {
		t.Errorf("FilterFailed() length = %d, want 2", len(failed))
	}
	if failed[0].Index != 1 {
		t.Errorf("failed[0].Index = %d, want 1", failed[0].Index)
	}
	if failed[1].Index != 3 {
		t.Errorf("failed[1].Index = %d, want 3", failed[1].Index)
	}
}

func TestNewRateLimiter(t *testing.T) {
	limiter := NewRateLimiter(10)
	defer limiter.Close()

	if limiter == nil {
		t.Fatal("NewRateLimiter returned nil")
	}
}

func TestRateLimiterWait(t *testing.T) {
	limiter := NewRateLimiter(100) // 100 requests per second
	defer limiter.Close()

	ctx := context.Background()
	start := time.Now()

	// First request should be immediate
	err := limiter.Wait(ctx)
	if err != nil {
		t.Errorf("Wait() error = %v, want nil", err)
	}

	elapsed := time.Since(start)
	if elapsed > 50*time.Millisecond {
		t.Errorf("First Wait() took %v, expected < 50ms", elapsed)
	}
}

func TestRateLimiterCancel(t *testing.T) {
	limiter := NewRateLimiter(1) // Very slow rate
	defer limiter.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := limiter.Wait(ctx)
	if err != context.Canceled {
		t.Errorf("Wait() error = %v, want context.Canceled", err)
	}
}

func ExampleBatch_Execute() {
	// Example showing basic batch execution
	// client, _ := earthengine.NewClient(...)
	//
	// batch := NewBatch(client, 10)
	// batch.Add(NewTreeCoverageQuery(45.5152, -122.6784))
	// batch.Add(NewElevationQuery(47.6062, -122.3321))
	//
	// ctx := context.Background()
	// results, err := batch.Execute(ctx)
	// if err != nil {
	//     fmt.Printf("Error: %v\n", err)
	//     return
	// }
	//
	// for i, result := range results {
	//     if result.Error != nil {
	//         fmt.Printf("Query %d failed: %v\n", i, result.Error)
	//         continue
	//     }
	//     fmt.Printf("Query %d result: %v\n", i, result.Value)
	// }
}

func ExampleBatch_ExecuteWithProgress() {
	// Example showing batch execution with progress reporting
	// client, _ := earthengine.NewClient(...)
	//
	// batch := NewBatch(client, 10)
	// for i := 0; i < 100; i++ {
	//     batch.Add(NewTreeCoverageQuery(45.0+float64(i)*0.01, -122.0))
	// }
	//
	// ctx := context.Background()
	// results, err := batch.ExecuteWithProgress(ctx, func(completed, total int) {
	//     fmt.Printf("Progress: %d/%d (%.1f%%)\n",
	//         completed, total, float64(completed)*100/float64(total))
	// })
}

func ExampleSummarize() {
	// Example showing result summarization
	// results, _ := batch.Execute(ctx)
	// summary := Summarize(results)
	// fmt.Printf("Success rate: %.1f%% (%d/%d)\n",
	//     summary.SuccessRate*100, summary.Succeeded, summary.Total)
}
