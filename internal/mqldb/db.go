package mqldb

import (
	"github.com/materials-commons/gomcdb/mcmodel"
)

type DB struct {
	ProjectID int
	Processes []*mcmodel.Activity
	samples   []*mcmodel.Entity
	//ProcessAttributesByProcessID map[int]*mcmodel.
}
