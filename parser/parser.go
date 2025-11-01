package parser

import (
	"fmt"

	"github.com/alecthomas/participle/v2"
)

// parser is the participle parser instance configured with the query lexer
var parser = participle.MustBuild[QueryAST](
	participle.Lexer(queryLexer),
	participle.Unquote("String"),
	participle.UseLookahead(2), // For better parsing
)

// ParseExprQuery parses an expression string into a Query structure
// Panics on parse or validation errors
func ParseExprQuery(exprStr string) Query {
	ast, err := parser.ParseString("", exprStr)
	if err != nil {
		panic(fmt.Sprintf("Parse error: %v", err))
	}
	q := traverseAST(ast)
	if err := validateQuery(q); err != nil {
		panic(fmt.Sprintf("Invalid query: %v", err))
	}
	return q
}

// ParseExprQueryWithCollection parses query and extracts collection from USE statement
// Returns the parsed Query and the collection name
// Panics on parse or validation errors
func ParseExprQueryWithCollection(exprStr string) (Query, string) {
	ast, err := parser.ParseString("", exprStr)
	if err != nil {
		panic(fmt.Sprintf("Parse error: %v", err))
	}

	collection := "default" // Default collection if no USE statement
	if ast.Use != nil {
		collection = ast.Use.Collection
	}

	q := traverseAST(ast)
	q.Collection = collection

	if err := validateQuery(q); err != nil {
		panic(fmt.Sprintf("Invalid query: %v", err))
	}
	return q, collection
}
