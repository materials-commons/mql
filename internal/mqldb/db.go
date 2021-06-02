package mqldb

import (
	"github.com/apex/log"
	"github.com/materials-commons/gomcdb/mcmodel"
	"gorm.io/gorm"
)

type DB struct {
	ProjectID                    int
	db                           *gorm.DB
	Processes                    []mcmodel.Activity
	AllProcessAttributes         []*mcmodel.Attribute
	ProcessAttributesByProcessID map[int]map[string]*mcmodel.Attribute
	Samples                      []mcmodel.Entity
	AllSampleAttributes          []*mcmodel.Attribute

	// A sample can have multiple states, thus this maps a sample id to a list of states, then each
	// state is a hash map of the attribute name to the attribute structure. For example, given a
	// sample with ID 1, that has 2 states (with IDs of 4, and 5), where each state has two attributes,
	// each name "AttrA" and "AttrB", then you would end up with SampleAttribytesBySampleIDAndStates
	// looking as follows:
	// [1] <-- Sample id key into hash map (map[int]...)
	//    [4] <-- First entity state id in entry for the first state (map[int]map[int]...)
	//       ["AttrA"] = &mcmodel.Attribute  <- AttrA is the key into map[string]
	//       ["AttrB"] = &mcmodel.Attribute
	//    [5]
	//       ["AttrA"] = &mcmodel.Attribute
	//       ["AttrB"] = &mcmodel.Attribute
	// ie map[1]map[4]map["AttrA"] = &mcmodel.Attribute{Name: "AttrA", Value...}
	//    map[1]map[4]map["AttrB"] = &mcmodel.Attribute{Name: "AttrB", Value...}
	//    map[1]map[5]map["AttrA"] = &mcmodel.Attribute{Name: "AttrA", Value (different than AttrA above)}
	//    map[1]map[5]map["AttrB"] = &mcmodel.Attribute{Name: "AttrB", Value (different than AttrB above)}

	SampleAttributesBySampleIDAndStates map[int]map[int]map[string]*mcmodel.Attribute
}

func NewDB(projectID int, db *gorm.DB) *DB {
	return &DB{
		ProjectID:                           projectID,
		db:                                  db,
		ProcessAttributesByProcessID:        make(map[int]map[string]*mcmodel.Attribute),
		SampleAttributesBySampleIDAndStates: make(map[int]map[int]map[string]*mcmodel.Attribute),
	}
}

func (db *DB) Load() error {
	if err := db.db.Where("project_id = ?", db.ProjectID).Find(&db.Processes).Error; err != nil {
		return err
	}

	err := db.db.Preload("AttributeValues").Where("attributable_type = ?", "App\\Models\\Activity").
		Where("attributable_id in (select id from activities where project_id = ?)", db.ProjectID).
		Find(&db.AllProcessAttributes).Error
	if err != nil {
		return err
	}

	for _, process := range db.Processes {
		db.ProcessAttributesByProcessID[process.ID] = make(map[string]*mcmodel.Attribute)
	}

	for i, attr := range db.AllProcessAttributes {
		db.ProcessAttributesByProcessID[attr.AttributableID][attr.Name] = db.AllProcessAttributes[i]
		if err := attr.LoadValues(); err != nil {
			log.Errorf("Failed converting attribute %d/%s values: %s", attr.ID, attr.Name, err)
		}
	}

	err = db.db.Preload("EntityStates").Where("project_id = ?", db.ProjectID).Find(&db.Samples).Error
	if err != nil {
		return err
	}

	err = db.db.Preload("AttributeValues").Where("attributable_type = ?", "App\\Models\\EntityState").
		Where(`attributable_id in 
                       (select distinct id from entity_states where entity_id in 
                               (select id from entities where project_id = ?))`, db.ProjectID).
		Find(&db.AllSampleAttributes).Error
	if err != nil {
		return err
	}

	// Create a map of entity state ids to sample state ids because the attributes are
	// all going to have an entity state id associated with them and we need to figure
	// out which sample that state is associated with. Also create hash entries for the
	// samples in SampleAttributeBySampleIDAndStates
	entityStateIDToSampleID := make(map[int]int)
	for _, sample := range db.Samples {
		db.SampleAttributesBySampleIDAndStates[sample.ID] = make(map[int]map[string]*mcmodel.Attribute)
		for _, entityState := range sample.EntityStates {
			entityStateIDToSampleID[entityState.ID] = sample.ID
			db.SampleAttributesBySampleIDAndStates[sample.ID][entityState.ID] = make(map[string]*mcmodel.Attribute)
		}
	}

	// Load the SampleAttributesBySampleIDAndStates map of values
	for i, attr := range db.AllSampleAttributes {
		// Here AttributableType == "App\Models\EntityState" and AttributableID == EntityState.ID
		sampleID := entityStateIDToSampleID[attr.AttributableID]
		db.SampleAttributesBySampleIDAndStates[sampleID][attr.AttributableID][attr.Name] = db.AllSampleAttributes[i]
	}

	return nil
}
