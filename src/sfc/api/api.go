// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation
package api

import (
	"fmt"
	"reflect"

	"github.com/gorilla/mux"
	"gitlab.com/project-emco/core/emco-base/src/sfc/pkg/module"
)

var moduleSfcIntentClient *module.SfcIntentClient
var moduleSfcClientSelectorIntentClient *module.SfcClientSelectorIntentClient
var moduleSfcProviderNetworkIntentClient *module.SfcProviderNetworkIntentClient

// Used to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type sfcIntentHandler struct {
	// Interface that implements SFC intent operations
	// We will set this variable with a mock interface for testing
	client module.SfcIntentManager
}
type sfcLinkIntentHandler struct {
	// Interface that implements SFC intent operations
	// We will set this variable with a mock interface for testing
	client module.SfcLinkIntentManager
}
type sfcClientSelectorIntentHandler struct {
	// Interface that implements SFC intent operations
	// We will set this variable with a mock interface for testing
	client module.SfcClientSelectorIntentManager
}
type sfcProviderNetworkIntentHandler struct {
	// Interface that implements SFC intent operations
	// We will set this variable with a mock interface for testing
	client module.SfcProviderNetworkIntentManager
}

// For the given client and testClient, if the testClient is not null and
// implements the client manager interface corresponding to client, then
// return the testClient, otherwise return the client.
func setClient(client, testClient interface{}) interface{} {
	switch cl := client.(type) {
	case *module.SfcIntentClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*module.SfcIntentManager)(nil)).Elem()) {
			c, ok := testClient.(module.SfcIntentManager)
			if ok {
				return c
			}
		}
	case *module.SfcClientSelectorIntentClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*module.SfcClientSelectorIntentManager)(nil)).Elem()) {
			c, ok := testClient.(module.SfcClientSelectorIntentManager)
			if ok {
				return c
			}
		}
	case *module.SfcProviderNetworkIntentClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*module.SfcProviderNetworkIntentManager)(nil)).Elem()) {
			c, ok := testClient.(module.SfcProviderNetworkIntentManager)
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

	moduleClient := module.NewClient()

	router := mux.NewRouter().PathPrefix("/v2").Subrouter()

	const sfcIntentsURL = "/projects/{project}/composite-apps/{compositeApp}/{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/network-chains"
	const sfcIntentsGetURL = sfcIntentsURL + "/{sfcIntent}"
	const sfcLinkIntentsURL = sfcIntentsGetURL + "/links"
	const sfcLinkIntentsGetURL = sfcLinkIntentsURL + "/{sfcLink}"
	const sfcClientSelectorIntentsURL = sfcIntentsGetURL + "/client-selectors"
	const sfcClientSelectorIntentsGetURL = sfcClientSelectorIntentsURL + "/{sfcClientSelector}"
	const sfcProviderNetworkIntentsURL = sfcIntentsGetURL + "/provider-networks"
	const sfcProviderNetworkIntentsGetURL = sfcProviderNetworkIntentsURL + "/{sfcProviderNetwork}"

	sfcHandler := sfcIntentHandler{
		client: setClient(moduleClient.SfcIntent, testClient).(module.SfcIntentManager),
	}
	router.HandleFunc(sfcIntentsURL, sfcHandler.createSfcHandler).Methods("POST")
	router.HandleFunc(sfcIntentsURL, sfcHandler.getSfcHandler).Methods("GET")
	router.HandleFunc(sfcIntentsGetURL, sfcHandler.putSfcHandler).Methods("PUT")
	router.HandleFunc(sfcIntentsGetURL, sfcHandler.getSfcHandler).Methods("GET")
	router.HandleFunc(sfcIntentsGetURL, sfcHandler.deleteSfcHandler).Methods("DELETE")

	sfcLinkHandler := sfcLinkIntentHandler{
		client: setClient(moduleClient.SfcLinkIntent, testClient).(module.SfcLinkIntentManager),
	}
	router.HandleFunc(sfcLinkIntentsURL, sfcLinkHandler.createLinkHandler).Methods("POST")
	router.HandleFunc(sfcLinkIntentsURL, sfcLinkHandler.getLinkHandler).Methods("GET")
	router.HandleFunc(sfcLinkIntentsGetURL, sfcLinkHandler.putLinkHandler).Methods("PUT")
	router.HandleFunc(sfcLinkIntentsGetURL, sfcLinkHandler.getLinkHandler).Methods("GET")
	router.HandleFunc(sfcLinkIntentsGetURL, sfcLinkHandler.deleteLinkHandler).Methods("DELETE")

	sfcClientSelectorHandler := sfcClientSelectorIntentHandler{
		client: setClient(moduleClient.SfcClientSelectorIntent, testClient).(module.SfcClientSelectorIntentManager),
	}
	router.HandleFunc(sfcClientSelectorIntentsURL, sfcClientSelectorHandler.createClientSelectorHandler).Methods("POST")
	router.HandleFunc(sfcClientSelectorIntentsURL, sfcClientSelectorHandler.getClientSelectorHandler).Methods("GET")
	router.HandleFunc(sfcClientSelectorIntentsGetURL, sfcClientSelectorHandler.putClientSelectorHandler).Methods("PUT")
	router.HandleFunc(sfcClientSelectorIntentsGetURL, sfcClientSelectorHandler.getClientSelectorHandler).Methods("GET")
	router.HandleFunc(sfcClientSelectorIntentsGetURL, sfcClientSelectorHandler.deleteClientSelectorHandler).Methods("DELETE")

	sfcProviderNetworkHandler := sfcProviderNetworkIntentHandler{
		client: setClient(moduleClient.SfcProviderNetworkIntent, testClient).(module.SfcProviderNetworkIntentManager),
	}
	router.HandleFunc(sfcProviderNetworkIntentsURL, sfcProviderNetworkHandler.createProviderNetworkHandler).Methods("POST")
	router.HandleFunc(sfcProviderNetworkIntentsURL, sfcProviderNetworkHandler.getProviderNetworkHandler).Methods("GET")
	router.HandleFunc(sfcProviderNetworkIntentsGetURL, sfcProviderNetworkHandler.putProviderNetworkHandler).Methods("PUT")
	router.HandleFunc(sfcProviderNetworkIntentsGetURL, sfcProviderNetworkHandler.getProviderNetworkHandler).Methods("GET")
	router.HandleFunc(sfcProviderNetworkIntentsGetURL, sfcProviderNetworkHandler.deleteProviderNetworkHandler).Methods("DELETE")

	return router
}
