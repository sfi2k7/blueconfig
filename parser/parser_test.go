package parser

import (
	"encoding/json"
	"testing"
)

// Helper function to compare queries
func queriesEqual(t *testing.T, expected, actual Query) bool {
	expectedJSON, _ := json.MarshalIndent(expected, "", "  ")
	actualJSON, _ := json.MarshalIndent(actual, "", "  ")

	if string(expectedJSON) != string(actualJSON) {
		t.Errorf("Queries do not match.\nExpected:\n%s\n\nActual:\n%s", expectedJSON, actualJSON)
		return false
	}
	return true
}

// TestBasicComparisons tests simple comparison operations
func TestBasicComparisons(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected Query
	}{
		{
			name:  "equality with string",
			query: "name == 'John'",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:    "==",
						Left:  &ConditionTerm{Property: "name"},
						Right: &ConditionTerm{Value: "John"},
					},
				},
			},
		},
		{
			name:  "inequality with string",
			query: "name != 'John'",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:     "==",
						Left:   &ConditionTerm{Property: "name"},
						Right:  &ConditionTerm{Value: "John"},
						Negate: true,
					},
				},
			},
		},
		{
			name:  "greater than with number",
			query: "age > 25",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:    ">",
						Left:  &ConditionTerm{Property: "age"},
						Right: &ConditionTerm{Value: float64(25)},
					},
				},
			},
		},
		{
			name:  "less than with float",
			query: "price < 99.99",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:    "<",
						Left:  &ConditionTerm{Property: "price"},
						Right: &ConditionTerm{Value: 99.99},
					},
				},
			},
		},
		{
			name:  "greater than or equal",
			query: "score >= 100",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:    ">=",
						Left:  &ConditionTerm{Property: "score"},
						Right: &ConditionTerm{Value: float64(100)},
					},
				},
			},
		},
		{
			name:  "less than or equal",
			query: "count <= 50",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:    "<=",
						Left:  &ConditionTerm{Property: "count"},
						Right: &ConditionTerm{Value: float64(50)},
					},
				},
			},
		},
		{
			name:  "boolean true",
			query: "active == true",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:    "==",
						Left:  &ConditionTerm{Property: "active"},
						Right: &ConditionTerm{Value: true},
					},
				},
			},
		},
		{
			name:  "boolean false",
			query: "deleted == false",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:    "==",
						Left:  &ConditionTerm{Property: "deleted"},
						Right: &ConditionTerm{Value: false},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseExprQuery(tt.query)
			queriesEqual(t, tt.expected, result)
		})
	}
}

// TestAndOperations tests AND logic
func TestAndOperations(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected Query
	}{
		{
			name:  "two conditions with AND",
			query: "age > 18 && status == 'active'",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:    ">",
						Left:  &ConditionTerm{Property: "age"},
						Right: &ConditionTerm{Value: float64(18)},
					},
					{
						Op:    "==",
						Left:  &ConditionTerm{Property: "status"},
						Right: &ConditionTerm{Value: "active"},
					},
				},
			},
		},
		{
			name:  "three conditions with AND",
			query: "age > 18 && age < 65 && status == 'active'",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:    ">",
						Left:  &ConditionTerm{Property: "age"},
						Right: &ConditionTerm{Value: float64(18)},
					},
					{
						Op:    "<",
						Left:  &ConditionTerm{Property: "age"},
						Right: &ConditionTerm{Value: float64(65)},
					},
					{
						Op:    "==",
						Left:  &ConditionTerm{Property: "status"},
						Right: &ConditionTerm{Value: "active"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseExprQuery(tt.query)
			queriesEqual(t, tt.expected, result)
		})
	}
}

// TestOrOperations tests OR logic
func TestOrOperations(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected Query
	}{
		{
			name:  "two conditions with OR",
			query: "age < 18 || age > 65",
			expected: Query{
				IsOr: true,
				SubQueries: []Query{
					{
						IsOr: false,
						Conditions: []Condition{
							{
								Op:    "<",
								Left:  &ConditionTerm{Property: "age"},
								Right: &ConditionTerm{Value: float64(18)},
							},
						},
					},
					{
						IsOr: false,
						Conditions: []Condition{
							{
								Op:    ">",
								Left:  &ConditionTerm{Property: "age"},
								Right: &ConditionTerm{Value: float64(65)},
							},
						},
					},
				},
			},
		},
		{
			name:  "three conditions with OR",
			query: "status == 'pending' || status == 'active' || status == 'completed'",
			expected: Query{
				IsOr: true,
				SubQueries: []Query{
					{
						IsOr: false,
						Conditions: []Condition{
							{
								Op:    "==",
								Left:  &ConditionTerm{Property: "status"},
								Right: &ConditionTerm{Value: "pending"},
							},
						},
					},
					{
						IsOr: false,
						Conditions: []Condition{
							{
								Op:    "==",
								Left:  &ConditionTerm{Property: "status"},
								Right: &ConditionTerm{Value: "active"},
							},
						},
					},
					{
						IsOr: false,
						Conditions: []Condition{
							{
								Op:    "==",
								Left:  &ConditionTerm{Property: "status"},
								Right: &ConditionTerm{Value: "completed"},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseExprQuery(tt.query)
			queriesEqual(t, tt.expected, result)
		})
	}
}

// TestNotOperations tests negation
func TestNotOperations(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected Query
	}{
		{
			name:  "simple NOT",
			query: "!(age > 18)",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:     ">",
						Left:   &ConditionTerm{Property: "age"},
						Right:  &ConditionTerm{Value: float64(18)},
						Negate: true,
					},
				},
			},
		},
		{
			name:  "NOT with AND - De Morgan's law",
			query: "!(age > 18 && status == 'active')",
			expected: Query{
				IsOr: true,
				SubQueries: []Query{
					{
						IsOr: false,
						Conditions: []Condition{
							{
								Op:     ">",
								Left:   &ConditionTerm{Property: "age"},
								Right:  &ConditionTerm{Value: float64(18)},
								Negate: true,
							},
						},
					},
					{
						IsOr: false,
						Conditions: []Condition{
							{
								Op:     "==",
								Left:   &ConditionTerm{Property: "status"},
								Right:  &ConditionTerm{Value: "active"},
								Negate: true,
							},
						},
					},
				},
			},
		},
		{
			name:  "NOT with OR - De Morgan's law",
			query: "!(age < 18 || age > 65)",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:     "<",
						Left:   &ConditionTerm{Property: "age"},
						Right:  &ConditionTerm{Value: float64(18)},
						Negate: true,
					},
					{
						Op:     ">",
						Left:   &ConditionTerm{Property: "age"},
						Right:  &ConditionTerm{Value: float64(65)},
						Negate: true,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseExprQuery(tt.query)
			queriesEqual(t, tt.expected, result)
		})
	}
}

// TestComplexQueries tests complex nested queries
func TestComplexQueries(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected Query
	}{
		{
			name:  "AND with OR precedence",
			query: "(age > 18 && age < 65) || status == 'vip'",
			expected: Query{
				IsOr: true,
				SubQueries: []Query{
					{
						IsOr: false,
						Conditions: []Condition{
							{
								Op:    ">",
								Left:  &ConditionTerm{Property: "age"},
								Right: &ConditionTerm{Value: float64(18)},
							},
							{
								Op:    "<",
								Left:  &ConditionTerm{Property: "age"},
								Right: &ConditionTerm{Value: float64(65)},
							},
						},
					},
					{
						IsOr: false,
						Conditions: []Condition{
							{
								Op:    "==",
								Left:  &ConditionTerm{Property: "status"},
								Right: &ConditionTerm{Value: "vip"},
							},
						},
					},
				},
			},
		},
		{
			name:  "nested OR and AND",
			query: "status == 'pending' || (status == 'active' && age > 18)",
			expected: Query{
				IsOr: true,
				SubQueries: []Query{
					{
						IsOr: false,
						Conditions: []Condition{
							{
								Op:    "==",
								Left:  &ConditionTerm{Property: "status"},
								Right: &ConditionTerm{Value: "pending"},
							},
						},
					},
					{
						IsOr: false,
						Conditions: []Condition{
							{
								Op:    "==",
								Left:  &ConditionTerm{Property: "status"},
								Right: &ConditionTerm{Value: "active"},
							},
							{
								Op:    ">",
								Left:  &ConditionTerm{Property: "age"},
								Right: &ConditionTerm{Value: float64(18)},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseExprQuery(tt.query)
			queriesEqual(t, tt.expected, result)
		})
	}
}

// TestFunctions tests function calls
func TestFunctions(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected Query
	}{
		{
			name:  "function as boolean predicate",
			query: "HasPrefix(name, 'Dr')",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op: "func_predicate",
						Left: &ConditionTerm{
							Function: &FunctionCall{
								Name: "HasPrefix",
								Args: []*ConditionTerm{
									{Property: "name"},
									{Value: "Dr"},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "function with comparison operator",
			query: "len(name) > 5",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op: ">",
						Left: &ConditionTerm{
							Function: &FunctionCall{
								Name: "len",
								Args: []*ConditionTerm{
									{Property: "name"},
								},
							},
						},
						Right: &ConditionTerm{Value: float64(5)},
					},
				},
			},
		},
		{
			name:  "function with equality",
			query: "len(items) == 10",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op: "==",
						Left: &ConditionTerm{
							Function: &FunctionCall{
								Name: "len",
								Args: []*ConditionTerm{
									{Property: "items"},
								},
							},
						},
						Right: &ConditionTerm{Value: float64(10)},
					},
				},
			},
		},
		{
			name:  "function in AND expression",
			query: "age > 18 && HasPrefix(name, 'Dr')",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:    ">",
						Left:  &ConditionTerm{Property: "age"},
						Right: &ConditionTerm{Value: float64(18)},
					},
					{
						Op: "func_predicate",
						Left: &ConditionTerm{
							Function: &FunctionCall{
								Name: "HasPrefix",
								Args: []*ConditionTerm{
									{Property: "name"},
									{Value: "Dr"},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseExprQuery(tt.query)
			queriesEqual(t, tt.expected, result)
		})
	}
}

// TestVariables tests variable references
func TestVariables(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected Query
	}{
		{
			name:  "variable in comparison",
			query: "age > $minAge",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:    ">",
						Left:  &ConditionTerm{Property: "age"},
						Right: &ConditionTerm{Variable: &VariableRef{Name: "minAge"}},
					},
				},
			},
		},
		{
			name:  "variable with equality",
			query: "status == $targetStatus",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:    "==",
						Left:  &ConditionTerm{Property: "status"},
						Right: &ConditionTerm{Variable: &VariableRef{Name: "targetStatus"}},
					},
				},
			},
		},
		{
			name:  "multiple variables",
			query: "age > $minAge && age < $maxAge",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:    ">",
						Left:  &ConditionTerm{Property: "age"},
						Right: &ConditionTerm{Variable: &VariableRef{Name: "minAge"}},
					},
					{
						Op:    "<",
						Left:  &ConditionTerm{Property: "age"},
						Right: &ConditionTerm{Variable: &VariableRef{Name: "maxAge"}},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseExprQuery(tt.query)
			queriesEqual(t, tt.expected, result)
		})
	}
}

// TestUseStatement tests USE collection statements
func TestUseStatement(t *testing.T) {
	tests := []struct {
		name               string
		query              string
		expectedQuery      Query
		expectedCollection string
	}{
		{
			name:  "USE with simple query",
			query: "USE users; age > 18",
			expectedQuery: Query{
				IsOr:       false,
				Collection: "users",
				Conditions: []Condition{
					{
						Op:    ">",
						Left:  &ConditionTerm{Property: "age"},
						Right: &ConditionTerm{Value: float64(18)},
					},
				},
			},
			expectedCollection: "users",
		},
		{
			name:  "USE with complex query",
			query: "USE orders; status == 'pending' && total > 100",
			expectedQuery: Query{
				IsOr:       false,
				Collection: "orders",
				Conditions: []Condition{
					{
						Op:    "==",
						Left:  &ConditionTerm{Property: "status"},
						Right: &ConditionTerm{Value: "pending"},
					},
					{
						Op:    ">",
						Left:  &ConditionTerm{Property: "total"},
						Right: &ConditionTerm{Value: float64(100)},
					},
				},
			},
			expectedCollection: "orders",
		},
		{
			name:  "no USE statement defaults to 'default'",
			query: "age > 18",
			expectedQuery: Query{
				IsOr:       false,
				Collection: "default",
				Conditions: []Condition{
					{
						Op:    ">",
						Left:  &ConditionTerm{Property: "age"},
						Right: &ConditionTerm{Value: float64(18)},
					},
				},
			},
			expectedCollection: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, collection := ParseExprQueryWithCollection(tt.query)
			if collection != tt.expectedCollection {
				t.Errorf("Expected collection %s, got %s", tt.expectedCollection, collection)
			}
			queriesEqual(t, tt.expectedQuery, result)
		})
	}
}

// TestDoubleQuotedStrings tests string parsing with double quotes
func TestDoubleQuotedStrings(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected Query
	}{
		{
			name:  "double quoted string",
			query: `name == "John Doe"`,
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:    "==",
						Left:  &ConditionTerm{Property: "name"},
						Right: &ConditionTerm{Value: "John Doe"},
					},
				},
			},
		},
		{
			name:  "single quoted string",
			query: `name == 'Jane Doe'`,
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:    "==",
						Left:  &ConditionTerm{Property: "name"},
						Right: &ConditionTerm{Value: "Jane Doe"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseExprQuery(tt.query)
			queriesEqual(t, tt.expected, result)
		})
	}
}

// TestBarewordIdentifiers tests bareword identifiers as properties
func TestBarewordIdentifiers(t *testing.T) {
	query := "status == active"
	expected := Query{
		IsOr: false,
		Conditions: []Condition{
			{
				Op:    "==",
				Left:  &ConditionTerm{Property: "status"},
				Right: &ConditionTerm{Property: "active"},
			},
		},
	}
	result := ParseExprQuery(query)
	queriesEqual(t, expected, result)
}

// TestParseErrors tests that invalid queries panic with errors
func TestParseErrors(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{
			name:  "incomplete comparison",
			query: "age >",
		},
		{
			name:  "missing operator",
			query: "age 18",
		},
		{
			name:  "unbalanced parentheses",
			query: "(age > 18",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("Expected panic for invalid query: %s", tt.query)
				}
			}()
			ParseExprQuery(tt.query)
		})
	}
}

// TestFunctionWithVariables tests functions that use variables in arguments
func TestFunctionWithVariables(t *testing.T) {
	query := "len(name) > $maxLength"
	expected := Query{
		IsOr: false,
		Conditions: []Condition{
			{
				Op: ">",
				Left: &ConditionTerm{
					Function: &FunctionCall{
						Name: "len",
						Args: []*ConditionTerm{
							{Property: "name"},
						},
					},
				},
				Right: &ConditionTerm{Variable: &VariableRef{Name: "maxLength"}},
			},
		},
	}
	result := ParseExprQuery(query)
	queriesEqual(t, expected, result)
}

// BenchmarkSimpleQuery benchmarks a simple query parse
func BenchmarkSimpleQuery(b *testing.B) {
	query := "age > 18 && status == 'active'"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseExprQuery(query)
	}
}

// BenchmarkComplexQuery benchmarks a complex query parse
func BenchmarkComplexQuery(b *testing.B) {
	query := "(age > 18 && age < 65) || (status == 'vip' && !(deleted == true))"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseExprQuery(query)
	}
}

// BenchmarkFunctionQuery benchmarks a query with functions
func BenchmarkFunctionQuery(b *testing.B) {
	query := "HasPrefix(name, 'Dr') && len(items) > 5"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseExprQuery(query)
	}
}

// TestFunctionsOnRightSide tests functions on the right side of comparisons
func TestFunctionsOnRightSide(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected Query
	}{
		{
			name:  "function on right side of equality",
			query: "name == upper('test')",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:   "==",
						Left: &ConditionTerm{Property: "name"},
						Right: &ConditionTerm{
							Function: &FunctionCall{
								Name: "upper",
								Args: []*ConditionTerm{
									{Value: "test"},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "function on right with variable argument",
			query: "status == lower($inputStatus)",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:   "==",
						Left: &ConditionTerm{Property: "status"},
						Right: &ConditionTerm{
							Function: &FunctionCall{
								Name: "lower",
								Args: []*ConditionTerm{
									{Variable: &VariableRef{Name: "inputStatus"}},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "function on right in AND expression",
			query: "name == upper('test') && age >= 18",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:   "==",
						Left: &ConditionTerm{Property: "name"},
						Right: &ConditionTerm{
							Function: &FunctionCall{
								Name: "upper",
								Args: []*ConditionTerm{
									{Value: "test"},
								},
							},
						},
					},
					{
						Op:    ">=",
						Left:  &ConditionTerm{Property: "age"},
						Right: &ConditionTerm{Value: float64(18)},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseExprQuery(tt.query)
			queriesEqual(t, tt.expected, result)
		})
	}
}

// TestVariablesOnLeftSide tests variables on the left side of comparisons
func TestVariablesOnLeftSide(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected Query
	}{
		{
			name:  "variable on left side",
			query: "$minAge < age",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:    "<",
						Left:  &ConditionTerm{Variable: &VariableRef{Name: "minAge"}},
						Right: &ConditionTerm{Property: "age"},
					},
				},
			},
		},
		{
			name:  "variable on left, variable on right",
			query: "$minValue <= $maxValue",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:    "<=",
						Left:  &ConditionTerm{Variable: &VariableRef{Name: "minValue"}},
						Right: &ConditionTerm{Variable: &VariableRef{Name: "maxValue"}},
					},
				},
			},
		},
		{
			name:  "range check with variables",
			query: "$minAge < age && age < $maxAge",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:    "<",
						Left:  &ConditionTerm{Variable: &VariableRef{Name: "minAge"}},
						Right: &ConditionTerm{Property: "age"},
					},
					{
						Op:    "<",
						Left:  &ConditionTerm{Property: "age"},
						Right: &ConditionTerm{Variable: &VariableRef{Name: "maxAge"}},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseExprQuery(tt.query)
			queriesEqual(t, tt.expected, result)
		})
	}
}

// TestComplexFunctionAndVariableCombinations tests complex scenarios
func TestComplexFunctionAndVariableCombinations(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected Query
	}{
		{
			name:  "original user query",
			query: "(name == upper('test') && age >= $age)",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:   "==",
						Left: &ConditionTerm{Property: "name"},
						Right: &ConditionTerm{
							Function: &FunctionCall{
								Name: "upper",
								Args: []*ConditionTerm{
									{Value: "test"},
								},
							},
						},
					},
					{
						Op:    ">=",
						Left:  &ConditionTerm{Property: "age"},
						Right: &ConditionTerm{Variable: &VariableRef{Name: "age"}},
					},
				},
			},
		},
		{
			name:  "function with variable in arguments",
			query: "len(items) > $minCount",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op: ">",
						Left: &ConditionTerm{
							Function: &FunctionCall{
								Name: "len",
								Args: []*ConditionTerm{
									{Property: "items"},
								},
							},
						},
						Right: &ConditionTerm{Variable: &VariableRef{Name: "minCount"}},
					},
				},
			},
		},
		{
			name:  "function comparing to function",
			query: "upper(name) == upper($targetName)",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op: "==",
						Left: &ConditionTerm{
							Function: &FunctionCall{
								Name: "upper",
								Args: []*ConditionTerm{
									{Property: "name"},
								},
							},
						},
						Right: &ConditionTerm{
							Function: &FunctionCall{
								Name: "upper",
								Args: []*ConditionTerm{
									{Variable: &VariableRef{Name: "targetName"}},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseExprQuery(tt.query)
			queriesEqual(t, tt.expected, result)
		})
	}
}

// TestNestedProperties tests nested property paths
func TestNestedProperties(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected Query
	}{
		{
			name:  "single level nested property",
			query: "user.age > 18",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:    ">",
						Left:  &ConditionTerm{Property: "user.age"},
						Right: &ConditionTerm{Value: float64(18)},
					},
				},
			},
		},
		{
			name:  "multi level nested property",
			query: "person.address.city == 'NYC'",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:    "==",
						Left:  &ConditionTerm{Property: "person.address.city"},
						Right: &ConditionTerm{Value: "NYC"},
					},
				},
			},
		},
		{
			name:  "nested property in function",
			query: "len(user.name) > 5",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op: ">",
						Left: &ConditionTerm{
							Function: &FunctionCall{
								Name: "len",
								Args: []*ConditionTerm{
									{Property: "user.name"},
								},
							},
						},
						Right: &ConditionTerm{Value: float64(5)},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseExprQuery(tt.query)
			queriesEqual(t, tt.expected, result)
		})
	}
}

// TestNestedVariables tests nested variable references
func TestNestedVariables(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected Query
	}{
		{
			name:  "single level nested variable",
			query: "age > $user.minAge",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:    ">",
						Left:  &ConditionTerm{Property: "age"},
						Right: &ConditionTerm{Variable: &VariableRef{Name: "user.minAge"}},
					},
				},
			},
		},
		{
			name:  "multi level nested variable",
			query: "salary >= $config.min.salary",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:    ">=",
						Left:  &ConditionTerm{Property: "salary"},
						Right: &ConditionTerm{Variable: &VariableRef{Name: "config.min.salary"}},
					},
				},
			},
		},
		{
			name:  "nested variable in function",
			query: "upper($user.name) == 'JOHN'",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op: "==",
						Left: &ConditionTerm{
							Function: &FunctionCall{
								Name: "upper",
								Args: []*ConditionTerm{
									{Variable: &VariableRef{Name: "user.name"}},
								},
							},
						},
						Right: &ConditionTerm{Value: "JOHN"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseExprQuery(tt.query)
			queriesEqual(t, tt.expected, result)
		})
	}
}

// TestNestedCombinations tests combinations of nested properties, variables, and functions
func TestNestedCombinations(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected Query
	}{
		{
			name:  "nested property with nested variable",
			query: "user.age >= $config.minAge",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:    ">=",
						Left:  &ConditionTerm{Property: "user.age"},
						Right: &ConditionTerm{Variable: &VariableRef{Name: "config.minAge"}},
					},
				},
			},
		},
		{
			name:  "nested property compared to function with nested variable",
			query: "user.name == upper($defaults.name)",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:   "==",
						Left: &ConditionTerm{Property: "user.name"},
						Right: &ConditionTerm{
							Function: &FunctionCall{
								Name: "upper",
								Args: []*ConditionTerm{
									{Variable: &VariableRef{Name: "defaults.name"}},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "function chain with nested properties",
			query: "len(upper(user.firstName)) > 3",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op: ">",
						Left: &ConditionTerm{
							Function: &FunctionCall{
								Name: "len",
								Args: []*ConditionTerm{
									{
										Function: &FunctionCall{
											Name: "upper",
											Args: []*ConditionTerm{
												{Property: "user.firstName"},
											},
										},
									},
								},
							},
						},
						Right: &ConditionTerm{Value: float64(3)},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseExprQuery(tt.query)
			queriesEqual(t, tt.expected, result)
		})
	}
}

// TestInOperator tests the IN operator
func TestInOperator(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected Query
	}{
		{
			name:  "IN with variable",
			query: "status IN $statuses",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:    "IN",
						Left:  &ConditionTerm{Property: "status"},
						Right: &ConditionTerm{Variable: &VariableRef{Name: "statuses"}},
					},
				},
			},
		},
		{
			name:  "IN with string values",
			query: "status IN ('active', 'pending', 'approved')",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:   "IN",
						Left: &ConditionTerm{Property: "status"},
						InValues: []*ConditionTerm{
							{Value: "active"},
							{Value: "pending"},
							{Value: "approved"},
						},
					},
				},
			},
		},
		{
			name:  "IN with numeric values",
			query: "age IN (18, 21, 25, 30)",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:   "IN",
						Left: &ConditionTerm{Property: "age"},
						InValues: []*ConditionTerm{
							{Value: float64(18)},
							{Value: float64(21)},
							{Value: float64(25)},
							{Value: float64(30)},
						},
					},
				},
			},
		},
		{
			name:  "NOT IN with variable",
			query: "id NOT IN $excludedIds",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:    "NOT_IN",
						Left:  &ConditionTerm{Property: "id"},
						Right: &ConditionTerm{Variable: &VariableRef{Name: "excludedIds"}},
					},
				},
			},
		},
		{
			name:  "NOT IN operator",
			query: "status NOT IN ('deleted', 'archived')",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:   "NOT_IN",
						Left: &ConditionTerm{Property: "status"},
						InValues: []*ConditionTerm{
							{Value: "deleted"},
							{Value: "archived"},
						},
					},
				},
			},
		},
		{
			name:  "IN with AND operator",
			query: "status IN ('active', 'pending') && age > 18",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:   "IN",
						Left: &ConditionTerm{Property: "status"},
						InValues: []*ConditionTerm{
							{Value: "active"},
							{Value: "pending"},
						},
					},
					{
						Op:    ">",
						Left:  &ConditionTerm{Property: "age"},
						Right: &ConditionTerm{Value: float64(18)},
					},
				},
			},
		},
		{
			name:  "IN with nested property",
			query: "user.role IN ('admin', 'moderator')",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:   "IN",
						Left: &ConditionTerm{Property: "user.role"},
						InValues: []*ConditionTerm{
							{Value: "admin"},
							{Value: "moderator"},
						},
					},
				},
			},
		},
		{
			name:  "IN with OR operator",
			query: "status IN ('active', 'pending') || priority > 5",
			expected: Query{
				IsOr: true,
				SubQueries: []Query{
					{
						IsOr: false,
						Conditions: []Condition{
							{
								Op:   "IN",
								Left: &ConditionTerm{Property: "status"},
								InValues: []*ConditionTerm{
									{Value: "active"},
									{Value: "pending"},
								},
							},
						},
					},
					{
						IsOr: false,
						Conditions: []Condition{
							{
								Op:    ">",
								Left:  &ConditionTerm{Property: "priority"},
								Right: &ConditionTerm{Value: float64(5)},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseExprQuery(tt.query)
			queriesEqual(t, tt.expected, result)
		})
	}
}

// TestNullChecks tests IS NULL and IS NOT NULL operators
func TestNullChecks(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected Query
	}{
		{
			name:  "IS NULL",
			query: "email IS NULL",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:   "IS_NULL",
						Left: &ConditionTerm{Property: "email"},
					},
				},
			},
		},
		{
			name:  "IS NOT NULL",
			query: "deletedAt IS NOT NULL",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:   "IS_NOT_NULL",
						Left: &ConditionTerm{Property: "deletedAt"},
					},
				},
			},
		},
		{
			name:  "nested property IS NULL",
			query: "user.profile IS NULL",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:   "IS_NULL",
						Left: &ConditionTerm{Property: "user.profile"},
					},
				},
			},
		},
		{
			name:  "IS NULL with AND",
			query: "email IS NULL && active == true",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:   "IS_NULL",
						Left: &ConditionTerm{Property: "email"},
					},
					{
						Op:    "==",
						Left:  &ConditionTerm{Property: "active"},
						Right: &ConditionTerm{Value: true},
					},
				},
			},
		},
		{
			name:  "IS NOT NULL with OR",
			query: "deletedAt IS NOT NULL || archived == false",
			expected: Query{
				IsOr: true,
				SubQueries: []Query{
					{
						IsOr: false,
						Conditions: []Condition{
							{
								Op:   "IS_NOT_NULL",
								Left: &ConditionTerm{Property: "deletedAt"},
							},
						},
					},
					{
						IsOr: false,
						Conditions: []Condition{
							{
								Op:    "==",
								Left:  &ConditionTerm{Property: "archived"},
								Right: &ConditionTerm{Value: false},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseExprQuery(tt.query)
			queriesEqual(t, tt.expected, result)
		})
	}
}

// TestLikeOperator tests the LIKE pattern matching operator
func TestLikeOperator(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected Query
	}{
		{
			name:  "LIKE with contains pattern",
			query: "name LIKE '%john%'",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:      "LIKE",
						Left:    &ConditionTerm{Property: "name"},
						Pattern: "%john%",
					},
				},
			},
		},
		{
			name:  "LIKE with prefix pattern",
			query: "email LIKE 'admin@%'",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:      "LIKE",
						Left:    &ConditionTerm{Property: "email"},
						Pattern: "admin@%",
					},
				},
			},
		},
		{
			name:  "LIKE with underscore wildcard",
			query: "code LIKE 'PRD-____'",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:      "LIKE",
						Left:    &ConditionTerm{Property: "code"},
						Pattern: "PRD-____",
					},
				},
			},
		},
		{
			name:  "LIKE with AND operator",
			query: "name LIKE '%smith%' && age > 18",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:      "LIKE",
						Left:    &ConditionTerm{Property: "name"},
						Pattern: "%smith%",
					},
					{
						Op:    ">",
						Left:  &ConditionTerm{Property: "age"},
						Right: &ConditionTerm{Value: float64(18)},
					},
				},
			},
		},
		{
			name:  "LIKE with OR operator",
			query: "email LIKE '%@gmail.com' || email LIKE '%@yahoo.com'",
			expected: Query{
				IsOr: true,
				SubQueries: []Query{
					{
						IsOr: false,
						Conditions: []Condition{
							{
								Op:      "LIKE",
								Left:    &ConditionTerm{Property: "email"},
								Pattern: "%@gmail.com",
							},
						},
					},
					{
						IsOr: false,
						Conditions: []Condition{
							{
								Op:      "LIKE",
								Left:    &ConditionTerm{Property: "email"},
								Pattern: "%@yahoo.com",
							},
						},
					},
				},
			},
		},
		{
			name:  "LIKE with nested property",
			query: "user.name LIKE 'Dr.%'",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:      "LIKE",
						Left:    &ConditionTerm{Property: "user.name"},
						Pattern: "Dr.%",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseExprQuery(tt.query)
			queriesEqual(t, tt.expected, result)
		})
	}
}

// TestBetweenOperator tests the BETWEEN operator
func TestBetweenOperator(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected Query
	}{
		{
			name:  "BETWEEN with integers",
			query: "age BETWEEN 18 AND 65",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:    "BETWEEN",
						Left:  &ConditionTerm{Property: "age"},
						Start: &ConditionTerm{Value: float64(18)},
						End:   &ConditionTerm{Value: float64(65)},
					},
				},
			},
		},
		{
			name:  "BETWEEN with floats",
			query: "price BETWEEN 10.99 AND 99.99",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:    "BETWEEN",
						Left:  &ConditionTerm{Property: "price"},
						Start: &ConditionTerm{Value: 10.99},
						End:   &ConditionTerm{Value: 99.99},
					},
				},
			},
		},
		{
			name:  "BETWEEN with variables",
			query: "price BETWEEN $minPrice AND $maxPrice",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:    "BETWEEN",
						Left:  &ConditionTerm{Property: "price"},
						Start: &ConditionTerm{Variable: &VariableRef{Name: "minPrice"}},
						End:   &ConditionTerm{Variable: &VariableRef{Name: "maxPrice"}},
					},
				},
			},
		},
		{
			name:  "BETWEEN with AND operator",
			query: "age BETWEEN 18 AND 65 && status == 'active'",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:    "BETWEEN",
						Left:  &ConditionTerm{Property: "age"},
						Start: &ConditionTerm{Value: float64(18)},
						End:   &ConditionTerm{Value: float64(65)},
					},
					{
						Op:    "==",
						Left:  &ConditionTerm{Property: "status"},
						Right: &ConditionTerm{Value: "active"},
					},
				},
			},
		},
		{
			name:  "BETWEEN with nested property",
			query: "user.age BETWEEN 21 AND 30",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:    "BETWEEN",
						Left:  &ConditionTerm{Property: "user.age"},
						Start: &ConditionTerm{Value: float64(21)},
						End:   &ConditionTerm{Value: float64(30)},
					},
				},
			},
		},
		{
			name:  "BETWEEN with OR operator",
			query: "age BETWEEN 18 AND 25 || age BETWEEN 60 AND 70",
			expected: Query{
				IsOr: true,
				SubQueries: []Query{
					{
						IsOr: false,
						Conditions: []Condition{
							{
								Op:    "BETWEEN",
								Left:  &ConditionTerm{Property: "age"},
								Start: &ConditionTerm{Value: float64(18)},
								End:   &ConditionTerm{Value: float64(25)},
							},
						},
					},
					{
						IsOr: false,
						Conditions: []Condition{
							{
								Op:    "BETWEEN",
								Left:  &ConditionTerm{Property: "age"},
								Start: &ConditionTerm{Value: float64(60)},
								End:   &ConditionTerm{Value: float64(70)},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseExprQuery(tt.query)
			queriesEqual(t, tt.expected, result)
		})
	}
}

// TestContainsOperator tests the CONTAINS operator for array/collection operations
func TestContainsOperator(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected Query
	}{
		{
			name:  "CONTAINS with string value",
			query: "tags CONTAINS 'urgent'",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:    "CONTAINS",
						Left:  &ConditionTerm{Property: "tags"},
						Right: &ConditionTerm{Value: "urgent"},
					},
				},
			},
		},
		{
			name:  "CONTAINS with numeric value",
			query: "items CONTAINS 123",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:    "CONTAINS",
						Left:  &ConditionTerm{Property: "items"},
						Right: &ConditionTerm{Value: float64(123)},
					},
				},
			},
		},
		{
			name:  "CONTAINS with variable",
			query: "roles CONTAINS $userRole",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:    "CONTAINS",
						Left:  &ConditionTerm{Property: "roles"},
						Right: &ConditionTerm{Variable: &VariableRef{Name: "userRole"}},
					},
				},
			},
		},
		{
			name:  "CONTAINS with AND operator",
			query: "tags CONTAINS 'urgent' && status == 'active'",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:    "CONTAINS",
						Left:  &ConditionTerm{Property: "tags"},
						Right: &ConditionTerm{Value: "urgent"},
					},
					{
						Op:    "==",
						Left:  &ConditionTerm{Property: "status"},
						Right: &ConditionTerm{Value: "active"},
					},
				},
			},
		},
		{
			name:  "CONTAINS with nested property",
			query: "user.tags CONTAINS 'featured'",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:    "CONTAINS",
						Left:  &ConditionTerm{Property: "user.tags"},
						Right: &ConditionTerm{Value: "featured"},
					},
				},
			},
		},
		{
			name:  "CONTAINS with OR operator",
			query: "tags CONTAINS 'urgent' || tags CONTAINS 'critical'",
			expected: Query{
				IsOr: true,
				SubQueries: []Query{
					{
						IsOr: false,
						Conditions: []Condition{
							{
								Op:    "CONTAINS",
								Left:  &ConditionTerm{Property: "tags"},
								Right: &ConditionTerm{Value: "urgent"},
							},
						},
					},
					{
						IsOr: false,
						Conditions: []Condition{
							{
								Op:    "CONTAINS",
								Left:  &ConditionTerm{Property: "tags"},
								Right: &ConditionTerm{Value: "critical"},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseExprQuery(tt.query)
			queriesEqual(t, tt.expected, result)
		})
	}
}

// TestAnyOfOperator tests the ANY_OF operator for array membership checks
func TestAnyOfOperator(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected Query
	}{
		{
			name:  "ANY_OF with variable",
			query: "id ANY_OF $ids",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:    "ANY_OF",
						Left:  &ConditionTerm{Property: "id"},
						Right: &ConditionTerm{Variable: &VariableRef{Name: "ids"}},
					},
				},
			},
		},
		{
			name:  "ANY_OF with string values",
			query: "roles ANY_OF ('admin', 'moderator', 'owner')",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:   "ANY_OF",
						Left: &ConditionTerm{Property: "roles"},
						InValues: []*ConditionTerm{
							{Value: "admin"},
							{Value: "moderator"},
							{Value: "owner"},
						},
					},
				},
			},
		},
		{
			name:  "ANY_OF with numeric values",
			query: "priority ANY_OF (1, 2, 3)",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:   "ANY_OF",
						Left: &ConditionTerm{Property: "priority"},
						InValues: []*ConditionTerm{
							{Value: float64(1)},
							{Value: float64(2)},
							{Value: float64(3)},
						},
					},
				},
			},
		},
		{
			name:  "ANY_OF with AND operator",
			query: "roles ANY_OF ('admin', 'moderator') && active == true",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:   "ANY_OF",
						Left: &ConditionTerm{Property: "roles"},
						InValues: []*ConditionTerm{
							{Value: "admin"},
							{Value: "moderator"},
						},
					},
					{
						Op:    "==",
						Left:  &ConditionTerm{Property: "active"},
						Right: &ConditionTerm{Value: true},
					},
				},
			},
		},
		{
			name:  "ANY_OF with nested property",
			query: "user.roles ANY_OF ('admin', 'superuser')",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:   "ANY_OF",
						Left: &ConditionTerm{Property: "user.roles"},
						InValues: []*ConditionTerm{
							{Value: "admin"},
							{Value: "superuser"},
						},
					},
				},
			},
		},
		{
			name:  "ANY_OF with OR operator",
			query: "tags ANY_OF ('urgent', 'critical') || priority > 5",
			expected: Query{
				IsOr: true,
				SubQueries: []Query{
					{
						IsOr: false,
						Conditions: []Condition{
							{
								Op:   "ANY_OF",
								Left: &ConditionTerm{Property: "tags"},
								InValues: []*ConditionTerm{
									{Value: "urgent"},
									{Value: "critical"},
								},
							},
						},
					},
					{
						IsOr: false,
						Conditions: []Condition{
							{
								Op:    ">",
								Left:  &ConditionTerm{Property: "priority"},
								Right: &ConditionTerm{Value: float64(5)},
							},
						},
					},
				},
			},
		},
		{
			name:  "ANY_OF with two values",
			query: "status ANY_OF ('active', 'pending')",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:   "ANY_OF",
						Left: &ConditionTerm{Property: "status"},
						InValues: []*ConditionTerm{
							{Value: "active"},
							{Value: "pending"},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseExprQuery(tt.query)
			queriesEqual(t, tt.expected, result)
		})
	}
}

// TestArithmeticOperations tests arithmetic expressions
func TestArithmeticOperations(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected Query
	}{
		{
			name:  "addition",
			query: "age + 5 > 30",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op: ">",
						Left: &ConditionTerm{
							Arithmetic: &ArithmeticExpr{
								Op:    "+",
								Left:  &ConditionTerm{Property: "age"},
								Right: &ConditionTerm{Value: float64(5)},
							},
						},
						Right: &ConditionTerm{Value: float64(30)},
					},
				},
			},
		},
		{
			name:  "subtraction",
			query: "salary - tax >= 50000",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op: ">=",
						Left: &ConditionTerm{
							Arithmetic: &ArithmeticExpr{
								Op:    "-",
								Left:  &ConditionTerm{Property: "salary"},
								Right: &ConditionTerm{Property: "tax"},
							},
						},
						Right: &ConditionTerm{Value: float64(50000)},
					},
				},
			},
		},
		{
			name:  "multiplication",
			query: "price * 1.1 < 100",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op: "<",
						Left: &ConditionTerm{
							Arithmetic: &ArithmeticExpr{
								Op:    "*",
								Left:  &ConditionTerm{Property: "price"},
								Right: &ConditionTerm{Value: 1.1},
							},
						},
						Right: &ConditionTerm{Value: float64(100)},
					},
				},
			},
		},
		{
			name:  "division",
			query: "total / count > 50",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op: ">",
						Left: &ConditionTerm{
							Arithmetic: &ArithmeticExpr{
								Op:    "/",
								Left:  &ConditionTerm{Property: "total"},
								Right: &ConditionTerm{Property: "count"},
							},
						},
						Right: &ConditionTerm{Value: float64(50)},
					},
				},
			},
		},
		{
			name:  "modulo",
			query: "id % 10 == 0",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op: "==",
						Left: &ConditionTerm{
							Arithmetic: &ArithmeticExpr{
								Op:    "%",
								Left:  &ConditionTerm{Property: "id"},
								Right: &ConditionTerm{Value: float64(10)},
							},
						},
						Right: &ConditionTerm{Value: float64(0)},
					},
				},
			},
		},
		{
			name:  "multiplication with two properties",
			query: "quantity * price > 1000",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op: ">",
						Left: &ConditionTerm{
							Arithmetic: &ArithmeticExpr{
								Op:    "*",
								Left:  &ConditionTerm{Property: "quantity"},
								Right: &ConditionTerm{Property: "price"},
							},
						},
						Right: &ConditionTerm{Value: float64(1000)},
					},
				},
			},
		},
		{
			name:  "arithmetic with variables",
			query: "price + $tax > 100",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op: ">",
						Left: &ConditionTerm{
							Arithmetic: &ArithmeticExpr{
								Op:    "+",
								Left:  &ConditionTerm{Property: "price"},
								Right: &ConditionTerm{Variable: &VariableRef{Name: "tax"}},
							},
						},
						Right: &ConditionTerm{Value: float64(100)},
					},
				},
			},
		},
		{
			name:  "arithmetic on right side",
			query: "100 < age + 5",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:   "<",
						Left: &ConditionTerm{Value: float64(100)},
						Right: &ConditionTerm{
							Arithmetic: &ArithmeticExpr{
								Op:    "+",
								Left:  &ConditionTerm{Property: "age"},
								Right: &ConditionTerm{Value: float64(5)},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseExprQuery(tt.query)
			queriesEqual(t, tt.expected, result)
		})
	}
}

// TestCastOperator tests the CAST type conversion operator
func TestCastOperator(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected Query
	}{
		{
			name:  "CAST to string",
			query: "CAST(age AS string) == '25'",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op: "==",
						Left: &ConditionTerm{
							Cast: &TypeCast{
								Value:      &ConditionTerm{Property: "age"},
								TargetType: "string",
							},
						},
						Right: &ConditionTerm{Value: "25"},
					},
				},
			},
		},
		{
			name:  "CAST to int",
			query: "CAST(stringValue AS int) > 100",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op: ">",
						Left: &ConditionTerm{
							Cast: &TypeCast{
								Value:      &ConditionTerm{Property: "stringValue"},
								TargetType: "int",
							},
						},
						Right: &ConditionTerm{Value: float64(100)},
					},
				},
			},
		},
		{
			name:  "CAST to float",
			query: "CAST(price AS float) >= 99.99",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op: ">=",
						Left: &ConditionTerm{
							Cast: &TypeCast{
								Value:      &ConditionTerm{Property: "price"},
								TargetType: "float",
							},
						},
						Right: &ConditionTerm{Value: 99.99},
					},
				},
			},
		},
		{
			name:  "CAST with variable",
			query: "CAST($userInput AS int) > 0",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op: ">",
						Left: &ConditionTerm{
							Cast: &TypeCast{
								Value:      &ConditionTerm{Variable: &VariableRef{Name: "userInput"}},
								TargetType: "int",
							},
						},
						Right: &ConditionTerm{Value: float64(0)},
					},
				},
			},
		},
		{
			name:  "CAST with arithmetic expression",
			query: "CAST(quantity * price AS int) > 1000",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op: ">",
						Left: &ConditionTerm{
							Cast: &TypeCast{
								Value: &ConditionTerm{
									Arithmetic: &ArithmeticExpr{
										Op:    "*",
										Left:  &ConditionTerm{Property: "quantity"},
										Right: &ConditionTerm{Property: "price"},
									},
								},
								TargetType: "int",
							},
						},
						Right: &ConditionTerm{Value: float64(1000)},
					},
				},
			},
		},
		{
			name:  "CAST on right side",
			query: "user.id == CAST(stringId AS int)",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:   "==",
						Left: &ConditionTerm{Property: "user.id"},
						Right: &ConditionTerm{
							Cast: &TypeCast{
								Value:      &ConditionTerm{Property: "stringId"},
								TargetType: "int",
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseExprQuery(tt.query)
			queriesEqual(t, tt.expected, result)
		})
	}
}

// TestDateTimeOperations tests NOW() and TODAY() date/time functions
func TestDateTimeOperations(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected Query
	}{
		{
			name:  "NOW function",
			query: "createdAt > NOW()",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:   ">",
						Left: &ConditionTerm{Property: "createdAt"},
						Right: &ConditionTerm{
							DateTime: &DateTimeValue{
								Type: "NOW",
							},
						},
					},
				},
			},
		},
		{
			name:  "TODAY function",
			query: "date == TODAY()",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:   "==",
						Left: &ConditionTerm{Property: "date"},
						Right: &ConditionTerm{
							DateTime: &DateTimeValue{
								Type: "TODAY",
							},
						},
					},
				},
			},
		},
		{
			name:  "NOW with less than",
			query: "updatedAt < NOW()",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:   "<",
						Left: &ConditionTerm{Property: "updatedAt"},
						Right: &ConditionTerm{
							DateTime: &DateTimeValue{
								Type: "NOW",
							},
						},
					},
				},
			},
		},
		{
			name:  "TODAY with greater than or equal",
			query: "expiryDate >= TODAY()",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:   ">=",
						Left: &ConditionTerm{Property: "expiryDate"},
						Right: &ConditionTerm{
							DateTime: &DateTimeValue{
								Type: "TODAY",
							},
						},
					},
				},
			},
		},
		{
			name:  "NOW with AND operator",
			query: "createdAt > NOW() && status == 'active'",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:   ">",
						Left: &ConditionTerm{Property: "createdAt"},
						Right: &ConditionTerm{
							DateTime: &DateTimeValue{
								Type: "NOW",
							},
						},
					},
					{
						Op:    "==",
						Left:  &ConditionTerm{Property: "status"},
						Right: &ConditionTerm{Value: "active"},
					},
				},
			},
		},
		{
			name:  "BETWEEN with NOW",
			query: "createdAt BETWEEN $startDate AND NOW()",
			expected: Query{
				IsOr: false,
				Conditions: []Condition{
					{
						Op:    "BETWEEN",
						Left:  &ConditionTerm{Property: "createdAt"},
						Start: &ConditionTerm{Variable: &VariableRef{Name: "startDate"}},
						End: &ConditionTerm{
							DateTime: &DateTimeValue{
								Type: "NOW",
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseExprQuery(tt.query)
			queriesEqual(t, tt.expected, result)
		})
	}
}
