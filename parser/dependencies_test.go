package parser

import (
	"testing"
)

// TestExtractQueryDependencies tests the dependency extraction functionality
func TestExtractQueryDependencies(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		wantProps []string
		wantVars  []string
		wantFuncs []string
	}{
		{
			name:      "simple property comparison",
			query:     "age > 25",
			wantProps: []string{"age"},
			wantVars:  []string{},
			wantFuncs: []string{},
		},
		{
			name:      "multiple properties with AND",
			query:     "age > 25 && name == 'John'",
			wantProps: []string{"age", "name"},
			wantVars:  []string{},
			wantFuncs: []string{},
		},
		{
			name:      "nested properties",
			query:     "location.city == 'New York' && location.state == 'NY'",
			wantProps: []string{"location.city", "location.state"},
			wantVars:  []string{},
			wantFuncs: []string{},
		},
		{
			name:      "with variables",
			query:     "age > $minAge && age < $maxAge",
			wantProps: []string{"age"},
			wantVars:  []string{"minAge", "maxAge"},
			wantFuncs: []string{},
		},
		{
			name:      "with functions",
			query:     "upper(name) == 'JOHN' && lower(email) == 'test@example.com'",
			wantProps: []string{"name", "email"},
			wantVars:  []string{},
			wantFuncs: []string{"upper", "lower"},
		},
		{
			name:      "complex query with all types",
			query:     "upper(name) == $targetName && age >= $minAge && location.city == 'NYC'",
			wantProps: []string{"name", "age", "location.city"},
			wantVars:  []string{"targetName", "minAge"},
			wantFuncs: []string{"upper"},
		},
		{
			name:      "arithmetic expressions",
			query:     "age + 10 > 40 && score * 2 >= 100",
			wantProps: []string{"age", "score"},
			wantVars:  []string{},
			wantFuncs: []string{},
		},
		{
			name:      "IN operator",
			query:     "status IN ('active', 'pending') && role == $allowedRole",
			wantProps: []string{"status", "role"},
			wantVars:  []string{"allowedRole"},
			wantFuncs: []string{},
		},
		{
			name:      "BETWEEN operator",
			query:     "age BETWEEN $minAge AND $maxAge",
			wantProps: []string{"age"},
			wantVars:  []string{"minAge", "maxAge"},
			wantFuncs: []string{},
		},
		{
			name:      "nested function calls",
			query:     "len(upper(name)) > 5",
			wantProps: []string{"name"},
			wantVars:  []string{},
			wantFuncs: []string{"len", "upper"},
		},
		{
			name:      "OR with different properties",
			query:     "age < 20 || age > 60 || status == 'premium'",
			wantProps: []string{"age", "status"},
			wantVars:  []string{},
			wantFuncs: []string{},
		},
		{
			name:      "complex nested query",
			query:     "(name == 'John' && age > $minAge) || (upper(status) == $targetStatus && location.city == 'NYC')",
			wantProps: []string{"name", "age", "status", "location.city"},
			wantVars:  []string{"minAge", "targetStatus"},
			wantFuncs: []string{"upper"},
		},
		{
			name:      "duplicate properties",
			query:     "age > 25 && age < 35 && age != 30",
			wantProps: []string{"age"},
			wantVars:  []string{},
			wantFuncs: []string{},
		},
		{
			name:      "duplicate variables",
			query:     "age > $threshold && score > $threshold",
			wantProps: []string{"age", "score"},
			wantVars:  []string{"threshold"},
			wantFuncs: []string{},
		},
		{
			name:      "duplicate functions",
			query:     "upper(name) == 'JOHN' && upper(email) == 'TEST@EXAMPLE.COM'",
			wantProps: []string{"name", "email"},
			wantVars:  []string{},
			wantFuncs: []string{"upper"},
		},
		{
			name:      "empty query conditions",
			query:     "true == true",
			wantProps: []string{},
			wantVars:  []string{},
			wantFuncs: []string{},
		},
		{
			name:      "CONTAINS operator",
			query:     "tags CONTAINS $searchTag",
			wantProps: []string{"tags"},
			wantVars:  []string{"searchTag"},
			wantFuncs: []string{},
		},
		{
			name:      "IS NULL operator",
			query:     "middleName IS NULL && lastName IS NOT NULL",
			wantProps: []string{"middleName", "lastName"},
			wantVars:  []string{},
			wantFuncs: []string{},
		},
		{
			name:      "function with variable argument",
			query:     "upper($dynamicField) == 'VALUE'",
			wantProps: []string{},
			wantVars:  []string{"dynamicField"},
			wantFuncs: []string{"upper"},
		},
		{
			name:      "arithmetic with variables",
			query:     "age + $offset > $threshold",
			wantProps: []string{"age"},
			wantVars:  []string{"offset", "threshold"},
			wantFuncs: []string{},
		},
		{
			name:      "complex arithmetic",
			query:     "price - discount * quantity > $minTotal",
			wantProps: []string{"price", "discount", "quantity"},
			wantVars:  []string{"minTotal"},
			wantFuncs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := ParseExprQuery(tt.query)
			deps := ExtractQueryDependencies(query)

			// Check properties
			if !stringSlicesEqual(deps.Properties, tt.wantProps) {
				t.Errorf("Properties mismatch.\nGot:  %v\nWant: %v", deps.Properties, tt.wantProps)
			}

			// Check variables
			if !stringSlicesEqual(deps.Variables, tt.wantVars) {
				t.Errorf("Variables mismatch.\nGot:  %v\nWant: %v", deps.Variables, tt.wantVars)
			}

			// Check functions
			if !stringSlicesEqual(deps.Functions, tt.wantFuncs) {
				t.Errorf("Functions mismatch.\nGot:  %v\nWant: %v", deps.Functions, tt.wantFuncs)
			}
		})
	}
}

// TestQueryDependenciesUsage demonstrates practical usage of ExtractQueryDependencies
func TestQueryDependenciesUsage(t *testing.T) {
	// Parse a complex query
	query := ParseExprQuery("upper(name) == $targetName && age >= $minAge && location.city == 'NYC'")

	// Extract dependencies
	deps := ExtractQueryDependencies(query)

	// Verify we can prepare context based on dependencies
	t.Run("can identify required properties", func(t *testing.T) {
		// Verify expected properties are extracted
		expectedProps := map[string]bool{
			"name":          true,
			"age":           true,
			"location.city": true,
		}

		for _, prop := range deps.Properties {
			if !expectedProps[prop] {
				t.Errorf("Unexpected property in dependencies: %s", prop)
			}
		}
	})

	t.Run("can identify required variables", func(t *testing.T) {
		vars := map[string]interface{}{
			"targetName": "JOHN DOE",
			"minAge":     18,
		}

		// Check if we have all required variables
		for _, varName := range deps.Variables {
			if _, exists := vars[varName]; !exists {
				t.Errorf("Missing required variable: %s", varName)
			}
		}
	})

	t.Run("can identify required functions", func(t *testing.T) {
		// Verify expected functions are extracted
		expectedFuncs := map[string]bool{
			"upper": true,
		}

		for _, funcName := range deps.Functions {
			if !expectedFuncs[funcName] {
				t.Errorf("Unexpected function in dependencies: %s", funcName)
			}
		}
	})
}

// TestDependenciesPrevalidation tests using dependencies for pre-validation
func TestDependenciesPrevalidation(t *testing.T) {
	query := ParseExprQuery("age >= $minAge && status == 'active' && location.city != ''")
	deps := ExtractQueryDependencies(query)

	// Verify we extracted the expected dependencies
	expectedProps := []string{"age", "status", "location.city"}
	if len(deps.Properties) != len(expectedProps) {
		t.Errorf("Expected %d properties, got %d", len(expectedProps), len(deps.Properties))
	}

	expectedVars := []string{"minAge"}
	if len(deps.Variables) != len(expectedVars) {
		t.Errorf("Expected %d variables, got %d", len(expectedVars), len(deps.Variables))
	}

	// Check that all expected properties are present
	for _, prop := range expectedProps {
		found := false
		for _, depProp := range deps.Properties {
			if depProp == prop {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected property %s not found in dependencies", prop)
		}
	}
}

// TestEmptyDependencies tests queries with no external dependencies
func TestEmptyDependencies(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{
			name:  "only literal values",
			query: "true == true",
		},
		{
			name:  "literal comparisons",
			query: "'hello' == 'hello'",
		},
		{
			name:  "number literals",
			query: "42 > 10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := ParseExprQuery(tt.query)
			deps := ExtractQueryDependencies(query)

			if len(deps.Properties) != 0 {
				t.Errorf("Expected no properties, got %v", deps.Properties)
			}
			if len(deps.Variables) != 0 {
				t.Errorf("Expected no variables, got %v", deps.Variables)
			}
			if len(deps.Functions) != 0 {
				t.Errorf("Expected no functions, got %v", deps.Functions)
			}
		})
	}
}

// Helper function to check if two string slices contain the same elements (order doesn't matter)
func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	// Create maps to count occurrences
	aMap := make(map[string]int)
	bMap := make(map[string]int)

	for _, item := range a {
		aMap[item]++
	}
	for _, item := range b {
		bMap[item]++
	}

	// Compare maps
	for key, count := range aMap {
		if bMap[key] != count {
			return false
		}
	}

	return true
}
