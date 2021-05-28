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
	// Ignored field that is here to distinguish json from "OrStatement"
	And   int       `json:"and"`
	Left  Statement `json:"left"`
	Right Statement `json:"right"`
}

func (s AndStatement) statementNode() {
}

type OrStatement struct {
	// Ignored field that is here to distinguish json from "AndStatement"
	Or    int       `json:"or"`
	Left  Statement `json:"left"`
	Right Statement `json:"right"`
}

func (s OrStatement) statementNode() {
}

type MatchStatement struct {
	FieldType int         `json:"field_type"`
	FieldName string      `json:"field_name"`
	Operation string      `json:"operation"`
	Value     interface{} `json:"value"`
}

func (s MatchStatement) statementNode() {
}

func EvalStatement(db *DB, statement Statement) ([]mcmodel.Activity, []mcmodel.Entity) {
	var matchingProcesses []mcmodel.Activity
	if hasProcessMatchStatement(statement) {
		matchingProcesses = evalMatchingProcesses(db, statement)
	}

	var matchingSamples []mcmodel.Entity
	if hasSampleMatchStatement(statement) {
		matchingSamples = evalMatchingSamples(db, statement)
	}

	return matchingProcesses, matchingSamples
}

func hasProcessMatchStatement(statement Statement) bool {
	switch s := statement.(type) {
	case MatchStatement:
		switch s.FieldType {
		case ProcessFieldType:
			return true
		case ProcessAttributeFieldType:
			return true
		default:
			return false
		}
	case AndStatement:
		if hasProcessMatchStatement(s.Left) {
			return true
		}
		if hasProcessMatchStatement(s.Right) {
			return true
		}
		return false
	case OrStatement:
		if hasProcessMatchStatement(s.Left) {
			return true
		}
		if hasProcessMatchStatement(s.Right) {
			return true
		}
		return false
	}

	return false
}

func hasSampleMatchStatement(statement Statement) bool {
	switch s := statement.(type) {
	case MatchStatement:
		switch s.FieldType {
		case SampleFieldType:
			return true
		case SampleAttributeFieldType:
			return true
		default:
			return false
		}
	case AndStatement:
		if hasSampleMatchStatement(s.Left) {
			return true
		}
		if hasSampleMatchStatement(s.Right) {
			return true
		}
		return false
	case OrStatement:
		if hasSampleMatchStatement(s.Left) {
			return true
		}
		if hasSampleMatchStatement(s.Right) {
			return true
		}
		return false
	}

	return false
}

func evalMatchingProcesses(db *DB, statement Statement) []mcmodel.Activity {
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

type SampleState struct {
	sample        *mcmodel.Entity
	EntityStateID int
}

func evalMatchingSamples(db *DB, statement Statement) []mcmodel.Entity {
	var matchingSamples []mcmodel.Entity
	uniqueSampleMatches := make(map[int]mcmodel.Entity)
	for _, sample := range db.Samples {
		for _, entityState := range sample.EntityStates {
			sampleState := SampleState{&sample, entityState.ID}
			if eval(db, nil, &sampleState, statement) {
				// Found a match on the sample, no need to check other sample states so break out of the state loop
				uniqueSampleMatches[sample.ID] = sample
				break
			}
		}
	}

	for _, sample := range uniqueSampleMatches {
		matchingSamples = append(matchingSamples, sample)
	}

	return matchingSamples
}

func eval(db *DB, process *mcmodel.Activity, sampleState *SampleState, statement Statement) bool {
	switch s := statement.(type) {
	case MatchStatement:
		return evalMatchStatement(db, process, sampleState, s)
	case AndStatement:
		return evalAndStatement(db, process, sampleState, s)
	case OrStatement:
		return evalOrStatement(db, process, sampleState, s)
	default:
		return false
	}
}

func evalAndStatement(db *DB, process *mcmodel.Activity, sampleState *SampleState, statement AndStatement) bool {
	if !eval(db, process, sampleState, statement.Left) {
		return false
	}
	return eval(db, process, sampleState, statement.Right)
}

func evalOrStatement(db *DB, process *mcmodel.Activity, sampleState *SampleState, statement OrStatement) bool {
	leftResult := eval(db, process, sampleState, statement.Left)
	rightResult := eval(db, process, sampleState, statement.Right)
	return leftResult || rightResult
}

func evalMatchStatement(db *DB, process *mcmodel.Activity, sampleState *SampleState, match MatchStatement) bool {
	switch match.FieldType {
	case ProcessFieldType:
		return evalProcessFieldMatch(process, match)
	case ProcessAttributeFieldType:
		return evalProcessAttributeFieldMatch(process, db, match)
	case SampleFieldType:
		return evalSampleFieldMatch(sampleState, match)
	case SampleAttributeFieldType:
		return evalSampleAttributeFieldMatch(sampleState, db, match)
	}

	return false
}

func evalProcessAttributeFieldMatch(process *mcmodel.Activity, db *DB, match MatchStatement) bool {
	if process == nil {
		return false
	}

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

func evalSampleAttributeFieldMatch(sampleState *SampleState, db *DB, match MatchStatement) bool {
	if sampleState == nil {
		return false
	}
	states, ok := db.SampleAttributesBySampleIDAndStates[sampleState.sample.ID]
	if !ok {
		fmt.Printf("    Sample %d/%s has no states\n", sampleState.sample.ID, sampleState.sample.Name)
		return false
	}

	attributes, ok := states[sampleState.EntityStateID]
	if !ok {
		fmt.Printf("     Sample %d/%s with state %d has no attributes\n", sampleState.sample.ID, sampleState.sample.Name, sampleState.EntityStateID)
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
	if process == nil {
		return false
	}
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

func evalSampleFieldMatch(sampleState *SampleState, match MatchStatement) bool {
	if sampleState == nil {
		return false
	}
	if match.FieldName == "name" {
		name, ok := match.Value.(string)
		if !ok {
			return false
		}
		return evalStringMatch(name, sampleState.sample.Name, match.Operation)
	}

	if match.FieldName == "id" {
		id, ok := match.Value.(int)
		if !ok {
			return false
		}
		return evalIntMatch(int64(id), int64(sampleState.sample.ID), match.Operation)
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
