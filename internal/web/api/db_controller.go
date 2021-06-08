package api

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/materials-commons/gomcdb/mcmodel"
	"github.com/materials-commons/mql/internal/mqldb"
	"gorm.io/gorm"
)

var (
	DB               *gorm.DB
	mutex            sync.Mutex
	mqlDBByProjectID map[int]*mqldb.DB
)

func Init(db *gorm.DB) {
	DB = db
	mqlDBByProjectID = make(map[int]*mqldb.DB)
}

func LoadProjectController(c echo.Context) error {
	var req struct {
		ProjectID int `json:"project_id"`
	}

	if err := c.Bind(&req); err != nil {
		return err
	}

	mutex.Lock()
	defer mutex.Unlock()

	if _, ok := mqlDBByProjectID[req.ProjectID]; ok {
		// Project already loaded nothing to do
		return nil
	}

	return loadProjectDB(req.ProjectID)
}

func ReloadProjectController(c echo.Context) error {
	var req struct {
		ProjectID int `json:"project_id"`
	}

	if err := c.Bind(&req); err != nil {
		return err
	}

	mutex.Lock()
	defer mutex.Unlock()

	return loadProjectDB(req.ProjectID)
}

func ExecuteQueryController(c echo.Context) error {
	var req struct {
		Statement       mqldb.Statement `json:"statement"`
		ProjectID       int             `json:"project_id"`
		SelectProcesses bool            `json:"select_processes"`
		SelectSamples   bool            `json:"select_samples"`
	}

	if err := c.Bind(&req); err != nil {
		return err
	}

	if req.ProjectID == 0 {
		return fmt.Errorf("unknown project: %d", req.ProjectID)
	}

	mutex.Lock()
	defer mutex.Unlock()

	db, ok := mqlDBByProjectID[req.ProjectID]
	if !ok {
		return fmt.Errorf("project not loaded")
	}

	selection := mqldb.Selection{
		SampleSelection: mqldb.SampleSelection{
			All: req.SelectSamples,
		},
		ProcessSelection: mqldb.ProcessSelection{
			All: req.SelectProcesses,
		},
	}
	var resp struct {
		Processes []mcmodel.Activity `json:"processes"`
		Samples   []mcmodel.Entity   `json:"samples"`
	}

	resp.Processes, resp.Samples = mqldb.EvalStatement(db, selection, req.Statement)

	return c.JSON(http.StatusOK, &resp)
}

// loadProjectDB will load the mqldb for the project and save it into mqlDBByProjectID. It does not attempt to lock
// access to mqlDBByProjectID. If this is important then the call must acquire the mutex.Lock().
func loadProjectDB(projectID int) error {
	db := mqldb.NewDB(projectID, DB)
	if err := db.Load(); err != nil {
		// do something
		return err
	}

	mqlDBByProjectID[projectID] = db
	return nil
}
