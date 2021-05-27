package mqldb

import (
	"github.com/materials-commons/gomcdb/mcmodel"
	"gorm.io/gorm"
)

type DB struct {
	ProjectID                    int
	db                           *gorm.DB
	Processes                    []mcmodel.Activity
	AllProcessAttributes         []*mcmodel.Attribute
	ProcessAttributesByProcessID map[int]map[string]*mcmodel.Attribute
	samples                      []*mcmodel.Entity
}

func NewDB(projectID int, db *gorm.DB) *DB {
	return &DB{
		ProjectID:                    projectID,
		db:                           db,
		ProcessAttributesByProcessID: make(map[int]map[string]*mcmodel.Attribute),
	}
}

func (db *DB) Load() error {
	if err := db.db.Where("project_id = ?", db.ProjectID).Find(&db.Processes).Error; err != nil {
		return err
	}

	err := db.db.Where("attributable_type = ?", "App\\Models\\Activity").
		Where("attributable_id in (select id from activities where project_id = ?)", db.ProjectID).
		Find(&db.AllProcessAttributes).Error
	if err != nil {
		return err
	}

	for _, process := range db.Processes {
		db.ProcessAttributesByProcessID[process.ID] = make(map[string]*mcmodel.Attribute)
	}

	for i, attr := range db.AllProcessAttributes {
		//fmt.Printf("Adding to ProcessAttributeByProcessID %d/%s\n", attr.AttributableID, attr.Name)

		db.ProcessAttributesByProcessID[attr.AttributableID][attr.Name] = db.AllProcessAttributes[i]
		if err := attr.LoadValue(); err != nil {
			//fmt.Printf("Failed to load value for attribute %s/%d/%s: %s\n", attr.Name, attr.ID, attr.Val, err)
		}
		//	fmt.Println("Attribute val:", attr.Val)
		//	switch attr.Value.ValueType {
		//	case mcmodel.ValueTypeFloat:
		//		fmt.Printf("    Attribute %s is type float with value: %f\n", attr.Name, attr.Value.ValueFloat)
		//	case mcmodel.ValueTypeString:
		//		fmt.Printf("    Attribute %s is type string with value: '%s'\n", attr.Name, attr.Value.ValueString)
		//	case mcmodel.ValueTypeInt:
		//		fmt.Printf("    Attribute %s is type int with value: %d\n", attr.Name, attr.Value.ValueInt)
		//	default:
		//		fmt.Printf("    Attribute %s has unknown value type: %d\n", attr.Name, attr.Value.ValueType)
		//	}
	}
	return nil
}
