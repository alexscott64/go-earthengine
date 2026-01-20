package apiv1

// This file contains type definitions generated from the Earth Engine API discovery document.
// Types are organized by category: expressions, assets, operations, requests/responses.

import "fmt"

// ===== Core Expression Types =====

// Expression represents an Earth Engine computation as a directed acyclic graph.
// The graph consists of value nodes referenced by string IDs, with one node
// designated as the result.
type Expression struct {
	// Result is the ID of the value node that produces the final result.
	Result string `json:"result"`

	// Values is a map of node IDs to their value definitions.
	Values map[string]*ValueNode `json:"values"`
}

// ValueNode represents a single node in an expression graph.
// Exactly one of the fields should be set.
type ValueNode struct {
	// ConstantValue is a constant value (primitive, array, or object).
	ConstantValue interface{} `json:"constantValue,omitempty"`

	// FunctionInvocationValue is a function call.
	FunctionInvocationValue *FunctionInvocation `json:"functionInvocationValue,omitempty"`

	// ValueReference is a reference to another node by ID.
	ValueReference string `json:"valueReference,omitempty"`

	// ArgumentReference is a reference to a function argument.
	ArgumentReference string `json:"argumentReference,omitempty"`

	// DictionaryValue is a dictionary/map value.
	DictionaryValue map[string]*ValueNode `json:"dictionaryValue,omitempty"`

	// ArrayValue is an array of value nodes.
	ArrayValue []*ValueNode `json:"arrayValue,omitempty"`
}

// FunctionInvocation represents a function call in an expression.
type FunctionInvocation struct {
	// FunctionName is the fully-qualified function name (e.g., "Image.load").
	FunctionName string `json:"functionName"`

	// Arguments maps argument names to their value nodes.
	Arguments map[string]*ValueNode `json:"arguments,omitempty"`
}

// ===== Asset Types =====

// EarthEngineAsset represents metadata about an Earth Engine asset.
type EarthEngineAsset struct {
	// Name is the asset's resource name (e.g., "projects/*/assets/**").
	Name string `json:"name,omitempty"`

	// Type is the asset type (e.g., "IMAGE", "IMAGE_COLLECTION", "TABLE", "FOLDER").
	Type string `json:"type,omitempty"`

	// Title is a human-readable title for the asset.
	Title string `json:"title,omitempty"`

	// Description is a longer description of the asset.
	Description string `json:"description,omitempty"`

	// Properties are custom metadata properties.
	Properties map[string]interface{} `json:"properties,omitempty"`

	// StartTime is the start time of the asset's temporal extent (RFC3339).
	StartTime string `json:"startTime,omitempty"`

	// EndTime is the end time of the asset's temporal extent (RFC3339).
	EndTime string `json:"endTime,omitempty"`

	// Geometry is the asset's spatial extent (GeoJSON).
	Geometry interface{} `json:"geometry,omitempty"`

	// Bands describes the image bands (for IMAGE and IMAGE_COLLECTION types).
	Bands []*ImageBand `json:"bands,omitempty"`

	// SizeBytes is the approximate size of the asset in bytes.
	SizeBytes int64 `json:"sizeBytes,omitempty,string"`

	// UpdateTime is when the asset was last updated (RFC3339).
	UpdateTime string `json:"updateTime,omitempty"`
}

// ImageBand represents a single band in an image.
type ImageBand struct {
	// ID is the band identifier.
	ID string `json:"id"`

	// DataType is the pixel data type.
	DataType *PixelDataType `json:"dataType,omitempty"`

	// Grid describes the band's pixel grid.
	Grid *PixelGrid `json:"grid,omitempty"`

	// PyramidingPolicy describes how to aggregate pixels at coarser resolutions.
	PyramidingPolicy string `json:"pyramidingPolicy,omitempty"`
}

// PixelDataType describes the data type of pixels in a band.
type PixelDataType struct {
	// Precision is the numeric precision (e.g., "INT", "FLOAT", "DOUBLE").
	Precision string `json:"precision,omitempty"`

	// Range specifies the valid value range (for integer types).
	Range *ValueRange `json:"range,omitempty"`
}

// ValueRange specifies a numeric range.
type ValueRange struct {
	Min float64 `json:"min,omitempty"`
	Max float64 `json:"max,omitempty"`
}

// PixelGrid describes the spatial properties of a raster band.
type PixelGrid struct {
	// CrsCode is the EPSG code (e.g., "EPSG:4326").
	CrsCode string `json:"crsCode,omitempty"`

	// CrsWkt is the WKT representation of the CRS.
	CrsWkt string `json:"crsWkt,omitempty"`

	// AffineTransform is the 6-parameter affine transform.
	AffineTransform *AffineTransform `json:"affineTransform,omitempty"`

	// Dimensions specifies the grid dimensions in pixels.
	Dimensions *GridDimensions `json:"dimensions,omitempty"`
}

// AffineTransform represents a 6-parameter affine transformation.
type AffineTransform struct {
	ScaleX     float64 `json:"scaleX,omitempty"`
	ShearX     float64 `json:"shearX,omitempty"`
	TranslateX float64 `json:"translateX,omitempty"`
	ShearY     float64 `json:"shearY,omitempty"`
	ScaleY     float64 `json:"scaleY,omitempty"`
	TranslateY float64 `json:"translateY,omitempty"`
}

// GridDimensions specifies raster dimensions.
type GridDimensions struct {
	Width  int64 `json:"width,omitempty,string"`
	Height int64 `json:"height,omitempty,string"`
}

// ===== Operation Types (Long-Running Operations) =====

// Operation represents a long-running operation.
type Operation struct {
	// Name is the operation resource name.
	Name string `json:"name,omitempty"`

	// Metadata contains operation metadata.
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Done indicates if the operation is complete.
	Done bool `json:"done,omitempty"`

	// Error contains error details if the operation failed.
	Error *Status `json:"error,omitempty"`

	// Response contains the operation result if successful.
	Response map[string]interface{} `json:"response,omitempty"`
}

// Status represents an error status.
type Status struct {
	// Code is the status code (corresponds to HTTP status codes).
	Code int `json:"code,omitempty"`

	// Message is a developer-facing error message.
	Message string `json:"message,omitempty"`

	// Details contains additional error details.
	Details []interface{} `json:"details,omitempty"`
}

// ===== API Error Type =====

// APIError represents an error returned by the Earth Engine API.
type APIError struct {
	ErrorInfo struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
		Details []struct {
			Type            string           `json:"@type"`
			FieldViolations []FieldViolation `json:"fieldViolations,omitempty"`
		} `json:"details,omitempty"`
	} `json:"error"`
}

// FieldViolation describes a single field validation error.
type FieldViolation struct {
	Field       string `json:"field"`
	Description string `json:"description"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("Earth Engine API error %d (%s): %s",
		e.ErrorInfo.Code, e.ErrorInfo.Status, e.ErrorInfo.Message)
}

// ===== Request/Response Types =====

// ComputeValueRequest is the request for computing a single value.
type ComputeValueRequest struct {
	// Expression is the computation to evaluate.
	Expression *Expression `json:"expression"`

	// FileFormat specifies the output format for binary data.
	FileFormat string `json:"fileFormat,omitempty"`

	// WorkloadTag is an optional tag for quota accounting.
	WorkloadTag string `json:"workloadTag,omitempty"`
}

// ComputeValueResponse is the response from computing a value.
type ComputeValueResponse struct {
	// Result contains the computed value.
	Result interface{} `json:"result,omitempty"`
}

// ComputePixelsRequest is the request for computing image pixels.
type ComputePixelsRequest struct {
	// Expression is the image expression to evaluate.
	Expression *Expression `json:"expression"`

	// FileFormat specifies the output format (e.g., "GEO_TIFF", "NPY").
	FileFormat string `json:"fileFormat,omitempty"`

	// Grid describes the output pixel grid.
	Grid *PixelGrid `json:"grid,omitempty"`

	// BandIds specifies which bands to include.
	BandIds []string `json:"bandIds,omitempty"`

	// VisualizationOptions specifies 8-bit RGB visualization parameters.
	VisualizationOptions *VisualizationOptions `json:"visualizationOptions,omitempty"`

	// WorkloadTag is an optional tag for quota accounting.
	WorkloadTag string `json:"workloadTag,omitempty"`
}

// VisualizationOptions specifies how to render an image as 8-bit RGB.
type VisualizationOptions struct {
	// Ranges specifies the value range for each band.
	Ranges []*ValueRange `json:"ranges,omitempty"`

	// Palette specifies colors for single-band visualization.
	Palette []string `json:"palette,omitempty"`

	// Gain multiplies pixel values.
	Gain []float64 `json:"gain,omitempty"`

	// Bias adds to pixel values.
	Bias []float64 `json:"bias,omitempty"`

	// Gamma applies gamma correction.
	Gamma []float64 `json:"gamma,omitempty"`
}

// ComputeFeaturesRequest is the request for computing features from a table.
type ComputeFeaturesRequest struct {
	// Expression is the feature collection expression.
	Expression *Expression `json:"expression"`

	// FileFormat specifies the output format (e.g., "GEO_JSON", "CSV").
	FileFormat string `json:"fileFormat,omitempty"`

	// WorkloadTag is an optional tag for quota accounting.
	WorkloadTag string `json:"workloadTag,omitempty"`
}

// ExportImageRequest is the request for exporting an image.
type ExportImageRequest struct {
	// Expression is the image expression to export.
	Expression *Expression `json:"expression"`

	// Description is a human-readable description of the export.
	Description string `json:"description,omitempty"`

	// FileExportOptions specifies export destination and format.
	FileExportOptions *FileExportOptions `json:"fileExportOptions,omitempty"`

	// Grid describes the output pixel grid.
	Grid *PixelGrid `json:"grid,omitempty"`

	// BandIds specifies which bands to export.
	BandIds []string `json:"bandIds,omitempty"`

	// RequestId is an optional client-provided request ID for idempotency.
	RequestId string `json:"requestId,omitempty"`

	// WorkloadTag is an optional tag for quota accounting.
	WorkloadTag string `json:"workloadTag,omitempty"`
}

// FileExportOptions specifies export destination and format.
type FileExportOptions struct {
	// FileFormat is the output file format (e.g., "GEO_TIFF", "NUMPY_NDARRAY").
	FileFormat string `json:"fileFormat,omitempty"`

	// DriveDestination exports to Google Drive.
	DriveDestination *DriveDestination `json:"driveDestination,omitempty"`

	// GcsDestination exports to Google Cloud Storage.
	GcsDestination *GcsDestination `json:"gcsDestination,omitempty"`
}

// DriveDestination specifies Google Drive export options.
type DriveDestination struct {
	// Folder is the Drive folder name.
	Folder string `json:"folder,omitempty"`

	// FilenamePrefix is the exported file name prefix.
	FilenamePrefix string `json:"filenamePrefix,omitempty"`
}

// GcsDestination specifies Google Cloud Storage export options.
type GcsDestination struct {
	// Bucket is the GCS bucket name.
	Bucket string `json:"bucket,omitempty"`

	// FilenamePrefix is the object name prefix.
	FilenamePrefix string `json:"filenamePrefix,omitempty"`

	// Permissions specifies access control.
	Permissions string `json:"permissions,omitempty"`
}

// ListAssetsResponse is the response from listing assets.
type ListAssetsResponse struct {
	// Assets is the list of assets.
	Assets []*EarthEngineAsset `json:"assets,omitempty"`

	// NextPageToken is the token for the next page.
	NextPageToken string `json:"nextPageToken,omitempty"`
}

// ListOperationsResponse is the response from listing operations.
type ListOperationsResponse struct {
	// Operations is the list of operations.
	Operations []*Operation `json:"operations,omitempty"`

	// NextPageToken is the token for the next page.
	NextPageToken string `json:"nextPageToken,omitempty"`
}

// ListAlgorithmsResponse is the response from listing algorithms.
type ListAlgorithmsResponse struct {
	// Algorithms is the list of algorithm names.
	Algorithms []string `json:"algorithms,omitempty"`
}

// Empty represents an empty response.
type Empty struct{}

// WaitOperationRequest is the request to wait for an operation to complete.
type WaitOperationRequest struct {
	// Timeout is the maximum time to wait.
	Timeout string `json:"timeout,omitempty"` // Duration string (e.g., "30s")
}
