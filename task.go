package earthengine

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// TaskState represents the state of an async task.
type TaskState string

const (
	// TaskStatePending indicates the task is queued but not started.
	TaskStatePending TaskState = "PENDING"
	// TaskStateRunning indicates the task is currently executing.
	TaskStateRunning TaskState = "RUNNING"
	// TaskStateCompleted indicates the task completed successfully.
	TaskStateCompleted TaskState = "COMPLETED"
	// TaskStateFailed indicates the task failed with an error.
	TaskStateFailed TaskState = "FAILED"
	// TaskStateCancelled indicates the task was cancelled.
	TaskStateCancelled TaskState = "CANCELLED"
)

// Task represents an async Earth Engine operation.
type Task struct {
	ID          string
	Type        string
	Description string
	State       TaskState
	Progress    float64 // 0.0 to 1.0
	StartTime   time.Time
	UpdateTime  time.Time
	Error       string

	// Internal
	client          *Client
	operationName   string
	completionChan  chan *Task
	progressChan    chan *TaskProgress
	cancelFunc      context.CancelFunc
	mu              sync.RWMutex
}

// TaskProgress represents progress information for a task.
type TaskProgress struct {
	TaskID      string
	State       TaskState
	Progress    float64
	Description string
	UpdateTime  time.Time
}

// TaskManager manages async tasks.
type TaskManager struct {
	client *Client
	tasks  map[string]*Task
	mu     sync.RWMutex
}

// NewTaskManager creates a new task manager.
func NewTaskManager(client *Client) *TaskManager {
	return &TaskManager{
		client: client,
		tasks:  make(map[string]*Task),
	}
}

// GetTask retrieves a task by ID.
func (tm *TaskManager) GetTask(taskID string) (*Task, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	task, exists := tm.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	return task, nil
}

// ListTasks returns all tasks.
func (tm *TaskManager) ListTasks() []*Task {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tasks := make([]*Task, 0, len(tm.tasks))
	for _, task := range tm.tasks {
		tasks = append(tasks, task)
	}

	return tasks
}

// RegisterTask registers a new task.
func (tm *TaskManager) RegisterTask(task *Task) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.tasks[task.ID] = task
}

// UnregisterTask removes a task from the manager.
func (tm *TaskManager) UnregisterTask(taskID string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	delete(tm.tasks, taskID)
}

// Wait waits for the task to complete.
//
// Returns the final task state or an error if the task failed.
func (t *Task) Wait(ctx context.Context) error {
	// Poll for task status
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := t.updateStatus(ctx); err != nil {
				return err
			}

			t.mu.RLock()
			state := t.State
			t.mu.RUnlock()

			switch state {
			case TaskStateCompleted:
				return nil
			case TaskStateFailed:
				t.mu.RLock()
				errMsg := t.Error
				t.mu.RUnlock()
				return fmt.Errorf("task failed: %s", errMsg)
			case TaskStateCancelled:
				return fmt.Errorf("task cancelled")
			}
		}
	}
}

// WaitWithProgress waits for the task and reports progress.
func (t *Task) WaitWithProgress(ctx context.Context, progressFn func(progress *TaskProgress)) error {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := t.updateStatus(ctx); err != nil {
				return err
			}

			t.mu.RLock()
			state := t.State
			progress := t.Progress
			desc := t.Description
			updateTime := t.UpdateTime
			t.mu.RUnlock()

			// Report progress
			if progressFn != nil {
				progressFn(&TaskProgress{
					TaskID:      t.ID,
					State:       state,
					Progress:    progress,
					Description: desc,
					UpdateTime:  updateTime,
				})
			}

			switch state {
			case TaskStateCompleted:
				return nil
			case TaskStateFailed:
				t.mu.RLock()
				errMsg := t.Error
				t.mu.RUnlock()
				return fmt.Errorf("task failed: %s", errMsg)
			case TaskStateCancelled:
				return fmt.Errorf("task cancelled")
			}
		}
	}
}

// Cancel cancels the task.
func (t *Task) Cancel(ctx context.Context) error {
	t.mu.Lock()
	if t.cancelFunc != nil {
		t.cancelFunc()
	}
	t.State = TaskStateCancelled
	t.UpdateTime = time.Now()
	t.mu.Unlock()

	// Call Earth Engine API to cancel if operation name exists
	if t.operationName != "" {
		// Would call: t.client.Operations.Cancel(ctx, t.operationName)
		// For now, just mark as cancelled locally
	}

	return nil
}

// GetProgress returns the current task progress.
func (t *Task) GetProgress() *TaskProgress {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return &TaskProgress{
		TaskID:      t.ID,
		State:       t.State,
		Progress:    t.Progress,
		Description: t.Description,
		UpdateTime:  t.UpdateTime,
	}
}

// updateStatus updates the task status from the server.
func (t *Task) updateStatus(ctx context.Context) error {
	// In a real implementation, this would call the Earth Engine API
	// to get the current status of the operation
	//
	// For now, we'll simulate progress for demonstration purposes
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.State == TaskStatePending {
		t.State = TaskStateRunning
		t.StartTime = time.Now()
	}

	if t.State == TaskStateRunning {
		// Simulate progress
		t.Progress += 0.1
		if t.Progress >= 1.0 {
			t.Progress = 1.0
			t.State = TaskStateCompleted
		}
		t.UpdateTime = time.Now()
	}

	return nil
}

// TaskFilter filters tasks by state.
type TaskFilter struct {
	States []TaskState
}

// FilterTasks returns tasks matching the filter.
func (tm *TaskManager) FilterTasks(filter TaskFilter) []*Task {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	var filtered []*Task
	for _, task := range tm.tasks {
		task.mu.RLock()
		state := task.State
		task.mu.RUnlock()

		if len(filter.States) == 0 {
			filtered = append(filtered, task)
			continue
		}

		for _, filterState := range filter.States {
			if state == filterState {
				filtered = append(filtered, task)
				break
			}
		}
	}

	return filtered
}

// CancelAll cancels all running tasks.
func (tm *TaskManager) CancelAll(ctx context.Context) error {
	tasks := tm.FilterTasks(TaskFilter{
		States: []TaskState{TaskStatePending, TaskStateRunning},
	})

	var firstErr error
	for _, task := range tasks {
		if err := task.Cancel(ctx); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

// Cleanup removes completed, failed, and cancelled tasks older than the specified duration.
func (tm *TaskManager) Cleanup(olderThan time.Duration) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	cutoff := time.Now().Add(-olderThan)
	for id, task := range tm.tasks {
		task.mu.RLock()
		state := task.State
		updateTime := task.UpdateTime
		task.mu.RUnlock()

		if (state == TaskStateCompleted || state == TaskStateFailed || state == TaskStateCancelled) &&
			updateTime.Before(cutoff) {
			delete(tm.tasks, id)
		}
	}
}
