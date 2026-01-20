package earthengine

import (
	"encoding/json"
	"testing"
)

func TestNewExpression(t *testing.T) {
	expr := NewExpression()
	if expr == nil {
		t.Fatal("NewExpression returned nil")
	}
	if expr.values == nil {
		t.Error("Expression values map is nil")
	}
	if expr.nextID != 0 {
		t.Errorf("Expected nextID to be 0, got %d", expr.nextID)
	}
}

func TestAddConstant(t *testing.T) {
	expr := NewExpression()
	nodeID := expr.AddConstant(42)

	if nodeID != "0" {
		t.Errorf("Expected first node ID to be '0', got '%s'", nodeID)
	}

	// Verify the node was added correctly
	node, ok := expr.values[nodeID]
	if !ok {
		t.Fatal("Node not found in values map")
	}

	nodeMap, ok := node.(map[string]interface{})
	if !ok {
		t.Fatal("Node is not a map")
	}

	if val, ok := nodeMap["constantValue"]; !ok || val != 42 {
		t.Errorf("Expected constantValue of 42, got %v", val)
	}
}

func TestAddFunctionCall(t *testing.T) {
	expr := NewExpression()
	args := map[string]interface{}{
		"id": "test-image",
	}
	nodeID := expr.AddFunctionCall("algorithms/Image.load", args)

	if nodeID != "0" {
		t.Errorf("Expected first node ID to be '0', got '%s'", nodeID)
	}

	node, ok := expr.values[nodeID]
	if !ok {
		t.Fatal("Node not found in values map")
	}

	nodeMap := node.(map[string]interface{})
	funcInv := nodeMap["functionInvocationValue"].(map[string]interface{})

	if funcInv["functionName"] != "algorithms/Image.load" {
		t.Errorf("Unexpected function name: %v", funcInv["functionName"])
	}
}

func TestExpressionMarshalJSON(t *testing.T) {
	expr := NewExpression()
	constantID := expr.AddConstant("test-value")
	expr.SetResult(constantID)

	data, err := json.Marshal(expr)
	if err != nil {
		t.Fatalf("Failed to marshal expression: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	exprData := result["expression"].(map[string]interface{})
	if exprData["result"] != "0" {
		t.Errorf("Expected result to be '0', got %v", exprData["result"])
	}

	values := exprData["values"].(map[string]interface{})
	if len(values) != 1 {
		t.Errorf("Expected 1 value node, got %d", len(values))
	}
}

func TestExpressionBuilder(t *testing.T) {
	builder := NewExpressionBuilder()

	// Build a simple expression: Image.load("test-image")
	imageID := builder.FunctionCall("algorithms/Image.load", map[string]interface{}{
		"id": "test-image",
	})

	expr := builder.Build(imageID)

	// Verify the expression
	if expr.result != "0" {
		t.Errorf("Expected result to be '0', got '%s'", expr.result)
	}

	if len(expr.values) != 1 {
		t.Errorf("Expected 1 value node, got %d", len(expr.values))
	}
}

func TestComplexExpression(t *testing.T) {
	// Build a complex expression similar to tree coverage query
	builder := NewExpressionBuilder()

	// Node: Image.load
	loadID := builder.FunctionCall(AlgorithmImageLoad, map[string]interface{}{
		"id": "USGS/NLCD_RELEASES/2021_REL/NLCD",
	})

	// Node: Image.select
	selectID := builder.FunctionCall(AlgorithmImageSelect, map[string]interface{}{
		"input": map[string]interface{}{
			"valueReference": loadID,
		},
		"bandSelectors": map[string]interface{}{
			"constantValue": []interface{}{"tree_canopy"},
		},
	})

	// Node: Point
	pointID := builder.FunctionCall(AlgorithmGeometryPoint, map[string]interface{}{
		"coordinates": map[string]interface{}{
			"constantValue": []interface{}{-120.9, 47.6},
		},
	})

	// Node: Reducer.first
	reducerID := builder.FunctionCall(AlgorithmReducerFirst, map[string]interface{}{})

	// Node: Image.reduceRegion
	reduceID := builder.FunctionCall(AlgorithmImageReduceRegion, map[string]interface{}{
		"image": map[string]interface{}{
			"valueReference": selectID,
		},
		"geometry": map[string]interface{}{
			"valueReference": pointID,
		},
		"reducer": map[string]interface{}{
			"valueReference": reducerID,
		},
		"scale": map[string]interface{}{
			"constantValue": 30,
		},
	})

	expr := builder.Build(reduceID)

	// Verify we have all 5 nodes
	if len(expr.values) != 5 {
		t.Errorf("Expected 5 nodes, got %d", len(expr.values))
	}

	// Verify the result points to the reduceRegion node
	if expr.result != "4" {
		t.Errorf("Expected result to be '4', got '%s'", expr.result)
	}

	// Verify it can be marshaled to JSON
	_, err := json.Marshal(expr)
	if err != nil {
		t.Fatalf("Failed to marshal complex expression: %v", err)
	}
}
