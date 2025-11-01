package models

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	parser "github.com/sfi2k7/blueconfig/parser"
)

// MethodFunc represents a user-defined function that can be called during evaluation
type MethodFunc func(args ...any) (any, error)

// Row represents a data row as a map of field names to RowValue pointers
type Row map[string]*RowValue

// Schema represents field type definitions
type Schema map[string]string

// Object represents a record/row with methods for querying and matching
type Object struct {
	data    Row
	schema  Schema
	methods map[string]MethodFunc
	vars    map[string]any
}

// NewObject creates a new Object with the given data (assumes data is already flattened)
func NewObject(data Row) *Object {
	return &Object{
		data:    data,
		schema:  make(Schema),
		methods: make(map[string]MethodFunc),
		vars:    make(map[string]any),
	}
}

// NewObjectFromMap creates a new Object from a map, automatically flattening nested structures
func NewObjectFromMap(data map[string]any) *Object {
	return NewObject(NewRow(data))
}

// NewObjectFromStruct creates a new Object from a struct, automatically flattening nested fields
func NewObjectFromStruct(data any) *Object {
	return NewObject(NewRowFromStruct(data))
}

// NewObjectWithContext creates a new Object with data, schema, methods, and variables (assumes data is already flattened)
func NewObjectWithContext(data Row, schema Schema, methods map[string]MethodFunc, vars map[string]any) *Object {
	obj := &Object{
		data:    data,
		schema:  make(Schema),
		methods: make(map[string]MethodFunc),
		vars:    make(map[string]any),
	}

	if schema != nil {
		obj.schema = schema
	}
	if methods != nil {
		obj.methods = methods
	}
	if vars != nil {
		obj.vars = vars
	}

	return obj
}

// NewObjectFromMapWithContext creates a new Object from a map with context, automatically flattening
func NewObjectFromMapWithContext(data map[string]any, schema Schema, methods map[string]MethodFunc, vars map[string]any) *Object {
	return NewObjectWithContext(NewRowWithSchema(data, schema), schema, methods, vars)
}

// NewObjectFromStructWithContext creates a new Object from a struct with context, automatically flattening
func NewObjectFromStructWithContext(data any, schema Schema, methods map[string]MethodFunc, vars map[string]any) *Object {
	return NewObjectWithContext(NewRowFromStructWithSchema(data, schema), schema, methods, vars)
}

// Reset clears the data map, allowing the Object to be reused with new data
// while preserving the schema, methods, and variables
func (o *Object) Reset() {
	o.data = make(Row)
}

// SetData replaces the current data with new data (assumes data is already flattened)
func (o *Object) SetData(data Row) {
	o.data = data
	// Apply schema types to values
	for key, val := range o.data {
		if schemaType, exists := o.schema[key]; exists {
			val.SetSchemaType(schemaType)
		}
	}
}

// SetDataFromMap replaces the current data with new data from a map, automatically flattening
func (o *Object) SetDataFromMap(data map[string]any) {
	o.SetData(NewRowWithSchema(data, o.schema))
}

// SetDataFromStruct replaces the current data with new data from a struct, automatically flattening
func (o *Object) SetDataFromStruct(data any) {
	o.SetData(NewRowFromStructWithSchema(data, o.schema))
}

// GetValue retrieves a RowValue by path (for the first level only)
func (o *Object) GetValue(path string) (*RowValue, bool) {
	val, ok := o.data[path]
	return val, ok
}

// ToMap converts the flattened Object data to a nested map
func (o *Object) ToMap() map[string]any {
	return RowToMap(o.data)
}

// ToStruct converts the flattened Object data to a struct
// target must be a pointer to a struct
func (o *Object) ToStruct(target any) error {
	return RowToStruct(o.data, target)
}

// Get retrieves a value by path (assumes flattened data with dot notation keys)
// For example, "address.city" is already a flattened key in the Row
func (o *Object) Get(path string) (any, bool) {
	if path == "" {
		return nil, false
	}

	// Since data is already flattened, just look up the key directly
	val, ok := o.data[path]
	if !ok {
		return nil, false
	}

	if val.IsNull() {
		return nil, true
	}

	return val.Val(), true
}

// GetType returns the type of a field as a string
func (o *Object) GetType(path string) string {
	val, ok := o.Get(path)
	if !ok {
		return "unknown"
	}

	// Check schema first if available
	if schemaType, exists := o.schema[path]; exists {
		return schemaType
	}

	// Infer from actual value
	return inferType(val)
}

// IsOfType checks if a field is of the specified type
func (o *Object) IsOfType(path, typeName string) bool {
	actualType := o.GetType(path)
	return actualType == typeName
}

// IsInt checks if a field is an integer type
func (o *Object) IsInt(path string) bool {
	val, ok := o.Get(path)
	if !ok {
		return false
	}

	switch val.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return true
	case float64:
		// Check if it's a whole number
		f := val.(float64)
		return f == float64(int64(f))
	default:
		return false
	}
}

// IsFloat checks if a field is a float type
func (o *Object) IsFloat(path string) bool {
	val, ok := o.Get(path)
	if !ok {
		return false
	}

	switch val.(type) {
	case float32, float64:
		return true
	default:
		return false
	}
}

// IsString checks if a field is a string type
func (o *Object) IsString(path string) bool {
	val, ok := o.Get(path)
	if !ok {
		return false
	}

	_, isString := val.(string)
	return isString
}

// IsBool checks if a field is a boolean type
func (o *Object) IsBool(path string) bool {
	val, ok := o.Get(path)
	if !ok {
		return false
	}

	_, isBool := val.(bool)
	return isBool
}

// Match evaluates a Query against the object's data
func (o *Object) Match(query parser.Query) (bool, error) {
	return o.matchQuery(query)
}

// matchQuery recursively evaluates a query
func (o *Object) matchQuery(query parser.Query) (bool, error) {
	if query.IsOr {
		// OR logic: at least one subquery must match
		for _, subQuery := range query.SubQueries {
			match, err := o.matchQuery(subQuery)
			if err != nil {
				return false, err
			}
			if match {
				return true, nil
			}
		}
		return false, nil
	}

	// AND logic: all conditions and subqueries must match
	for _, condition := range query.Conditions {
		match, err := o.matchCondition(condition)
		if err != nil {
			return false, err
		}
		if !match {
			return false, nil
		}
	}

	for _, subQuery := range query.SubQueries {
		match, err := o.matchQuery(subQuery)
		if err != nil {
			return false, err
		}
		if !match {
			return false, nil
		}
	}

	return true, nil
}

// matchCondition evaluates a single condition
func (o *Object) matchCondition(cond parser.Condition) (bool, error) {
	var result bool
	var err error

	switch cond.Op {
	case "==":
		result, err = o.matchComparison(cond, "==")
	case ">":
		result, err = o.matchComparison(cond, ">")
	case "<":
		result, err = o.matchComparison(cond, "<")
	case ">=":
		result, err = o.matchComparison(cond, ">=")
	case "<=":
		result, err = o.matchComparison(cond, "<=")
	case "IN":
		result, err = o.matchIn(cond)
	case "NOT_IN":
		result, err = o.matchNotIn(cond)
	case "IS_NULL":
		result, err = o.matchIsNull(cond)
	case "IS_NOT_NULL":
		result, err = o.matchIsNotNull(cond)
	case "LIKE":
		result, err = o.matchLike(cond)
	case "BETWEEN":
		result, err = o.matchBetween(cond)
	case "CONTAINS":
		result, err = o.matchContains(cond)
	case "ANY_OF":
		result, err = o.matchAnyOf(cond)
	default:
		return false, fmt.Errorf("unsupported operator: %s", cond.Op)
	}

	if err != nil {
		return false, err
	}

	// Apply negation if needed
	if cond.Negate {
		result = !result
	}

	return result, nil
}

// evaluateTerm evaluates a ConditionTerm and returns its value
func (o *Object) evaluateTerm(term *parser.ConditionTerm) (any, error) {
	if term == nil {
		return nil, fmt.Errorf("term is nil")
	}

	// Property reference
	if term.Property != "" {
		val, ok := o.Get(term.Property)
		if !ok {
			return nil, nil // Property doesn't exist, return nil
		}
		return val, nil
	}

	// Literal value
	if term.Value != nil {
		return term.Value, nil
	}

	// Variable reference
	if term.Variable != nil {
		if val, ok := o.vars[term.Variable.Name]; ok {
			return val, nil
		}
		return nil, fmt.Errorf("variable not found: %s", term.Variable.Name)
	}

	// Function call
	if term.Function != nil {
		return o.evaluateFunction(term.Function)
	}

	// Arithmetic expression
	if term.Arithmetic != nil {
		return o.evaluateArithmetic(term.Arithmetic)
	}

	// Type cast
	if term.Cast != nil {
		return o.evaluateCast(term.Cast)
	}

	// DateTime
	if term.DateTime != nil {
		return o.evaluateDateTime(term.DateTime)
	}

	return nil, fmt.Errorf("unable to evaluate term")
}

// evaluateFunction calls a user-defined or built-in function
func (o *Object) evaluateFunction(fn *parser.FunctionCall) (any, error) {
	// Evaluate arguments
	args := make([]any, len(fn.Args))
	for i, argTerm := range fn.Args {
		val, err := o.evaluateTerm(argTerm)
		if err != nil {
			return nil, err
		}
		args[i] = val
	}

	// Check if function exists in methods
	if method, ok := o.methods[fn.Name]; ok {
		return method(args...)
	}

	// Built-in functions
	return o.evaluateBuiltinFunction(fn.Name, args)
}

// evaluateBuiltinFunction handles built-in functions
func (o *Object) evaluateBuiltinFunction(name string, args []any) (any, error) {
	switch name {
	case "upper":
		if len(args) != 1 {
			return nil, fmt.Errorf("upper() expects 1 argument, got %d", len(args))
		}
		str, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("upper() expects string argument")
		}
		return strings.ToUpper(str), nil

	case "lower":
		if len(args) != 1 {
			return nil, fmt.Errorf("lower() expects 1 argument, got %d", len(args))
		}
		str, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("lower() expects string argument")
		}
		return strings.ToLower(str), nil

	case "len":
		if len(args) != 1 {
			return nil, fmt.Errorf("len() expects 1 argument, got %d", len(args))
		}
		switch v := args[0].(type) {
		case string:
			return float64(len(v)), nil
		case []any:
			return float64(len(v)), nil
		default:
			return nil, fmt.Errorf("len() expects string or array argument")
		}

	default:
		return nil, fmt.Errorf("unknown function: %s", name)
	}
}

// evaluateArithmetic evaluates an arithmetic expression
func (o *Object) evaluateArithmetic(expr *parser.ArithmeticExpr) (any, error) {
	left, err := o.evaluateTerm(expr.Left)
	if err != nil {
		return nil, err
	}

	right, err := o.evaluateTerm(expr.Right)
	if err != nil {
		return nil, err
	}

	// Convert to float64 for arithmetic
	leftNum, err := toFloat64(left)
	if err != nil {
		return nil, fmt.Errorf("left operand is not a number: %v", left)
	}

	rightNum, err := toFloat64(right)
	if err != nil {
		return nil, fmt.Errorf("right operand is not a number: %v", right)
	}

	switch expr.Op {
	case "+":
		return leftNum + rightNum, nil
	case "-":
		return leftNum - rightNum, nil
	case "*":
		return leftNum * rightNum, nil
	case "/":
		if rightNum == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return leftNum / rightNum, nil
	case "%":
		if rightNum == 0 {
			return nil, fmt.Errorf("modulo by zero")
		}
		return float64(int64(leftNum) % int64(rightNum)), nil
	default:
		return nil, fmt.Errorf("unsupported arithmetic operator: %s", expr.Op)
	}
}

// evaluateCast performs type casting
func (o *Object) evaluateCast(cast *parser.TypeCast) (any, error) {
	val, err := o.evaluateTerm(cast.Value)
	if err != nil {
		return nil, err
	}

	switch cast.TargetType {
	case "int":
		return toInt64(val)
	case "float":
		return toFloat64(val)
	case "string":
		return toString(val), nil
	case "bool":
		return toBool(val), nil
	default:
		return nil, fmt.Errorf("unsupported cast type: %s", cast.TargetType)
	}
}

// evaluateDateTime evaluates a date/time expression
func (o *Object) evaluateDateTime(dt *parser.DateTimeValue) (any, error) {
	// This is a placeholder - in a real implementation, you'd use time.Now(), etc.
	// For now, we'll return string representations
	switch dt.Type {
	case "NOW":
		return "NOW()", nil // Placeholder
	case "TODAY":
		return "TODAY()", nil // Placeholder
	default:
		return nil, fmt.Errorf("unsupported datetime type: %s", dt.Type)
	}
}

// matchComparison handles comparison operators
func (o *Object) matchComparison(cond parser.Condition, op string) (bool, error) {
	left, err := o.evaluateTerm(cond.Left)
	if err != nil {
		return false, err
	}

	right, err := o.evaluateTerm(cond.Right)
	if err != nil {
		return false, err
	}

	return compare(left, right, op)
}

// matchIn checks if left value is in the list of values
func (o *Object) matchIn(cond parser.Condition) (bool, error) {
	left, err := o.evaluateTerm(cond.Left)
	if err != nil {
		return false, err
	}

	for _, inTerm := range cond.InValues {
		val, err := o.evaluateTerm(inTerm)
		if err != nil {
			return false, err
		}
		if equals(left, val) {
			return true, nil
		}
	}

	return false, nil
}

// Helper functions

// wrapValue wraps a value in a RowValue for comparison
func wrapValue(val any) *RowValue {
	if v, ok := val.(*RowValue); ok {
		return v
	}
	return NewValue(val)
}

func inferType(val any) string {
	if val == nil {
		return "null"
	}

	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.String:
		// Check if it's a JSON-encoded array or object
		str := val.(string)
		if len(str) > 0 {
			trimmed := strings.TrimSpace(str)
			if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
				return "array"
			}
			if strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}") {
				return "map"
			}
		}
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "int"
	case reflect.Float32, reflect.Float64:
		return "float"
	case reflect.Bool:
		return "bool"
	case reflect.Map:
		return "map"
	case reflect.Slice, reflect.Array:
		return "array"
	default:
		return "unknown"
	}
}

func toFloat64(val any) (float64, error) {
	if val == nil {
		return 0, fmt.Errorf("cannot convert nil to float64")
	}

	switch v := val.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, err
		}
		return f, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", val)
	}
}

func toInt64(val any) (int64, error) {
	if val == nil {
		return 0, fmt.Errorf("cannot convert nil to int64")
	}

	switch v := val.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case uint:
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint64:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case float32:
		return int64(v), nil
	case string:
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, err
		}
		return i, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int64", val)
	}
}

func toString(val any) string {
	if val == nil {
		return ""
	}
	return fmt.Sprintf("%v", val)
}

func toBool(val any) bool {
	if val == nil {
		return false
	}

	switch v := val.(type) {
	case bool:
		return v
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return v != 0
	case float32, float64:
		return v != 0.0
	case string:
		return v != ""
	default:
		return false
	}
}

// matchContains checks if a collection contains a value
func (o *Object) matchContains(cond parser.Condition) (bool, error) {
	left, err := o.evaluateTerm(cond.Left)
	if err != nil {
		return false, err
	}

	right, err := o.evaluateTerm(cond.Right)
	if err != nil {
		return false, err
	}

	// Check if left (array/slice) contains right (value)
	switch leftVal := left.(type) {
	case []any:
		for _, item := range leftVal {
			if equals(item, right) {
				return true, nil
			}
		}
		return false, nil
	case string:
		// String contains
		rightStr, ok := right.(string)
		if !ok {
			return false, nil
		}
		return strings.Contains(leftVal, rightStr), nil
	default:
		return false, fmt.Errorf("CONTAINS operator requires array or string on left side")
	}
}

// Helper functions

func (o *Object) matchNotIn(cond parser.Condition) (bool, error) {
	result, err := o.matchIn(cond)
	if err != nil {
		return false, err
	}
	return !result, nil
}

// matchIsNull checks if a value is null/nil
func (o *Object) matchIsNull(cond parser.Condition) (bool, error) {
	val, err := o.evaluateTerm(cond.Left)
	if err != nil {
		return false, err
	}
	return val == nil, nil
}

// matchIsNotNull checks if a field is not null
func (o *Object) matchIsNotNull(cond parser.Condition) (bool, error) {
	result, err := o.matchIsNull(cond)
	if err != nil {
		return false, err
	}
	return !result, nil
}

// matchLike performs pattern matching (simple wildcard matching)
func (o *Object) matchLike(cond parser.Condition) (bool, error) {
	left, err := o.evaluateTerm(cond.Left)
	if err != nil {
		return false, err
	}

	str, ok := left.(string)
	if !ok {
		return false, nil
	}

	// Simple wildcard matching: % matches any sequence of characters
	pattern := cond.Pattern
	pattern = strings.ReplaceAll(pattern, "%", ".*")
	pattern = "^" + pattern + "$"

	// For simplicity, just use string contains for now
	// In a real implementation, you'd use regexp
	if strings.Contains(pattern, ".*") {
		// Has wildcards
		return strings.Contains(str, strings.ReplaceAll(cond.Pattern, "%", "")), nil
	}

	return str == cond.Pattern, nil
}

// matchBetween checks if a value is between start and end (inclusive)
func (o *Object) matchBetween(cond parser.Condition) (bool, error) {
	val, err := o.evaluateTerm(cond.Left)
	if err != nil {
		return false, err
	}

	start, err := o.evaluateTerm(cond.Start)
	if err != nil {
		return false, err
	}

	end, err := o.evaluateTerm(cond.End)
	if err != nil {
		return false, err
	}

	// Check if val >= start && val <= end
	greaterOrEqual, err := compare(val, start, ">=")
	if err != nil {
		return false, err
	}

	lessOrEqual, err := compare(val, end, "<=")
	if err != nil {
		return false, err
	}

	return greaterOrEqual && lessOrEqual, nil
}

// matchAnyOf checks if any value in left array is in right array
func (o *Object) matchAnyOf(cond parser.Condition) (bool, error) {
	left, err := o.evaluateTerm(cond.Left)
	if err != nil {
		return false, err
	}

	leftArray, ok := left.([]any)
	if !ok {
		return false, fmt.Errorf("ANY_OF operator requires array on left side")
	}

	for _, leftItem := range leftArray {
		for _, rightTerm := range cond.InValues {
			rightVal, err := o.evaluateTerm(rightTerm)
			if err != nil {
				return false, err
			}
			if equals(leftItem, rightVal) {
				return true, nil
			}
		}
	}
	return false, nil
}

// TestExtractQueryDependencies tests the dependency extraction functionality

func compare(left, right any, op string) (bool, error) {
	// Wrap values in RowValue for proper type conversion
	leftVal := wrapValue(left)
	rightVal := wrapValue(right)

	return leftVal.Compare(rightVal, op)
}

func equals(left, right any) bool {
	// Wrap values in RowValue for proper type conversion
	leftVal := wrapValue(left)
	rightVal := wrapValue(right)

	return leftVal.Equals(rightVal)
}
