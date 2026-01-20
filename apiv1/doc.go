// Package apiv1 provides a low-level client for the Google Earth Engine API v1.
//
// This package contains auto-generated types and methods based on the Earth Engine
// API discovery document. For most use cases, prefer the high-level helpers in
// the parent package.
//
// # Service Client
//
// The Service type is the main entry point for API calls:
//
//	ctx := context.Background()
//	service, err := apiv1.NewService(ctx,
//	    apiv1.WithServiceAccountFile("credentials.json"),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Use the service
//	resp, err := service.Projects.Value.Compute(ctx, "projects/my-project", req)
//
// # Resource Hierarchy
//
// The API is organized into resources:
//   - Projects.Value - Compute single values from expressions
//   - Projects.Image - Image import, export, and pixel computation
//   - Projects.Table - Table/vector data operations
//   - Projects.Assets - Asset management (CRUD)
//   - Projects.Operations - Long-running operation management
//
// # Authentication
//
// Authentication uses Google Cloud service accounts. See the parent package
// documentation for authentication details.
//
// # Error Handling
//
// All methods return errors that can be type-asserted to *googleapi.Error
// for detailed error information including HTTP status codes.
package apiv1
