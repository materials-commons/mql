package mqldb

import (
	"fmt"
	"testing"

	"github.com/materials-commons/gomcdb/mcmodel"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestWritingProcessQuery(t *testing.T) {
	dsn := "mc:mcpw@tcp(127.0.0.1:3306)/mc?charset=utf8mb4&parseTime=True&loc=Local"
	mysqldb, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Errorf("Failed to open db: %s", err)
	}

	db := NewDB(77, mysqldb)
	if err := db.Load(); err != nil {
		t.Fatalf("Unable to load database: %s", err)
	}

	m := MatchStatement{
		FieldType: ProcessFieldType,
		FieldName: "name",
		Operation: "=",
		Value:     "Texture",
	}
	s := SingleStatement{
		Statement: m,
	}

	//matching := EvalStatement(db, s)
	//fmt.Printf("Number of matches = %d\n", len(matching))

	m.FieldType = ProcessAttributeFieldType
	m.FieldName = "frames per second"
	m.Operation = ">"
	m.Value = 50
	s.Statement = m
	matching := EvalStatement(db, s)
	fmt.Printf("Number of matches on 'frames per second' > 50 = %d\n", len(matching))
}

const (
	ProcessFieldType          = 1
	SampleFieldType           = 2
	ProcessAttributeFieldType = 3
	SampleAttributeFieldType  = 4
)

type MatchStatement struct {
	FieldType int
	FieldName string
	Operation string
	Value     interface{}
}

type Statement interface {
	statementNode()
}

type AndStatement struct {
	Left  MatchStatement
	Right Statement
}

func (s AndStatement) statementNode() {
}

type OrStatement struct {
	Left  MatchStatement
	Right Statement
}

func (s OrStatement) statementNode() {
}

type SingleStatement struct {
	Statement MatchStatement
}

func (s SingleStatement) statementNode() {
}

func EvalStatement(db *DB, statement Statement) []mcmodel.Activity {
	var matchingProcesses []mcmodel.Activity
	for _, process := range db.Processes {
		if eval(db, process, statement) {
			matchingProcesses = append(matchingProcesses, process)
		}
	}

	return matchingProcesses
}

func eval(db *DB, process mcmodel.Activity, statement Statement) bool {
	switch s := statement.(type) {
	case SingleStatement:
		return evalSingleStatement(db, process, s)
	case AndStatement:
		return evalAndStatement(db, process, s)
	case OrStatement:
		return evalOrStatement(db, process, s)
	default:
		return false
	}
}

func evalSingleStatement(db *DB, process mcmodel.Activity, statement SingleStatement) bool {
	return evalMatchStatement(db, process, statement.Statement)
}

func evalAndStatement(db *DB, process mcmodel.Activity, statement AndStatement) bool {
	if !evalMatchStatement(db, process, statement.Left) {
		return false
	}
	return eval(db, process, statement.Right)
}

func evalOrStatement(db *DB, process mcmodel.Activity, statement OrStatement) bool {
	leftResult := evalMatchStatement(db, process, statement.Left)
	rightResult := eval(db, process, statement.Right)
	return leftResult || rightResult
}

func evalMatchStatement(db *DB, process mcmodel.Activity, match MatchStatement) bool {
	switch match.FieldType {
	case ProcessFieldType:
		return evalProcessFieldMatch(process, match)
	case ProcessAttributeFieldType:
		return evalProcessAttributeFieldMatch(process, db, match)
	}

	return false
}

func evalProcessAttributeFieldMatch(process mcmodel.Activity, db *DB, match MatchStatement) bool {
	attributes, ok := db.ProcessAttributesByProcessID[process.ID]
	if !ok {
		return false
	}

	attribute, ok := attributes[match.FieldName]
	if !ok {
		// process doesn't have this attribute
		return false
	}

	switch attribute.Value.ValueType {
	case mcmodel.ValueTypeInt:
		fmt.Printf("Checking %s:%d\n", attribute.Name, attribute.Value.ValueInt)
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

func evalProcessFieldMatch(process mcmodel.Activity, match MatchStatement) bool {
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
