package models

import (
	"strings"
	"testing"

	parser "github.com/sfi2k7/blueconfig/parser"
)

// TestObjectBasicGet tests the Get method with simple and nested paths
func TestObjectBasicGet(t *testing.T) {
	user := NewRow(map[string]any{
		"id":    1,
		"name":  "John Doe",
		"email": "john.doe@example.com",
		"age":   30,
		"location": map[string]any{
			"city":    "New York",
			"state":   "NY",
			"country": "USA",
		},
	})

	obj := NewObject(user)

	tests := []struct {
		name     string
		path     string
		expected any
		exists   bool
	}{
		{"get id", "id", 1, true},
		{"get name", "name", "John Doe", true},
		{"get age", "age", 30, true},
		{"get nested city", "location.city", "New York", true},
		{"get nested state", "location.state", "NY", true},
		{"get nested country", "location.country", "USA", true},
		{"get non-existent field", "nonexistent", nil, false},
		{"get non-existent nested", "location.zipcode", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, exists := obj.Get(tt.path)
			if exists != tt.exists {
				t.Errorf("Get(%s) exists = %v, want %v", tt.path, exists, tt.exists)
			}
			if exists && val != tt.expected {
				t.Errorf("Get(%s) = %v, want %v", tt.path, val, tt.expected)
			}
		})
	}
}

// TestObjectTypeChecking tests type checking methods
func TestObjectTypeChecking(t *testing.T) {
	user := NewRow(map[string]any{
		"id":     1,
		"name":   "John Doe",
		"age":    30,
		"score":  95.5,
		"active": true,
		"tags":   []any{"developer", "golang"},
		// Note: nested map will be flattened to "profile.bio"
		"profile": map[string]any{"bio": "Software engineer"},
	})

	obj := NewObject(user)

	t.Run("IsInt", func(t *testing.T) {
		if !obj.IsInt("id") {
			t.Error("id should be int")
		}
		if !obj.IsInt("age") {
			t.Error("age should be int")
		}
		if obj.IsInt("score") {
			t.Error("score should not be int (it's float)")
		}
		if obj.IsInt("name") {
			t.Error("name should not be int")
		}
	})

	t.Run("IsString", func(t *testing.T) {
		if !obj.IsString("name") {
			t.Error("name should be string")
		}
		if obj.IsString("age") {
			t.Error("age should not be string")
		}
	})

	t.Run("IsBool", func(t *testing.T) {
		if !obj.IsBool("active") {
			t.Error("active should be bool")
		}
		if obj.IsBool("name") {
			t.Error("name should not be bool")
		}
	})

	t.Run("GetType", func(t *testing.T) {
		tests := []struct {
			path     string
			expected string
		}{
			{"id", "int"},
			{"name", "string"},
			{"score", "float"},
			{"active", "bool"},
			{"tags", "array"},
			{"profile.bio", "string"}, // Flattened from nested map
			{"nonexistent", "unknown"},
		}

		for _, tt := range tests {
			actualType := obj.GetType(tt.path)
			if actualType != tt.expected {
				t.Errorf("GetType(%s) = %s, want %s", tt.path, actualType, tt.expected)
			}
		}
	})
}

// TestObjectMatchSimpleConditions tests matching simple conditions
func TestObjectMatchSimpleConditions(t *testing.T) {
	user := NewRow(map[string]any{
		"name":  "John Doe",
		"age":   30,
		"email": "john.doe@example.com",
	})

	obj := NewObject(user)

	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{"equality match", "name == 'John Doe'", true},
		{"equality no match", "name == 'Jane Doe'", false},
		{"greater than true", "age > 25", true},
		{"greater than false", "age > 35", false},
		{"less than true", "age < 35", true},
		{"less than false", "age < 25", false},
		{"greater or equal true", "age >= 30", true},
		{"greater or equal false", "age >= 31", false},
		{"less or equal true", "age <= 30", true},
		{"less or equal false", "age <= 29", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := parser.ParseExprQuery(tt.query)
			match, err := obj.Match(query)
			if err != nil {
				t.Fatalf("Match() error = %v", err)
			}
			if match != tt.expected {
				t.Errorf("Match(%s) = %v, want %v", tt.query, match, tt.expected)
			}
		})
	}
}

// TestObjectMatchAndConditions tests AND logic
func TestObjectMatchAndConditions(t *testing.T) {
	user := NewRow(map[string]any{
		"name": "John Doe",
		"age":  30,
	})

	obj := NewObject(user)

	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{"both conditions true", "name == 'John Doe' && age == 30", true},
		{"first false", "name == 'Jane Doe' && age == 30", false},
		{"second false", "name == 'John Doe' && age == 25", false},
		{"both false", "name == 'Jane Doe' && age == 25", false},
		{"multiple AND all true", "name == 'John Doe' && age > 25 && age < 35", true},
		{"multiple AND one false", "name == 'John Doe' && age > 25 && age < 28", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := parser.ParseExprQuery(tt.query)
			match, err := obj.Match(query)
			if err != nil {
				t.Fatalf("Match() error = %v", err)
			}
			if match != tt.expected {
				t.Errorf("Match(%s) = %v, want %v", tt.query, match, tt.expected)
			}
		})
	}
}

// TestObjectMatchOrConditions tests OR logic
func TestObjectMatchOrConditions(t *testing.T) {
	user := NewRow(map[string]any{
		"name": "John Doe",
		"age":  30,
	})

	obj := NewObject(user)

	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{"both true", "name == 'John Doe' || age == 30", true},
		{"first true second false", "name == 'John Doe' || age == 25", true},
		{"first false second true", "name == 'Jane Doe' || age == 30", true},
		{"both false", "name == 'Jane Doe' || age == 25", false},
		{"multiple OR one true", "name == 'Jane Doe' || age == 25 || age == 30", true},
		{"multiple OR all false", "name == 'Jane Doe' || age == 25 || age == 35", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := parser.ParseExprQuery(tt.query)
			match, err := obj.Match(query)
			if err != nil {
				t.Fatalf("Match() error = %v", err)
			}
			if match != tt.expected {
				t.Errorf("Match(%s) = %v, want %v", tt.query, match, tt.expected)
			}
		})
	}
}

// TestObjectMatchComplexConditions tests complex nested conditions
func TestObjectMatchComplexConditions(t *testing.T) {
	user := NewRow(map[string]any{
		"name": "John Doe",
		"age":  30,
	})

	obj := NewObject(user)

	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{
			"AND with OR subquery - true",
			"name == 'John Doe' && (age < 20 || age > 25)",
			true,
		},
		{
			"AND with OR subquery - false",
			"name == 'John Doe' && (age < 20 || age > 35)",
			false,
		},
		{
			"OR with AND subquery - true",
			"name == 'Jane Doe' || (age > 25 && age < 35)",
			true,
		},
		{
			"OR with AND subquery - false",
			"name == 'Jane Doe' || (age > 35 && age < 40)",
			false,
		},
		{
			"deeply nested - true",
			"(name == 'John Doe' && age > 25) || (name == 'Jane Doe' && age < 25)",
			true,
		},
		{
			"deeply nested - false",
			"(name == 'John Doe' && age > 35) || (name == 'Jane Doe' && age < 25)",
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := parser.ParseExprQuery(tt.query)
			match, err := obj.Match(query)
			if err != nil {
				t.Fatalf("Match() error = %v", err)
			}
			if match != tt.expected {
				t.Errorf("Match(%s) = %v, want %v", tt.query, match, tt.expected)
			}
		})
	}
}

// TestObjectMatchWithVariables tests matching with variables
func TestObjectMatchWithVariables(t *testing.T) {
	user := NewRow(map[string]any{
		"name": "John Doe",
		"age":  30,
	})

	vars := map[string]any{
		"maxAge":     40,
		"minAge":     18,
		"targetName": "John Doe",
	}

	obj := NewObjectWithContext(user, nil, nil, vars)

	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{"variable comparison true", "age < $maxAge", true},
		{"variable comparison false", "age > $maxAge", false},
		{"variable equality true", "name == $targetName", true},
		{"variable in range", "age > $minAge && age < $maxAge", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := parser.ParseExprQuery(tt.query)
			match, err := obj.Match(query)
			if err != nil {
				t.Fatalf("Match() error = %v", err)
			}
			if match != tt.expected {
				t.Errorf("Match(%s) = %v, want %v", tt.query, match, tt.expected)
			}
		})
	}
}

// TestObjectMatchWithFunctions tests matching with function calls
func TestObjectMatchWithFunctions(t *testing.T) {
	user := NewRow(map[string]any{
		"name":  "John Doe",
		"email": "JOHN.DOE@EXAMPLE.COM",
	})

	// Custom methods
	methods := map[string]MethodFunc{
		"customUpper": func(args ...any) (any, error) {
			if len(args) != 1 {
				return nil, nil
			}
			str, ok := args[0].(string)
			if !ok {
				return nil, nil
			}
			return strings.ToUpper(str), nil
		},
	}

	obj := NewObjectWithContext(user, nil, methods, nil)

	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{"built-in upper function true", "upper(name) == 'JOHN DOE'", true},
		{"built-in upper function false", "upper(name) == 'john doe'", false},
		{"built-in lower function true", "lower(email) == 'john.doe@example.com'", true},
		{"custom function true", "customUpper(name) == 'JOHN DOE'", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := parser.ParseExprQuery(tt.query)
			match, err := obj.Match(query)
			if err != nil {
				t.Fatalf("Match() error = %v", err)
			}
			if match != tt.expected {
				t.Errorf("Match(%s) = %v, want %v", tt.query, match, tt.expected)
			}
		})
	}
}

// TestObjectMatchArithmetic tests arithmetic expressions
func TestObjectMatchArithmetic(t *testing.T) {
	user := NewRow(map[string]any{
		"age":   30,
		"score": 85,
	})

	obj := NewObject(user)

	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{"addition", "age + 10 == 40", true},
		{"subtraction", "age - 10 == 20", true},
		{"multiplication", "score * 2 == 170", true},
		{"division", "score / 5 == 17", true},
		{"complex arithmetic", "age * 2 + 10 == 70", true},
		{"arithmetic with field comparison", "age + score > 100", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := parser.ParseExprQuery(tt.query)
			match, err := obj.Match(query)
			if err != nil {
				t.Fatalf("Match() error = %v", err)
			}
			if match != tt.expected {
				t.Errorf("Match(%s) = %v, want %v", tt.query, match, tt.expected)
			}
		})
	}
}

// TestObjectMatchInOperator tests IN operator
func TestObjectMatchInOperator(t *testing.T) {
	user := NewRow(map[string]any{
		"status": "active",
		"role":   "admin",
		"age":    30,
	})

	obj := NewObject(user)

	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{"IN with string - true", "status IN ('active', 'pending')", true},
		{"IN with string - false", "status IN ('deleted', 'archived')", false},
		{"IN with number - true", "age IN (25, 30, 35)", true},
		{"IN with number - false", "age IN (20, 25, 35)", false},
		{"NOT IN - true", "status NOT IN ('deleted', 'archived')", true},
		{"NOT IN - false", "status NOT IN ('active', 'pending')", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := parser.ParseExprQuery(tt.query)
			match, err := obj.Match(query)
			if err != nil {
				t.Fatalf("Match() error = %v", err)
			}
			if match != tt.expected {
				t.Errorf("Match(%s) = %v, want %v", tt.query, match, tt.expected)
			}
		})
	}
}

// TestObjectMatchNullOperators tests IS NULL and IS NOT NULL
func TestObjectMatchNullOperators(t *testing.T) {
	user := NewRow(map[string]any{
		"name":       "John Doe",
		"middleName": nil,
	})

	obj := NewObject(user)

	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{"IS NULL - true", "middleName IS NULL", true},
		{"IS NULL - false", "name IS NULL", false},
		{"IS NOT NULL - true", "name IS NOT NULL", true},
		{"IS NOT NULL - false", "middleName IS NOT NULL", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := parser.ParseExprQuery(tt.query)
			match, err := obj.Match(query)
			if err != nil {
				t.Fatalf("Match() error = %v", err)
			}
			if match != tt.expected {
				t.Errorf("Match(%s) = %v, want %v", tt.query, match, tt.expected)
			}
		})
	}
}

// TestObjectMatchBetween tests BETWEEN operator
func TestObjectMatchBetween(t *testing.T) {
	user := NewRow(map[string]any{
		"age":   30,
		"score": 85,
	})

	obj := NewObject(user)

	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{"BETWEEN - true", "age BETWEEN 25 AND 35", true},
		{"BETWEEN - false lower", "age BETWEEN 35 AND 40", false},
		{"BETWEEN - false upper", "age BETWEEN 20 AND 25", false},
		{"BETWEEN at boundary", "age BETWEEN 30 AND 40", true},
		{"BETWEEN score true", "score BETWEEN 80 AND 90", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := parser.ParseExprQuery(tt.query)
			match, err := obj.Match(query)
			if err != nil {
				t.Fatalf("Match() error = %v", err)
			}
			if match != tt.expected {
				t.Errorf("Match(%s) = %v, want %v", tt.query, match, tt.expected)
			}
		})
	}
}

// TestObjectMatchContains tests CONTAINS operator
func TestObjectMatchContains(t *testing.T) {
	user := NewRow(map[string]any{
		"tags":  []any{"golang", "python", "javascript"},
		"email": "john.doe@example.com",
	})

	obj := NewObject(user)

	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{"CONTAINS in array - true", "tags CONTAINS 'golang'", true},
		{"CONTAINS in array - false", "tags CONTAINS 'ruby'", false},
		{"CONTAINS in string - true", "email CONTAINS 'example'", true},
		{"CONTAINS in string - false", "email CONTAINS 'test'", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := parser.ParseExprQuery(tt.query)
			match, err := obj.Match(query)
			if err != nil {
				t.Fatalf("Match() error = %v", err)
			}
			if match != tt.expected {
				t.Errorf("Match(%s) = %v, want %v", tt.query, match, tt.expected)
			}
		})
	}
}

// TestObjectMatchNegation tests negation with ! operator
func TestObjectMatchNegation(t *testing.T) {
	user := NewRow(map[string]any{
		"name": "John Doe",
		"age":  30,
	})

	obj := NewObject(user)

	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{"negation simple - true", "!(name == 'Jane Doe')", true},
		{"negation simple - false", "!(name == 'John Doe')", false},
		{"negation complex - true", "!(age < 20 || age > 40)", true},
		{"negation complex - false", "!(age < 20 || age > 25)", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := parser.ParseExprQuery(tt.query)
			match, err := obj.Match(query)
			if err != nil {
				t.Fatalf("Match() error = %v", err)
			}
			if match != tt.expected {
				t.Errorf("Match(%s) = %v, want %v", tt.query, match, tt.expected)
			}
		})
	}
}

// TestObjectMatchNestedFields tests matching with nested field access
func TestObjectMatchNestedFields(t *testing.T) {
	user := NewRow(map[string]any{
		"id":   1,
		"name": "John Doe",
		"location": map[string]any{
			"city":    "New York",
			"state":   "NY",
			"country": "USA",
		},
	})

	obj := NewObject(user)

	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{"nested equality - true", "location.city == 'New York'", true},
		{"nested equality - false", "location.city == 'Boston'", false},
		{"nested with AND - true", "location.city == 'New York' && location.state == 'NY'", true},
		{"nested with AND - false", "location.city == 'New York' && location.state == 'CA'", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := parser.ParseExprQuery(tt.query)
			match, err := obj.Match(query)
			if err != nil {
				t.Fatalf("Match() error = %v", err)
			}
			if match != tt.expected {
				t.Errorf("Match(%s) = %v, want %v", tt.query, match, tt.expected)
			}
		})
	}
}

// TestObjectPseudocodeExample tests the exact example from the pseudocode
func TestObjectPseudocodeExample(t *testing.T) {
	// ROW / RECORD
	user := NewRow(map[string]any{
		"id":    1,
		"name":  "John Doe",
		"email": "john.doe@example.com",
		"age":   30,
		"location": map[string]any{
			"city":    "New York",
			"state":   "NY",
			"country": "USA",
		},
	})

	// Variables
	vars := map[string]any{
		"maxAge": 30,
		"minAge": 18,
	}

	// Methods
	methods := map[string]MethodFunc{
		"upper": func(args ...any) (any, error) {
			if len(args) != 1 {
				return nil, nil
			}
			str, ok := args[0].(string)
			if !ok {
				return nil, nil
			}
			return strings.ToUpper(str), nil
		},
		"lower": func(args ...any) (any, error) {
			if len(args) != 1 {
				return nil, nil
			}
			str, ok := args[0].(string)
			if !ok {
				return nil, nil
			}
			return strings.ToLower(str), nil
		},
	}

	// Schema
	schema := Schema{
		"name":     "string",
		"id":       "int",
		"email":    "string",
		"age":      "int",
		"location": "schema_location",
	}

	// Create the object and load it with data
	obj := NewObjectWithContext(user, schema, methods, vars)

	// Test Get
	id, ok := obj.Get("id")
	if !ok || id != 1 {
		t.Errorf("Get(id) failed: got %v, want 1", id)
	}

	city, ok := obj.Get("location.city")
	if !ok || city != "New York" {
		t.Errorf("Get(location.city) failed: got %v, want 'New York'", city)
	}

	// Test GetType
	emailType := obj.GetType("email")
	if emailType != "string" {
		t.Errorf("GetType(email) = %s, want 'string'", emailType)
	}

	// Test IsOfType
	if !obj.IsOfType("age", "int") {
		t.Error("IsOfType(age, int) should be true")
	}

	// Test IsInt
	if !obj.IsInt("age") {
		t.Error("IsInt(age) should be true")
	}

	// Test Match with a condition
	query := parser.ParseExprQuery("name == upper('test')")
	match, err := obj.Match(query)
	if err != nil {
		t.Fatalf("Match() error = %v", err)
	}
	// Should be false because name is "John Doe" not "TEST"
	if match {
		t.Error("Expected no match for 'name == upper('test')'")
	}

	// Test a matching condition
	query2 := parser.ParseExprQuery("age <= $maxAge && age >= $minAge")
	match2, err := obj.Match(query2)
	if err != nil {
		t.Fatalf("Match() error = %v", err)
	}
	if !match2 {
		t.Error("Expected match for age in range")
	}

	// Test the complex example from the thread
	query3 := parser.ParseExprQuery("name == upper('test') && (age < 20 || age > 40)")
	match3, err := obj.Match(query3)
	if err != nil {
		t.Fatalf("Match() error = %v", err)
	}
	if match3 {
		t.Error("Expected no match for complex condition")
	}
}

/*
 Aggregation:
 	{ $match: 'users > 10 '}
  	{ $sort : 'age desc' },
    { $limit: 10 },
    { project: { "age as howold, name as split(firstname,' ') as firstname"  }},
    { group:{ "count = $sum(1), average = $avg(age), max = $max(age), min = $min(age)"  }}


*/
