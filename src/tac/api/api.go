// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

// defines all the routes to place a workflow intent and workflow hook intent.
package api

import (
	"fmt"
	"reflect"

	"github.com/gorilla/mux"
	"gitlab.com/project-emco/core/emco-base/src/tac/pkg/module"
)

// NewRouter creates a router that registers the various routes.
func NewRouter(mockClient interface{}) *mux.Router {
	const baseURL string = "/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/temporal-action-controller"

	r := mux.NewRouter().PathPrefix("/v2").Subrouter()
	c := module.NewClient()
	h := intentHandler{
		client: setClient(c.WorkflowIntentClient, mockClient).(module.WorkflowIntentManager),
	}
	w := workerHandler{
		client: setClient(c.WorkerIntentClient, mockClient).(module.WorkerIntentManager),
	}

	// Temporal Action Hook Intent APIs Unit Test Cases for front end and back end
	r.HandleFunc(baseURL, h.handleTacIntentCreate).Methods("POST")
	r.HandleFunc(baseURL+"/{tac-intent}", h.handleTacIntentGet).Methods("GET")
	r.HandleFunc(baseURL, h.handleTacIntentGet).Methods("GET")
	r.HandleFunc(baseURL+"/{tac-intent}", h.handleTacIntentDelete).Methods("DELETE")
	r.HandleFunc(baseURL+"/{tac-intent}", h.handleTacIntentPut).Methods("PUT")
	// Cancel or get the status of a temporal action controller intent
	r.HandleFunc(baseURL+"/{tac-intent}/cancel", h.handleTemporalWorkflowHookCancel).Methods("POST")
	r.HandleFunc(baseURL+"/{tac-intent}/status", h.handleTemporalWorkflowHookStatus).Methods("GET")
	// Worker APIs - Used to register workers in DIGs so TAC can dynamically deploy and terminate them.
	r.HandleFunc(baseURL+"/{tac-intent}/workers", w.handleWorkerCreate).Methods("POST")
	r.HandleFunc(baseURL+"/{tac-intent}/workers", w.handleWorkerGet).Methods("GET")
	r.HandleFunc(baseURL+"/{tac-intent}/workers/{workers}", w.handleWorkerGet).Methods("GET")
	r.HandleFunc(baseURL+"/{tac-intent}/workers/{workers}", w.handleWorkerUpdate).Methods("PUT")
	r.HandleFunc(baseURL+"/{tac-intent}/workers/{workers}", w.handleWorkerDelete).Methods("DELETE")

	return r
}

// setClient set the client and its corresponding manager interface.
// If the mockClient parameter is not nil and implements the manager interface
// corresponding to the client return the mockClient. Otherwise, return the client.
func setClient(client, mockClient interface{}) interface{} {
	switch cl := client.(type) {
	case *module.WorkflowIntentClient:
		if mockClient != nil && reflect.TypeOf(mockClient).Implements(reflect.TypeOf((*module.WorkflowIntentManager)(nil)).Elem()) {
			c, ok := mockClient.(module.WorkflowIntentManager)
			if ok {
				return c
			}
		}
	case *module.WorkerIntentClient:
		if mockClient != nil && reflect.TypeOf(mockClient).Implements(reflect.TypeOf((*module.WorkerIntentManager)(nil)).Elem()) {
			c, ok := mockClient.(module.WorkerIntentManager)
			if ok {
				return c
			}
		}
	default:
		fmt.Printf("unknown type %T\n", cl)
	}
	return client
}
