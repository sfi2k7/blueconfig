package parser

import (
	"github.com/alecthomas/participle/v2/lexer"
)

// queryLexer defines the lexer rules for tokenizing query strings
var queryLexer = lexer.MustSimple([]lexer.SimpleRule{
	{Name: "String", Pattern: `'[^']*'|"[^"]*"`},
	{Name: "Keyword", Pattern: `\b(IN|NOT|IS|NULL|LIKE|BETWEEN|AND|CONTAINS|ANY_OF|CAST|AS|NOW|TODAY)\b`},
	{Name: "Ident", Pattern: `[a-zA-Z_][a-zA-Z0-9_]*(\.[a-zA-Z_][a-zA-Z0-9_]*)*`},
	{Name: "Float", Pattern: `\d+\.\d+`},
	{Name: "Int", Pattern: `\d+`},
	{Name: "GTE", Pattern: `>=`},
	{Name: "LTE", Pattern: `<=`},
	{Name: "EQ", Pattern: `==`},
	{Name: "NEQ", Pattern: `!=`},
	{Name: "Punct", Pattern: `[(){}[\],;$]`},
	{Name: "Plus", Pattern: `\+`},
	{Name: "Minus", Pattern: `-`},
	{Name: "Multiply", Pattern: `\*`},
	{Name: "Divide", Pattern: `/`},
	{Name: "Modulo", Pattern: `%`},
	{Name: "GT", Pattern: `>`},
	{Name: "LT", Pattern: `<`},
	{Name: "And", Pattern: `&&`},
	{Name: "Or", Pattern: `\|\|`},
	{Name: "Not", Pattern: `!`},
	{Name: "whitespace", Pattern: `\s+`},
})
