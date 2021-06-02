package mqldb

import (
	"fmt"
	"testing"

	"github.com/materials-commons/gomcdb/mcmodel"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func createTestDB() *DB {
	db := &DB{
		ProjectID:                           1,
		ProcessAttributesByProcessID:        make(map[int]map[string]*mcmodel.Attribute),
		SampleAttributesBySampleIDAndStates: make(map[int]map[int]map[string]*mcmodel.Attribute),
	}

	// Fill out processes
	db.Processes = append(db.Processes, mcmodel.Activity{
		ID:   1,
		Name: "EBSD",
	})

	db.Processes = append(db.Processes, mcmodel.Activity{
		ID:   2,
		Name: "EBSD",
	})

	db.Processes = append(db.Processes, mcmodel.Activity{
		ID:   3,
		Name: "Texture",
	})

	db.Processes = append(db.Processes, mcmodel.Activity{
		ID:   4,
		Name: "Texture",
	})

	// First EBSD process attributes
	db.ProcessAttributesByProcessID[1] = make(map[string]*mcmodel.Attribute)
	db.ProcessAttributesByProcessID[1]["Beam Type"] = &mcmodel.Attribute{
		Name: "Beam Type",
		AttributeValues: []mcmodel.AttributeValue{
			{
				ValueType:   mcmodel.ValueTypeString,
				ValueString: "Wide",
			},
		},
	}
	db.ProcessAttributesByProcessID[1]["frames per second"] = &mcmodel.Attribute{
		Name: "frames per second",
		AttributeValues: []mcmodel.AttributeValue{
			{
				ValueType: mcmodel.ValueTypeInt,
				ValueInt:  5,
			},
		},
	}
	db.ProcessAttributesByProcessID[1]["note"] = &mcmodel.Attribute{
		Name: "note",
		AttributeValues: []mcmodel.AttributeValue{
			{
				ValueType:   mcmodel.ValueTypeString,
				ValueString: "ignore these results",
			},
		},
	}

	// Second EBSD process attributes
	db.ProcessAttributesByProcessID[2] = make(map[string]*mcmodel.Attribute)
	db.ProcessAttributesByProcessID[2]["Beam Type"] = &mcmodel.Attribute{
		Name: "Beam Type",
		AttributeValues: []mcmodel.AttributeValue{
			{
				ValueType:   mcmodel.ValueTypeString,
				ValueString: "Thin",
			},
		},
	}
	db.ProcessAttributesByProcessID[2]["frames per second"] = &mcmodel.Attribute{
		Name: "frames per second",
		AttributeValues: []mcmodel.AttributeValue{
			{
				ValueType: mcmodel.ValueTypeInt,
				ValueInt:  3,
			},
		},
	}

	// First Texture process attributes
	db.ProcessAttributesByProcessID[3] = make(map[string]*mcmodel.Attribute)
	db.ProcessAttributesByProcessID[3]["PF scale max"] = &mcmodel.Attribute{
		Name: "PF scale max",
		AttributeValues: []mcmodel.AttributeValue{
			{
				ValueType: mcmodel.ValueTypeInt,
				ValueInt:  2,
			},
		},
	}
	db.ProcessAttributesByProcessID[3]["note"] = &mcmodel.Attribute{
		Name: "note",
		AttributeValues: []mcmodel.AttributeValue{
			{
				ValueType:   mcmodel.ValueTypeString,
				ValueString: "ignore these results",
			},
		},
	}

	// Second Texture process attributes
	db.ProcessAttributesByProcessID[4] = make(map[string]*mcmodel.Attribute)
	db.ProcessAttributesByProcessID[4]["PF scale max"] = &mcmodel.Attribute{
		Name: "PF scale max",
		AttributeValues: []mcmodel.AttributeValue{
			{
				ValueType: mcmodel.ValueTypeInt,
				ValueInt:  3,
			},
		},
	}

	///////////////// Fill out samples

	db.Samples = append(db.Samples, mcmodel.Entity{
		Name: "S1",
		ID:   1,
		EntityStates: []mcmodel.EntityState{
			{ID: 1},
			{ID: 2},
		},
	})
	db.SampleAttributesBySampleIDAndStates[1] = make(map[int]map[string]*mcmodel.Attribute)

	// Allocate for entity states
	db.SampleAttributesBySampleIDAndStates[1][1] = make(map[string]*mcmodel.Attribute)
	db.SampleAttributesBySampleIDAndStates[1][1]["zn"] = &mcmodel.Attribute{
		Name: "zn",
		AttributeValues: []mcmodel.AttributeValue{
			{
				ValueType:  mcmodel.ValueTypeFloat,
				ValueFloat: 0.5,
			},
		},
	}
	db.SampleAttributesBySampleIDAndStates[1][1]["mg"] = &mcmodel.Attribute{
		Name: "mg",
		AttributeValues: []mcmodel.AttributeValue{
			{
				ValueType:  mcmodel.ValueTypeFloat,
				ValueFloat: 0.4,
			},
		},
	}

	db.SampleAttributesBySampleIDAndStates[1][2] = make(map[string]*mcmodel.Attribute)
	db.SampleAttributesBySampleIDAndStates[1][2]["zn"] = &mcmodel.Attribute{
		Name: "zn",
		AttributeValues: []mcmodel.AttributeValue{
			{
				ValueType:  mcmodel.ValueTypeFloat,
				ValueFloat: 0.5,
			},
		},
	}
	db.SampleAttributesBySampleIDAndStates[1][2]["mg"] = &mcmodel.Attribute{
		Name: "mg",
		AttributeValues: []mcmodel.AttributeValue{
			{
				ValueType:  mcmodel.ValueTypeFloat,
				ValueFloat: 0.5,
			},
		},
	}
	db.SampleAttributesBySampleIDAndStates[1][2]["hardness"] = &mcmodel.Attribute{
		Name: "hardness",
		AttributeValues: []mcmodel.AttributeValue{
			{
				ValueType: mcmodel.ValueTypeInt,
				ValueInt:  1,
			},
		},
	}

	db.Samples = append(db.Samples, mcmodel.Entity{
		Name: "S2",
		ID:   2,
		EntityStates: []mcmodel.EntityState{
			{ID: 3},
			{ID: 4},
		},
	})
	db.SampleAttributesBySampleIDAndStates[2] = make(map[int]map[string]*mcmodel.Attribute)

	// Allocate for entity states
	db.SampleAttributesBySampleIDAndStates[2][3] = make(map[string]*mcmodel.Attribute)
	db.SampleAttributesBySampleIDAndStates[2][3]["zn"] = &mcmodel.Attribute{
		Name: "zn",
		AttributeValues: []mcmodel.AttributeValue{
			{
				ValueType:  mcmodel.ValueTypeFloat,
				ValueFloat: 0.5,
			},
		},
	}
	db.SampleAttributesBySampleIDAndStates[2][3]["mg"] = &mcmodel.Attribute{
		Name: "mg",
		AttributeValues: []mcmodel.AttributeValue{
			{
				ValueType:  mcmodel.ValueTypeFloat,
				ValueFloat: 0.4,
			},
		},
	}
	db.SampleAttributesBySampleIDAndStates[2][3]["ductility"] = &mcmodel.Attribute{
		Name: "ductility",
		AttributeValues: []mcmodel.AttributeValue{
			{
				ValueType:  mcmodel.ValueTypeFloat,
				ValueFloat: 0.81,
			},
		},
	}
	db.SampleAttributesBySampleIDAndStates[2][3]["alloy"] = &mcmodel.Attribute{
		Name: "alloy",
		AttributeValues: []mcmodel.AttributeValue{
			{
				ValueType:   mcmodel.ValueTypeString,
				ValueString: "zn45",
			},
		},
	}

	db.SampleAttributesBySampleIDAndStates[2][4] = make(map[string]*mcmodel.Attribute)
	db.SampleAttributesBySampleIDAndStates[2][4]["zn"] = &mcmodel.Attribute{
		Name: "zn",
		AttributeValues: []mcmodel.AttributeValue{
			{
				ValueType:  mcmodel.ValueTypeFloat,
				ValueFloat: 0.6,
			},
		},
	}
	db.SampleAttributesBySampleIDAndStates[2][4]["mg"] = &mcmodel.Attribute{
		Name: "mg",
		AttributeValues: []mcmodel.AttributeValue{
			{
				ValueType:  mcmodel.ValueTypeFloat,
				ValueFloat: 0.3,
			},
		},
	}
	db.SampleAttributesBySampleIDAndStates[2][4]["bend"] = &mcmodel.Attribute{
		Name: "bend",
		AttributeValues: []mcmodel.AttributeValue{
			{
				ValueType:   mcmodel.ValueTypeString,
				ValueString: "Right",
			},
		},
	}

	db.Samples = append(db.Samples, mcmodel.Entity{
		Name: "S3",
		ID:   3,
		EntityStates: []mcmodel.EntityState{
			{ID: 5},
			{ID: 6},
		},
	})
	db.SampleAttributesBySampleIDAndStates[3] = make(map[int]map[string]*mcmodel.Attribute)

	// Allocate for entity states
	db.SampleAttributesBySampleIDAndStates[3][5] = make(map[string]*mcmodel.Attribute)
	db.SampleAttributesBySampleIDAndStates[3][5]["zn"] = &mcmodel.Attribute{
		Name: "zn",
		AttributeValues: []mcmodel.AttributeValue{
			{
				ValueType:  mcmodel.ValueTypeFloat,
				ValueFloat: 0.68,
			},
		},
	}
	db.SampleAttributesBySampleIDAndStates[3][5]["mg"] = &mcmodel.Attribute{
		Name: "mg",
		AttributeValues: []mcmodel.AttributeValue{
			{
				ValueType:  mcmodel.ValueTypeFloat,
				ValueFloat: 0.32,
			},
		},
	}

	db.SampleAttributesBySampleIDAndStates[3][6] = make(map[string]*mcmodel.Attribute)
	db.SampleAttributesBySampleIDAndStates[3][6]["zn"] = &mcmodel.Attribute{
		Name: "zn",
		AttributeValues: []mcmodel.AttributeValue{
			{
				ValueType:  mcmodel.ValueTypeFloat,
				ValueFloat: 0.45,
			},
		},
	}
	db.SampleAttributesBySampleIDAndStates[3][6]["mg"] = &mcmodel.Attribute{
		Name: "mg",
		AttributeValues: []mcmodel.AttributeValue{
			{
				ValueType:  mcmodel.ValueTypeFloat,
				ValueFloat: 0.45,
			},
		},
	}

	return db
}

func TestSimpleProcessQueries(t *testing.T) {
	db := createTestDB()
	processMatchStatement := MatchStatement{
		FieldType: ProcessFieldType,
		FieldName: "name",
		Operation: "=",
		Value:     "Texture",
	}

	matchingProcesses, _ := EvalStatement(db, processMatchStatement)
	if len(matchingProcesses) != 2 {
		t.Fatalf("Expected 2 matches on name = 'Texture', but got %d", len(matchingProcesses))
	}

	////////////////////////////////////

	processAttributeMatchStatement := MatchStatement{
		FieldType: ProcessAttributeFieldType,
		FieldName: "frames per second",
		Operation: ">",
		Value:     3,
	}

	matchingProcesses, _ = EvalStatement(db, processAttributeMatchStatement)
	if len(matchingProcesses) != 1 {
		t.Fatalf("Expected 1 match on attribute 'frames per second' > 2, but got %d", len(matchingProcesses))
	}

	orStatement := OrStatement{
		Left:  processMatchStatement,
		Right: processAttributeMatchStatement,
	}

	matchingProcesses, _ = EvalStatement(db, orStatement)
	if len(matchingProcesses) != 3 {
		t.Fatalf("Expected 3 matches on: name = 'Texture' or attribute 'frames per second' > 2, but got %d", len(matchingProcesses))
	}

	////////////////////////////////////

	andStatement := AndStatement{
		Left: MatchStatement{
			FieldType: ProcessAttributeFieldType,
			FieldName: "note",
			Operation: "=",
			Value:     "ignore these results",
		},
		Right: MatchStatement{
			FieldType: ProcessFieldType,
			FieldName: "name",
			Operation: "=",
			Value:     "Texture",
		},
	}

	matchingProcesses, _ = EvalStatement(db, andStatement)
	if len(matchingProcesses) != 1 {
		t.Fatalf("Expected 1 match on: process name = 'Texture' and process attribute 'note' = 'ignore these results', but got %d", len(matchingProcesses))
	}

	orStatement.Left = andStatement
	orStatement.Right = MatchStatement{
		FieldType: ProcessAttributeFieldType,
		FieldName: "Beam Type",
		Operation: "=",
		Value:     "Wide",
	}

	matchingProcesses, _ = EvalStatement(db, orStatement)
	if len(matchingProcesses) != 2 {
		t.Fatalf("Expected 2 matches on: (process name = 'Texture' and process attribute 'note' = 'ignore these results') or process attribute 'Beam Type' = 'Wide', but got %d", len(matchingProcesses))
	}
}

func TestComplexAndOrStatementProcessQuery(t *testing.T) {
	db := createTestDB()
	leftSideOfOrStatement := AndStatement{
		Left: MatchStatement{
			FieldType: ProcessAttributeFieldType,
			FieldName: "note",
			Operation: "=",
			Value:     "ignore these results",
		},
		Right: MatchStatement{
			FieldType: ProcessFieldType,
			FieldName: "name",
			Operation: "=",
			Value:     "Texture",
		},
	}

	rightSideOfOrStatement := OrStatement{
		Left: MatchStatement{
			FieldType: ProcessAttributeFieldType,
			FieldName: "Beam Type",
			Operation: "=",
			Value:     "Wide",
		},
		Right: MatchStatement{
			FieldType: ProcessAttributeFieldType,
			FieldName: "frames per second",
			Operation: "=",
			Value:     3,
		},
	}

	orStatement := OrStatement{
		Left:  leftSideOfOrStatement,
		Right: rightSideOfOrStatement,
	}

	matchingProcesses, _ := EvalStatement(db, orStatement)
	if len(matchingProcesses) != 3 {
		fmt.Printf("matching = %+v\n", matchingProcesses)
		t.Fatalf(`Expected 3 matches on: 
(process name = 'Texture' and process attribute 'note' = 'ignore these results') or
(process attribute 'Beam Type' = 'Wide' or process attribute 'frames per second' = 3), but got %d`, len(matchingProcesses))
	}
}

func TestSimpleSampleQueries(t *testing.T) {
	db := createTestDB()
	// Test simple match on sample name
	sampleNameMatchStatement := MatchStatement{
		FieldType: SampleFieldType,
		FieldName: "name",
		Operation: "=",
		Value:     "S1",
	}

	_, matchingSamples := EvalStatement(db, sampleNameMatchStatement)
	if len(matchingSamples) != 1 {
		t.Fatalf("Expected 1 match on: name = 'S1', got %d", len(matchingSamples))
	}

	// Test simple match on sample attribute
	sampleAttributeMatchStatement := MatchStatement{
		FieldType: SampleAttributeFieldType,
		FieldName: "alloy",
		Operation: "=",
		Value:     "zn45",
	}

	_, matchingSamples = EvalStatement(db, sampleAttributeMatchStatement)
	if len(matchingSamples) != 1 {
		t.Fatalf("Expected 1 match on: sample attribute 'alloy' = 'zn45', got %d", len(matchingSamples))
	}

	// Test simple or statement using the above two statements
	orStatement := OrStatement{
		Left:  sampleNameMatchStatement,
		Right: sampleAttributeMatchStatement,
	}

	_, matchingSamples = EvalStatement(db, orStatement)
	if len(matchingSamples) != 2 {
		t.Fatalf(`Expected 1 match on: 
name = 'S1' or
sample attribute 'alloy' = 'zn45', got %d`, len(matchingSamples))
	}
}

func TestComplexAndOrStatementSampleQuery(t *testing.T) {
	db := createTestDB()
	// Matches sample S1 in entity state 2 attributes
	leftSideOfOrStatement := AndStatement{
		Left: MatchStatement{
			FieldType: SampleAttributeFieldType,
			FieldName: "zn",
			Operation: "=",
			Value:     0.5,
		},
		Right: MatchStatement{
			FieldType: SampleAttributeFieldType,
			FieldName: "mg",
			Operation: "=",
			Value:     0.5,
		},
	}

	// Matches S2, entity state 3 for Left, and matches nothing on right
	rightSideOfOrStatement := OrStatement{
		Left: MatchStatement{
			FieldType: SampleAttributeFieldType,
			FieldName: "ductility",
			Operation: "=",
			Value:     0.81,
		},
		Right: MatchStatement{
			FieldType: SampleAttributeFieldType,
			FieldName: "no-such",
			Operation: "=",
			Value:     0.5,
		},
	}

	orStatement := OrStatement{
		Left:  leftSideOfOrStatement,
		Right: rightSideOfOrStatement,
	}

	_, matchingSamples := EvalStatement(db, orStatement)
	if len(matchingSamples) != 2 {
		t.Fatalf(`Expected x matches on:
(sample attribute 'zn' = 0.5 and sample attribute 'mg' = 0.5) or
(sample attribute 'ductility' = 0.81 or sample attribute 'no-such' = 0.5)
, got %d`, len(matchingSamples))
	}
}

func TestLoadingFromSQLDB(t *testing.T) {
	dsn := "mc:mcpw@tcp(127.0.0.1:3306)/mc?charset=utf8mb4&parseTime=True&loc=Local"
	mysqldb, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Errorf("Failed to open db: %s", err)
	}

	db := NewDB(77, mysqldb)
	if err := db.Load(); err != nil {
		t.Fatalf("Failed loading database: %s", err)
	}
}
