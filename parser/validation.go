package parser

import (
	"encoding/json"
	"fmt"
)

// validateQuery checks if the parsed query is valid
func validateQuery(q Query) error {
	if q.IsOr {
		if len(q.SubQueries) == 0 {
			return fmt.Errorf("OR query has no sub-queries")
		}
		for i, sub := range q.SubQueries {
			if err := validateQuery(sub); err != nil {
				return fmt.Errorf("sub-query %d: %w", i, err)
			}
		}
	} else {
		if len(q.Conditions) == 0 && len(q.SubQueries) == 0 {
			return fmt.Errorf("AND query has no conditions or sub-queries")
		}
		for i, cond := range q.Conditions {
			if cond.Op == "" {
				fmt.Println("Condition")
				jsoned, _ := json.Marshal(cond)
				fmt.Println(string(jsoned))
				return fmt.Errorf("condition %d: missing operator", i)
			}
		}
		for i, sub := range q.SubQueries {
			if err := validateQuery(sub); err != nil {
				return fmt.Errorf("sub-query %d: %w", i, err)
			}
		}
	}
	return nil
}
