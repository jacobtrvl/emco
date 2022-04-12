// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation
package api

import (
	"fmt"
	"reflect"

	"github.com/gorilla/mux"
	moduleLib "gitlab.com/project-emco/core/emco-base/src/app-config/pkg/module"
)

var moduleClient *moduleLib.Client

// For the given client and testClient, if the testClient is not null and
// implements the client manager interface corresponding to client, then
// return the testClient, otherwise return the client.
func setClient(client, testClient interface{}) interface{} {
	switch cl := client.(type) {
	case *moduleLib.ResTypeClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*moduleLib.ResTypeManager)(nil)).Elem()) {
			c, ok := testClient.(moduleLib.ResTypeManager)
			if ok {
				return c
			}
		}
	case *moduleLib.AppConfigClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*moduleLib.AppConfigManager)(nil)).Elem()) {
			c, ok := testClient.(moduleLib.AppConfigManager)
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
func NewRouter(testClient interface{}) *mux.Router {
	moduleLib.InitAppCfgState()
	moduleClient = moduleLib.NewClient()

	router := mux.NewRouter().PathPrefix("/v2").Subrouter()
	baseResTypeHandler := restypeHandler{
		client: setClient(moduleClient.BaseResType, testClient).(moduleLib.ResTypeManager),
	}
	fmt.Printf("Entering the app-config mux router")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/resource-type", baseResTypeHandler.createResTypeHandler).Methods("POST")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/resource-type/{restype}", baseResTypeHandler.getResTypeHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/resource-type", baseResTypeHandler.putResTypeHandler).Methods("PUT")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/resource-type/{restype}", baseResTypeHandler.deleteResTypeHandler).Methods("DELETE")
	baseAppConfigHandler := AppConfigHandler{
		client: setClient(moduleClient.BaseAppConfigType, testClient).(moduleLib.AppConfigManager),
	}
	fmt.Printf("Entering the app-config-dig mux router")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/app-config", baseAppConfigHandler.createAppConfigHandler).Methods("POST")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/app-config/{appConfig}", baseAppConfigHandler.getAppConfigHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/app-config/{appConfig}", baseAppConfigHandler.deleteAppConfigHandler).Methods("DELETE")

	return router
}
