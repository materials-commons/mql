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
		ProjectID:                    1,
		ProcessAttributesByProcessID: make(map[int]map[string]*mcmodel.Attribute),
	}

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
		Value: mcmodel.AttributeValue{
			ValueType:   mcmodel.ValueTypeString,
			ValueString: "Wide",
		},
	}
	db.ProcessAttributesByProcessID[1]["frames per second"] = &mcmodel.Attribute{
		Name: "frames per second",
		Value: mcmodel.AttributeValue{
			ValueType: mcmodel.ValueTypeInt,
			ValueInt:  5,
		},
	}
	db.ProcessAttributesByProcessID[1]["note"] = &mcmodel.Attribute{
		Name: "note",
		Value: mcmodel.AttributeValue{
			ValueType:   mcmodel.ValueTypeString,
			ValueString: "ignore these results",
		},
	}

	// Second EBSD process attributes
	db.ProcessAttributesByProcessID[2] = make(map[string]*mcmodel.Attribute)
	db.ProcessAttributesByProcessID[2]["Beam Type"] = &mcmodel.Attribute{
		Name: "Beam Type",
		Value: mcmodel.AttributeValue{
			ValueType:   mcmodel.ValueTypeString,
			ValueString: "Thin",
		},
	}
	db.ProcessAttributesByProcessID[2]["frames per second"] = &mcmodel.Attribute{
		Name: "frames per second",
		Value: mcmodel.AttributeValue{
			ValueType: mcmodel.ValueTypeInt,
			ValueInt:  3,
		},
	}

	// First Texture process attributes
	db.ProcessAttributesByProcessID[3] = make(map[string]*mcmodel.Attribute)
	db.ProcessAttributesByProcessID[3]["PF scale max"] = &mcmodel.Attribute{
		Name: "PF scale max",
		Value: mcmodel.AttributeValue{
			ValueType: mcmodel.ValueTypeInt,
			ValueInt:  2,
		},
	}
	db.ProcessAttributesByProcessID[3]["note"] = &mcmodel.Attribute{
		Name: "note",
		Value: mcmodel.AttributeValue{
			ValueType:   mcmodel.ValueTypeString,
			ValueString: "ignore these results",
		},
	}

	// Second Texture process attributes
	db.ProcessAttributesByProcessID[4] = make(map[string]*mcmodel.Attribute)
	db.ProcessAttributesByProcessID[4]["PF scale max"] = &mcmodel.Attribute{
		Name: "PF scale max",
		Value: mcmodel.AttributeValue{
			ValueType: mcmodel.ValueTypeInt,
			ValueInt:  3,
		},
	}

	return db
}

func TestWritingProcessQuery(t *testing.T) {
	db := createTestDB()
	processMatchStatement := MatchStatement{
		FieldType: ProcessFieldType,
		FieldName: "name",
		Operation: "=",
		Value:     "Texture",
	}

	matching := EvalStatement(db, processMatchStatement)
	if len(matching) != 2 {
		t.Fatalf("Expected 2 matches on name = 'Texture', but got %d", len(matching))
	}

	////////////////////////////////////

	processAttributeMatchStatement := MatchStatement{
		FieldType: ProcessAttributeFieldType,
		FieldName: "frames per second",
		Operation: ">",
		Value:     3,
	}

	matching = EvalStatement(db, processAttributeMatchStatement)
	if len(matching) != 1 {
		t.Fatalf("Expected 1 match on attribute 'frames per second' > 2, but got %d", len(matching))
	}

	orStatement := OrStatement{
		Left:  processMatchStatement,
		Right: processAttributeMatchStatement,
	}

	matching = EvalStatement(db, orStatement)
	if len(matching) != 3 {
		t.Fatalf("Expected 3 matches on: name = 'Texture' or attribute 'frames per second' > 2, but got %d", len(matching))
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

	matching = EvalStatement(db, andStatement)
	if len(matching) != 1 {
		t.Fatalf("Expected 1 match on: process name = 'Texture' and process attribute 'note' = 'ignore these results', but got %d", len(matching))
	}

	orStatement.Left = andStatement
	orStatement.Right = MatchStatement{
		FieldType: ProcessAttributeFieldType,
		FieldName: "Beam Type",
		Operation: "=",
		Value:     "Wide",
	}

	matching = EvalStatement(db, orStatement)
	if len(matching) != 2 {
		t.Fatalf("Expected 2 matches on: (process name = 'Texture' and process attribute 'note' = 'ignore these results') or process attribute 'Beam Type' = 'Wide', but got %d", len(matching))
	}
}

func TestComplexAndOrStatement(t *testing.T) {
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

	matching := EvalStatement(db, orStatement)
	if len(matching) != 3 {
		fmt.Printf("matching = %+v\n", matching)
		t.Fatalf(`Expected 3 matches on: 
(process name = 'Texture' and process attribute 'note' = 'ignore these results') or
(process attribute 'Beam Type' = 'Wide' or process attribute 'frames per second' = 3), but got %d`, len(matching))
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
