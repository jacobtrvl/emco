// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"encoding/json"
	"time"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"

	pkgerrors "github.com/pkg/errors"
)

// LogicalCloud contains the parameters needed for a Logical Cloud
type LogicalCloud struct {
	MetaData      MetaDataList `json:"metadata"`
	Specification Spec         `json:"spec"`
}

// MetaData contains the parameters needed for metadata
type MetaDataList struct {
	LogicalCloudName string `json:"name"`
	Description      string `json:"description"`
	UserData1        string `json:"userData1"`
	UserData2        string `json:"userData2"`
}

// Spec contains the parameters needed for spec
type Spec struct {
	NameSpace string            `json:"namespace"`
	Labels    map[string]string `json:"labels"`
	Level     string            `json:"level"`
	User      UserData          `json:"user"`
}

// UserData contains the parameters needed for user
type UserData struct {
	UserName string `json:"userName"`
	Type     string `json:"type"`
}

// LogicalCloudKey is the key structure that is used in the database
type LogicalCloudKey struct {
	Project          string `json:"project"`
	LogicalCloudName string `json:"logicalCloud"`
}

// LogicalCloudManager is an interface that exposes the connection
// functionality
type LogicalCloudManager interface {
	Create(project string, c LogicalCloud) (LogicalCloud, error)
	Get(project, name string) (LogicalCloud, error)
	GetAll(project string) ([]LogicalCloud, error)
	GetState(p string, lc string) (state.StateInfo, error)
	Delete(project, name string) error
	Update(project, name string, c LogicalCloud) (LogicalCloud, error)
}

// LogicalCloudClient implements the LogicalCloudManager
// It will also be used to maintain some localized state
type LogicalCloudClient struct {
	storeName string
	tagMeta   string
	tagState  string
}

// LogicalCloudClient returns an instance of the LogicalCloudClient
// which implements the LogicalCloudManager
func NewLogicalCloudClient() *LogicalCloudClient {
	return &LogicalCloudClient{
		storeName: "resources",
		tagMeta:   "data",
		tagState:  "stateInfo",
	}
}

// Create entry for the logical cloud resource in the database
func (v *LogicalCloudClient) Create(project string, c LogicalCloud) (LogicalCloud, error) {

	//Construct key consisting of name
	key := LogicalCloudKey{
		Project:          project,
		LogicalCloudName: c.MetaData.LogicalCloudName,
	}

	//Check if this Logical Cloud already exists
	_, err := v.Get(project, c.MetaData.LogicalCloudName)
	if err == nil {
		return LogicalCloud{}, pkgerrors.New("Logical Cloud already exists")
	}

	// if Logical Cloud Level is not specified, it defaults to 1:
	if c.Specification.Level == "" {
		c.Specification.Level = "1"
	}

	err = db.DBconn.Insert(v.storeName, key, nil, v.tagMeta, c)
	if err != nil {
		return LogicalCloud{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	// Add the state info record (initial state)
	s := state.StateInfo{}
	a := state.ActionEntry{
		State:     state.StateEnum.Created,
		ContextId: "",
		TimeStamp: time.Now(),
	}
	s.Actions = append(s.Actions, a)

	err = db.DBconn.Insert(v.storeName, key, nil, v.tagState, s)
	if err != nil {
		return LogicalCloud{}, pkgerrors.Wrap(err, "Error updating the state info of the LogicalCloud: "+c.MetaData.LogicalCloudName)
	}

	return c, nil
}

// Get returns Logical Cloud corresponding to logical cloud name
func (v *LogicalCloudClient) Get(project, logicalCloudName string) (LogicalCloud, error) {

	//Construct the composite key to select the entry
	key := LogicalCloudKey{
		Project:          project,
		LogicalCloudName: logicalCloudName,
	}
	value, err := db.DBconn.Find(v.storeName, key, v.tagMeta)
	if err != nil {
		return LogicalCloud{}, err
	}

	if len(value) == 0 {
		return LogicalCloud{}, pkgerrors.New("Logical Cloud not found")
	}

	//value is a byte array
	if value != nil {
		lc := LogicalCloud{}
		err = db.DBconn.Unmarshal(value[0], &lc)
		if err != nil {
			return LogicalCloud{}, err
		}
		return lc, nil
	}

	return LogicalCloud{}, pkgerrors.New("Unknown Error")
}

// GetAll returns Logical Clouds in the project
func (v *LogicalCloudClient) GetAll(project string) ([]LogicalCloud, error) {

	//Construct the composite key to select the entry
	key := LogicalCloudKey{
		Project:          project,
		LogicalCloudName: "",
	}

	var resp []LogicalCloud
	values, err := db.DBconn.Find(v.storeName, key, v.tagMeta)
	if err != nil {
		return []LogicalCloud{}, err
	}

	for _, value := range values {
		lc := LogicalCloud{}
		err = db.DBconn.Unmarshal(value, &lc)
		if err != nil {
			return []LogicalCloud{}, err
		}
		resp = append(resp, lc)
	}

	return resp, nil
}

// GetState returns the LogicalCloud StateInfo with a given logical cloud name and project
func (v *LogicalCloudClient) GetState(p string, lc string) (state.StateInfo, error) {

	key := LogicalCloudKey{
		Project:          p,
		LogicalCloudName: lc,
	}

	result, err := db.DBconn.Find(v.storeName, key, v.tagState)
	if err != nil {
		return state.StateInfo{}, err
	}

	if len(result) == 0 {
		return state.StateInfo{}, pkgerrors.New("LogicalCloud StateInfo not found")
	}

	if result != nil {
		s := state.StateInfo{}
		err = db.DBconn.Unmarshal(result[0], &s)
		if err != nil {
			return state.StateInfo{}, err
		}
		return s, nil
	}

	return state.StateInfo{}, pkgerrors.New("Unknown Error")
}

// Delete the Logical Cloud entry from database
func (v *LogicalCloudClient) Delete(project, logicalCloudName string) error {

	//Construct the composite key to select the entry
	key := LogicalCloudKey{
		Project:          project,
		LogicalCloudName: logicalCloudName,
	}
	//Check if this Logical Cloud exists
	_, err := v.Get(project, logicalCloudName)
	if err != nil {
		return err
	}

	// Check if there was a previous context for this logical cloud
	s, err := v.GetState(project, logicalCloudName)
	if err != nil {
		return err
	}
	cid := state.GetLastContextIdFromStateInfo(s)

	// If there's no context for Logical Cloud, just go ahead and delete it now
	if cid == "" {
		err = db.DBconn.Remove(v.storeName, key)
		if err != nil {
			return pkgerrors.Wrap(err, "Error when deleting Logical Cloud (scenario with no context)")
		}
		return nil
	}

	ac, err := state.GetAppContextFromId(cid)
	if err != nil {
		return err
	}

	// Make sure rsync status for this logical cloud is Terminated,
	// otherwise we can't re-instantiate logical cloud yet
	acStatus, err := GetAppContextStatus(ac)
	if err != nil {
		return err
	}
	switch acStatus.Status {
	case appcontext.AppContextStatusEnum.Terminating:
		log.Error("The Logical Cloud can't be deleted yet, it is being terminated", log.Fields{"logicalcloud": logicalCloudName})
		return pkgerrors.New("The Logical Cloud can't be deleted yet, it is being terminated")
	case appcontext.AppContextStatusEnum.Instantiated:
		log.Error("The Logical Cloud is instantiated, please terminate first", log.Fields{"logicalcloud": logicalCloudName})
		return pkgerrors.New("The Logical Cloud is instantiated, please terminate first")
	case appcontext.AppContextStatusEnum.Instantiating:
		log.Error("The Logical Cloud is instantiating, please wait and then terminate", log.Fields{"logicalcloud": logicalCloudName})
		return pkgerrors.New("The Logical Cloud is instantiating, please wait and then terminate")
	case appcontext.AppContextStatusEnum.InstantiateFailed:
		log.Error("The Logical Cloud has failed instantiating, for safety please terminate and try again", log.Fields{"logicalcloud": logicalCloudName})
		return pkgerrors.New("The Logical Cloud has failed instantiating, for safety please terminate and try again")
	case appcontext.AppContextStatusEnum.TerminateFailed:
		log.Info("The Logical Cloud has failed terminating, proceeding with the delete operation", log.Fields{"logicalcloud": logicalCloudName})
		// try to delete anyway since termination failed
		fallthrough
	case appcontext.AppContextStatusEnum.Terminated:
		// remove the appcontext
		err := ac.DeleteCompositeApp()
		if err != nil {
			log.Error("Error deleting AppContext CompositeApp Logical Cloud", log.Fields{"logicalcloud": logicalCloudName})
			return pkgerrors.Wrap(err, "Error deleting AppContext CompositeApp Logical Cloud")
		}

		err = db.DBconn.Remove(v.storeName, key)
		if err != nil {
			log.Error("Error when deleting Logical Cloud (scenario with Terminated status)", log.Fields{"logicalcloud": logicalCloudName})
			return pkgerrors.Wrap(err, "Error when deleting Logical Cloud (scenario with Terminated status)")
		}
		log.Info("Deleted Logical Cloud", log.Fields{"logicalcloud": logicalCloudName})
		return nil
	default:
		log.Error("The Logical Cloud isn't in an expected status so not taking any action", log.Fields{"logicalcloud": logicalCloudName, "status": acStatus.Status})
		return pkgerrors.New("The Logical Cloud isn't in an expected status so not taking any action")
	}
}

// Update an entry for the Logical Cloud in the database
func (v *LogicalCloudClient) Update(project, logicalCloudName string, c LogicalCloud) (LogicalCloud, error) {

	key := LogicalCloudKey{
		Project:          project,
		LogicalCloudName: logicalCloudName,
	}
	// Check for mismatch, logicalCloudName and payload logical cloud name
	if c.MetaData.LogicalCloudName != logicalCloudName {
		return LogicalCloud{}, pkgerrors.New("Logical Cloud name mismatch")
	}
	//Check if this Logical Cloud exists
	_, err := v.Get(project, logicalCloudName)
	if err != nil {
		return LogicalCloud{}, err
	}
	err = db.DBconn.Insert(v.storeName, key, nil, v.tagMeta, c)
	if err != nil {
		return LogicalCloud{}, pkgerrors.Wrap(err, "Updating DB Entry")
	}
	return c, nil
}

// GetAppContextStatus returns the Status for a particular AppContext
func GetAppContextStatus(ac appcontext.AppContext) (*appcontext.AppContextStatus, error) {

	h, err := ac.GetCompositeAppHandle()
	if err != nil {
		return nil, err
	}
	sh, err := ac.GetLevelHandle(h, "status")
	if err != nil {
		return nil, err
	}
	s, err := ac.GetValue(sh)
	if err != nil {
		return nil, err
	}
	acStatus := appcontext.AppContextStatus{}
	js, _ := json.Marshal(s)
	json.Unmarshal(js, &acStatus)

	return &acStatus, nil
}
