package mqldb

import (
	"strconv"

	"github.com/materials-commons/gomcdb/mcmodel"
)

type MatchStatement struct {
	FieldType int         `json:"field_type"`
	FieldName string      `json:"field_name"`
	Operation string      `json:"operation"`
	Value     interface{} `json:"value"`
}

func (s MatchStatement) statementNode() {
}

func (s MatchStatement) tryEvalAttributeIntMatch(val1 int64) bool {
	val2, ok := s.matchValToInt()
	if !ok {
		return false
	}
	return evalIntMatch(val1, val2, s.Operation)
}

func (s MatchStatement) matchValToInt() (int64, bool) {
	switch s.Value.(type) {
	case int64:
		return s.Value.(int64), true
	case int:
		return int64(s.Value.(int)), true
	case float64:
		return int64(s.Value.(float64)), true
	case float32:
		return int64(s.Value.(float32)), true
	case string:
		val, err := strconv.ParseFloat(s.Value.(string), 64)
		if err != nil {
			return -1, false
		}

		return int64(val), true
	default:
		return -1, false
	}
}

func (s MatchStatement) tryEvalAttributeFloatMatch(val1 float64) bool {
	val2, ok := s.matchValToFloat()
	if !ok {
		return false
	}

	return evalFloatMatch(val1, val2, s.Operation)
}

func (s MatchStatement) matchValToFloat() (float64, bool) {
	switch s.Value.(type) {
	case int64:
		return float64(s.Value.(int64)), true
	case int:
		return float64(s.Value.(int)), true
	case float64:
		return s.Value.(float64), true
	case float32:
		return float64(s.Value.(float32)), true
	case string:
		val, err := strconv.ParseFloat(s.Value.(string), 64)
		if err != nil {
			return -1, false
		}

		return val, true
	default:
		return -1, false
	}
}

func (s MatchStatement) tryEvalAttributeStringMatch(val1 string) bool {
	val2, ok := s.Value.(string)
	if !ok {
		return false
	}
	return evalStringMatch(val1, val2, s.Operation)
}

func (s MatchStatement) evalProcessFieldMatch(process *mcmodel.Activity) bool {
	if process == nil {
		return false
	}
	if s.FieldName == "name" {
		name, ok := s.Value.(string)
		if !ok {
			return false
		}
		return evalStringMatch(name, process.Name, s.Operation)
	}

	if s.FieldName == "id" {
		id, ok := s.Value.(int)
		if !ok {
			return false
		}
		return evalIntMatch(int64(id), int64(process.ID), s.Operation)
	}

	return false
}

func (s MatchStatement) evalSampleFieldMatch(sampleState *SampleState) bool {
	if sampleState == nil {
		return false
	}
	if s.FieldName == "name" {
		name, ok := s.Value.(string)
		if !ok {
			return false
		}
		return evalStringMatch(name, sampleState.sample.Name, s.Operation)
	}

	if s.FieldName == "id" {
		id, ok := s.Value.(int)
		if !ok {
			return false
		}
		return evalIntMatch(int64(id), int64(sampleState.sample.ID), s.Operation)
	}

	return false
}

func evalStringMatch(val1, val2, operation string) bool {
	switch operation {
	case "=":
		return val1 == val2
	case "<>":
		return val1 != val2
	default:
		return false
	}
}

func evalIntMatch(val1, val2 int64, operation string) bool {
	switch operation {
	case "=":
		return val1 == val2
	case "<>":
		return val1 != val2
	case ">":
		return val1 > val2
	case ">=":
		return val1 >= val2
	case "<":
		return val1 < val2
	case "<=":
		return val1 <= val2
	default:
		return false
	}
}

func evalFloatMatch(val1, val2 float64, operation string) bool {
	switch operation {
	case "=":
		return val1 == val2
	case "<>":
		return val1 != val2
	case ">":
		return val1 > val2
	case ">=":
		return val1 >= val2
	case "<":
		return val1 < val2
	case "<=":
		return val1 <= val2
	default:
		return false
	}
}
