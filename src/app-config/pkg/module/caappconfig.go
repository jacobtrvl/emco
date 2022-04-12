package module

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	orchmodule "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"
	readynotifypb "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/grpc/readynotify"
)

// AppConfig consists of metadata and Spec
type AppConfig struct {
	Metadata Metadata      `json:"metadata"`
	Spec     appConfigSpec `json:"spec"`
}

// appConfigSpec consists of ClusterSpecific and ClusterInfo
type appConfigSpec struct {
	ResType         string      `json:"resourceType"`
	FileType        string      `json:"fileType"`
	ClusterSpecific string      `json:"clusterSpecific"`
	ClusterInfo     ClusterInfo `json:"clusterInfo"`
}

// ClusterInfo consists of scope, Clusterprovider, ClusterName, ClusterLabel and Mode
type ClusterInfo struct {
	Scope           string `json:"scope"`
	ClusterProvider string `json:"clusterProvider"`
	ClusterName     string `json:"cluster"`
	ClusterLabel    string `json:"clusterLabel"`
	Mode            string `json:"mode"`
}

// SpecFileContent contains the array of file contents
type SpecFileContent struct {
	FileContents []string
	FileNames    []string
}

// appConfigKey consists of AppConfigName, project, CompApp, CompAppVersion, DeploymentIntentGroupName, GenericIntentName, ResourceName
type appConfigKey struct {
	AppConfig           string `json:"appConfig"`
	Project             string `json:"project"`
	CompositeApp        string `json:"compositeApp"`
	CompositeAppVersion string `json:"compositeAppVersion"`
	DigName             string `json:"deploymentIntentGroup"`
}

// AppConfigManager exposes all the functionalities of appconfig
type AppConfigManager interface {
	CreateAppConfig(c AppConfig, t SpecFileContent, p, ca, cv, dig string, exists bool) (AppConfig, error)
	GetAppConfig(c, p, ca, cv, dig string) (AppConfig, error)
	GetAppConfigContent(c, p, ca, cv, dig string) (SpecFileContent, error)
	GetAllAppConfig(p, ca, cv, dig string) ([]AppConfig, error)
	DeleteAppConfig(c, p, ca, cv, dig string) error
}

// appConfigClientDbInfo consists of tableName and columns
type appConfigClientDbInfo struct {
	storeName  string // name of the mongodb collection to use for client documents
	tagMeta    string // attribute key name for the json data of a client document
	tagContent string // attribute key name for the file data of a client document
	tagState   string // attribute key name for the file data of a client document
}

// AppConfigClient consists of appConfigClientDbInfo
type AppConfigClient struct {
	db appConfigClientDbInfo
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (ck appConfigKey) String() string {
	out, err := json.Marshal(ck)
	if err != nil {
		return ""
	}
	return string(out)
}

// NewAppConfigClient returns an instance of the AppConfigClient
func NewAppConfigClient() *AppConfigClient {
	return &AppConfigClient{
		db: appConfigClientDbInfo{
			storeName:  "resources",
			tagMeta:    "data",
			tagContent: "caappconfigcontent",
			tagState:   "stateInfo",
		},
	}
}

type AppCfgContext struct {
	Lock   *sync.Mutex
	Client readynotifypb.ReadyNotifyClient
	App    string
}

type AppCfgContextData struct {
	Data map[string]*AppCfgContext
	sync.Mutex
}

var AppCfgState map[string]*AppCfgContextData

func InitAppCfgState() {
	AppCfgState = make(map[string]*AppCfgContextData)
}

func CreateAppCfgState(c, p, ca, dig string) (*AppCfgContextData, bool) {
	acKey := fmt.Sprintf("%s-%s-%s", p, ca, dig)
	_, ok := AppCfgState[acKey]
	if !ok {
		AppCfgState[acKey] = &AppCfgContextData{Data: map[string]*AppCfgContext{}}
		AppCfgState[acKey].Data[c] = &AppCfgContext{}
		AppCfgState[acKey].Lock()
		AppCfgState[acKey].Data[c].Lock = &sync.Mutex{}
		AppCfgState[acKey].Data[c].Client = nil
		AppCfgState[acKey].Unlock()
		return AppCfgState[acKey], true
	} else {
		_, lok := AppCfgState[acKey].Data[c]
		if !lok {
			AppCfgState[acKey].Data[c] = &AppCfgContext{}
			AppCfgState[acKey].Lock()
			AppCfgState[acKey].Data[c].Lock = &sync.Mutex{}
			AppCfgState[acKey].Data[c].Client = nil
			AppCfgState[acKey].Unlock()
			return AppCfgState[acKey], true
		}
	}
	return AppCfgState[acKey], false

}

func runAppCfgStateClient(acid, c, p, ca, dig string) error {
	acKey := fmt.Sprintf("%s-%s-%s", p, ca, dig)
	acContext, ok := AppCfgState[acKey].Data[c]
	if !ok {
		return pkgerrors.New("AppCfg state is empty...")
	}
	if acContext.Client != nil {
		log.Info("Calling global client unsubscribe context from terminate appconfig", log.Fields{})
		acContext.Lock.Lock()
		cl := acContext.Client
		acContext.Client = nil
		acContext.Lock.Unlock()
		cl.Unsubscribe(context.Background(), &readynotifypb.Topic{ClientName: c, AppContext: acid})
	}
	return nil
}

func InstantiateAppConfig(config, p, ca, cv, dig string) (ContextForAppConfig, interface{}, error) {

	aci := acInstantiator{
		project:          p,
		compositeApp:     ca,
		compAppVersion:   cv,
		deploymentIntent: dig,
	}
	cca, err := aci.MakeAppConfigContext(config)
	return cca, cca.ctxval, err

}

func WaitUpdateAppConfigState(ctxid string, statusVal state.StateValue, ac, p, ca, cv, dig string) error {

	acClient := NewAppConfigClient()
	s, err := acClient.GetAppConfigState(ac, p, ca, cv, dig)
	if err != nil {
		return pkgerrors.Wrap(err, "WaitUpdateAppConfigState Cannot Get app config state")
	}
	ai := state.ActionEntry{
		State:     statusVal,
		ContextId: ctxid,
		TimeStamp: time.Now(),
		Revision:  1,
	}
	s.StatusContextId = ctxid
	s.Actions = append(s.Actions, ai)
	key := appConfigKey{
		AppConfig:           ac,
		Project:             p,
		CompositeApp:        ca,
		CompositeAppVersion: cv,
		DigName:             dig,
	}
	err = db.DBconn.Insert(acClient.db.storeName, key, nil, acClient.db.tagState, s)
	if err != nil {
		return pkgerrors.Wrap(err, "Creating DB Entry")
	}
	return nil

}

// CreateAppConfig creates a new AppConfig
func (cc *AppConfigClient) CreateAppConfig(c AppConfig, t SpecFileContent, p, ca, cv, dig string, exists bool) (AppConfig, error) {

	key := appConfigKey{
		AppConfig:           c.Metadata.Name,
		Project:             p,
		CompositeApp:        ca,
		CompositeAppVersion: cv,
		DigName:             dig,
	}

	log.Info("Creating appconfig...", log.Fields{})
	_, err := cc.GetAppConfig(c.Metadata.Name, p, ca, cv, dig)
	if err == nil && !exists {
		return AppConfig{}, pkgerrors.New("appConfig already exists")
	}
	err = db.DBconn.Insert(cc.db.storeName, key, nil, cc.db.tagMeta, c)
	if err != nil {
		return AppConfig{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	err = db.DBconn.Insert(cc.db.storeName, key, nil, cc.db.tagContent, t)
	if err != nil {
		return AppConfig{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	CreateAppCfgState(c.Metadata.Name, p, ca, dig)

	// Add the stateInfo record for the first time
	s := state.StateInfo{}
	a := state.ActionEntry{
		State:     state.StateEnum.Created,
		ContextId: "",
		TimeStamp: time.Now(),
	}
	s.Actions = append(s.Actions, a)

	err = db.DBconn.Insert(cc.db.storeName, key, nil, cc.db.tagState, s)
	if err != nil {
		return AppConfig{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	caStat, err := orchmodule.NewDeploymentIntentGroupClient().GetDeploymentIntentGroupState(dig, p, ca, cv)
	if err != nil {
		log.Info("DeploymentIntentGroup has no state info ", log.Fields{"DeploymentIntentGroup: ": dig})
		return AppConfig{}, pkgerrors.Wrap(err, "DeploymentIntentGroup has no state info")
	}

	caStateVal, err := state.GetCurrentStateFromStateInfo(caStat)
	if err != nil {
		log.Info("Error getting current state from state info ", log.Fields{"DeploymentIntentGroup: ": dig})
	}

	if caStateVal == state.StateEnum.Instantiated {
		log.Info("compapp instantiated :"+ca+" hence applying the app config", log.Fields{})
		cca, _, err := InstantiateAppConfig(c.Metadata.Name, p, ca, cv, dig)
		if err != nil {
			return AppConfig{}, pkgerrors.Wrap(err, "Error creating appConfig context")
		}
		WaitUpdateAppConfigState(cca.ctxval.(string), state.StateEnum.Instantiated, c.Metadata.Name, p, ca, cv, dig)
	} else {
		log.Info("composite APP for app config not instantiated... Hence will return later", log.Fields{})
	}

	return c, nil

}

// GetAppConfig returns AppConfig
func (cc *AppConfigClient) GetAppConfig(c, p, ca, cv, dig string) (AppConfig, error) {

	key := appConfigKey{
		AppConfig:           c,
		Project:             p,
		CompositeApp:        ca,
		CompositeAppVersion: cv,
		DigName:             dig,
	}

	value, err := db.DBconn.Find(cc.db.storeName, key, cc.db.tagMeta)
	if err != nil {
		return AppConfig{}, err
	}

	if len(value) == 0 {
		return AppConfig{}, pkgerrors.New("appConfig not found")
	}

	//value is a byte array
	if value != nil {
		c := AppConfig{}
		err = db.DBconn.Unmarshal(value[0], &c)
		if err != nil {
			return AppConfig{}, err
		}
		return c, nil
	}

	return AppConfig{}, pkgerrors.New("Unknown Error")

}

// GetAllappConfig returns all the AppConfig objects
func (cc *AppConfigClient) GetAllAppConfig(p, ca, cv, dig string) ([]AppConfig, error) {

	key := appConfigKey{
		AppConfig:           "",
		Project:             p,
		CompositeApp:        ca,
		CompositeAppVersion: cv,
		DigName:             dig,
	}

	var czs []AppConfig
	values, err := db.DBconn.Find(cc.db.storeName, key, cc.db.tagMeta)
	if err != nil {
		return []AppConfig{}, err
	}

	for _, value := range values {
		cz := AppConfig{}
		err = db.DBconn.Unmarshal(value, &cz)
		if err != nil {
			return []AppConfig{}, err
		}
		czs = append(czs, cz)
	}

	return czs, nil
}

// GetAppConfigContent returns the AppConfigContent
func (cc *AppConfigClient) GetAppConfigContent(c, p, ca, cv, dig string) (SpecFileContent, error) {
	key := appConfigKey{
		AppConfig:           c,
		Project:             p,
		CompositeApp:        ca,
		CompositeAppVersion: cv,
		DigName:             dig,
	}

	value, err := db.DBconn.Find(cc.db.storeName, key, cc.db.tagContent)
	if err != nil {
		return SpecFileContent{}, err
	}

	if len(value) == 0 {
		return SpecFileContent{}, pkgerrors.New("appConfig Spec File Content not found")
	}

	if value != nil {
		sFileContent := SpecFileContent{}

		err = db.DBconn.Unmarshal(value[0], &sFileContent)
		if err != nil {
			return SpecFileContent{}, err
		}
		return sFileContent, nil
	}

	return SpecFileContent{}, pkgerrors.New("Unknown Error")
}

// GetAppConfigState returns AppConfig StateInfo
func (cc *AppConfigClient) GetAppConfigState(c, p, ca, cv, dig string) (state.StateInfo, error) {

	key := appConfigKey{
		AppConfig:           c,
		Project:             p,
		CompositeApp:        ca,
		CompositeAppVersion: cv,
		DigName:             dig,
	}

	result, err := db.DBconn.Find(cc.db.storeName, key, cc.db.tagState)
	if err != nil {
		return state.StateInfo{}, err
	} else if len(result) == 0 {
		return state.StateInfo{}, pkgerrors.New("AppConfig StateInfo not found")
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

// TerminateappConfig terminates appConfig
func TerminateAppConfig(acid, c, p, ca, cv, dig string) error {
	err := runAppCfgStateClient(acid, c, p, ca, dig)
	if err != nil {
		log.Warn("Failed to run the unsubscribe client for appcfg", log.Fields{})
	}
	aConfig, err := NewAppConfigClient().GetAppConfig(c, p, ca, cv, dig)
	if err != nil {
		return err
	}
	s, err := NewAppConfigClient().GetAppConfigState(c, p, ca, cv, dig)
	if err != nil {
		return pkgerrors.Wrap(err, "Error getting app config state while Terminate")
	}

	stateVal, err := state.GetCurrentStateFromStateInfo(s)
	if err != nil {
		return pkgerrors.Wrap(err, "Error getting current state from AppConfig stateInfo: ")
	}

	if stateVal == state.StateEnum.Instantiated {
		appCfgCtxId := state.GetLastContextIdFromStateInfo(s)
		log.Info("Terminating app config with id :"+appCfgCtxId, log.Fields{})
		err = SyncTerminateAppConfig(appCfgCtxId)
		if err != nil {
			log.Error("Failed to terminate the appconfig", log.Fields{})
		}

		WaitUpdateAppConfigState(appCfgCtxId, state.StateEnum.Terminated, aConfig.Metadata.Name, p, ca, cv, dig)
	}
	return nil
}

// DeleteappConfig deletes appConfig
func (cc *AppConfigClient) DeleteAppConfig(c, p, ca, cv, dig string) error {
	caStat, err := orchmodule.NewDeploymentIntentGroupClient().GetDeploymentIntentGroupState(dig, p, ca, cv)
	if err != nil {
		log.Info("DeploymentIntentGroup has no state info ", log.Fields{"DeploymentIntentGroup: ": dig})
		return pkgerrors.Wrap(err, "DeploymentIntentGroup has no state info")
	}

	appCID := state.GetLastContextIdFromStateInfo(caStat)
	if appCID == "" {
		return pkgerrors.New("Failed to get the app context ID for composite App")
	}
	_, err = cc.GetAppConfig(c, p, ca, cv, dig)
	if err != nil {
		return err
	}
	s, err := cc.GetAppConfigState(c, p, ca, cv, dig)
	if err != nil {
		return err
	}

	stateVal, err := state.GetCurrentStateFromStateInfo(s)
	if err != nil {
		return pkgerrors.Wrap(err, "Error getting current state from AppConfig stateInfo: ")
	}

	err = runAppCfgStateClient(appCID, c, p, ca, dig)
	if err != nil {
		log.Warn("Failed to run the unsubscribe client for appcfg", log.Fields{})
	}
	appCfgCtxId := state.GetLastContextIdFromStateInfo(s)

	if stateVal == state.StateEnum.Instantiated {
		log.Info("state is now in instantiated", log.Fields{})
		err = TerminateAppConfig(appCID, c, p, ca, cv, dig)
		if err != nil {
			log.Error("Failed to terminate... hence exiting...", log.Fields{})
			return pkgerrors.Wrap(err, "Failed to terminate appconfig")
		}

		for {
			acStatus, err := state.GetAppContextStatus(appCfgCtxId)
			if err != nil {
				log.Warn("Failed to get the app context status", log.Fields{})
				continue
			}
			if acStatus.Status != appcontext.AppContextStatusEnum.Terminated {
				time.Sleep(1 * time.Second)
			} else {
				log.Info("app config context terminated", log.Fields{})
				break
			}
		}
	} else {
		log.Warn("stae is not in instantiated state before termination", log.Fields{})
	}
	if stateVal == state.StateEnum.Terminated {
		context, err := state.GetAppContextFromId(appCfgCtxId)
		if err != nil {
			return pkgerrors.Wrap(err, "Error getting appcontext from DeploymentIntentGroup StateInfo")
		}
		err = context.DeleteCompositeApp()
		if err != nil {
			return pkgerrors.Wrap(err, "Error deleting appcontext for DeploymentIntentGroup")
		}

		key := appConfigKey{
			AppConfig:           c,
			Project:             p,
			CompositeApp:        ca,
			CompositeAppVersion: cv,
			DigName:             dig,
		}
		err = db.DBconn.Remove(cc.db.storeName, key)
		return err
	} else {
		log.Warn("app config state is not terminated properly", log.Fields{})
		return pkgerrors.New("app config is not terminated properly")
	}
	return nil
}
