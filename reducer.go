package earthengine

// Reducer represents a reducer operation in Earth Engine.
type Reducer interface {
	// NodeID returns the node ID for this reducer in the expression graph.
	NodeID(expr *ExpressionBuilder) string
}

// SimpleReducer represents a basic reducer with no arguments.
type SimpleReducer struct {
	algorithmName string
}

// NodeID implements the Reducer interface for SimpleReducer.
func (r SimpleReducer) NodeID(expr *ExpressionBuilder) string {
	return expr.FunctionCall(r.algorithmName, map[string]interface{}{})
}

// ReducerFirst returns a reducer that gets the first value.
func ReducerFirst() Reducer {
	return SimpleReducer{algorithmName: AlgorithmReducerFirst}
}

// ReducerMean returns a reducer that calculates the mean value.
func ReducerMean() Reducer {
	return SimpleReducer{algorithmName: AlgorithmReducerMean}
}

// ReducerSum returns a reducer that calculates the sum of values.
func ReducerSum() Reducer {
	return SimpleReducer{algorithmName: AlgorithmReducerSum}
}

// ReducerMin returns a reducer that finds the minimum value.
func ReducerMin() Reducer {
	return SimpleReducer{algorithmName: AlgorithmReducerMin}
}

// ReducerMax returns a reducer that finds the maximum value.
func ReducerMax() Reducer {
	return SimpleReducer{algorithmName: AlgorithmReducerMax}
}

// ReducerCount returns a reducer that counts values.
func ReducerCount() Reducer {
	return SimpleReducer{algorithmName: AlgorithmReducerCount}
}
