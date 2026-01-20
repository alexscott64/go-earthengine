package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// Checkpoint represents a saved state of batch execution.
type Checkpoint struct {
	TotalQueries     int              `json:"total_queries"`
	CompletedIndices []int            `json:"completed_indices"`
	Results          []CheckpointResult `json:"results"`
	FilePath         string           `json:"file_path"`
}

// CheckpointResult represents a saved result.
type CheckpointResult struct {
	Index int         `json:"index"`
	Value interface{} `json:"value"`
	Error string      `json:"error,omitempty"`
}

// CheckpointManager manages checkpoints for batch operations.
type CheckpointManager struct {
	mu           sync.Mutex
	checkpoint   *Checkpoint
	saveInterval int // Save after every N completions
}

// NewCheckpointManager creates a new checkpoint manager.
func NewCheckpointManager(filepath string, saveInterval int) *CheckpointManager {
	if saveInterval <= 0 {
		saveInterval = 10 // Default: save every 10 completions
	}

	return &CheckpointManager{
		checkpoint: &Checkpoint{
			FilePath:         filepath,
			CompletedIndices: make([]int, 0),
			Results:          make([]CheckpointResult, 0),
		},
		saveInterval: saveInterval,
	}
}

// LoadCheckpoint loads a checkpoint from disk.
func LoadCheckpoint(filepath string) (*Checkpoint, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No checkpoint exists
		}
		return nil, fmt.Errorf("failed to read checkpoint: %w", err)
	}

	var checkpoint Checkpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return nil, fmt.Errorf("failed to parse checkpoint: %w", err)
	}

	checkpoint.FilePath = filepath
	return &checkpoint, nil
}

// SaveCheckpoint saves the checkpoint to disk.
func (cm *CheckpointManager) SaveCheckpoint() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	data, err := json.MarshalIndent(cm.checkpoint, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal checkpoint: %w", err)
	}

	if err := os.WriteFile(cm.checkpoint.FilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write checkpoint: %w", err)
	}

	return nil
}

// RecordResult records a completed query result.
func (cm *CheckpointManager) RecordResult(index int, value interface{}, err error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	result := CheckpointResult{
		Index: index,
		Value: value,
	}
	if err != nil {
		result.Error = err.Error()
	}

	cm.checkpoint.CompletedIndices = append(cm.checkpoint.CompletedIndices, index)
	cm.checkpoint.Results = append(cm.checkpoint.Results, result)
}

// IsCompleted checks if a query index has already been completed.
func (cm *CheckpointManager) IsCompleted(index int) bool {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for _, idx := range cm.checkpoint.CompletedIndices {
		if idx == index {
			return true
		}
	}
	return false
}

// GetCompletedCount returns the number of completed queries.
func (cm *CheckpointManager) GetCompletedCount() int {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	return len(cm.checkpoint.CompletedIndices)
}

// ExecuteWithCheckpoint executes a batch with checkpoint support.
//
// If a checkpoint exists, it will resume from where it left off.
// Progress is automatically saved to the checkpoint file.
//
// Example:
//
//	batch := helpers.NewBatch(client, 10)
//	for _, loc := range locations {
//	    batch.Add(helpers.NewTreeCoverageQuery(loc.Lat, loc.Lon))
//	}
//
//	// Execute with checkpoint - will resume if interrupted
//	results, err := helpers.ExecuteWithCheckpoint(ctx, batch, "progress.json", 10)
func ExecuteWithCheckpoint(ctx context.Context, batch *Batch, checkpointPath string, saveInterval int) ([]Result, error) {
	// Try to load existing checkpoint
	existingCheckpoint, err := LoadCheckpoint(checkpointPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load checkpoint: %w", err)
	}

	// Create checkpoint manager
	cm := NewCheckpointManager(checkpointPath, saveInterval)
	if existingCheckpoint != nil {
		cm.checkpoint = existingCheckpoint
	}

	cm.checkpoint.TotalQueries = batch.Size()

	// Initialize results array
	results := make([]Result, batch.Size())

	// Restore completed results from checkpoint
	if existingCheckpoint != nil {
		for _, cr := range existingCheckpoint.Results {
			results[cr.Index] = Result{
				Index: cr.Index,
				Value: cr.Value,
			}
			if cr.Error != "" {
				results[cr.Index].Error = fmt.Errorf("%s", cr.Error)
			}
		}
	}

	// Track completion for auto-save
	completedSinceLastSave := 0
	var mu sync.Mutex

	// Create a semaphore to limit concurrency
	sem := make(chan struct{}, batch.concurrency)
	var wg sync.WaitGroup

	// Execute queries that haven't been completed
	for i, query := range batch.queries {
		// Skip if already completed
		if cm.IsCompleted(i) {
			continue
		}

		wg.Add(1)
		go func(index int, q Query) {
			defer wg.Done()

			// Acquire semaphore
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				results[index] = Result{
					Index: index,
					Error: ctx.Err(),
				}
				return
			}

			// Execute query
			value, err := q.Execute(ctx, batch.client)
			results[index] = Result{
				Value: value,
				Error: err,
				Index: index,
			}

			// Record result in checkpoint
			cm.RecordResult(index, value, err)

			// Auto-save checkpoint
			mu.Lock()
			completedSinceLastSave++
			shouldSave := completedSinceLastSave >= cm.saveInterval
			if shouldSave {
				completedSinceLastSave = 0
			}
			mu.Unlock()

			if shouldSave {
				if saveErr := cm.SaveCheckpoint(); saveErr != nil {
					// Log error but don't fail the batch
					fmt.Fprintf(os.Stderr, "Warning: failed to save checkpoint: %v\n", saveErr)
				}
			}
		}(i, query)
	}

	// Wait for all queries to complete
	wg.Wait()

	// Final save
	if err := cm.SaveCheckpoint(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to save final checkpoint: %v\n", err)
	}

	// Check if context was canceled
	if ctx.Err() != nil {
		return results, ctx.Err()
	}

	return results, nil
}

// RemoveCheckpoint deletes a checkpoint file.
//
// Call this after successful completion to clean up.
//
// Example:
//
//	results, err := helpers.ExecuteWithCheckpoint(ctx, batch, "progress.json", 10)
//	if err == nil {
//	    helpers.RemoveCheckpoint("progress.json")
//	}
func RemoveCheckpoint(checkpointPath string) error {
	if err := os.Remove(checkpointPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove checkpoint: %w", err)
	}
	return nil
}

// GetCheckpointProgress returns the progress from a checkpoint file.
//
// Returns (completed, total, error).
//
// Example:
//
//	completed, total, err := helpers.GetCheckpointProgress("progress.json")
//	if err == nil {
//	    fmt.Printf("Progress: %d/%d (%.1f%%)\n",
//	        completed, total, float64(completed)*100/float64(total))
//	}
func GetCheckpointProgress(checkpointPath string) (int, int, error) {
	checkpoint, err := LoadCheckpoint(checkpointPath)
	if err != nil {
		return 0, 0, err
	}
	if checkpoint == nil {
		return 0, 0, nil
	}

	return len(checkpoint.CompletedIndices), checkpoint.TotalQueries, nil
}
