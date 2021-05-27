package mqldb

import (
	"fmt"

	"github.com/materials-commons/gomcdb/mcmodel"
)

const (
	ProcessFieldType          = 1
	SampleFieldType           = 2
	ProcessAttributeFieldType = 3
	SampleAttributeFieldType  = 4
)

//type MatchStatement struct {
//	FieldType int
//	FieldName string
//	Operation string
//	Value     interface{}
//}

type Statement interface {
	statementNode()
}

type AndStatement struct {
	Left  Statement
	Right Statement
}

func (s AndStatement) statementNode() {
}

type OrStatement struct {
	Left  Statement
	Right Statement
}

func (s OrStatement) statementNode() {
}

type MatchStatement struct {
	FieldType int
	FieldName string
	Operation string
	Value     interface{}
}

func (s MatchStatement) statementNode() {
}

func EvalStatement(db *DB, statement Statement) []mcmodel.Activity {
	var matchingProcesses []mcmodel.Activity
	uniqueProcessMatches := make(map[int]mcmodel.Activity)
	for _, process := range db.Processes {
		if eval(db, &process, nil, statement) {
			uniqueProcessMatches[process.ID] = process
		}
	}

	for _, process := range uniqueProcessMatches {
		matchingProcesses = append(matchingProcesses, process)
	}
	return matchingProcesses
}

func eval(db *DB, process *mcmodel.Activity, sample *mcmodel.Entity, statement Statement) bool {
	switch s := statement.(type) {
	case MatchStatement:
		return evalMatchStatement(db, process, sample, s)
	case AndStatement:
		return evalAndStatement(db, process, sample, s)
	case OrStatement:
		return evalOrStatement(db, process, sample, s)
	default:
		return false
	}
}

func evalAndStatement(db *DB, process *mcmodel.Activity, sample *mcmodel.Entity, statement AndStatement) bool {
	if !eval(db, process, sample, statement.Left) {
		return false
	}
	return eval(db, process, sample, statement.Right)
}

func evalOrStatement(db *DB, process *mcmodel.Activity, sample *mcmodel.Entity, statement OrStatement) bool {
	leftResult := eval(db, process, sample, statement.Left)
	rightResult := eval(db, process, sample, statement.Right)
	return leftResult || rightResult
}

func evalMatchStatement(db *DB, process *mcmodel.Activity, sample *mcmodel.Entity, match MatchStatement) bool {
	switch match.FieldType {
	case ProcessFieldType:
		return evalProcessFieldMatch(process, match)
	case ProcessAttributeFieldType:
		return evalProcessAttributeFieldMatch(process, db, match)
	case SampleFieldType:
		return evalSampleFieldMatch(sample, match)
	case SampleAttributeFieldType:
		return evalSampleAttributeFieldMatch(sample, db, match)
	}

	return false
}

func evalProcessAttributeFieldMatch(process *mcmodel.Activity, db *DB, match MatchStatement) bool {
	attributes, ok := db.ProcessAttributesByProcessID[process.ID]
	if !ok {
		fmt.Printf("    Process %d/%s has no attributes\n", process.ID, process.Name)
		return false
	}

	attribute, ok := attributes[match.FieldName]
	if !ok {
		return false
	}

	switch attribute.Value.ValueType {
	case mcmodel.ValueTypeInt:
		return tryEvalAttributeIntMatch(attribute.Value.ValueInt, match)
	case mcmodel.ValueTypeFloat:
		return tryEvalAttributeFloatMatch(attribute.Value.ValueFloat, match)
	case mcmodel.ValueTypeString:
		return tryEvalAttributeStringMatch(attribute.Value.ValueString, match)
	default:
		return false
	}
}

func evalSampleAttributeFieldMatch(sample *mcmodel.Entity, db *DB, match MatchStatement) bool {
	attributes, ok := db.ProcessAttributesByProcessID[sample.ID]
	if !ok {
		fmt.Printf("    Process %d/%s has no attributes\n", sample.ID, sample.Name)
		return false
	}

	attribute, ok := attributes[match.FieldName]
	if !ok {
		return false
	}

	switch attribute.Value.ValueType {
	case mcmodel.ValueTypeInt:
		return tryEvalAttributeIntMatch(attribute.Value.ValueInt, match)
	case mcmodel.ValueTypeFloat:
		return tryEvalAttributeFloatMatch(attribute.Value.ValueFloat, match)
	case mcmodel.ValueTypeString:
		return tryEvalAttributeStringMatch(attribute.Value.ValueString, match)
	default:
		return false
	}
}

func tryEvalAttributeIntMatch(val1 int64, match MatchStatement) bool {
	val2, ok := match.Value.(int)
	if !ok {
		return false
	}
	return evalIntMatch(val1, int64(val2), match.Operation)
}

func tryEvalAttributeFloatMatch(val1 float64, match MatchStatement) bool {
	val2, ok := match.Value.(float64)
	if !ok {
		val2As32, ok := match.Value.(float32)
		if !ok {
			return false
		}
		return evalFloatMatch(val1, float64(val2As32), match.Operation)
	}

	return evalFloatMatch(val1, val2, match.Operation)
}

func tryEvalAttributeStringMatch(val1 string, match MatchStatement) bool {
	val2, ok := match.Value.(string)
	if !ok {
		return false
	}
	return evalStringMatch(val1, val2, match.Operation)
}

func evalProcessFieldMatch(process *mcmodel.Activity, match MatchStatement) bool {
	if match.FieldName == "name" {
		name, ok := match.Value.(string)
		if !ok {
			return false
		}
		return evalStringMatch(name, process.Name, match.Operation)
	}

	if match.FieldName == "id" {
		id, ok := match.Value.(int)
		if !ok {
			return false
		}
		return evalIntMatch(int64(id), int64(process.ID), match.Operation)
	}

	return false
}

func evalSampleFieldMatch(sample *mcmodel.Entity, match MatchStatement) bool {
	if match.FieldName == "name" {
		name, ok := match.Value.(string)
		if !ok {
			return false
		}
		return evalStringMatch(name, sample.Name, match.Operation)
	}

	if match.FieldName == "id" {
		id, ok := match.Value.(int)
		if !ok {
			return false
		}
		return evalIntMatch(int64(id), int64(sample.ID), match.Operation)
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
