package parser

// QueryAST represents the root of the parsed expression with optional USE statement
type QueryAST struct {
	Use  *UseStatement `@@?`
	Expr *Expr         `@@`
}

// UseStatement represents "USE collection_name;" syntax
type UseStatement struct {
	Collection string `"USE" @Ident ";"`
}

// Expr is the top-level expression
type Expr struct {
	Or *OrExpr `@@`
}

// OrExpr handles OR operations (left-associative)
type OrExpr struct {
	And []*AndExpr `@@ ( "||" @@ )*`
}

// AndExpr handles AND operations
type AndExpr struct {
	Not []*NotExpr `@@ ( "&&" @@ )*`
}

// NotExpr handles negation
type NotExpr struct {
	Not   string      `@"!"?`
	Paren *Expr       `( "(" @@ ")" )`
	Comp  *Comparison `| @@`
}

// Comparison represents binary comparisons and IN operator
type Comparison struct {
	Left *ArithExpr `@@`

	// Standard comparison operators
	Op    string     `( @( "==" | "!=" | ">" | "<" | ">=" | "<=" )`
	Right *ArithExpr `  @@`

	// IN operator
	InOp      string    `| @"IN"`
	InList    []*Value  `  ( "(" @@ ( "," @@ )* ")"`
	InVarName *Variable `  | @@ )`

	// NOT IN operator
	NotInOp      string    `| @"NOT" "IN"`
	NotInList    []*Value  `  ( "(" @@ ( "," @@ )* ")"`
	NotInVarName *Variable `  | @@ )`

	// IS NULL
	IsNullOp string `| @"IS" @"NULL"`

	// IS NOT NULL
	IsNotNullOp string `| @"IS" @"NOT" @"NULL"`

	// LIKE operator
	LikeOp      string `| @"LIKE"`
	LikePattern *Value `  @@`

	// BETWEEN operator
	BetweenOp    string     `| @"BETWEEN"`
	BetweenStart *ArithExpr `  @@`
	BetweenAnd   string     `  @"AND"`
	BetweenEnd   *ArithExpr `  @@`

	// CONTAINS operator
	ContainsOp    string `| @"CONTAINS"`
	ContainsValue *Term  `  @@`

	// ANY_OF operator
	AnyOfOp      string    `| @"ANY_OF"`
	AnyOfList    []*Value  `  ( "(" @@ ( "," @@ )* ")"`
	AnyOfVarName *Variable `  | @@ ) )?`
}

// ArithExpr handles arithmetic operations (addition, subtraction)
type ArithExpr struct {
	Left  *ArithTerm `@@`
	Op    string     `( @( "+" | "-" )`
	Right *ArithExpr `  @@ )?`
}

// ArithTerm handles multiplication, division, modulo
type ArithTerm struct {
	Left  *Term      `@@`
	Op    string     `( @( "*" | "/" | "%" )`
	Right *ArithTerm `  @@ )?`
}

// Term is a value, identifier, variable, function, cast, datetime, or parenthesized expression
type Term struct {
	Val      *Value        `  @@`
	Func     *Function     `| @@`
	Var      *Variable     `| @@`
	Cast     *CastExpr     `| @@`
	DateTime *DateTimeExpr `| @@`
	Ident    string        `| @Ident`
	Paren    *ArithExpr    `| "(" @@ ")"`
}

// CastExpr represents type casting
type CastExpr struct {
	CastKeyword string     `@"CAST" "("`
	Value       *ArithExpr `@@`
	AsKeyword   string     `@"AS"`
	TargetType  string     `@Ident ")"`
}

// DateTimeExpr represents date/time operations
type DateTimeExpr struct {
	Now   string `  @"NOW" "(" ")"`
	Today string `| @"TODAY" "(" ")"`
}

// Function represents function calls
type Function struct {
	Name string  `@Ident`
	Args []*Term `"(" ( @@ ( "," @@ )* )? ")"`
}

// Variable represents a runtime variable reference with $var syntax
type Variable struct {
	Name string `"$" @Ident`
}

// Value represents literals
type Value struct {
	Number *float64 `@Float | @Int`
	String *string  `| @String`
	Bool   *string  `| @"true" | @"false"`
}
