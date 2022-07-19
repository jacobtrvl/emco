// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package module

import (
	"time"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/common/emcoerror"
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

// StateManager exposes all the caCert state functionalities
type StateManager interface {
	Create(contextID string) error
	Get() (state.StateInfo, error)
	Update(newState state.StateValue, contextID string, createIfNotExists bool) error
}

// StateClient holds the client properties
type StateClient struct {
	dbInfo db.DbInfo
	dbKey  interface{}
}

// NewStateClient returns an instance of the StateClient which implements the Manager
func NewStateClient(dbKey interface{}) *StateClient {
	return &StateClient{
		dbInfo: db.DbInfo{
			StoreName: "resources",
			TagState:  "stateInfo"},
		dbKey: dbKey}
}

// Create the stateInfo resource in mongo
func (c *StateClient) Create(contextID string) error {
	// create the stateInfo
	a := state.ActionEntry{
		State:     state.StateEnum.Created,
		ContextId: contextID,
		TimeStamp: time.Now(),
	}

	s := state.StateInfo{}
	s.Actions = append(s.Actions, a)

	return db.DBconn.Insert(c.dbInfo.StoreName, c.dbKey, nil, c.dbInfo.TagState, s)
}

// Get the stateInfo from mongo
func (c *StateClient) Get() (state.StateInfo, error) {
	values, err := db.DBconn.Find(c.dbInfo.StoreName, c.dbKey, c.dbInfo.TagState)
	if err != nil {
		return state.StateInfo{}, err
	}

	if len(values) == 0 ||
		(len(values) > 0 &&
			values[0] == nil) {
		return state.StateInfo{}, &emcoerror.Error{
			Message: emcoerror.StateInfoNotFound,
			Type:    emcoerror.NotFound,
		}
	}

	if len(values) > 0 &&
		values[0] != nil {
		s := state.StateInfo{}
		if err = db.DBconn.Unmarshal(values[0], &s); err != nil {
			return state.StateInfo{}, err
		}
		return s, nil
	}

	return state.StateInfo{}, &emcoerror.Error{
		Message: emcoerror.UnknownErrorMessage,
		Type:    emcoerror.Unknown,
	}
}

// Update the stateInfo
func (c *StateClient) Update(newState state.StateValue,
	contextID string, createIfNotExists bool) error {
	s, err := c.Get()
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

		if err = db.DBconn.Insert(c.dbInfo.StoreName, c.dbKey, nil, c.dbInfo.TagState, s); err != nil {
			return err
		}

		return nil
	}

	switch e := err.(type) {
	case *emcoerror.Error:
		if e.Type == emcoerror.NotFound &&
			createIfNotExists {
			return c.Create(contextID)
		}
	}

	return err
}

// Delete the stateInfo
func (c *StateClient) Delete() error {
	return db.DBconn.Remove(c.dbInfo.StoreName, c.dbKey)
}

// VerifyState verifies the enrollment\distribution state
func (sc *StateClient) VerifyState(event LifeCycleEvent) (string, error) {
	var contextID string
	// check for previous instantiation state
	s, err := sc.Get()
	if err != nil {
		return contextID, err
	}

	contextID = state.GetLastContextIdFromStateInfo(s)
	if contextID != "" {
		status, err := state.GetAppContextStatus(contextID)
		if err != nil {
			return contextID, err
		}

		switch status.Status {
		case appcontext.AppContextStatusEnum.Terminating:
			err := &emcoerror.Error{
				Message: (&emcoerror.StateError{
					Resource: "CaCert",
					Event:    string(event),
					Status:   status.Status,
				}).Error(),
				Type: emcoerror.Conflict,
			}
			logutils.Error("",
				logutils.Fields{
					"Error":     err.Error(),
					"ContextID": contextID})
			return contextID, err
		case appcontext.AppContextStatusEnum.Instantiating:
			err := &emcoerror.Error{
				Message: (&emcoerror.StateError{
					Resource: "CaCert",
					Event:    string(event),
					Status:   status.Status}).Error(),
				Type: emcoerror.Conflict,
			}
			logutils.Error("",
				logutils.Fields{
					"Error":     err.Error(),
					"ContextID": contextID})
			return contextID, err
		case appcontext.AppContextStatusEnum.TerminateFailed:
			err := &emcoerror.Error{
				Message: (&emcoerror.StateError{
					Resource: "CaCert",
					Event:    string(event),
					Status:   status.Status}).Error(),
				Type: emcoerror.Conflict,
			}
			logutils.Error("",
				logutils.Fields{
					"Error":     err.Error(),
					"ContextID": contextID})
			return contextID, err
		case appcontext.AppContextStatusEnum.Terminated:
			// handle events specific use cases
			switch event {
			case InstantiateEvent:
				return contextID, nil
			case TerminateEvent:
				err := &emcoerror.Error{
					Message: (&emcoerror.StateError{
						Resource: "CaCert",
						Event:    string(event),
						Status:   status.Status}).Error(),
					Type: emcoerror.Conflict,
				}
				logutils.Error("",
					logutils.Fields{
						"Error":     err.Error(),
						"ContextID": contextID})
				return contextID, err
			}
		case appcontext.AppContextStatusEnum.Instantiated:
			switch event {
			case InstantiateEvent:
				err := &emcoerror.Error{
					Message: (&emcoerror.StateError{
						Resource: "CaCert",
						Event:    string(event),
						Status:   status.Status}).Error(),
					Type: emcoerror.Conflict,
				}
				logutils.Error("",
					logutils.Fields{
						"Error":     err.Error(),
						"ContextID": contextID})
				return contextID, err
			case TerminateEvent:
				return contextID, nil
			}
		case appcontext.AppContextStatusEnum.InstantiateFailed:
			switch event {
			case InstantiateEvent:
				err := &emcoerror.Error{
					Message: (&emcoerror.StateError{
						Resource: "CaCert",
						Event:    string(event),
						Status:   status.Status}).Error(),
					Type: emcoerror.Conflict,
				}
				logutils.Error("",
					logutils.Fields{
						"Error":     err.Error(),
						"ContextID": contextID})
				return contextID, err
			case TerminateEvent:
				// Terminate anyway
				return contextID, nil
			}
		default:
			err := &emcoerror.Error{
				Message: (&emcoerror.StateError{
					Resource: "CaCert",
					Event:    string(event),
					Status:   status.Status}).Error(),
				Type: emcoerror.Conflict,
			}
			logutils.Error("",
				logutils.Fields{
					"Error":     err.Error(),
					"ContextID": contextID,
					"Status":    status.Status})
			return contextID, err
		}
	}

	return contextID, nil
}
