// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation
package api

import (
	"fmt"
	"reflect"

	"github.com/gorilla/mux"
	"gitlab.com/project-emco/core/emco-base/src/dtc/pkg/module"
	controller "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/controller"
)

var moduleClient *module.Client
var moduleController *controller.ControllerClient

// For the given client and testClient, if the testClient is not null and
// implements the client manager interface corresponding to client, then
// return the testClient, otherwise return the client.
func setClient(client, testClient interface{}) interface{} {
	switch cl := client.(type) {
	case *module.TrafficGroupIntentDbClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*module.TrafficGroupIntentManager)(nil)).Elem()) {
			c, ok := testClient.(module.TrafficGroupIntentManager)
			if ok {
				return c
			}
		}
	case *module.InboundServerIntentDbClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*module.InboundServerIntentManager)(nil)).Elem()) {
			c, ok := testClient.(module.InboundServerIntentManager)
			if ok {
				return c
			}
		}
	case *module.InboundClientsIntentDbClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*module.InboundClientsIntentManager)(nil)).Elem()) {
			c, ok := testClient.(module.InboundClientsIntentManager)
			if ok {
				return c
			}
		}
	case *module.InboundClientsAccessIntentDbClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*module.InboundClientsAccessIntentManager)(nil)).Elem()) {
			c, ok := testClient.(module.InboundClientsAccessIntentManager)
			if ok {
				return c
			}
		}
	case *controller.ControllerClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*controller.ControllerManager)(nil)).Elem()) {
			c, ok := testClient.(controller.ControllerManager)
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

	moduleClient = module.NewClient()

	router := mux.NewRouter().PathPrefix("/v2").Subrouter()
	trafficgroupintentHandler := trafficgroupintentHandler{
		client: setClient(moduleClient.TrafficGroupIntent, testClient).(module.TrafficGroupIntentManager),
	}
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents", trafficgroupintentHandler.createHandler).Methods("POST")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents", trafficgroupintentHandler.getHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents/{trafficGroupIntent}", trafficgroupintentHandler.getHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents/{trafficGroupIntent}", trafficgroupintentHandler.putHandler).Methods("PUT")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents/{trafficGroupIntent}", trafficgroupintentHandler.deleteHandler).Methods("DELETE")

	inboundserverintentHandler := inboundserverintentHandler{
		client: setClient(moduleClient.ServerInboundIntent, testClient).(module.InboundServerIntentManager),
	}
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents/{trafficGroupIntent}/inbound-intents", inboundserverintentHandler.createHandler).Methods("POST")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents/{trafficGroupIntent}/inbound-intents/{inboundServerIntent}", inboundserverintentHandler.getHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents/{trafficGroupIntent}/inbound-intents", inboundserverintentHandler.getHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents/{trafficGroupIntent}/inbound-intents/{inboundServerIntent}", inboundserverintentHandler.putHandler).Methods("PUT")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents/{trafficGroupIntent}/inbound-intents/{inboundServerIntent}", inboundserverintentHandler.deleteHandler).Methods("DELETE")

	inboundclientsintentHandler := inboundclientsintentHandler{
		client: setClient(moduleClient.ClientsInboundIntent, testClient).(module.InboundClientsIntentManager),
	}
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents/{trafficGroupIntent}/inbound-intents/{inboundServerIntent}/clients", inboundclientsintentHandler.createHandler).Methods("POST")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents/{trafficGroupIntent}/inbound-intents/{inboundServerIntent}/clients", inboundclientsintentHandler.getHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents/{trafficGroupIntent}/inbound-intents/{inboundServerIntent}/clients/{inboundClientsIntent}", inboundclientsintentHandler.getHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents/{trafficGroupIntent}/inbound-intents/{inboundServerIntent}/clients/{inboundClientsIntent}", inboundclientsintentHandler.putHandler).Methods("PUT")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents/{trafficGroupIntent}/inbound-intents/{inboundServerIntent}/clients/{inboundClientsIntent}", inboundclientsintentHandler.deleteHandler).Methods("DELETE")

	inboundclientsaccessintentHandler := inboundclientsaccessintentHandler{
		client: setClient(moduleClient.ClientsAccessInboundIntent, testClient).(module.InboundClientsAccessIntentManager),
	}
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents/{trafficGroupIntent}/inbound-intents/{inboundServerIntent}/clients/{inboundClientsIntent}/access-points", inboundclientsaccessintentHandler.createHandler).Methods("POST")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents/{trafficGroupIntent}/inbound-intents/{inboundServerIntent}/clients/{inboundClientsIntent}/access-points", inboundclientsaccessintentHandler.getHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents/{trafficGroupIntent}/inbound-intents/{inboundServerIntent}/clients/{inboundClientsIntent}/access-points/{inboundClientsAccessIntent}", inboundclientsaccessintentHandler.getHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents/{trafficGroupIntent}/inbound-intents/{inboundServerIntent}/clients/{inboundClientsIntent}/access-points/{inboundClientsAccessIntent}", inboundclientsaccessintentHandler.putHandler).Methods("PUT")
	router.HandleFunc("/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/traffic-group-intents/{trafficGroupIntent}/inbound-intents/{inboundServerIntent}/clients/{inboundClientsIntent}/access-points/{inboundClientsAccessIntent}", inboundclientsaccessintentHandler.deleteHandler).Methods("DELETE")

	controlHandler := controllerHandler{
		client: setClient(moduleClient.Controller, testClient).(controller.ControllerManager),
	}
	router.HandleFunc("/dtc-controllers", controlHandler.createHandler).Methods("POST")
	router.HandleFunc("/dtc-controllers", controlHandler.getHandler).Methods("GET")
	router.HandleFunc("/dtc-controllers/{dtcController}", controlHandler.putHandler).Methods("PUT")
	router.HandleFunc("/dtc-controllers/{dtcController}", controlHandler.getHandler).Methods("GET")
	router.HandleFunc("/dtc-controllers/{dtcController}", controlHandler.deleteHandler).Methods("DELETE")

	return router
}
