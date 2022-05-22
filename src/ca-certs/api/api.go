// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package api

import (
	"fmt"
	"reflect"

	"github.com/gorilla/mux"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/client"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/client/clusterprovider"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/client/logicalcloud"
)

// NewRouter returns the mux router after plugging in all the handlers
func NewRouter(mockClient interface{}) *mux.Router {
	r := route{
		router: mux.NewRouter().PathPrefix("/v2").Subrouter(),
		client: client.NewClient(),
		mock:   mockClient}

	// set routes for adding CA cert intent and cluster groups for cluster-provoder
	r.setClusterProviderRoutes()
	// set routes for adding CA cert intent, logical-cloud and cluster groups for logical-cloud
	r.setLogicalCloudRoutes()

	return r.router
}

// setClient set the client and its corresponding manager interface
// If the mockClient parameter is not nil and implements the manager interface
// corresponding to the client, return the mockClient. Otherwise, return the client
func setClient(client, mockClient interface{}) interface{} {
	if mockClient == nil {
		return client
	}

	switch cl := client.(type) {
	case *clusterprovider.CertClient:
		if mockClient != nil && reflect.TypeOf(mockClient).Implements(reflect.TypeOf((*clusterprovider.CertManager)(nil)).Elem()) {
			c, ok := mockClient.(clusterprovider.CertManager)
			if ok {
				return c
			}
		}

	case *clusterprovider.ClusterClient:
		if mockClient != nil && reflect.TypeOf(mockClient).Implements(reflect.TypeOf((*clusterprovider.ClusterManager)(nil)).Elem()) {
			c, ok := mockClient.(clusterprovider.ClusterManager)
			if ok {
				return c
			}
		}

	case *logicalcloud.CertClient:
		if mockClient != nil && reflect.TypeOf(mockClient).Implements(reflect.TypeOf((*logicalcloud.CertManager)(nil)).Elem()) {
			c, ok := mockClient.(logicalcloud.CertManager)
			if ok {
				return c
			}
		}

	case *logicalcloud.LogicalCloudClient:
		if mockClient != nil && reflect.TypeOf(mockClient).Implements(reflect.TypeOf((*logicalcloud.LogicalCloudManager)(nil)).Elem()) {
			c, ok := mockClient.(logicalcloud.LogicalCloudManager)
			if ok {
				return c
			}
		}
	case *logicalcloud.ClusterClient:
		if mockClient != nil && reflect.TypeOf(mockClient).Implements(reflect.TypeOf((*logicalcloud.ClusterManager)(nil)).Elem()) {
			c, ok := mockClient.(logicalcloud.ClusterManager)
			if ok {
				return c
			}
		}
	default:
		fmt.Printf("unknown type %T\n", cl)
	}

	return client
}
