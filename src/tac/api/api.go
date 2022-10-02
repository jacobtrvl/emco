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

	router := mux.NewRouter()
	v2Router := router.PathPrefix("/v2").Subrouter()
	c := module.NewClient()
	h := intentHandler{
		client: setClient(c.WorkflowIntentClient, mockClient).(module.WorkflowIntentManager),
	}

	// Temporal Action Hook Intent APIs Unit Test Cases for front end and back end
	v2Router.HandleFunc(baseURL, h.handleTacIntentCreate).Methods("POST")
	v2Router.HandleFunc(baseURL+"/{tac-intent}", h.handleTacIntentGet).Methods("GET")
	v2Router.HandleFunc(baseURL, h.handleTacIntentGet).Methods("GET")
	v2Router.HandleFunc(baseURL+"/{tac-intent}", h.handleTacIntentDelete).Methods("DELETE")
	v2Router.HandleFunc(baseURL+"/{tac-intent}", h.handleTacIntentPut).Methods("PUT")
	// Cancel or get the status of a temporal action controller intent
	v2Router.HandleFunc(baseURL+"/{tac-intent}/cancel", h.handleTemporalWorkflowHookCancel).Methods("POST")
	v2Router.HandleFunc(baseURL+"/{tac-intent}/status", h.handleTemporalWorkflowHookStatus).Methods("GET")

	return router
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
	default:
		fmt.Printf("unknown type %T\n", cl)
	}
	return client
}
