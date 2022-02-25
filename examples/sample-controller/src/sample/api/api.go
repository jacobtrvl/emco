// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

// Package api defines all the routes and their associated handler functions.
// This example implements two HTTP methods.
// It registers two routes to create and retrieve the intents associated with a deployment group.
package api

import (
	"fmt"
	"reflect"

	"github.com/gorilla/mux"
	"gitlab.com/project-emco/core/emco-base/examples/sample-controller/pkg/module"
)

// NewRouter creates a router that registers the various routes.
// If the mockClient parameter is not nil, the router is configured with a mock handler.
func NewRouter(mockClient interface{}) *mux.Router {
	const baseURL string = "/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/sampleIntents"

	r := mux.NewRouter().PathPrefix("/v2").Subrouter()
	c := module.NewClient()
	h := intentHandler{
		client: setClient(c.SampleIntent, mockClient).(module.SampleIntentManager),
	}

	// You can have multiple handlers based on the requirement and its implementation.
	// ref: https://gitlab.com/project-emco/core/emco-base/-/blob/main/src/hpa-plc/api/api.go

	r.HandleFunc(baseURL, h.handleSampleIntentCreate).Methods("POST")
	r.HandleFunc(baseURL+"/{sampleIntent}", h.handleSampleIntentGet).Methods("GET")

	return r
}

// setClient set the client and its corresponding manager interface.
// If the mockClient parameter is not nil and implements the manager interface
// corresponding to the client return the mockClient. Otherwise, return the client.
func setClient(client, mockClient interface{}) interface{} {
	switch cl := client.(type) {
	case *module.SampleIntentClient:
		if mockClient != nil && reflect.TypeOf(mockClient).Implements(reflect.TypeOf((*module.SampleIntentManager)(nil)).Elem()) {
			c, ok := mockClient.(module.SampleIntentManager)
			if ok {
				return c
			}
		}
	default:
		fmt.Printf("unknown type %T\n", cl)
	}
	return client
}
