package mqldb

import "fmt"

func MapToStatement(m map[string]interface{}) Statement {
	_, hasAnd := m["and"]
	_, hasOr := m["or"]
	_, hasFieldName := m["field_name"]
	switch {
	case hasAnd:
		andStatement := AndStatement{}
		left, hasLeft := m["left"]
		if hasLeft {
			andStatement.Left = MapToStatement(left.(map[string]interface{}))
		}

		right, hasRight := m["right"]
		if hasRight {
			andStatement.Right = MapToStatement(right.(map[string]interface{}))
		}

		return andStatement

	case hasOr:
		orStatement := OrStatement{}
		left, hasLeft := m["left"]
		if hasLeft {
			orStatement.Left = MapToStatement(left.(map[string]interface{}))
		}

		right, hasRight := m["right"]
		if hasRight {
			orStatement.Right = MapToStatement(right.(map[string]interface{}))
		}

		return orStatement

	case hasFieldName:
		fmt.Printf("m = %+v\n", m)
		return MatchStatement{
			FieldType: int(m["field_type"].(float64)),
			FieldName: m["field_name"].(string),
			Operation: m["operation"].(string),
			Value:     m["value"],
		}
	}

	return nil
}
