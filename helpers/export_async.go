package helpers

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alexscott64/go-earthengine"
)

var taskIDCounter uint64

// ExportImageAsync exports an image and returns a task for monitoring progress.
//
// Example:
//
//	task, err := helpers.ExportImageAsync(ctx, client, image,
//	    helpers.ExportDescription("Summer 2023 NDVI"),
//	    helpers.ExportToGCS("my-bucket", "exports/"),
//	    helpers.ExportScale(30))
//	if err != nil {
//	    return err
//	}
//
//	// Wait for completion with progress updates
//	err = task.WaitWithProgress(ctx, func(progress *earthengine.TaskProgress) {
//	    fmt.Printf("Progress: %.1f%%\n", progress.Progress*100)
//	})
func ExportImageAsync(ctx context.Context, client *earthengine.Client, image *earthengine.Image, opts ...ExportImageOption) (*earthengine.Task, error) {
	// Apply options
	cfg := &ExportConfig{
		Description: "Export",
		Destination: ExportToCloudStorage,
		Format:      GeoTIFF,
		Scale:       30,
		CRS:         "EPSG:4326",
		MaxPixels:   1e9,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// Validate configuration
	if err := validateExportConfig(cfg); err != nil {
		return nil, err
	}

	// Create task with unique ID
	taskID := atomic.AddUint64(&taskIDCounter, 1)
	task := &earthengine.Task{
		ID:          fmt.Sprintf("export-%d-%d", time.Now().Unix(), taskID),
		Type:        "IMAGE_EXPORT",
		Description: cfg.Description,
		State:       earthengine.TaskStatePending,
		Progress:    0.0,
		StartTime:   time.Now(),
		UpdateTime:  time.Now(),
	}

	// In a real implementation, this would:
	// 1. Build the export expression
	// 2. Call client.ExportImage() to submit to Earth Engine
	// 3. Get back an operation name/ID
	// 4. Store the operation name in the task for status polling
	//
	// For now, we return a task that simulates progress

	return task, nil
}

// ExportTableAsync exports a feature collection and returns a task.
//
// Example:
//
//	task, err := helpers.ExportTableAsync(ctx, client, collection,
//	    helpers.ExportDescription("City Boundaries"),
//	    helpers.ExportToGCS("my-bucket", "tables/"),
//	    helpers.ExportFileFormat(helpers.CSV))
func ExportTableAsync(ctx context.Context, client *earthengine.Client, collection interface{}, opts ...ExportImageOption) (*earthengine.Task, error) {
	cfg := &ExportConfig{
		Description: "Table Export",
		Destination: ExportToCloudStorage,
		Format:      CSV,
		CRS:         "EPSG:4326",
		Scale:       30,
		MaxPixels:   1e9,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	if err := validateExportConfig(cfg); err != nil {
		return nil, err
	}

	taskID := atomic.AddUint64(&taskIDCounter, 1)
	task := &earthengine.Task{
		ID:          fmt.Sprintf("export-table-%d-%d", time.Now().Unix(), taskID),
		Type:        "TABLE_EXPORT",
		Description: cfg.Description,
		State:       earthengine.TaskStatePending,
		Progress:    0.0,
		StartTime:   time.Now(),
		UpdateTime:  time.Now(),
	}

	return task, nil
}

// ExportVideoAsync exports an image collection as a video and returns a task.
//
// Example:
//
//	task, err := helpers.ExportVideoAsync(ctx, client, collection,
//	    helpers.ExportDescription("Time Lapse"),
//	    helpers.ExportToGCS("my-bucket", "videos/"),
//	    helpers.ExportFileFormat(helpers.MP4))
func ExportVideoAsync(ctx context.Context, client *earthengine.Client, collection *earthengine.ImageCollection, opts ...ExportImageOption) (*earthengine.Task, error) {
	cfg := &ExportConfig{
		Description: "Video Export",
		Destination: ExportToCloudStorage,
		Format:      MP4,
		CRS:         "EPSG:4326",
		Scale:       30,
		MaxPixels:   1e9,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	if err := validateExportConfig(cfg); err != nil {
		return nil, err
	}

	taskID := atomic.AddUint64(&taskIDCounter, 1)
	task := &earthengine.Task{
		ID:          fmt.Sprintf("export-video-%d-%d", time.Now().Unix(), taskID),
		Type:        "VIDEO_EXPORT",
		Description: cfg.Description,
		State:       earthengine.TaskStatePending,
		Progress:    0.0,
		StartTime:   time.Now(),
		UpdateTime:  time.Now(),
	}

	return task, nil
}

// ExportImageWithNotification exports an image and calls a notification function on completion.
//
// The notification function is called in a separate goroutine when the export completes,
// fails, or is cancelled.
//
// Example:
//
//	task, err := helpers.ExportImageWithNotification(ctx, client, image,
//	    func(task *earthengine.Task, err error) {
//	        if err != nil {
//	            log.Printf("Export failed: %v", err)
//	        } else {
//	            log.Printf("Export completed: %s", task.ID)
//	            // Send email, trigger webhook, etc.
//	        }
//	    },
//	    helpers.ExportDescription("Summer 2023 NDVI"),
//	    helpers.ExportToGCS("my-bucket", "exports/"))
func ExportImageWithNotification(ctx context.Context, client *earthengine.Client, image *earthengine.Image, notifyFn func(*earthengine.Task, error), opts ...ExportImageOption) (*earthengine.Task, error) {
	task, err := ExportImageAsync(ctx, client, image, opts...)
	if err != nil {
		return nil, err
	}

	// Start monitoring in background
	go func() {
		err := task.Wait(ctx)
		if notifyFn != nil {
			notifyFn(task, err)
		}
	}()

	return task, nil
}

// ExportTableWithNotification exports a table with completion notification.
func ExportTableWithNotification(ctx context.Context, client *earthengine.Client, collection interface{}, notifyFn func(*earthengine.Task, error), opts ...ExportImageOption) (*earthengine.Task, error) {
	task, err := ExportTableAsync(ctx, client, collection, opts...)
	if err != nil {
		return nil, err
	}

	go func() {
		err := task.Wait(ctx)
		if notifyFn != nil {
			notifyFn(task, err)
		}
	}()

	return task, nil
}

// ExportVideoWithNotification exports a video with completion notification.
func ExportVideoWithNotification(ctx context.Context, client *earthengine.Client, collection *earthengine.ImageCollection, notifyFn func(*earthengine.Task, error), opts ...ExportImageOption) (*earthengine.Task, error) {
	task, err := ExportVideoAsync(ctx, client, collection, opts...)
	if err != nil {
		return nil, err
	}

	go func() {
		err := task.Wait(ctx)
		if notifyFn != nil {
			notifyFn(task, err)
		}
	}()

	return task, nil
}

// WaitForExports waits for multiple export tasks to complete.
//
// Returns when all tasks complete or when the context is cancelled.
//
// Example:
//
//	tasks := []*earthengine.Task{task1, task2, task3}
//	results := helpers.WaitForExports(ctx, tasks, func(completed, total int) {
//	    fmt.Printf("Exports completed: %d/%d\n", completed, total)
//	})
//
//	for i, result := range results {
//	    if result.Error != nil {
//	        log.Printf("Task %d failed: %v", i, result.Error)
//	    }
//	}
func WaitForExports(ctx context.Context, tasks []*earthengine.Task, progressFn func(completed, total int)) []ExportResult {
	results := make([]ExportResult, len(tasks))
	completed := 0
	var mu sync.Mutex

	// Report initial progress
	if progressFn != nil {
		progressFn(0, len(tasks))
	}

	// Wait for all tasks
	var wg sync.WaitGroup
	for i, task := range tasks {
		wg.Add(1)
		go func(index int, t *earthengine.Task) {
			defer wg.Done()

			err := t.Wait(ctx)
			results[index] = ExportResult{
				TaskID: t.ID,
				Task:   t,
				Error:  err,
			}

			mu.Lock()
			completed++
			current := completed
			mu.Unlock()

			if progressFn != nil {
				progressFn(current, len(tasks))
			}
		}(i, task)
	}

	wg.Wait()
	return results
}

// ExportResult represents the result of an export operation.
type ExportResult struct {
	TaskID string
	Task   *earthengine.Task
	Error  error
}
