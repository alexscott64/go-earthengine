package apiv1

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// ProjectsOperationsService handles long-running operations.
//
// Operations are returned by export and other long-running API calls.
// Use this service to check status, wait for completion, or cancel operations.
type ProjectsOperationsService struct {
	s *Service
}

// Get retrieves the current status of an operation.
//
// Example:
//
//	op, err := service.Projects.Operations.Get(ctx, "projects/my-project/operations/abc123")
//	if op.Done {
//	    if op.Error != nil {
//	        log.Printf("Operation failed: %v", op.Error.Message)
//	    } else {
//	        log.Printf("Operation completed: %v", op.Response)
//	    }
//	}
func (r *ProjectsOperationsService) Get(ctx context.Context, name string, opts ...CallOption) (*Operation, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	resp := &Operation{}
	if err := r.s.makeRequest(ctx, "GET", name, nil, resp, opts...); err != nil {
		return nil, err
	}

	return resp, nil
}

// List lists operations that match the specified filter.
//
// Use pageToken for pagination.
//
// Example:
//
//	resp, err := service.Projects.Operations.List(ctx, "projects/my-project", nil)
//	for _, op := range resp.Operations {
//	    fmt.Printf("Operation: %s, Done: %v\n", op.Name, op.Done)
//	}
func (r *ProjectsOperationsService) List(ctx context.Context, name string, pageSize int, pageToken string, filter string, opts ...CallOption) (*ListOperationsResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	// Build URL with query parameters
	u, err := url.Parse(r.s.BasePath + name + "/operations")
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	q := u.Query()
	if pageSize > 0 {
		q.Set("pageSize", fmt.Sprintf("%d", pageSize))
	}
	if pageToken != "" {
		q.Set("pageToken", pageToken)
	}
	if filter != "" {
		q.Set("filter", filter)
	}
	u.RawQuery = q.Encode()

	resp := &ListOperationsResponse{}
	if err := r.s.makeRequest(ctx, "GET", name+"/operations", nil, resp, opts...); err != nil {
		return nil, err
	}

	return resp, nil
}

// Wait waits for an operation to complete, with an optional timeout.
//
// This is a blocking call that polls the operation until it's done or the timeout expires.
//
// Example:
//
//	req := &apiv1.WaitOperationRequest{Timeout: "3600s"}
//	op, err := service.Projects.Operations.Wait(ctx, "projects/my-project/operations/abc123", req)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if op.Error != nil {
//	    log.Fatalf("Operation failed: %v", op.Error.Message)
//	}
func (r *ProjectsOperationsService) Wait(ctx context.Context, name string, req *WaitOperationRequest, opts ...CallOption) (*Operation, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	urlPath := name + ":wait"
	resp := &Operation{}

	if err := r.s.makeRequest(ctx, "POST", urlPath, req, resp, opts...); err != nil {
		return nil, err
	}

	return resp, nil
}

// WaitWithPolling waits for an operation to complete using client-side polling.
//
// This is useful when the server-side :wait method times out.
// It polls the operation status at the specified interval until complete or context is cancelled.
//
// Example:
//
//	op, err := service.Projects.Operations.WaitWithPolling(ctx, opName, 5*time.Second)
func (r *ProjectsOperationsService) WaitWithPolling(ctx context.Context, name string, pollInterval time.Duration) (*Operation, error) {
	if pollInterval == 0 {
		pollInterval = 5 * time.Second
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			op, err := r.Get(ctx, name)
			if err != nil {
				return nil, err
			}
			if op.Done {
				return op, nil
			}
		}
	}
}

// Cancel cancels a long-running operation.
//
// Note: Cancellation is not always guaranteed - some operations may complete
// before the cancellation takes effect.
//
// Example:
//
//	err := service.Projects.Operations.Cancel(ctx, "projects/my-project/operations/abc123")
func (r *ProjectsOperationsService) Cancel(ctx context.Context, name string, opts ...CallOption) error {
	if name == "" {
		return fmt.Errorf("name is required")
	}

	urlPath := name + ":cancel"
	return r.s.makeRequest(ctx, "POST", urlPath, &Empty{}, nil, opts...)
}

// Delete deletes a long-running operation.
//
// This removes the operation from the server. It does not cancel the operation.
//
// Example:
//
//	err := service.Projects.Operations.Delete(ctx, "projects/my-project/operations/abc123")
func (r *ProjectsOperationsService) Delete(ctx context.Context, name string, opts ...CallOption) error {
	if name == "" {
		return fmt.Errorf("name is required")
	}

	return r.s.makeRequest(ctx, "DELETE", name, nil, nil, opts...)
}
