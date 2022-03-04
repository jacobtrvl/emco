// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation
package api

import (
	"fmt"
	"reflect"

	"github.com/gorilla/mux"

	moduleLib "gitlab.com/project-emco/core/emco-base/src/workflowmgr/pkg/module"
)

var moduleClient *moduleLib.Client

// For the given client and testClient, if the testClient is not null and
// implements the client manager interface corresponding to client, then
// return the testClient, otherwise return the client.
func setClient(client, testClient interface{}) interface{} {
	switch cl := client.(type) {
	case *moduleLib.WorkflowIntentClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*moduleLib.WorkflowIntentManager)(nil)).Elem()) {
			c, ok := testClient.(moduleLib.WorkflowIntentManager)
			if ok {
				return c
			}
		}
	default:
		fmt.Printf("unknown type %T\n", cl)
	}
	return client
}

// NewRouter creates a router that registers the various urls that are supported
// testClient parameter allows unit testing for a given client
func NewRouter(testClient interface{}) *mux.Router {

	moduleClient = moduleLib.NewClient()

	router := mux.NewRouter().PathPrefix("/v2").Subrouter()

	wfIntentHandler := workflowIntentHandler{
		client: setClient(moduleClient.WorkflowIntentClient, testClient).(moduleLib.WorkflowIntentManager),
	}

	baseUrl := "/projects/{project}/composite-apps/{compositeApp}/" +
		"{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/" +
		"temporal-workflow-intents"
	nameUrl := baseUrl + "/{workflow-intent-name}"
	startUrl := nameUrl + "/start"
	statusUrl := nameUrl + "/status"
	cancelUrl := nameUrl + "/cancel"

	router.HandleFunc(baseUrl, wfIntentHandler.createHandler).Methods("POST")
	router.HandleFunc(baseUrl, wfIntentHandler.getHandler).Methods("GET")
	router.HandleFunc(nameUrl, wfIntentHandler.getHandler).Methods("GET")
	router.HandleFunc(nameUrl, wfIntentHandler.deleteHandler).Methods("DELETE")
	router.HandleFunc(startUrl, wfIntentHandler.startHandler).Methods("POST")
	router.HandleFunc(statusUrl, wfIntentHandler.statusHandler).Methods("GET")
	router.HandleFunc(cancelUrl, wfIntentHandler.cancelHandler).Methods("POST")

	return router
}
