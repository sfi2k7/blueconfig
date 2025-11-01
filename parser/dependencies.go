package parser

// QueryDependencies represents all the external dependencies needed to evaluate a query
type QueryDependencies struct {
	Properties []string // Field/property names referenced in the query
	Variables  []string // Variable names (without $ prefix) referenced in the query
	Functions  []string // Function names called in the query
}

// ExtractQueryDependencies analyzes a Query and returns all dependencies needed to evaluate it.
// This is a stateless method that helps identify what context (properties, variables, methods)
// is required before attempting to match records.
//
// Example usage:
//
//	query := ParseExprQuery("upper(name) == $targetName && age >= $minAge && location.city == 'NYC'")
//	deps := ExtractQueryDependencies(query)
//	// deps.Properties = ["name", "age", "location.city"]
//	// deps.Variables = ["targetName", "minAge"]
//	// deps.Functions = ["upper"]
func ExtractQueryDependencies(query Query) QueryDependencies {
	deps := QueryDependencies{
		Properties: []string{},
		Variables:  []string{},
		Functions:  []string{},
	}

	// Use maps to track unique values
	propMap := make(map[string]bool)
	varMap := make(map[string]bool)
	funcMap := make(map[string]bool)

	// Recursively extract dependencies
	extractFromQuery(query, propMap, varMap, funcMap)

	// Convert maps to slices
	for prop := range propMap {
		deps.Properties = append(deps.Properties, prop)
	}
	for varName := range varMap {
		deps.Variables = append(deps.Variables, varName)
	}
	for funcName := range funcMap {
		deps.Functions = append(deps.Functions, funcName)
	}

	return deps
}

// extractFromQuery recursively extracts dependencies from a Query
func extractFromQuery(query Query, props, vars, funcs map[string]bool) {
	// Extract from conditions
	for _, condition := range query.Conditions {
		extractFromCondition(condition, props, vars, funcs)
	}

	// Extract from subqueries
	for _, subQuery := range query.SubQueries {
		extractFromQuery(subQuery, props, vars, funcs)
	}
}

// extractFromCondition extracts dependencies from a Condition
func extractFromCondition(cond Condition, props, vars, funcs map[string]bool) {
	// Extract from left term
	if cond.Left != nil {
		extractFromTerm(cond.Left, props, vars, funcs)
	}

	// Extract from right term
	if cond.Right != nil {
		extractFromTerm(cond.Right, props, vars, funcs)
	}

	// Extract from IN values
	for _, inValue := range cond.InValues {
		extractFromTerm(inValue, props, vars, funcs)
	}

	// Extract from BETWEEN start/end
	if cond.Start != nil {
		extractFromTerm(cond.Start, props, vars, funcs)
	}
	if cond.End != nil {
		extractFromTerm(cond.End, props, vars, funcs)
	}
}

// extractFromTerm extracts dependencies from a ConditionTerm
func extractFromTerm(term *ConditionTerm, props, vars, funcs map[string]bool) {
	if term == nil {
		return
	}

	// Property reference
	if term.Property != "" {
		props[term.Property] = true
	}

	// Variable reference
	if term.Variable != nil {
		vars[term.Variable.Name] = true
	}

	// Function call
	if term.Function != nil {
		funcs[term.Function.Name] = true
		// Extract from function arguments
		for _, arg := range term.Function.Args {
			extractFromTerm(arg, props, vars, funcs)
		}
	}

	// Arithmetic expression
	if term.Arithmetic != nil {
		extractFromTerm(term.Arithmetic.Left, props, vars, funcs)
		extractFromTerm(term.Arithmetic.Right, props, vars, funcs)
	}

	// Type cast
	if term.Cast != nil {
		extractFromTerm(term.Cast.Value, props, vars, funcs)
	}
}
