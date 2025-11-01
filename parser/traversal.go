package parser

// traverseAST converts the parsed AST into a Query structure
func traverseAST(ast *QueryAST) Query {
	q := traverseExpr(ast.Expr)

	// Extract collection from USE statement if present
	if ast.Use != nil {
		q.Collection = ast.Use.Collection
	}

	return q
}

// traverseExpr processes the top-level expression
func traverseExpr(expr *Expr) Query {
	return traverseOr(expr.Or)
}

// traverseOr processes OR expressions
func traverseOr(or *OrExpr) Query {
	if len(or.And) == 1 {
		return traverseAnd(or.And[0])
	}
	// OR: create subqueries
	subs := make([]Query, len(or.And))
	for i, and := range or.And {
		subs[i] = traverseAnd(and)
	}
	return Query{IsOr: true, SubQueries: subs}
}

// traverseAnd processes AND expressions
func traverseAnd(and *AndExpr) Query {
	if len(and.Not) == 1 {
		return traverseNot(and.Not[0])
	}
	// AND: flatten simple conditions, preserve complex subqueries
	var conds []Condition
	var subQueries []Query
	for _, not := range and.Not {
		q := traverseNot(not)
		// If it's a simple query (just conditions, no subqueries), flatten it
		if !q.IsOr && len(q.SubQueries) == 0 && len(q.Conditions) > 0 {
			conds = append(conds, q.Conditions...)
		} else if q.IsOr || len(q.SubQueries) > 0 {
			// If it's a complex query (OR or has subqueries), preserve it as a subquery
			subQueries = append(subQueries, q)
		} else if len(q.Conditions) > 0 {
			// Edge case: AND query with conditions but no subqueries
			conds = append(conds, q.Conditions...)
		}
	}
	return Query{IsOr: false, Conditions: conds, SubQueries: subQueries}
}

// traverseNot processes NOT expressions and parenthesized expressions
func traverseNot(not *NotExpr) Query {
	isNegated := not.Not == "!"

	// Handle parenthesized expression
	if not.Paren != nil {
		sub := traverseExpr(not.Paren)
		if isNegated {
			return negateQuery(sub)
		}
		return sub
	}

	// Handle comparison
	if not.Comp != nil {
		cond := traverseComparison(not.Comp)
		if isNegated {
			cond.Negate = !cond.Negate
		}
		return Query{IsOr: false, Conditions: []Condition{cond}}
	}

	return Query{}
}

// negateQuery applies De Morgan's law: !(a || b) = !a && !b, !(a && b) = !a || !b
func negateQuery(q Query) Query {
	if q.IsOr {
		// !(a || b || c) = !a && !b && !c
		var allConds []Condition
		for _, subQ := range q.SubQueries {
			negated := negateQuery(subQ)
			allConds = append(allConds, negated.Conditions...)
		}
		return Query{IsOr: false, Conditions: allConds}
	} else {
		// !(a && b && c) = !a || !b || !c
		var subQueries []Query
		for _, cond := range q.Conditions {
			cond.Negate = !cond.Negate
			subQueries = append(subQueries, Query{IsOr: false, Conditions: []Condition{cond}})
		}
		if len(subQueries) == 1 {
			return subQueries[0]
		}
		return Query{IsOr: true, SubQueries: subQueries}
	}
}

// traverseComparison converts a Comparison AST node to a Condition
func traverseComparison(comp *Comparison) Condition {
	if comp == nil {
		return Condition{}
	}

	left := traverseArithExpr(comp.Left)
	op := comp.Op

	// Handle IN operator
	if comp.InOp != "" {
		// Check if using variable
		if comp.InVarName != nil {
			return Condition{
				Op:    "IN",
				Left:  left,
				Right: &ConditionTerm{Variable: &VariableRef{Name: comp.InVarName.Name}},
			}
		}

		// Otherwise use list
		if comp.InList != nil {
			values := make([]*ConditionTerm, len(comp.InList))
			for i, v := range comp.InList {
				val := extractValue(v)
				values[i] = &ConditionTerm{Value: val}
			}
			return Condition{
				Op:       "IN",
				Left:     left,
				InValues: values,
			}
		}
	}

	// Handle NOT IN operator
	if comp.NotInOp != "" {
		// Check if using variable
		if comp.NotInVarName != nil {
			return Condition{
				Op:    "NOT_IN",
				Left:  left,
				Right: &ConditionTerm{Variable: &VariableRef{Name: comp.NotInVarName.Name}},
			}
		}

		// Otherwise use list
		if comp.NotInList != nil {
			values := make([]*ConditionTerm, len(comp.NotInList))
			for i, v := range comp.NotInList {
				val := extractValue(v)
				values[i] = &ConditionTerm{Value: val}
			}
			return Condition{
				Op:       "NOT_IN",
				Left:     left,
				InValues: values,
			}
		}
	}

	// Handle IS NULL
	if comp.IsNullOp != "" {
		return Condition{
			Op:   "IS_NULL",
			Left: left,
		}
	}

	// Handle IS NOT NULL
	if comp.IsNotNullOp != "" {
		return Condition{
			Op:   "IS_NOT_NULL",
			Left: left,
		}
	}

	// Handle LIKE operator
	if comp.LikeOp != "" && comp.LikePattern != nil {
		pattern := extractValue(comp.LikePattern)
		return Condition{
			Op:      "LIKE",
			Left:    left,
			Pattern: pattern.(string),
		}
	}

	// Handle BETWEEN operator
	if comp.BetweenOp != "" && comp.BetweenStart != nil && comp.BetweenEnd != nil {
		return Condition{
			Op:    "BETWEEN",
			Left:  left,
			Start: traverseArithExpr(comp.BetweenStart),
			End:   traverseArithExpr(comp.BetweenEnd),
		}
	}

	// Handle CONTAINS operator
	if comp.ContainsOp != "" && comp.ContainsValue != nil {
		// ContainsValue is a Term, need to convert through ArithExpr
		valueTerm := &ArithExpr{Left: &ArithTerm{Left: comp.ContainsValue}}
		return Condition{
			Op:    "CONTAINS",
			Left:  left,
			Right: traverseArithExpr(valueTerm),
		}
	}

	// Handle ANY_OF operator
	if comp.AnyOfOp != "" {
		// Check if using variable
		if comp.AnyOfVarName != nil {
			return Condition{
				Op:    "ANY_OF",
				Left:  left,
				Right: &ConditionTerm{Variable: &VariableRef{Name: comp.AnyOfVarName.Name}},
			}
		}

		// Otherwise use list
		if comp.AnyOfList != nil {
			values := make([]*ConditionTerm, len(comp.AnyOfList))
			for i, v := range comp.AnyOfList {
				val := extractValue(v)
				values[i] = &ConditionTerm{Value: val}
			}
			return Condition{
				Op:       "ANY_OF",
				Left:     left,
				InValues: values,
			}
		}
	}

	// Handle function as boolean predicate (e.g., HasPrefix(name, "x"))
	if left.Function != nil && op == "" {
		return Condition{
			Op:   "func_predicate",
			Left: left,
		}
	}

	// Validate right term exists for binary operations
	if comp.Right == nil {
		return Condition{}
	}

	right := traverseArithExpr(comp.Right)
	negate := false

	if op == "!=" {
		op = "=="
		negate = true
	}

	return Condition{
		Op:     op,
		Left:   left,
		Right:  right,
		Negate: negate,
	}
}

// traverseArithExpr handles arithmetic expressions (addition, subtraction)
func traverseArithExpr(arith *ArithExpr) *ConditionTerm {
	if arith == nil {
		return nil
	}

	left := traverseArithTerm(arith.Left)

	if arith.Op != "" && arith.Right != nil {
		right := traverseArithExpr(arith.Right)
		return &ConditionTerm{
			Arithmetic: &ArithmeticExpr{
				Op:    arith.Op,
				Left:  left,
				Right: right,
			},
		}
	}

	return left
}

// traverseArithTerm handles arithmetic terms (multiplication, division, modulo)
func traverseArithTerm(term *ArithTerm) *ConditionTerm {
	if term == nil {
		return nil
	}

	left := traverseTermToCondition(term.Left)

	if term.Op != "" && term.Right != nil {
		right := traverseArithTerm(term.Right)
		return &ConditionTerm{
			Arithmetic: &ArithmeticExpr{
				Op:    term.Op,
				Left:  left,
				Right: right,
			},
		}
	}

	return left
}

// traverseTermToCondition converts a Term AST node to a ConditionTerm
func traverseTermToCondition(term *Term) *ConditionTerm {
	if term == nil {
		return nil
	}

	ct := &ConditionTerm{}

	if term.Val != nil {
		ct.Value = extractValue(term.Val)
	} else if term.Var != nil {
		ct.Variable = &VariableRef{Name: term.Var.Name}
	} else if term.Ident != "" {
		ct.Property = term.Ident
	} else if term.Func != nil {
		args := buildFunctionArgs(term.Func.Args)
		ct.Function = &FunctionCall{Name: term.Func.Name, Args: args}
	} else if term.Cast != nil {
		ct.Cast = &TypeCast{
			Value:      traverseArithExpr(term.Cast.Value),
			TargetType: term.Cast.TargetType,
		}
	} else if term.DateTime != nil {
		dtType := ""
		if term.DateTime.Now != "" {
			dtType = "NOW"
		} else if term.DateTime.Today != "" {
			dtType = "TODAY"
		}
		ct.DateTime = &DateTimeValue{Type: dtType}
	} else if term.Paren != nil {
		return traverseArithExpr(term.Paren)
	}

	return ct
}

// buildFunctionArgs converts function argument terms to ConditionTerm values
func buildFunctionArgs(args []*Term) []*ConditionTerm {
	result := []*ConditionTerm{}
	for _, arg := range args {
		result = append(result, traverseTermToCondition(arg))
	}
	return result
}

// extractValue extracts the actual value from a Value AST node
func extractValue(val *Value) interface{} {
	if val == nil {
		return nil
	}
	if val.String != nil {
		return *val.String
	}
	if val.Number != nil {
		return *val.Number
	}
	if val.Bool != nil {
		// Bool is captured as string "true" or "false" from grammar
		return *val.Bool == "true"
	}
	return nil
}
