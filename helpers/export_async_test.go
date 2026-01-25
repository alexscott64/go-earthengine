package helpers

import (
	"context"
	"testing"
	"time"

	"github.com/alexscott64/go-earthengine"
)

func TestExportImageAsync(t *testing.T) {
	ctx := context.Background()
	client := &earthengine.Client{}
	image := &earthengine.Image{}

	task, err := ExportImageAsync(ctx, client, image,
		ExportDescription("Test Export"),
		ExportToGCS("test-bucket", "exports/"),
		ExportScale(30))

	if err != nil {
		t.Fatalf("ExportImageAsync failed: %v", err)
	}

	if task == nil {
		t.Fatal("Task is nil")
	}

	if task.Type != "IMAGE_EXPORT" {
		t.Errorf("Task type = %s, want IMAGE_EXPORT", task.Type)
	}

	if task.Description != "Test Export" {
		t.Errorf("Task description = %s, want Test Export", task.Description)
	}

	if task.State != earthengine.TaskStatePending {
		t.Errorf("Task state = %s, want %s", task.State, earthengine.TaskStatePending)
	}
}

func TestExportImageAsyncValidation(t *testing.T) {
	ctx := context.Background()
	client := &earthengine.Client{}
	image := &earthengine.Image{}

	// Missing bucket for GCS export
	_, err := ExportImageAsync(ctx, client, image,
		ExportDescription("Test"),
		ExportToGCS("", ""))

	if err == nil {
		t.Error("Expected validation error for missing bucket")
	}

	// Invalid scale
	_, err = ExportImageAsync(ctx, client, image,
		ExportDescription("Test"),
		ExportToGCS("bucket", "prefix"),
		ExportScale(-1))

	if err == nil {
		t.Error("Expected validation error for negative scale")
	}

	// Empty description should use default
	task, err := ExportImageAsync(ctx, client, image,
		ExportToGCS("bucket", "prefix"))

	if err != nil {
		t.Errorf("Export with default description should succeed: %v", err)
	}

	if task.Description != "Export" {
		t.Errorf("Default description = %s, want Export", task.Description)
	}
}

func TestExportTableAsync(t *testing.T) {
	ctx := context.Background()
	client := &earthengine.Client{}

	task, err := ExportTableAsync(ctx, client, nil,
		ExportDescription("Table Export"),
		ExportToGCS("test-bucket", "tables/"),
		ExportFileFormat(CSV))

	if err != nil {
		t.Fatalf("ExportTableAsync failed: %v", err)
	}

	if task.Type != "TABLE_EXPORT" {
		t.Errorf("Task type = %s, want TABLE_EXPORT", task.Type)
	}
}

func TestExportVideoAsync(t *testing.T) {
	ctx := context.Background()
	client := &earthengine.Client{}
	collection := &earthengine.ImageCollection{}

	task, err := ExportVideoAsync(ctx, client, collection,
		ExportDescription("Video Export"),
		ExportToGCS("test-bucket", "videos/"),
		ExportFileFormat(MP4))

	if err != nil {
		t.Fatalf("ExportVideoAsync failed: %v", err)
	}

	if task.Type != "VIDEO_EXPORT" {
		t.Errorf("Task type = %s, want VIDEO_EXPORT", task.Type)
	}
}

func TestExportToDrive(t *testing.T) {
	ctx := context.Background()
	client := &earthengine.Client{}
	image := &earthengine.Image{}

	task, err := ExportImageAsync(ctx, client, image,
		ExportDescription("Drive Export"),
		ExportToGoogleDrive("My Folder"),
		ExportScale(30))

	if err != nil {
		t.Fatalf("ExportImageAsync to Drive failed: %v", err)
	}

	if task == nil {
		t.Fatal("Task is nil")
	}
}

func TestExportToAsset(t *testing.T) {
	ctx := context.Background()
	client := &earthengine.Client{}
	image := &earthengine.Image{}

	task, err := ExportImageAsync(ctx, client, image,
		ExportDescription("Asset Export"),
		ExportToEEAsset("projects/my-project/assets/my-image"),
		ExportScale(30))

	if err != nil {
		t.Fatalf("ExportImageAsync to Asset failed: %v", err)
	}

	if task == nil {
		t.Fatal("Task is nil")
	}
}

func TestExportImageWithNotification(t *testing.T) {
	ctx := context.Background()
	client := &earthengine.Client{}
	image := &earthengine.Image{}

	notified := false
	var notifiedTask *earthengine.Task
	var notifiedErr error

	task, err := ExportImageWithNotification(ctx, client, image,
		func(t *earthengine.Task, e error) {
			notified = true
			notifiedTask = t
			notifiedErr = e
		},
		ExportDescription("Test Export"),
		ExportToGCS("bucket", "prefix"))

	if err != nil {
		t.Fatalf("ExportImageWithNotification failed: %v", err)
	}

	// Manually complete the task to speed up the test
	go func() {
		time.Sleep(100 * time.Millisecond)
		task.State = earthengine.TaskStateCompleted
		task.Progress = 1.0
	}()

	// Wait for notification goroutine to complete
	time.Sleep(3 * time.Second)

	if !notified {
		t.Error("Notification function was not called")
	}

	if notifiedTask == nil {
		t.Error("Notified task is nil")
	}

	if notifiedTask != nil && notifiedTask.ID != task.ID {
		t.Errorf("Notified task ID = %s, want %s", notifiedTask.ID, task.ID)
	}

	// Error should be nil for successful completion
	if notifiedErr != nil {
		t.Errorf("Notified error = %v, want nil", notifiedErr)
	}
}

func TestExportTableWithNotification(t *testing.T) {
	ctx := context.Background()
	client := &earthengine.Client{}

	notified := false

	task, err := ExportTableWithNotification(ctx, client, nil,
		func(t *earthengine.Task, e error) {
			notified = true
		},
		ExportDescription("Table Export"),
		ExportToGCS("bucket", "prefix"))

	if err != nil {
		t.Fatalf("ExportTableWithNotification failed: %v", err)
	}

	// Complete the task quickly
	go func() {
		time.Sleep(100 * time.Millisecond)
		task.State = earthengine.TaskStateCompleted
		task.Progress = 1.0
	}()

	time.Sleep(3 * time.Second)

	if !notified {
		t.Error("Notification function was not called")
	}

	_ = task
}

func TestExportVideoWithNotification(t *testing.T) {
	ctx := context.Background()
	client := &earthengine.Client{}
	collection := &earthengine.ImageCollection{}

	notified := false

	task, err := ExportVideoWithNotification(ctx, client, collection,
		func(t *earthengine.Task, e error) {
			notified = true
		},
		ExportDescription("Video Export"),
		ExportToGCS("bucket", "prefix"))

	if err != nil {
		t.Fatalf("ExportVideoWithNotification failed: %v", err)
	}

	// Complete the task quickly
	go func() {
		time.Sleep(100 * time.Millisecond)
		task.State = earthengine.TaskStateCompleted
		task.Progress = 1.0
	}()

	time.Sleep(3 * time.Second)

	if !notified {
		t.Error("Notification function was not called")
	}

	_ = task
}

func TestWaitForExports(t *testing.T) {
	ctx := context.Background()

	tasks := []*earthengine.Task{
		{ID: "task-1", State: earthengine.TaskStatePending},
		{ID: "task-2", State: earthengine.TaskStatePending},
		{ID: "task-3", State: earthengine.TaskStatePending},
	}

	// Mark tasks as completed after a delay
	for _, task := range tasks {
		go func(t *earthengine.Task) {
			time.Sleep(50 * time.Millisecond)
			t.State = earthengine.TaskStateCompleted
		}(task)
	}

	progressUpdates := 0
	results := WaitForExports(ctx, tasks, func(completed, total int) {
		progressUpdates++
		if completed > total {
			t.Errorf("Completed (%d) > Total (%d)", completed, total)
		}
	})

	if len(results) != len(tasks) {
		t.Errorf("Results length = %d, want %d", len(results), len(tasks))
	}

	for i, result := range results {
		if result.TaskID != tasks[i].ID {
			t.Errorf("Result %d TaskID = %s, want %s", i, result.TaskID, tasks[i].ID)
		}
	}

	if progressUpdates == 0 {
		t.Error("No progress updates received")
	}
}

func TestWaitForExportsWithFailure(t *testing.T) {
	ctx := context.Background()

	tasks := []*earthengine.Task{
		{ID: "task-1", State: earthengine.TaskStatePending},
		{ID: "task-2", State: earthengine.TaskStatePending},
	}

	// First task completes successfully
	go func() {
		time.Sleep(50 * time.Millisecond)
		tasks[0].State = earthengine.TaskStateCompleted
	}()

	// Second task fails
	go func() {
		time.Sleep(50 * time.Millisecond)
		tasks[1].State = earthengine.TaskStateFailed
		tasks[1].Error = "Export failed"
	}()

	results := WaitForExports(ctx, tasks, nil)

	// First task should have no error
	if results[0].Error != nil {
		t.Errorf("Task 1 error = %v, want nil", results[0].Error)
	}

	// Second task should have error
	if results[1].Error == nil {
		t.Error("Task 2 should have error")
	}
}

func TestExportWithAllOptions(t *testing.T) {
	ctx := context.Background()
	client := &earthengine.Client{}
	image := &earthengine.Image{}

	task, err := ExportImageAsync(ctx, client, image,
		ExportDescription("Full Options Export"),
		ExportToGCS("my-bucket", "exports/"),
		ExportScale(10),
		ExportCRS("EPSG:3857"),
		ExportMaxPixels(1e10),
		ExportFileFormat(GeoTIFF))

	if err != nil {
		t.Fatalf("Export with all options failed: %v", err)
	}

	if task == nil {
		t.Fatal("Task is nil")
	}

	if task.Description != "Full Options Export" {
		t.Errorf("Task description = %s, want Full Options Export", task.Description)
	}
}

func TestExportResult(t *testing.T) {
	task := &earthengine.Task{
		ID:    "test-task",
		State: earthengine.TaskStateCompleted,
	}

	result := ExportResult{
		TaskID: task.ID,
		Task:   task,
		Error:  nil,
	}

	if result.TaskID != "test-task" {
		t.Errorf("Result TaskID = %s, want test-task", result.TaskID)
	}

	if result.Task != task {
		t.Error("Result Task doesn't match")
	}

	if result.Error != nil {
		t.Errorf("Result Error = %v, want nil", result.Error)
	}
}

func TestConcurrentExports(t *testing.T) {
	ctx := context.Background()
	client := &earthengine.Client{}

	// Start multiple exports concurrently
	var tasks []*earthengine.Task
	for i := 0; i < 10; i++ {
		image := &earthengine.Image{}
		task, err := ExportImageAsync(ctx, client, image,
			ExportDescription("Concurrent Export"),
			ExportToGCS("bucket", "prefix"))

		if err != nil {
			t.Fatalf("Export %d failed: %v", i, err)
		}

		tasks = append(tasks, task)
	}

	if len(tasks) != 10 {
		t.Errorf("Created %d tasks, want 10", len(tasks))
	}

	// Verify all tasks are unique
	ids := make(map[string]bool)
	for _, task := range tasks {
		if ids[task.ID] {
			t.Errorf("Duplicate task ID: %s", task.ID)
		}
		ids[task.ID] = true
	}
}
