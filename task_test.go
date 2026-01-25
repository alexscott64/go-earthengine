package earthengine

import (
	"context"
	"testing"
	"time"
)

func TestNewTaskManager(t *testing.T) {
	client := &Client{}
	tm := NewTaskManager(client)

	if tm == nil {
		t.Fatal("NewTaskManager returned nil")
	}

	if tm.client != client {
		t.Error("TaskManager client not set correctly")
	}

	if len(tm.tasks) != 0 {
		t.Errorf("TaskManager should start with 0 tasks, got %d", len(tm.tasks))
	}
}

func TestRegisterAndGetTask(t *testing.T) {
	client := &Client{}
	tm := NewTaskManager(client)

	task := &Task{
		ID:          "test-task-1",
		Type:        "TEST",
		Description: "Test task",
		State:       TaskStatePending,
	}

	tm.RegisterTask(task)

	retrieved, err := tm.GetTask("test-task-1")
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}

	if retrieved.ID != task.ID {
		t.Errorf("Retrieved task ID = %s, want %s", retrieved.ID, task.ID)
	}
}

func TestGetTaskNotFound(t *testing.T) {
	client := &Client{}
	tm := NewTaskManager(client)

	_, err := tm.GetTask("nonexistent")
	if err == nil {
		t.Error("GetTask should return error for nonexistent task")
	}
}

func TestListTasks(t *testing.T) {
	client := &Client{}
	tm := NewTaskManager(client)

	task1 := &Task{ID: "task-1", State: TaskStatePending}
	task2 := &Task{ID: "task-2", State: TaskStateRunning}

	tm.RegisterTask(task1)
	tm.RegisterTask(task2)

	tasks := tm.ListTasks()
	if len(tasks) != 2 {
		t.Errorf("ListTasks returned %d tasks, want 2", len(tasks))
	}
}

func TestUnregisterTask(t *testing.T) {
	client := &Client{}
	tm := NewTaskManager(client)

	task := &Task{ID: "task-1"}
	tm.RegisterTask(task)

	tm.UnregisterTask("task-1")

	_, err := tm.GetTask("task-1")
	if err == nil {
		t.Error("Task should not exist after unregister")
	}
}

func TestTaskGetProgress(t *testing.T) {
	task := &Task{
		ID:          "test-task",
		State:       TaskStateRunning,
		Progress:    0.5,
		Description: "Test",
		UpdateTime:  time.Now(),
	}

	progress := task.GetProgress()

	if progress.TaskID != task.ID {
		t.Errorf("Progress TaskID = %s, want %s", progress.TaskID, task.ID)
	}

	if progress.State != task.State {
		t.Errorf("Progress State = %s, want %s", progress.State, task.State)
	}

	if progress.Progress != task.Progress {
		t.Errorf("Progress = %f, want %f", progress.Progress, task.Progress)
	}
}

func TestTaskCancel(t *testing.T) {
	ctx := context.Background()
	task := &Task{
		ID:    "test-task",
		State: TaskStateRunning,
	}

	err := task.Cancel(ctx)
	if err != nil {
		t.Errorf("Cancel failed: %v", err)
	}

	if task.State != TaskStateCancelled {
		t.Errorf("Task state = %s, want %s", task.State, TaskStateCancelled)
	}
}

func TestFilterTasks(t *testing.T) {
	client := &Client{}
	tm := NewTaskManager(client)

	tm.RegisterTask(&Task{ID: "pending-1", State: TaskStatePending})
	tm.RegisterTask(&Task{ID: "running-1", State: TaskStateRunning})
	tm.RegisterTask(&Task{ID: "completed-1", State: TaskStateCompleted})
	tm.RegisterTask(&Task{ID: "failed-1", State: TaskStateFailed})

	// Filter running tasks
	running := tm.FilterTasks(TaskFilter{
		States: []TaskState{TaskStateRunning},
	})
	if len(running) != 1 {
		t.Errorf("FilterTasks(Running) returned %d tasks, want 1", len(running))
	}

	// Filter pending and running
	active := tm.FilterTasks(TaskFilter{
		States: []TaskState{TaskStatePending, TaskStateRunning},
	})
	if len(active) != 2 {
		t.Errorf("FilterTasks(Pending+Running) returned %d tasks, want 2", len(active))
	}

	// No filter returns all
	all := tm.FilterTasks(TaskFilter{})
	if len(all) != 4 {
		t.Errorf("FilterTasks(no filter) returned %d tasks, want 4", len(all))
	}
}

func TestCancelAll(t *testing.T) {
	ctx := context.Background()
	client := &Client{}
	tm := NewTaskManager(client)

	tm.RegisterTask(&Task{ID: "pending-1", State: TaskStatePending})
	tm.RegisterTask(&Task{ID: "running-1", State: TaskStateRunning})
	tm.RegisterTask(&Task{ID: "completed-1", State: TaskStateCompleted})

	err := tm.CancelAll(ctx)
	if err != nil {
		t.Errorf("CancelAll failed: %v", err)
	}

	// Check that pending and running tasks were cancelled
	pending, _ := tm.GetTask("pending-1")
	if pending.State != TaskStateCancelled {
		t.Errorf("Pending task state = %s, want %s", pending.State, TaskStateCancelled)
	}

	running, _ := tm.GetTask("running-1")
	if running.State != TaskStateCancelled {
		t.Errorf("Running task state = %s, want %s", running.State, TaskStateCancelled)
	}

	// Completed task should not be cancelled
	completed, _ := tm.GetTask("completed-1")
	if completed.State != TaskStateCompleted {
		t.Errorf("Completed task state changed to %s", completed.State)
	}
}

func TestCleanup(t *testing.T) {
	client := &Client{}
	tm := NewTaskManager(client)

	// Add old completed task
	oldTask := &Task{
		ID:         "old-task",
		State:      TaskStateCompleted,
		UpdateTime: time.Now().Add(-2 * time.Hour),
	}
	tm.RegisterTask(oldTask)

	// Add recent completed task
	recentTask := &Task{
		ID:         "recent-task",
		State:      TaskStateCompleted,
		UpdateTime: time.Now(),
	}
	tm.RegisterTask(recentTask)

	// Add running task
	runningTask := &Task{
		ID:         "running-task",
		State:      TaskStateRunning,
		UpdateTime: time.Now().Add(-2 * time.Hour),
	}
	tm.RegisterTask(runningTask)

	// Cleanup tasks older than 1 hour
	tm.Cleanup(1 * time.Hour)

	// Old completed task should be removed
	_, err := tm.GetTask("old-task")
	if err == nil {
		t.Error("Old completed task should be removed")
	}

	// Recent completed task should remain
	_, err = tm.GetTask("recent-task")
	if err != nil {
		t.Error("Recent completed task should not be removed")
	}

	// Running task should not be removed even if old
	_, err = tm.GetTask("running-task")
	if err != nil {
		t.Error("Running task should not be removed")
	}
}

func TestTaskWait(t *testing.T) {
	ctx := context.Background()
	task := &Task{
		ID:    "test-task",
		State: TaskStateRunning,
	}

	// Manually complete the task after a delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		task.mu.Lock()
		task.State = TaskStateCompleted
		task.mu.Unlock()
	}()

	err := task.Wait(ctx)
	if err != nil {
		t.Errorf("Wait failed: %v", err)
	}

	if task.State != TaskStateCompleted {
		t.Errorf("Task state = %s, want %s", task.State, TaskStateCompleted)
	}
}

func TestTaskWaitWithProgress(t *testing.T) {
	ctx := context.Background()
	task := &Task{
		ID:          "test-task",
		State:       TaskStateRunning,
		Description: "Test task",
	}

	progressUpdates := 0

	// Complete task after delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		task.mu.Lock()
		task.State = TaskStateCompleted
		task.Progress = 1.0
		task.mu.Unlock()
	}()

	err := task.WaitWithProgress(ctx, func(progress *TaskProgress) {
		progressUpdates++
		if progress.TaskID != task.ID {
			t.Errorf("Progress TaskID = %s, want %s", progress.TaskID, task.ID)
		}
	})

	if err != nil {
		t.Errorf("WaitWithProgress failed: %v", err)
	}

	if progressUpdates == 0 {
		t.Error("No progress updates received")
	}
}

func TestTaskWaitCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	task := &Task{
		ID:    "test-task",
		State: TaskStateRunning,
	}

	// Cancel context immediately
	cancel()

	err := task.Wait(ctx)
	if err != context.Canceled {
		t.Errorf("Wait error = %v, want %v", err, context.Canceled)
	}
}

func TestTaskUpdateStatus(t *testing.T) {
	ctx := context.Background()
	task := &Task{
		ID:       "test-task",
		State:    TaskStatePending,
		Progress: 0.0,
	}

	// First update: should transition to running
	err := task.updateStatus(ctx)
	if err != nil {
		t.Errorf("updateStatus failed: %v", err)
	}

	if task.State != TaskStateRunning {
		t.Errorf("Task state = %s, want %s", task.State, TaskStateRunning)
	}

	// Multiple updates: should increase progress
	for i := 0; i < 12; i++ {
		task.updateStatus(ctx)
	}

	if task.Progress < 1.0 {
		t.Errorf("Task progress = %f, want >= 1.0", task.Progress)
	}

	if task.State != TaskStateCompleted {
		t.Errorf("Task state = %s, want %s after progress complete", task.State, TaskStateCompleted)
	}
}
