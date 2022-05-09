// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package common

import (
	"time"

	"github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"
)

type LifeCycleEvent string

const (
	InstantiateEvent LifeCycleEvent = "Instantiate"
	TerminateEvent   LifeCycleEvent = "Terminate"
)

// type ResourceName string

// const (
// 	CertEnrollment   ResourceName = "cert-enrollment"
// 	CertDistribution ResourceName = "cert-distribution"
// )

// StateClient
type StateClient struct {
	db DbInfo
}

// NewStateClient
func NewStateClient() *StateClient {
	return &StateClient{
		db: DbInfo{
			storeName: "resources",
			tagState:  "stateInfo"}}
}

// CreateState
func (c *StateClient) CreateState(key interface{}, contextID string) error {
	// create the stateInfo
	a := state.ActionEntry{
		State:     state.StateEnum.Created,
		ContextId: contextID,
		TimeStamp: time.Now(),
	}

	s := state.StateInfo{}
	s.Actions = append(s.Actions, a)

	return db.DBconn.Insert(c.db.storeName, key, nil, c.db.tagState, s)
}

// GetState
func (c *StateClient) GetState(key interface{}) (state.StateInfo, error) {
	values, err := db.DBconn.Find(c.db.storeName, key, c.db.tagState)
	if err != nil {
		return state.StateInfo{}, err
	}
	// TODO: VErify why there are multiple state info
	// Why the keyId is not used
	if len(values) == 0 ||
		(len(values) > 0 &&
			values[0] == nil) {
		return state.StateInfo{}, errors.New("StateInfo not found")
	}

	if len(values) > 0 &&
		values[0] != nil {
		s := state.StateInfo{}
		if err = db.DBconn.Unmarshal(values[0], &s); err != nil {
			return state.StateInfo{}, err
		}
		return s, nil
	}

	return state.StateInfo{}, errors.New("Unknown Error")
}

// UpdateState
func (c *StateClient) UpdateState(key interface{}, newState state.StateValue,
	contextID string, createIfNotExists bool) error {
	s, err := c.GetState(key)
	if err == nil { // state exists
		revision, err := state.GetLatestRevisionFromStateInfo(s)
		if err != nil {
			return err
		}

		a := state.ActionEntry{
			State:     newState,
			ContextId: contextID,
			TimeStamp: time.Now(),
			Revision:  revision + 1,
		}

		s.StatusContextId = contextID
		s.Actions = append(s.Actions, a)

		if err = db.DBconn.Insert(c.db.storeName, key, nil, c.db.tagState, s); err != nil {
			return err
		}

		return nil
	}

	if err.Error() == "StateInfo not found" &&
		createIfNotExists {
		return c.CreateState(key, contextID)

	}

	return err
}

// VerifyState
// func (c *StateClient) VerifyState(key interface{}, event LifeCycleEvent, resource ResourceName) (string, error) {
func (c *StateClient) VerifyState(key interface{}, event LifeCycleEvent) (string, error) {
	var log = func(message, contextID string, status appcontext.StatusValue, err error) {
		fields := make(logutils.Fields)
		fields["AppContextID"] = contextID
		if err != nil {
			fields["Error"] = err.Error()
		}
		if len(status) > 0 {
			fields["Status"] = status
		}
		logutils.Error(message, fields)
	}
	var contextID string
	// check for previous instantiation state TODO: revisit the logic here
	s, err := c.GetState(key)
	if err != nil {
		return contextID, err
	}
	contextID = state.GetLastContextIdFromStateInfo(s)
	if contextID != "" {
		status, err := state.GetAppContextStatus(contextID)
		if err != nil {
			return contextID, err
		}

		// Make sure rsync status for this certificate enrollment is Terminated
		switch status.Status {
		case appcontext.AppContextStatusEnum.Terminating:
			// log(fmt.Sprintf("The %s is being terminated", resource), contextID, "", nil)
			// return contextID, errors.Errorf("failed to %s. The ertificate enrollement is being terminated", event)
			log("The resource is being terminated", contextID, "", nil)
			return contextID, errors.Errorf("failed to %s. The resource is being terminated", event)

		case appcontext.AppContextStatusEnum.Instantiating:
			// log(fmt.Sprintf("The %s is in instantiating status", resource), contextID, "", nil)
			// return contextID, errors.Errorf("failed to %s. The %s is in instantiating status", event)
			log("The resource is in instantiating status", contextID, "", nil)
			return contextID, errors.Errorf("failed to %s. The resource is in instantiating status", event)
		case appcontext.AppContextStatusEnum.TerminateFailed:
			// log(fmt.Sprintf("The %s has failed terminating, please delete the %s", resource, resource), contextID, "", nil)
			// return contextID, errors.Errorf("failed to %s. The %s has failed terminating, please delete the %s", event)
			log("The resource has failed terminating, please delete the resource", contextID, "", nil)
			return contextID, errors.Errorf("failed to %s. The resource has failed terminating, please delete the resource", event)
		case appcontext.AppContextStatusEnum.Terminated:
			// Handle events specific use cases
			switch event {
			case InstantiateEvent:
				// Fully delete the old AppContext and continue with the Instantiation
				appContext, err := state.GetAppContextFromId(contextID)
				if err != nil {
					return contextID, err
				}
				if err := appContext.DeleteCompositeApp(); err != nil { // TODO: Confirm this is reuired or not
					// log(fmt.Sprintf("Failed to delete the app context for the %s", resource), contextID, "", err)
					log("Failed to delete the app context for the resource", contextID, "", err)
					return contextID, err
				}
				return contextID, nil
			case TerminateEvent:
				return contextID, errors.New("the certificate enrollment has already been terminated")
			}

		case appcontext.AppContextStatusEnum.Instantiated:
			switch event {
			case InstantiateEvent:
				// log(fmt.Sprintf("The %s is already instantiated", resource), contextID, "", nil)
				// return contextID, errors.New("The %s is already instantiated")
				log("The resource is already instantiated", contextID, "", nil)
				return contextID, errors.New("the resource is already instantiated")
			case TerminateEvent:
				return contextID, nil
			}
		case appcontext.AppContextStatusEnum.InstantiateFailed:
			switch event {
			case InstantiateEvent:
				// log(fmt.Sprintf("The %s has failed instantiating before, please terminate and try again", resource), contextID, "", nil)
				// return contextID, errors.New("The %s has failed instantiating before, please terminate and try again")
				log("The resource has failed instantiating before, please terminate and try again", contextID, "", nil)
				return contextID, errors.New("the resource has failed instantiating before, please terminate and try again")
			case TerminateEvent:
				// Terminate anyway
				return contextID, nil
			}

		default:
			// log(fmt.Sprintf("The %s isn't in an expected status so not taking any action", resource), contextID, status.Status, nil)
			// return contextID, errors.New("The %s isn't in an expected status so not taking any action")
			log("The resource isn't in an expected status so not taking any action", contextID, status.Status, nil)
			return contextID, errors.New("the resource isn't in an expected status so not taking any action")
		}
	}

	return contextID, nil
}
