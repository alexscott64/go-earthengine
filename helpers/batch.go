package helpers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/yourusername/go-earthengine"
)

// Batch represents a batch of Earth Engine queries that can be executed in parallel.
//
// Example:
//
//	batch := helpers.NewBatch(client, 10) // 10 concurrent queries
//	for _, loc := range locations {
//	    batch.Add(helpers.NewTreeCoverageQuery(loc.Lat, loc.Lon))
//	}
//	results, err := batch.Execute(ctx)
type Batch struct {
	client      *earthengine.Client
	queries     []Query
	concurrency int
}

// Result represents the result of a single query in a batch.
type Result struct {
	Value interface{} // The result value (type depends on query)
	Error error       // Any error that occurred
	Index int         // Index of the query in the batch
}

// BatchOption configures a Batch.
type BatchOption func(*Batch)

// WithConcurrency sets the number of concurrent queries to execute.
func WithConcurrency(n int) BatchOption {
	return func(b *Batch) {
		b.concurrency = n
	}
}

// NewBatch creates a new batch executor.
//
// The concurrency parameter controls how many queries run in parallel.
// A good default is 10-20 to balance speed and API rate limits.
//
// Example:
//
//	batch := helpers.NewBatch(client, 10)
func NewBatch(client *earthengine.Client, concurrency int) *Batch {
	if concurrency <= 0 {
		concurrency = 10 // Default concurrency
	}
	return &Batch{
		client:      client,
		queries:     make([]Query, 0),
		concurrency: concurrency,
	}
}

// Add adds a query to the batch.
//
// Example:
//
//	batch.Add(helpers.NewTreeCoverageQuery(45.5152, -122.6784))
//	batch.Add(helpers.NewElevationQuery(47.6062, -122.3321))
func (b *Batch) Add(q Query) *Batch {
	b.queries = append(b.queries, q)
	return b
}

// Size returns the number of queries in the batch.
func (b *Batch) Size() int {
	return len(b.queries)
}

// Execute runs all queries in the batch with controlled concurrency.
//
// Results are returned in the same order as queries were added.
// If a query fails, its Result will have a non-nil Error field.
//
// Example:
//
//	results, err := batch.Execute(ctx)
//	if err != nil {
//	    // Fatal error (e.g., context canceled)
//	    return err
//	}
//	for i, result := range results {
//	    if result.Error != nil {
//	        fmt.Printf("Query %d failed: %v\n", i, result.Error)
//	        continue
//	    }
//	    fmt.Printf("Query %d result: %v\n", i, result.Value)
//	}
func (b *Batch) Execute(ctx context.Context) ([]Result, error) {
	if len(b.queries) == 0 {
		return []Result{}, nil
	}

	results := make([]Result, len(b.queries))

	// Create a semaphore to limit concurrency
	sem := make(chan struct{}, b.concurrency)

	// Create a wait group to wait for all queries
	var wg sync.WaitGroup

	// Execute each query
	for i, query := range b.queries {
		wg.Add(1)
		go func(index int, q Query) {
			defer wg.Done()

			// Acquire semaphore
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }() // Release semaphore
			case <-ctx.Done():
				results[index] = Result{
					Index: index,
					Error: ctx.Err(),
				}
				return
			}

			// Execute query
			value, err := q.Execute(ctx, b.client)
			results[index] = Result{
				Value: value,
				Error: err,
				Index: index,
			}
		}(i, query)
	}

	// Wait for all queries to complete
	wg.Wait()

	// Check if context was canceled
	if ctx.Err() != nil {
		return results, ctx.Err()
	}

	return results, nil
}

// ExecuteWithProgress runs all queries and reports progress via a callback.
//
// The progress callback is called periodically with the number of completed queries
// and the total number of queries.
//
// Example:
//
//	results, err := batch.ExecuteWithProgress(ctx, func(completed, total int) {
//	    fmt.Printf("Progress: %d/%d (%.1f%%)\n", completed, total, float64(completed)*100/float64(total))
//	})
func (b *Batch) ExecuteWithProgress(ctx context.Context, progressFn func(completed, total int)) ([]Result, error) {
	if len(b.queries) == 0 {
		return []Result{}, nil
	}

	results := make([]Result, len(b.queries))
	completed := 0
	var mu sync.Mutex

	// Create a semaphore to limit concurrency
	sem := make(chan struct{}, b.concurrency)

	// Create a wait group to wait for all queries
	var wg sync.WaitGroup

	// Report initial progress
	if progressFn != nil {
		progressFn(0, len(b.queries))
	}

	// Execute each query
	for i, query := range b.queries {
		wg.Add(1)
		go func(index int, q Query) {
			defer wg.Done()

			// Acquire semaphore
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }() // Release semaphore
			case <-ctx.Done():
				results[index] = Result{
					Index: index,
					Error: ctx.Err(),
				}
				return
			}

			// Execute query
			value, err := q.Execute(ctx, b.client)
			results[index] = Result{
				Value: value,
				Error: err,
				Index: index,
			}

			// Update progress
			mu.Lock()
			completed++
			current := completed
			mu.Unlock()

			if progressFn != nil {
				progressFn(current, len(b.queries))
			}
		}(i, query)
	}

	// Wait for all queries to complete
	wg.Wait()

	// Check if context was canceled
	if ctx.Err() != nil {
		return results, ctx.Err()
	}

	return results, nil
}

// ExecuteWithRetry executes queries with automatic retry on failure.
//
// Failed queries will be retried up to maxRetries times with exponential backoff.
//
// Example:
//
//	results, err := batch.ExecuteWithRetry(ctx, 3, 100*time.Millisecond)
func (b *Batch) ExecuteWithRetry(ctx context.Context, maxRetries int, initialBackoff time.Duration) ([]Result, error) {
	if len(b.queries) == 0 {
		return []Result{}, nil
	}

	results := make([]Result, len(b.queries))

	// Create a semaphore to limit concurrency
	sem := make(chan struct{}, b.concurrency)

	// Create a wait group to wait for all queries
	var wg sync.WaitGroup

	// Execute each query with retry
	for i, query := range b.queries {
		wg.Add(1)
		go func(index int, q Query) {
			defer wg.Done()

			// Acquire semaphore
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }() // Release semaphore
			case <-ctx.Done():
				results[index] = Result{
					Index: index,
					Error: ctx.Err(),
				}
				return
			}

			// Try executing with retry
			var value interface{}
			var err error
			backoff := initialBackoff

			for attempt := 0; attempt <= maxRetries; attempt++ {
				value, err = q.Execute(ctx, b.client)
				if err == nil {
					break // Success
				}

				// If this wasn't the last attempt, sleep before retrying
				if attempt < maxRetries {
					select {
					case <-time.After(backoff):
						backoff *= 2 // Exponential backoff
					case <-ctx.Done():
						results[index] = Result{
							Index: index,
							Error: ctx.Err(),
						}
						return
					}
				}
			}

			results[index] = Result{
				Value: value,
				Error: err,
				Index: index,
			}
		}(i, query)
	}

	// Wait for all queries to complete
	wg.Wait()

	// Check if context was canceled
	if ctx.Err() != nil {
		return results, ctx.Err()
	}

	return results, nil
}

// Summary returns statistics about the batch execution results.
type Summary struct {
	Total     int     // Total number of queries
	Succeeded int     // Number of successful queries
	Failed    int     // Number of failed queries
	SuccessRate float64 // Success rate (0-1)
}

// Summarize returns statistics about batch execution results.
//
// Example:
//
//	results, _ := batch.Execute(ctx)
//	summary := helpers.Summarize(results)
//	fmt.Printf("Success rate: %.1f%% (%d/%d)\n",
//	    summary.SuccessRate*100, summary.Succeeded, summary.Total)
func Summarize(results []Result) Summary {
	total := len(results)
	succeeded := 0

	for _, r := range results {
		if r.Error == nil {
			succeeded++
		}
	}

	var successRate float64
	if total > 0 {
		successRate = float64(succeeded) / float64(total)
	}

	return Summary{
		Total:       total,
		Succeeded:   succeeded,
		Failed:      total - succeeded,
		SuccessRate: successRate,
	}
}

// FilterSuccessful returns only the successful results.
//
// Example:
//
//	results, _ := batch.Execute(ctx)
//	successful := helpers.FilterSuccessful(results)
//	for _, result := range successful {
//	    fmt.Printf("Value: %v\n", result.Value)
//	}
func FilterSuccessful(results []Result) []Result {
	filtered := make([]Result, 0, len(results))
	for _, r := range results {
		if r.Error == nil {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// FilterFailed returns only the failed results.
//
// Example:
//
//	results, _ := batch.Execute(ctx)
//	failed := helpers.FilterFailed(results)
//	for _, result := range failed {
//	    fmt.Printf("Query %d failed: %v\n", result.Index, result.Error)
//	}
func FilterFailed(results []Result) []Result {
	filtered := make([]Result, 0)
	for _, r := range results {
		if r.Error != nil {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// RateLimiter helps control the rate of API requests.
type RateLimiter struct {
	requests  chan struct{}
	interval  time.Duration
	ticker    *time.Ticker
	closeChan chan struct{}
}

// NewRateLimiter creates a new rate limiter.
//
// requestsPerSecond controls how many requests are allowed per second.
//
// Example:
//
//	limiter := helpers.NewRateLimiter(10) // 10 requests per second
//	defer limiter.Close()
//
//	for _, loc := range locations {
//	    limiter.Wait(ctx)
//	    result, err := helpers.TreeCoverage(client, loc.Lat, loc.Lon)
//	}
func NewRateLimiter(requestsPerSecond float64) *RateLimiter {
	if requestsPerSecond <= 0 {
		requestsPerSecond = 10 // Default
	}

	interval := time.Duration(float64(time.Second) / requestsPerSecond)
	limiter := &RateLimiter{
		requests:  make(chan struct{}, 1),
		interval:  interval,
		ticker:    time.NewTicker(interval),
		closeChan: make(chan struct{}),
	}

	// Start the ticker goroutine
	go func() {
		for {
			select {
			case <-limiter.ticker.C:
				select {
				case limiter.requests <- struct{}{}:
				default:
				}
			case <-limiter.closeChan:
				return
			}
		}
	}()

	return limiter
}

// Wait blocks until a request token is available.
func (rl *RateLimiter) Wait(ctx context.Context) error {
	select {
	case <-rl.requests:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Close stops the rate limiter.
func (rl *RateLimiter) Close() {
	close(rl.closeChan)
	rl.ticker.Stop()
}

// Example usage helper function
func exampleBatchUsage() error {
	// This is just for documentation - won't compile without a real client
	return fmt.Errorf("example only")

	// client, _ := earthengine.NewClient(...)
	//
	// // Create locations to query
	// locations := []struct{ Lat, Lon float64 }{
	//     {45.5152, -122.6784}, // Portland
	//     {47.6062, -122.3321}, // Seattle
	//     {37.7749, -122.4194}, // San Francisco
	// }
	//
	// // Create batch
	// batch := NewBatch(client, 10)
	// for _, loc := range locations {
	//     batch.Add(NewTreeCoverageQuery(loc.Lat, loc.Lon))
	// }
	//
	// // Execute with progress
	// ctx := context.Background()
	// results, err := batch.ExecuteWithProgress(ctx, func(completed, total int) {
	//     fmt.Printf("Progress: %d/%d\n", completed, total)
	// })
	//
	// // Summarize results
	// summary := Summarize(results)
	// fmt.Printf("Success rate: %.1f%%\n", summary.SuccessRate*100)
}
