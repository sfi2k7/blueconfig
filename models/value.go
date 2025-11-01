package models

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type RowValue struct {
	raw          any
	schemaType   string
	converted    any
	isNull       bool
	hasConverted bool
}

func NewValue(raw any) *RowValue {
	return &RowValue{
		raw:    raw,
		isNull: raw == nil,
	}
}

func NewValueWithSchema(raw any, schemaType string) *RowValue {
	return &RowValue{
		raw:        raw,
		schemaType: schemaType,
		isNull:     raw == nil,
	}
}

func (v *RowValue) Base() any {
	return v.raw
}

func (v *RowValue) IsNull() bool {
	return v.isNull
}

func (v *RowValue) Val() any {
	if v.isNull {
		return nil
	}

	if v.hasConverted {
		return v.converted
	}

	if v.schemaType == "" {
		return v.raw
	}

	var result any
	var err error

	switch v.schemaType {
	case "int", "int64", "integer":
		result, err = v.AsInt64()
	case "float", "float64", "double":
		result, err = v.AsFloat64()
	case "string", "text":
		result = v.AsString()
	case "bool", "boolean":
		result, err = v.AsBool()
	default:
		return v.raw
	}

	if err != nil {
		return v.raw
	}

	v.converted = result
	v.hasConverted = true
	return result
}

func (v *RowValue) AsInt64() (int64, error) {
	if v.isNull {
		return 0, fmt.Errorf("cannot convert nil to int64")
	}

	switch val := v.raw.(type) {
	case int64:
		return val, nil
	case int:
		return int64(val), nil
	case int8:
		return int64(val), nil
	case int16:
		return int64(val), nil
	case int32:
		return int64(val), nil
	case uint:
		return int64(val), nil
	case uint8:
		return int64(val), nil
	case uint16:
		return int64(val), nil
	case uint32:
		return int64(val), nil
	case uint64:
		return int64(val), nil
	case float64:
		return int64(val), nil
	case float32:
		return int64(val), nil
	case string:
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return 0, err
		}
		return i, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int64", v.raw)
	}
}

func (v *RowValue) AsFloat64() (float64, error) {
	if v.isNull {
		return 0, fmt.Errorf("cannot convert nil to float64")
	}

	switch val := v.raw.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int8:
		return float64(val), nil
	case int16:
		return float64(val), nil
	case int32:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case uint:
		return float64(val), nil
	case uint8:
		return float64(val), nil
	case uint16:
		return float64(val), nil
	case uint32:
		return float64(val), nil
	case uint64:
		return float64(val), nil
	case string:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return 0, err
		}
		return f, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v.raw)
	}
}

func (v *RowValue) AsString() string {
	if v.isNull {
		return ""
	}
	return fmt.Sprintf("%v", v.raw)
}

func (v *RowValue) AsBool() (bool, error) {
	if v.isNull {
		return false, nil
	}

	switch val := v.raw.(type) {
	case bool:
		return val, nil
	case int:
		return val != 0, nil
	case int8:
		return val != 0, nil
	case int16:
		return val != 0, nil
	case int32:
		return val != 0, nil
	case int64:
		return val != 0, nil
	case uint:
		return val != 0, nil
	case uint8:
		return val != 0, nil
	case uint16:
		return val != 0, nil
	case uint32:
		return val != 0, nil
	case uint64:
		return val != 0, nil
	case float32:
		return val != 0.0, nil
	case float64:
		return val != 0.0, nil
	case string:
		if val == "true" || val == "1" || val == "yes" || val == "t" {
			return true, nil
		}
		if val == "false" || val == "0" || val == "no" || val == "f" || val == "" {
			return false, nil
		}
		return false, fmt.Errorf("cannot convert string '%s' to bool", val)
	default:
		return false, fmt.Errorf("cannot convert %T to bool", v.raw)
	}
}

func (v *RowValue) Type() string {
	if v.schemaType != "" {
		return v.schemaType
	}
	return v.InferredType()
}

func (v *RowValue) InferredType() string {
	if v.isNull {
		return "null"
	}

	val := reflect.ValueOf(v.raw)
	switch val.Kind() {
	case reflect.String:
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

func (v *RowValue) SchemaType() string {
	return v.schemaType
}

func (v *RowValue) SetSchemaType(schemaType string) {
	v.schemaType = schemaType
	v.hasConverted = false
	v.converted = nil
}

func (v *RowValue) Compare(other *RowValue, op string) (bool, error) {
	if op == "==" {
		return v.Equals(other), nil
	}

	leftVal := v.Val()
	rightVal := other.Val()

	leftNum, leftErr := v.AsFloat64()
	rightNum, rightErr := other.AsFloat64()

	if leftErr == nil && rightErr == nil {
		switch op {
		case ">":
			return leftNum > rightNum, nil
		case "<":
			return leftNum < rightNum, nil
		case ">=":
			return leftNum >= rightNum, nil
		case "<=":
			return leftNum <= rightNum, nil
		default:
			return false, fmt.Errorf("unsupported comparison operator: %s", op)
		}
	}

	leftStr := v.AsString()
	rightStr := other.AsString()

	switch op {
	case ">":
		return leftStr > rightStr, nil
	case "<":
		return leftStr < rightStr, nil
	case ">=":
		return leftStr >= rightStr, nil
	case "<=":
		return leftStr <= rightStr, nil
	default:
		return false, fmt.Errorf("cannot compare %T and %T with operator %s", leftVal, rightVal, op)
	}
}

func (v *RowValue) Equals(other *RowValue) bool {
	if v.isNull && other.isNull {
		return true
	}
	if v.isNull || other.isNull {
		return false
	}

	leftVal := v.Val()
	rightVal := other.Val()

	if leftVal == rightVal {
		return true
	}

	leftNum, leftErr := v.AsFloat64()
	rightNum, rightErr := other.AsFloat64()
	if leftErr == nil && rightErr == nil {
		return leftNum == rightNum
	}

	return v.AsString() == other.AsString()
}

func (v *RowValue) String() string {
	if v.isNull {
		return "<nil>"
	}
	if v.schemaType != "" {
		return fmt.Sprintf("Value{raw=%v, type=%s, val=%v}", v.raw, v.schemaType, v.Val())
	}
	return fmt.Sprintf("Value{raw=%v}", v.raw)
}

func (v *RowValue) MustInt64() int64 {
	val, err := v.AsInt64()
	if err != nil {
		panic(err)
	}
	return val
}

func (v *RowValue) MustFloat64() float64 {
	val, err := v.AsFloat64()
	if err != nil {
		panic(err)
	}
	return val
}

func (v *RowValue) IsZero() bool {
	if v.isNull {
		return true
	}

	switch val := v.Val().(type) {
	case int, int8, int16, int32, int64:
		return val == 0
	case uint, uint8, uint16, uint32, uint64:
		return val == 0
	case float32, float64:
		return val == 0.0
	case string:
		return val == ""
	case bool:
		return !val
	default:
		return false
	}
}

func (v *RowValue) IsNumeric() bool {
	_, err := v.AsFloat64()
	return err == nil
}

func (v *RowValue) Clone() *RowValue {
	return &RowValue{
		raw:          v.raw,
		schemaType:   v.schemaType,
		converted:    v.converted,
		isNull:       v.isNull,
		hasConverted: v.hasConverted,
	}
}

// flattenMap flattens a nested map into dot notation
func flattenMap(data map[string]any, prefix string) map[string]any {
	result := make(map[string]any)

	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		if value == nil {
			result[fullKey] = nil
			continue
		}

		// Check if value is a nested map
		val := reflect.ValueOf(value)
		if val.Kind() == reflect.Map && val.Type().Key().Kind() == reflect.String {
			// Convert to map[string]any
			nestedMap := make(map[string]any)
			iter := val.MapRange()
			for iter.Next() {
				k := iter.Key().String()
				v := iter.Value().Interface()
				nestedMap[k] = v
			}
			// Recursively flatten
			nested := flattenMap(nestedMap, fullKey)
			for k, v := range nested {
				result[k] = v
			}
		} else if val.Kind() == reflect.Struct {
			// Flatten struct fields
			nested := flattenStruct(value, fullKey)
			for k, v := range nested {
				result[k] = v
			}
		} else if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
			// Arrays/slices are stored as JSON strings
			jsonBytes, err := json.Marshal(value)
			if err != nil {
				result[fullKey] = fmt.Sprintf("%v", value)
			} else {
				result[fullKey] = string(jsonBytes)
			}
		} else {
			// Primitive value
			result[fullKey] = value
		}
	}

	return result
}

// flattenStruct flattens a struct into dot notation
func flattenStruct(data any, prefix string) map[string]any {
	result := make(map[string]any)
	val := reflect.ValueOf(data)

	// Handle pointer to struct
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return result
		}
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return result
	}

	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get field name (check for json tag first)
		fieldName := field.Name
		omitEmpty := false

		if tag := field.Tag.Get("json"); tag != "" {
			parts := strings.Split(tag, ",")

			// Check if field should be skipped
			if parts[0] == "-" {
				continue
			}

			// Use tag name if provided, otherwise use field name
			if parts[0] != "" {
				fieldName = parts[0]
			}

			// Check for omitempty option
			for _, option := range parts[1:] {
				if option == "omitempty" {
					omitEmpty = true
					break
				}
			}
		}

		// Skip if omitempty and value is zero/nil/empty
		if omitEmpty {
			if fieldValue.IsZero() {
				continue
			}
			// Also check for empty slices/arrays
			if (fieldValue.Kind() == reflect.Slice || fieldValue.Kind() == reflect.Array) && fieldValue.Len() == 0 {
				continue
			}
		}

		fullKey := fieldName
		if prefix != "" {
			fullKey = prefix + "." + fieldName
		}

		// Handle different field types
		// Handle pointers first
		if fieldValue.Kind() == reflect.Ptr {
			if fieldValue.IsNil() {
				result[fullKey] = nil
			} else {
				// Dereference and process the actual value
				deref := fieldValue.Elem()
				if deref.Kind() == reflect.Struct {
					// Nested struct via pointer
					nested := flattenStruct(deref.Interface(), fullKey)
					for k, v := range nested {
						result[k] = v
					}
				} else if deref.Kind() == reflect.Map && deref.Type().Key().Kind() == reflect.String {
					// Nested map via pointer
					nestedMap := make(map[string]any)
					iter := deref.MapRange()
					for iter.Next() {
						k := iter.Key().String()
						v := iter.Value().Interface()
						nestedMap[k] = v
					}
					nested := flattenMap(nestedMap, fullKey)
					for k, v := range nested {
						result[k] = v
					}
				} else {
					// Primitive via pointer
					result[fullKey] = deref.Interface()
				}
			}
		} else if fieldValue.Kind() == reflect.Struct {
			// Nested struct
			nested := flattenStruct(fieldValue.Interface(), fullKey)
			for k, v := range nested {
				result[k] = v
			}
		} else if fieldValue.Kind() == reflect.Map && fieldValue.Type().Key().Kind() == reflect.String {
			// Nested map
			nestedMap := make(map[string]any)
			iter := fieldValue.MapRange()
			for iter.Next() {
				k := iter.Key().String()
				v := iter.Value().Interface()
				nestedMap[k] = v
			}
			nested := flattenMap(nestedMap, fullKey)
			for k, v := range nested {
				result[k] = v
			}
		} else if fieldValue.Kind() == reflect.Slice || fieldValue.Kind() == reflect.Array {
			// Arrays/slices as JSON
			jsonBytes, err := json.Marshal(fieldValue.Interface())
			if err != nil {
				result[fullKey] = fmt.Sprintf("%v", fieldValue.Interface())
			} else {
				result[fullKey] = string(jsonBytes)
			}
		} else {
			// Primitive value
			result[fullKey] = fieldValue.Interface()
		}
	}

	return result
}

// NewRow creates a Row from a map, automatically flattening nested structures
func NewRow(data map[string]any) Row {
	// Flatten the data first
	flattened := flattenMap(data, "")

	row := make(Row)
	for key, val := range flattened {
		row[key] = NewValue(val)
	}
	return row
}

// NewRowWithSchema creates a Row from a map with schema, automatically flattening
func NewRowWithSchema(data map[string]any, schema Schema) Row {
	// Flatten the data first
	flattened := flattenMap(data, "")

	row := make(Row)
	for key, val := range flattened {
		if schemaType, exists := schema[key]; exists {
			row[key] = NewValueWithSchema(val, schemaType)
		} else {
			row[key] = NewValue(val)
		}
	}
	return row
}

// NewRowFromStruct creates a flattened Row from a struct
func NewRowFromStruct(data any) Row {
	// Flatten the struct
	flattened := flattenStruct(data, "")

	row := make(Row)
	for key, val := range flattened {
		row[key] = NewValue(val)
	}
	return row
}

// NewRowFromStructWithSchema creates a flattened Row from a struct with schema
func NewRowFromStructWithSchema(data any, schema Schema) Row {
	// Flatten the struct
	flattened := flattenStruct(data, "")

	row := make(Row)
	for key, val := range flattened {
		if schemaType, exists := schema[key]; exists {
			row[key] = NewValueWithSchema(val, schemaType)
		} else {
			row[key] = NewValue(val)
		}
	}
	return row
}

// unflattenMap converts a flattened map back into a nested map
func unflattenMap(flattened map[string]any) map[string]any {
	result := make(map[string]any)

	for key, value := range flattened {
		parts := strings.Split(key, ".")
		current := result

		// Navigate/create the nested structure
		for i := 0; i < len(parts)-1; i++ {
			part := parts[i]
			if _, exists := current[part]; !exists {
				current[part] = make(map[string]any)
			}
			// Type assertion - if it's not a map, we have a conflict
			if nested, ok := current[part].(map[string]any); ok {
				current = nested
			} else {
				// Conflict: this path is both a value and a container
				// Skip this key or overwrite - we'll skip for safety
				break
			}
		}

		// Set the final value
		if len(parts) > 0 {
			lastPart := parts[len(parts)-1]
			current[lastPart] = value
		}
	}

	return result
}

// unflattenToStruct converts a flattened Row into a struct
// The target must be a pointer to a struct
func unflattenToStruct(row Row, target any) error {
	targetVal := reflect.ValueOf(target)
	if targetVal.Kind() != reflect.Ptr || targetVal.IsNil() {
		return fmt.Errorf("target must be a non-nil pointer to struct")
	}

	targetVal = targetVal.Elem()
	if targetVal.Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct, got pointer to %s", targetVal.Kind())
	}

	// First, convert Row to map[string]any
	flatMap := make(map[string]any)
	for key, rowVal := range row {
		if rowVal.IsNull() {
			flatMap[key] = nil
		} else {
			flatMap[key] = rowVal.Val()
		}
	}

	// Unflatten the map
	nestedMap := unflattenMap(flatMap)

	// Now populate the struct from the nested map
	return populateStructFromMap(targetVal, nestedMap)
}

// populateStructFromMap populates a struct from a map
func populateStructFromMap(structVal reflect.Value, data map[string]any) error {
	structType := structVal.Type()

	for i := 0; i < structVal.NumField(); i++ {
		field := structType.Field(i)
		fieldValue := structVal.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get field name from json tag or use field name
		fieldName := field.Name
		omitEmpty := false

		if tag := field.Tag.Get("json"); tag != "" {
			parts := strings.Split(tag, ",")
			if parts[0] == "-" {
				continue // Skip this field
			}
			if parts[0] != "" {
				fieldName = parts[0]
			}
			for _, option := range parts[1:] {
				if option == "omitempty" {
					omitEmpty = true
					break
				}
			}
		}

		// Get value from data map
		value, exists := data[fieldName]
		if !exists {
			if omitEmpty {
				continue
			}
			// Field not in data, leave as zero value
			continue
		}

		// Handle nil values
		if value == nil {
			if fieldValue.Kind() == reflect.Ptr {
				fieldValue.Set(reflect.Zero(fieldValue.Type()))
			}
			continue
		}

		// Set the field value
		if err := setFieldValue(fieldValue, value); err != nil {
			return fmt.Errorf("error setting field %s: %w", fieldName, err)
		}
	}

	return nil
}

// setFieldValue sets a struct field value from an interface{}
func setFieldValue(fieldValue reflect.Value, value any) error {
	if !fieldValue.CanSet() {
		return fmt.Errorf("field cannot be set")
	}

	valueVal := reflect.ValueOf(value)
	fieldType := fieldValue.Type()

	// Handle pointer fields
	if fieldType.Kind() == reflect.Ptr {
		if value == nil {
			fieldValue.Set(reflect.Zero(fieldType))
			return nil
		}
		// Create a new pointer and set the element
		newPtr := reflect.New(fieldType.Elem())
		if err := setFieldValue(newPtr.Elem(), value); err != nil {
			return err
		}
		fieldValue.Set(newPtr)
		return nil
	}

	// Handle different field types
	switch fieldType.Kind() {
	case reflect.String:
		// Check if value is a JSON-encoded array/object string
		if str, ok := value.(string); ok {
			// Try to detect if it's JSON
			if (strings.HasPrefix(str, "[") && strings.HasSuffix(str, "]")) ||
				(strings.HasPrefix(str, "{") && strings.HasSuffix(str, "}")) {
				// It might be JSON, but field expects string, so just set it
				fieldValue.SetString(str)
			} else {
				fieldValue.SetString(str)
			}
		} else {
			fieldValue.SetString(fmt.Sprintf("%v", value))
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := toInt64FromAny(value)
		if err != nil {
			return err
		}
		fieldValue.SetInt(intVal)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		intVal, err := toInt64FromAny(value)
		if err != nil {
			return err
		}
		fieldValue.SetUint(uint64(intVal))

	case reflect.Float32, reflect.Float64:
		floatVal, err := toFloat64FromAny(value)
		if err != nil {
			return err
		}
		fieldValue.SetFloat(floatVal)

	case reflect.Bool:
		boolVal, err := toBoolFromAny(value)
		if err != nil {
			return err
		}
		fieldValue.SetBool(boolVal)

	case reflect.Slice, reflect.Array:
		// If value is a JSON string, unmarshal it
		if str, ok := value.(string); ok && strings.HasPrefix(str, "[") {
			// Try to unmarshal JSON array
			sliceVal := reflect.New(fieldType)
			if err := json.Unmarshal([]byte(str), sliceVal.Interface()); err != nil {
				return fmt.Errorf("failed to unmarshal JSON array: %w", err)
			}
			fieldValue.Set(sliceVal.Elem())
		} else if valueVal.Kind() == reflect.Slice || valueVal.Kind() == reflect.Array {
			// Direct slice assignment
			fieldValue.Set(valueVal.Convert(fieldType))
		} else {
			return fmt.Errorf("cannot convert %T to %s", value, fieldType.Kind())
		}

	case reflect.Map:
		// If value is already a map
		if valueVal.Kind() == reflect.Map {
			fieldValue.Set(valueVal.Convert(fieldType))
		} else if str, ok := value.(string); ok && strings.HasPrefix(str, "{") {
			// Try to unmarshal JSON object
			mapVal := reflect.New(fieldType)
			if err := json.Unmarshal([]byte(str), mapVal.Interface()); err != nil {
				return fmt.Errorf("failed to unmarshal JSON object: %w", err)
			}
			fieldValue.Set(mapVal.Elem())
		} else {
			return fmt.Errorf("cannot convert %T to map", value)
		}

	case reflect.Struct:
		// If value is a nested map, populate the struct
		if nestedMap, ok := value.(map[string]any); ok {
			if err := populateStructFromMap(fieldValue, nestedMap); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("cannot convert %T to struct", value)
		}

	default:
		// Try direct assignment if types match
		if valueVal.Type().AssignableTo(fieldType) {
			fieldValue.Set(valueVal)
		} else if valueVal.Type().ConvertibleTo(fieldType) {
			fieldValue.Set(valueVal.Convert(fieldType))
		} else {
			return fmt.Errorf("cannot assign %T to field of type %s", value, fieldType)
		}
	}

	return nil
}

// Helper conversion functions
func toInt64FromAny(value any) (int64, error) {
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
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
	case float32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to int64", value)
	}
}

func toFloat64FromAny(value any) (float64, error) {
	switch v := value.(type) {
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
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", value)
	}
}

func toBoolFromAny(value any) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(v).Int() != 0, nil
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(v).Uint() != 0, nil
	case float32, float64:
		return reflect.ValueOf(v).Float() != 0, nil
	case string:
		return strconv.ParseBool(v)
	default:
		return false, fmt.Errorf("cannot convert %T to bool", value)
	}
}

// RowToMap converts a flattened Row to a nested map
func RowToMap(row Row) map[string]any {
	flatMap := make(map[string]any)
	for key, rowVal := range row {
		if rowVal.IsNull() {
			flatMap[key] = nil
		} else {
			flatMap[key] = rowVal.Val()
		}
	}
	return unflattenMap(flatMap)
}

// RowToStruct converts a flattened Row to a struct
func RowToStruct(row Row, target any) error {
	return unflattenToStruct(row, target)
}

// RowToRawMap converts a Row to map[string]string for raw API
// This is used when interfacing with the raw database layer
func RowToRawMap(row Row) map[string]string {
	result := make(map[string]string)
	for key, val := range row {
		if val.IsNull() {
			result[key] = ""
		} else {
			result[key] = val.AsString()
		}
	}
	return result
}

// RawMapToRow converts a map[string]string from raw API to Row
// This is used when retrieving data from the raw database layer
func RawMapToRow(rawMap map[string]string) Row {
	row := make(Row)
	for key, val := range rawMap {
		if val == "" {
			row[key] = NewValue(nil)
		} else {
			row[key] = NewValue(val)
		}
	}
	return row
}
