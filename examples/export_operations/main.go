package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/alexscott64/go-earthengine"
	"github.com/alexscott64/go-earthengine/helpers"
)

// This example demonstrates how to use the export API for exporting
// Earth Engine data to various destinations with progress tracking.
//
// Use cases:
// - Export processed imagery to Cloud Storage
// - Export analysis results to Google Drive
// - Save feature collections as CSV/Shapefiles
// - Create time-lapse videos
// - Monitor long-running export operations

func main() {
	ctx := context.Background()

	// Initialize Earth Engine client
	// Note: This example requires proper authentication setup
	client := &earthengine.Client{}
	_ = ctx // For demonstration only

	fmt.Println("Earth Engine Export Operations Examples")
	fmt.Println("=======================================")
	fmt.Println()

	// Example 1: Simple image export
	example1_SimpleImageExport(ctx, client)

	// Example 2: Export with progress tracking
	example2_ExportWithProgress(ctx, client)

	// Example 3: Multiple exports with batch monitoring
	example3_BatchExports(ctx, client)

	// Example 4: Export with notification
	example4_ExportWithNotification(ctx, client)

	// Example 5: Export to different destinations
	example5_ExportDestinations(ctx, client)

	// Example 6: Table and video exports
	example6_OtherExportTypes(ctx, client)

	// Example 7: Task management
	example7_TaskManagement(ctx, client)
}

func example1_SimpleImageExport(ctx context.Context, client *earthengine.Client) {
	fmt.Println("Example 1: Simple Image Export")
	fmt.Println("-------------------------------")

	// Create an NDVI composite
	startDate := "2023-06-01"
	endDate := "2023-08-31"

	// Create median composite
	composite := helpers.Composite(client, startDate, endDate,
		helpers.MedianComposite,
		helpers.Sentinel2(),
		helpers.CloudMask(20))

	// Export to Cloud Storage
	task, err := helpers.ExportImageAsync(ctx, client, composite,
		helpers.ExportDescription("Summer 2023 NDVI Composite"),
		helpers.ExportToGCS("my-ee-exports", "composites/"),
		helpers.ExportScale(10),
		helpers.ExportCRS("EPSG:4326"),
		helpers.ExportFileFormat(helpers.GeoTIFF))

	if err != nil {
		log.Printf("Export failed: %v", err)
		return
	}

	fmt.Printf("Export started: %s\n", task.ID)
	fmt.Printf("Description: %s\n", task.Description)
	fmt.Printf("Type: %s\n", task.Type)
	fmt.Printf("State: %s\n", task.State)
	fmt.Println()
}

func example2_ExportWithProgress(ctx context.Context, client *earthengine.Client) {
	fmt.Println("Example 2: Export with Progress Tracking")
	fmt.Println("-----------------------------------------")

	// Create a simple image
	image := client.Image("USGS/SRTMGL1_003")

	// Start export
	task, err := helpers.ExportImageAsync(ctx, client, image,
		helpers.ExportDescription("SRTM Elevation Export"),
		helpers.ExportToGCS("my-ee-exports", "elevation/"),
		helpers.ExportScale(30))

	if err != nil {
		log.Printf("Export failed: %v", err)
		return
	}

	fmt.Printf("Export task created: %s\n", task.ID)
	fmt.Println("Monitoring progress...")
	fmt.Println()

	// Wait with progress updates
	err = task.WaitWithProgress(ctx, func(progress *earthengine.TaskProgress) {
		fmt.Printf("\rProgress: %.1f%% | State: %s",
			progress.Progress*100,
			progress.State)
	})

	fmt.Println()

	if err != nil {
		log.Printf("Export failed: %v", err)
	} else {
		fmt.Println("Export completed successfully!")
	}

	fmt.Println()
}

func example3_BatchExports(ctx context.Context, client *earthengine.Client) {
	fmt.Println("Example 3: Multiple Exports with Batch Monitoring")
	fmt.Println("-------------------------------------------------")

	// Define regions to export
	regions := []struct {
		name string
		bounds [4]float64 // [west, south, east, north]
	}{
		{"Region A", [4]float64{-122.5, 37.5, -122.0, 38.0}},
		{"Region B", [4]float64{-122.0, 37.5, -121.5, 38.0}},
		{"Region C", [4]float64{-121.5, 37.5, -121.0, 38.0}},
	}

	// Create export tasks for each region
	var tasks []*earthengine.Task
	for _, region := range regions {
		// Create an image for this region
		image := client.Image("USGS/SRTMGL1_003")

		task, err := helpers.ExportImageAsync(ctx, client, image,
			helpers.ExportDescription(fmt.Sprintf("Export %s", region.name)),
			helpers.ExportToGCS("my-ee-exports", fmt.Sprintf("regions/%s/", region.name)),
			helpers.ExportScale(30))

		if err != nil {
			log.Printf("Failed to start export for %s: %v", region.name, err)
			continue
		}

		tasks = append(tasks, task)
		fmt.Printf("Started export: %s (%s)\n", region.name, task.ID)
	}

	fmt.Println()
	fmt.Println("Waiting for all exports to complete...")

	// Wait for all exports
	results := helpers.WaitForExports(ctx, tasks, func(completed, total int) {
		fmt.Printf("\rExports completed: %d/%d (%.1f%%)",
			completed, total, float64(completed)*100/float64(total))
	})

	fmt.Println()
	fmt.Println()

	// Summary
	successful := 0
	for i, result := range results {
		if result.Error != nil {
			fmt.Printf("❌ Export %d failed: %v\n", i+1, result.Error)
		} else {
			fmt.Printf("✓ Export %d completed: %s\n", i+1, result.TaskID)
			successful++
		}
	}

	fmt.Printf("\nSummary: %d/%d exports successful\n", successful, len(results))
	fmt.Println()
}

func example4_ExportWithNotification(ctx context.Context, client *earthengine.Client) {
	fmt.Println("Example 4: Export with Completion Notification")
	fmt.Println("-----------------------------------------------")

	image := client.Image("USGS/SRTMGL1_003")

	// Export with notification callback
	task, err := helpers.ExportImageWithNotification(ctx, client, image,
		func(t *earthengine.Task, err error) {
			if err != nil {
				log.Printf("❌ Export failed: %v", err)
				// Could send email, trigger webhook, etc.
			} else {
				log.Printf("✓ Export completed successfully: %s", t.ID)
				// Could send success notification
			}
		},
		helpers.ExportDescription("Export with Notification"),
		helpers.ExportToGCS("my-ee-exports", "notifications/"))

	if err != nil {
		log.Printf("Failed to start export: %v", err)
		return
	}

	fmt.Printf("Export started: %s\n", task.ID)
	fmt.Println("Notification will be triggered on completion...")
	fmt.Println("(Continuing with other work while export runs in background)")
	fmt.Println()

	// Simulate doing other work
	time.Sleep(2 * time.Second)

	fmt.Println("Other work completed")
	fmt.Println()
}

func example5_ExportDestinations(ctx context.Context, client *earthengine.Client) {
	fmt.Println("Example 5: Export to Different Destinations")
	fmt.Println("--------------------------------------------")

	image := client.Image("USGS/SRTMGL1_003")

	// Export to Google Cloud Storage
	taskGCS, err := helpers.ExportImageAsync(ctx, client, image,
		helpers.ExportDescription("Export to GCS"),
		helpers.ExportToGCS("my-bucket", "exports/gcs/"),
		helpers.ExportScale(30))
	if err == nil {
		fmt.Printf("✓ GCS export started: %s\n", taskGCS.ID)
	}

	// Export to Google Drive
	taskDrive, err := helpers.ExportImageAsync(ctx, client, image,
		helpers.ExportDescription("Export to Drive"),
		helpers.ExportToGoogleDrive("Earth Engine Exports"),
		helpers.ExportScale(30))
	if err == nil {
		fmt.Printf("✓ Drive export started: %s\n", taskDrive.ID)
	}

	// Export as Earth Engine Asset
	taskAsset, err := helpers.ExportImageAsync(ctx, client, image,
		helpers.ExportDescription("Export as Asset"),
		helpers.ExportToEEAsset("projects/my-project/assets/elevation"),
		helpers.ExportScale(30))
	if err == nil {
		fmt.Printf("✓ Asset export started: %s\n", taskAsset.ID)
	}

	fmt.Println()
}

func example6_OtherExportTypes(ctx context.Context, client *earthengine.Client) {
	fmt.Println("Example 6: Table and Video Exports")
	fmt.Println("-----------------------------------")

	// Export table (feature collection)
	// In a real application, you would have an actual feature collection
	tableTask, err := helpers.ExportTableAsync(ctx, client, nil,
		helpers.ExportDescription("Export Feature Collection"),
		helpers.ExportToGCS("my-bucket", "tables/"),
		helpers.ExportFileFormat(helpers.CSV))
	if err == nil {
		fmt.Printf("✓ Table export started: %s\n", tableTask.ID)
	}

	// Export video (image collection time series)
	// In a real application, you would have an actual image collection
	collection := client.ImageCollection("COPERNICUS/S2_SR")
	videoTask, err := helpers.ExportVideoAsync(ctx, client, collection,
		helpers.ExportDescription("Time Lapse Video"),
		helpers.ExportToGCS("my-bucket", "videos/"),
		helpers.ExportFileFormat(helpers.MP4))
	if err == nil {
		fmt.Printf("✓ Video export started: %s\n", videoTask.ID)
	}

	fmt.Println()
}

func example7_TaskManagement(ctx context.Context, client *earthengine.Client) {
	fmt.Println("Example 7: Task Management")
	fmt.Println("--------------------------")

	// Create a task manager
	tm := earthengine.NewTaskManager(client)

	// Start several exports
	for i := 0; i < 3; i++ {
		image := client.Image("USGS/SRTMGL1_003")
		task, err := helpers.ExportImageAsync(ctx, client, image,
			helpers.ExportDescription(fmt.Sprintf("Export %d", i+1)),
			helpers.ExportToGCS("my-bucket", fmt.Sprintf("task%d/", i)))

		if err != nil {
			log.Printf("Failed to start export %d: %v", i, err)
			continue
		}

		tm.RegisterTask(task)
		fmt.Printf("Registered task: %s\n", task.ID)
	}

	fmt.Println()

	// List all tasks
	allTasks := tm.ListTasks()
	fmt.Printf("Total tasks: %d\n", len(allTasks))

	// Filter running/pending tasks
	activeTasks := tm.FilterTasks(earthengine.TaskFilter{
		States: []earthengine.TaskState{
			earthengine.TaskStatePending,
			earthengine.TaskStateRunning,
		},
	})
	fmt.Printf("Active tasks: %d\n", len(activeTasks))

	// Get specific task
	if len(allTasks) > 0 {
		task, err := tm.GetTask(allTasks[0].ID)
		if err == nil {
			progress := task.GetProgress()
			fmt.Printf("Task %s: %.1f%% complete\n", task.ID, progress.Progress*100)
		}
	}

	fmt.Println()

	// Cancel all active tasks
	fmt.Println("Cancelling all active tasks...")
	err := tm.CancelAll(ctx)
	if err != nil {
		log.Printf("Error cancelling tasks: %v", err)
	}

	// Cleanup old tasks
	fmt.Println("Cleaning up old tasks...")
	tm.Cleanup(24 * time.Hour) // Remove tasks older than 24 hours

	fmt.Println("Task management complete")
	fmt.Println()
}

// Helper functions for real-world scenarios

func exportNDVITimeSeries(ctx context.Context, client *earthengine.Client, startDate, endDate string, bounds [4]float64) (*earthengine.Task, error) {
	// Create NDVI time series
	collection := client.ImageCollection("COPERNICUS/S2_SR").
		FilterDate(startDate, endDate).
		FilterMetadata("CLOUDY_PIXEL_PERCENTAGE", "less_than", 20)

	// Calculate NDVI for each image
	// In a real implementation, you would map an NDVI calculation over the collection
	_ = collection // Placeholder for demonstration

	// Export as multi-band image or video
	return helpers.ExportImageAsync(ctx, client, nil,
		helpers.ExportDescription("NDVI Time Series"),
		helpers.ExportToGCS("my-bucket", "ndvi-series/"),
		helpers.ExportScale(10))
}

func exportClassificationResult(ctx context.Context, client *earthengine.Client, classified *earthengine.Image) (*earthengine.Task, error) {
	// Export classified image with palette
	return helpers.ExportImageAsync(ctx, client, classified,
		helpers.ExportDescription("Land Cover Classification"),
		helpers.ExportToGCS("my-bucket", "classifications/"),
		helpers.ExportScale(30),
		helpers.ExportCRS("EPSG:4326"),
		helpers.ExportMaxPixels(1e9))
}

func exportStatistics(ctx context.Context, client *earthengine.Client) (*earthengine.Task, error) {
	// Export zonal statistics as table
	return helpers.ExportTableAsync(ctx, client, nil,
		helpers.ExportDescription("Zonal Statistics"),
		helpers.ExportToGCS("my-bucket", "statistics/"),
		helpers.ExportFileFormat(helpers.CSV))
}
