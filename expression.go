package earthengine

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Expression represents an Earth Engine expression with a graph-based structure.
// The expression consists of a map of value nodes and a result node ID.
type Expression struct {
	values map[string]interface{} // Node map: nodeID -> value definition
	nextID int                     // Counter for generating unique node IDs
	result string                  // ID of the result node
}

// NewExpression creates a new empty expression.
func NewExpression() *Expression {
	return &Expression{
		values: make(map[string]interface{}),
		nextID: 0,
	}
}

// ValueNode represents different types of value nodes in the expression graph.
type ValueNode struct {
	ConstantValue           interface{}            `json:"constantValue,omitempty"`
	FunctionInvocationValue *FunctionInvocation    `json:"functionInvocationValue,omitempty"`
	ValueReference          string                 `json:"valueReference,omitempty"`
	ArgumentReference       string                 `json:"argumentReference,omitempty"`
}

// FunctionInvocation represents a function call in the expression graph.
type FunctionInvocation struct {
	FunctionName string                 `json:"functionName"`
	Arguments    map[string]interface{} `json:"arguments,omitempty"`
}

// AddConstant adds a constant value node to the expression and returns its ID.
func (e *Expression) AddConstant(value interface{}) string {
	nodeID := e.getNextID()
	e.values[nodeID] = map[string]interface{}{
		"constantValue": value,
	}
	return nodeID
}

// AddValueReference adds a reference to another node and returns its ID.
func (e *Expression) AddValueReference(refID string) string {
	nodeID := e.getNextID()
	e.values[nodeID] = map[string]interface{}{
		"valueReference": refID,
	}
	return nodeID
}

// AddFunctionCall adds a function invocation node to the expression and returns its ID.
func (e *Expression) AddFunctionCall(functionName string, args map[string]interface{}) string {
	nodeID := e.getNextID()
	e.values[nodeID] = map[string]interface{}{
		"functionInvocationValue": map[string]interface{}{
			"functionName": functionName,
			"arguments":    args,
		},
	}
	return nodeID
}

// SetResult sets the result node ID for the expression.
func (e *Expression) SetResult(nodeID string) {
	e.result = nodeID
}

// MarshalJSON implements json.Marshaler for Expression.
func (e *Expression) MarshalJSON() ([]byte, error) {
	if e.result == "" {
		return nil, fmt.Errorf("expression has no result node set")
	}

	return json.Marshal(map[string]interface{}{
		"expression": map[string]interface{}{
			"result": e.result,
			"values": e.values,
		},
	})
}

// getNextID generates the next unique node ID.
func (e *Expression) getNextID() string {
	id := strconv.Itoa(e.nextID)
	e.nextID++
	return id
}

// ExpressionBuilder provides a helper for building complex expressions.
type ExpressionBuilder struct {
	expr *Expression
}

// NewExpressionBuilder creates a new expression builder.
func NewExpressionBuilder() *ExpressionBuilder {
	return &ExpressionBuilder{
		expr: NewExpression(),
	}
}

// Constant adds a constant value and returns its node ID.
func (eb *ExpressionBuilder) Constant(value interface{}) string {
	return eb.expr.AddConstant(value)
}

// FunctionCall adds a function call and returns its node ID.
func (eb *ExpressionBuilder) FunctionCall(functionName string, args map[string]interface{}) string {
	return eb.expr.AddFunctionCall(functionName, args)
}

// Reference adds a value reference and returns its node ID.
func (eb *ExpressionBuilder) Reference(nodeID string) string {
	return eb.expr.AddValueReference(nodeID)
}

// Build sets the result node and returns the completed expression.
func (eb *ExpressionBuilder) Build(resultNodeID string) *Expression {
	eb.expr.SetResult(resultNodeID)
	return eb.expr
}
