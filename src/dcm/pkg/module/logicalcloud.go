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
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/status"

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
	GenericStatus(project, name, qStatusInstance, qType, qOutput string, fClusters, fResources []string) (status.StatusResult, error)
	StatusClusters(p, lc, qStatusInstance string) (status.LogicalCloudClustersStatus, error)
	StatusResources(p, lc, qStatusInstance, qType string, fClusters []string) (status.LogicalCloudResourcesStatus, error)
	Status(p, lc, qStatusInstance, qType, qOutput string, fClusters, fResources []string) (status.LogicalCloudStatus, error)
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
	logicalCloud, err := v.Get(project, logicalCloudName)
	if err != nil {
		return LogicalCloud{}, err
	}
	err = db.DBconn.Insert(v.storeName, key, nil, v.tagMeta, c)
	if err != nil {
		return LogicalCloud{}, pkgerrors.Wrap(err, "Updating DB Entry")
	}

	// If Logical Cloud was already instantiated, then prepare new appcontext to give to rsync
	lcStateInfo, err := v.GetState(project, logicalCloudName)
	if err != nil {
		return LogicalCloud{}, err
	}
	log.Debug("", log.Fields{"lcStateInfo": lcStateInfo})
	oldCID := state.GetLastContextIdFromStateInfo(lcStateInfo)
	if oldCID != "" {
		log.Debug("", log.Fields{"oldCID": oldCID})
		oldContext, err := state.GetAppContextFromId(oldCID)
		if err != nil {
			return LogicalCloud{}, err
		}

		// Since there's a context associated, if the logical cloud isn't fully Terminated then prevent
		// clusters from being added since this is a functional scenario not currently supported
		contextStatus, err := GetAppContextStatus(oldContext)
		if err != nil {
			return LogicalCloud{}, pkgerrors.New("Logical Cloud is not in a state where a cluster can be created")
		}
		// If Logical Cloud is instantiated, it's safe to proceed with an updated-instantiation
		if contextStatus.Status == appcontext.AppContextStatusEnum.Instantiated {
			// We need to know which clusters to instantiate on
			clusterList, err := NewClusterClient().GetAllClusters(project, logicalCloudName)
			if err != nil {
				return LogicalCloud{}, err
			}

			// We need to know the Level as that influences how to build the appcontext
			level := logicalCloud.Specification.Level

			var newCID string
			// Prepare new appcontext to replace previous one
			if level == "1" {
				// For L1, we need to know what quotas and user permissions to use
				quotaList, err := NewQuotaClient().GetAllQuotas(project, logicalCloudName)
				if err != nil {
					return LogicalCloud{}, err
				}
				userPermissionList, err := NewUserPermissionClient().GetAllUserPerms(project, logicalCloudName)
				if err != nil {
					return LogicalCloud{}, err
				}
				_, newCID, err = blindInstantiateL1(project, logicalCloud, v, clusterList, quotaList, userPermissionList)
				log.Debug("", log.Fields{"newCID": newCID})
			} else if level == "0" {
				_, newCID, err = blindInstantiateL0(project, logicalCloud, v, clusterList)
				log.Debug("", log.Fields{"newCID": newCID})
			}
			if err != nil {
				return LogicalCloud{}, err
			}

			// Update DB Status CID
			err = state.UpdateAppContextStatusContextID(newCID, oldCID)
			if err != nil {
				return LogicalCloud{}, err
			}

			// Call rsync to update Logical Cloud in clusters (calculate differences and bring clusters up to DCM state)
			err = callRsyncUpdate(oldCID, newCID)
			if err != nil {
				log.Error("Failed calling rsync update", log.Fields{"err": err})
				return LogicalCloud{}, pkgerrors.Wrap(err, "Failed calling rsync update")
			}

			latestRev, err := state.GetLatestRevisionFromStateInfo(lcStateInfo)
			if err != nil {
				log.Error("Latest revision not found", log.Fields{})
				return LogicalCloud{}, err
			}
			// TODO: make atomic
			newRev := latestRev + 1

			a := state.ActionEntry{
				State:     state.StateEnum.Updated,
				ContextId: newCID,
				TimeStamp: time.Now(),
				Revision:  newRev,
			}
			lcStateInfo.StatusContextId = newCID
			lcStateInfo.Actions = append(lcStateInfo.Actions, a)

			err = db.DBconn.Insert(v.storeName, key, nil, v.tagState, lcStateInfo)
			if err != nil {
				log.Error("Error updating the state info of the LogicalCloud: ", log.Fields{"logicalCloud": logicalCloud})
				return LogicalCloud{}, err
			}

			// TODO: enhancement: also check if any L1 cluster actually got added, if not then no need for ReadyNotify:
			if level == "1" {
				// Call rsync grpc streaming api, which launches a goroutine to wait for the response of
				// every cluster (function should know how many clusters are expected and only finish when
				// all respective certificates have been obtained and all kubeconfigs stored in CloudConfig)
				err = callRsyncReadyNotify(newCID)
				if err != nil {
					log.Error("Failed calling rsync ready-notify", log.Fields{"err": err})
					return LogicalCloud{}, pkgerrors.Wrap(err, "Failed calling rsync ready-notify")
				}
			}
		}
		return LogicalCloud{}, nil
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

func (v *LogicalCloudClient) GenericStatus(p, lc, qStatusInstance, qType, qOutput string, fClusters, fResources []string) (status.StatusResult, error) {
	stateInfo, err := v.GetState(p, lc)
	if err != nil {
		return status.StatusResult{}, pkgerrors.Wrap(err, "LogicalCloud state not found: "+lc)
	}

	qInstance, err := state.GetContextIdForStatusContextId(stateInfo, qStatusInstance)
	if err != nil {
		return status.StatusResult{}, err
	}

	statusResponse, err := status.GenericPrepareStatusResult(status.LcStatusQuery, stateInfo, qInstance, qType, qOutput, make([]string, 0), fClusters, fResources)
	if err != nil {
		return status.StatusResult{}, err
	}
	statusResponse.Name = lc

	return statusResponse, nil
}

func (v *LogicalCloudClient) StatusClusters(p, lc, qStatusInstance string) (status.LogicalCloudClustersStatus, error) {
	lcState, err := v.GetState(p, lc)
	if err != nil {
		return status.LogicalCloudClustersStatus{}, pkgerrors.Wrap(err, "Logical Cloud state not found")
	}

	statusResponse, err := status.PrepareClustersByAppStatusResult(lcState, qStatusInstance, []string{"logical-cloud"})
	if err != nil {
		return status.LogicalCloudClustersStatus{}, err
	}
	statusResponse.Name = lc
	lcStatus := status.LogicalCloudClustersStatus{
		Project:             p,
		LogicalCloud:        lc,
		ClustersByAppResult: statusResponse,
	}

	return lcStatus, nil
}

func (v *LogicalCloudClient) StatusResources(p, lc, qStatusInstance, qType string, fClusters []string) (status.LogicalCloudResourcesStatus, error) {
	lcState, err := v.GetState(p, lc)
	if err != nil {
		return status.LogicalCloudResourcesStatus{}, pkgerrors.Wrap(err, "Logical Cloud state not found")
	}

	statusResponse, err := status.PrepareResourcesByAppStatusResult(lcState, qStatusInstance, qType, []string{"logical-cloud"}, fClusters)
	if err != nil {
		return status.LogicalCloudResourcesStatus{}, err
	}
	statusResponse.Name = lc
	lcStatus := status.LogicalCloudResourcesStatus{
		Project:              p,
		LogicalCloud:         lc,
		ResourcesByAppResult: statusResponse,
	}

	return lcStatus, nil
}

func (v *LogicalCloudClient) Status(p, lc, qStatusInstance, qType, qOutput string, fClusters, fResources []string) (status.LogicalCloudStatus, error) {
	lcState, err := v.GetState(p, lc)
	if err != nil {
		return status.LogicalCloudStatus{}, pkgerrors.Wrap(err, "Logical Cloud state not found")
	}

	statusResponse, err := status.PrepareStatusResult(lcState, qStatusInstance, qType, qOutput, make([]string, 0), fClusters, fResources)
	if err != nil {
		return status.LogicalCloudStatus{}, err
	}
	statusResponse.Name = lc
	lcStatus := status.LogicalCloudStatus{
		Project:      p,
		LogicalCloud: lc,
		StatusResult: statusResponse,
	}

	return lcStatus, nil
}
