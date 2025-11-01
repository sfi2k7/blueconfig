package parser

// Query represents the parsed query structure for execution
type Query struct {
	IsOr       bool        `json:"isOr"`       // True if OR disjunction
	Conditions []Condition `json:"conditions"` // Flat conditions for AND
	SubQueries []Query     `json:"subQueries"` // Child queries for OR
	Collection string      `json:"collection"` // Collection name from USE statement
}

// Condition represents a single condition in a query
type Condition struct {
	Op       string           `json:"op"`                 // Operator: ==, !=, >, <, >=, <=, func_predicate, IN, NOT_IN, IS_NULL, IS_NOT_NULL, LIKE, BETWEEN, CONTAINS, ANY_OF
	Left     *ConditionTerm   `json:"left,omitempty"`     // Left side of comparison
	Right    *ConditionTerm   `json:"right,omitempty"`    // Right side of comparison
	Negate   bool             `json:"negate,omitempty"`   // True if condition is negated
	InValues []*ConditionTerm `json:"inValues,omitempty"` // For IN, NOT_IN, ANY_OF operators
	Pattern  string           `json:"pattern,omitempty"`  // For LIKE operator
	Start    *ConditionTerm   `json:"start,omitempty"`    // For BETWEEN operator (start range)
	End      *ConditionTerm   `json:"end,omitempty"`      // For BETWEEN operator (end range)
}

// ConditionTerm represents a term in a condition (property, value, variable, function, arithmetic, cast, or datetime)
type ConditionTerm struct {
	Property   string          `json:"property,omitempty"`   // Field name (e.g., "name", "age")
	Value      interface{}     `json:"value,omitempty"`      // Literal value
	Variable   *VariableRef    `json:"variable,omitempty"`   // Variable reference
	Function   *FunctionCall   `json:"function,omitempty"`   // Function call
	Arithmetic *ArithmeticExpr `json:"arithmetic,omitempty"` // Arithmetic expression
	Cast       *TypeCast       `json:"cast,omitempty"`       // Type cast expression
	DateTime   *DateTimeValue  `json:"dateTime,omitempty"`   // Date/time expression
}

// VariableRef represents a reference to a runtime variable
type VariableRef struct {
	Name string `json:"name"` // Variable name (without the $ prefix)
}

// FunctionCall represents a function call with arguments
type FunctionCall struct {
	Name string           `json:"name"` // Function name
	Args []*ConditionTerm `json:"args"` // Function arguments
}

// ArithmeticExpr represents an arithmetic operation
type ArithmeticExpr struct {
	Op    string         `json:"op"`    // Operator: +, -, *, /, %
	Left  *ConditionTerm `json:"left"`  // Left operand
	Right *ConditionTerm `json:"right"` // Right operand
}

// TypeCast represents a type conversion operation
type TypeCast struct {
	Value      *ConditionTerm `json:"value"`      // Value to cast
	TargetType string         `json:"targetType"` // Target type: int, float, string, bool, etc.
}

// DateTimeValue represents date/time literals
type DateTimeValue struct {
	Type string `json:"type"` // Type: NOW, TODAY
}

// TermResult is a helper struct used during AST traversal
// (kept here for backwards compatibility with traversal code)
type TermResult struct {
	Ident   string
	Val     *Value
	Func    *Function
	VarName string // For variable references
}
