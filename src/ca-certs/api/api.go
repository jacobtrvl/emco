// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package api

import (
	"github.com/gorilla/mux"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/clusterprovider"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/logicalcloud"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"

	"fmt"
	"reflect"
)

// NewRouter returns the mux router after plugging in all the handlers
func NewRouter(mockClient interface{}) *mux.Router {
	// routes
	const (
		// cluster-provider
		clusterProviderCertURL             string = "/cluster-providers/{clusterProvider}/ca-certs"
		clusterProviderCertDistributionURL string = clusterProviderCertURL + "/{caCert}/distribution"
		clusterProviderCertEnrollmentURL   string = clusterProviderCertURL + "/{caCert}/enrollment"
		// logical-cloud
		logicalCloudCertURL             string = "/projects/{project}/ca-certs"
		logicalCloudCertDistributionURL string = logicalCloudCertURL + "/{caCert}/distribution"
		logicalCloudCertEnrollmentURL   string = logicalCloudCertURL + "/{caCert}/enrollment"
	)

	client := module.NewClient()
	router := mux.NewRouter().PathPrefix("/v2").Subrouter()

	// route to create the cluster-provider CA cert intent
	cpCertHandler := clusterProviderCertHandler{
		manager: setClient(client.ClusterProviderCert, mockClient).(clusterprovider.CertManager)}
	router.HandleFunc(clusterProviderCertURL, cpCertHandler.handleCertificateGet).Methods("GET")
	router.HandleFunc(clusterProviderCertURL, cpCertHandler.handleCertificateCreate).Methods("POST")
	router.HandleFunc(clusterProviderCertURL+"/{caCert}", cpCertHandler.handleCertificateGet).Methods("GET")
	router.HandleFunc(clusterProviderCertURL+"/{caCert}", cpCertHandler.handleCertificateDelete).Methods("DELETE")
	router.HandleFunc(clusterProviderCertURL+"/{caCert}", cpCertHandler.handleCertificateUpdate).Methods("PUT")
	// route to add selected clusters within the cluster-provider
	cpClusterHandler := clusterProviderClusterHandler{
		manager: setClient(client.ClusterProviderCluster, mockClient).(clusterprovider.ClusterManager)}
	router.HandleFunc(clusterProviderCertURL+"/{caCert}/clusters", cpClusterHandler.handleClusterGet).Methods("GET")
	router.HandleFunc(clusterProviderCertURL+"/{caCert}/clusters", cpClusterHandler.handleClusterCreate).Methods("POST")
	router.HandleFunc(clusterProviderCertURL+"/{caCert}/clusters/{cluster}", cpClusterHandler.handleClusterGet).Methods("GET")
	router.HandleFunc(clusterProviderCertURL+"/{caCert}/clusters/{cluster}", cpClusterHandler.handleClusterDelete).Methods("DELETE")
	router.HandleFunc(clusterProviderCertURL+"/{caCert}/clusters/{cluster}", cpClusterHandler.handleClusterUpdate).Methods("PUT")

	// route to enroll the cluster-provider CA cert
	cpCertEnrollmentHandler := clusterProviderCertEnrollmentHandler{
		manager: setClient(client.ClusterProviderCertEnrollment, mockClient).(clusterprovider.CertEnrollmentManager)}
	router.HandleFunc(clusterProviderCertEnrollmentURL+"/status", cpCertEnrollmentHandler.handleStatus).Methods("GET")
	router.HandleFunc(clusterProviderCertEnrollmentURL+"/instantiate", cpCertEnrollmentHandler.handleInstantiate).Methods("POST")
	router.HandleFunc(clusterProviderCertEnrollmentURL+"/terminate", cpCertEnrollmentHandler.handleTerminate).Methods("POST")
	router.HandleFunc(clusterProviderCertEnrollmentURL+"/update", cpCertEnrollmentHandler.handleUpdate).Methods("POST")

	// route to distribute the cluster-provider CA cert
	cpCertDistributionHandler := clusterProviderCertDistributionHandler{
		manager: setClient(client.ClusterProviderCertDistribution, mockClient).(clusterprovider.CertDistributionManager)}
	router.HandleFunc(clusterProviderCertDistributionURL+"/status", cpCertDistributionHandler.handleStatus).Methods("GET")
	router.HandleFunc(clusterProviderCertDistributionURL+"/instantiate", cpCertDistributionHandler.handleInstantiate).Methods("POST")
	router.HandleFunc(clusterProviderCertDistributionURL+"/terminate", cpCertDistributionHandler.handleTerminate).Methods("POST")
	router.HandleFunc(clusterProviderCertDistributionURL+"/update", cpCertDistributionHandler.handleUpdate).Methods("POST")

	// route to create the logical-cloud CA cert intent
	lcCertHandler := logicalCloudCertHandler{
		manager: setClient(client.LogicalCloudCert, mockClient).(logicalcloud.CertManager)}
	router.HandleFunc(logicalCloudCertURL, lcCertHandler.handleCertificateGet).Methods("GET")
	router.HandleFunc(logicalCloudCertURL, lcCertHandler.handleCertificateCreate).Methods("POST")
	router.HandleFunc(logicalCloudCertURL+"/{caCert}", lcCertHandler.handleCertificateGet).Methods("GET")
	router.HandleFunc(logicalCloudCertURL+"/{caCert}", lcCertHandler.handleCertificateDelete).Methods("DELETE")
	router.HandleFunc(logicalCloudCertURL+"/{caCert}", lcCertHandler.handleCertificateUpdate).Methods("PUT")

	// route to add logical-cloud to the CA cert intent
	lcHandler := logicalCloudHandler{
		manager: setClient(client.LogicalCloud, mockClient).(logicalcloud.LogicalCloudManager)}
	router.HandleFunc(logicalCloudCertURL+"/{caCert}/logical-clouds", lcHandler.handleLogicalCloudGet).Methods("GET")
	router.HandleFunc(logicalCloudCertURL+"/{caCert}/logical-clouds", lcHandler.handleLogicalCloudCreate).Methods("POST")
	router.HandleFunc(logicalCloudCertURL+"/{caCert}/logical-clouds/{logicalCloud}", lcHandler.handleLogicalCloudGet).Methods("GET")
	router.HandleFunc(logicalCloudCertURL+"/{caCert}/logical-clouds/{logicalCloud}", lcHandler.handleLogicalCloudDelete).Methods("DELETE")
	router.HandleFunc(logicalCloudCertURL+"/{caCert}/logical-clouds/{logicalCloud}", lcHandler.handleLogicalCloudUpdate).Methods("PUT")

	// route to add selected clusters within the logical-cloud
	lcClusterHandler := logicalCloudClusterHandler{
		manager: setClient(client.LogicalCloudCluster, mockClient).(logicalcloud.ClusterManager)}
	router.HandleFunc(logicalCloudCertURL+"/{caCert}/logical-clouds/{logicalCloud}/clusters", lcClusterHandler.handleClusterGet).Methods("GET")
	router.HandleFunc(logicalCloudCertURL+"/{caCert}/logical-clouds/{logicalCloud}/clusters", lcClusterHandler.handleClusterCreate).Methods("POST")
	router.HandleFunc(logicalCloudCertURL+"/{caCert}/logical-clouds/{logicalCloud}/clusters/{cluster}", lcClusterHandler.handleClusterGet).Methods("GET")
	router.HandleFunc(logicalCloudCertURL+"/{caCert}/logical-clouds/{logicalCloud}/clusters/{cluster}", lcClusterHandler.handleClusterDelete).Methods("DELETE")
	router.HandleFunc(logicalCloudCertURL+"/{caCert}/logical-clouds/{logicalCloud}/clusters/{cluster}", lcClusterHandler.handleClusterUpdate).Methods("PUT")

	// route to enroll the logical-cloud CA cert
	lcCertEnrollmentHandler := logicalCloudCertEnrollmentHandler{
		manager: setClient(client.LogicalCloudCertEnrollment, mockClient).(logicalcloud.CertEnrollmentManager)}
	router.HandleFunc(logicalCloudCertEnrollmentURL+"/status", lcCertEnrollmentHandler.handleStatus).Methods("GET")
	router.HandleFunc(logicalCloudCertEnrollmentURL+"/instantiate", lcCertEnrollmentHandler.handleInstantiate).Methods("POST")
	router.HandleFunc(logicalCloudCertEnrollmentURL+"/terminate", lcCertEnrollmentHandler.handleTerminate).Methods("POST")
	router.HandleFunc(logicalCloudCertEnrollmentURL+"/update", lcCertEnrollmentHandler.handleUpdate).Methods("POST")

	// route to distribute the logical-cloud CA cert
	lcCertDistributionHandler := logicalCloudCertDistributionHandler{
		manager: setClient(client.LogicalCloudCertDistribution, mockClient).(logicalcloud.CertDistributionManager)}
	router.HandleFunc(logicalCloudCertDistributionURL+"/status", lcCertDistributionHandler.handleStatus).Methods("GET")
	router.HandleFunc(logicalCloudCertDistributionURL+"/instantiate", lcCertDistributionHandler.handleInstantiate).Methods("POST")
	router.HandleFunc(logicalCloudCertDistributionURL+"/terminate", lcCertDistributionHandler.handleTerminate).Methods("POST")
	router.HandleFunc(logicalCloudCertDistributionURL+"/update", lcCertDistributionHandler.handleUpdate).Methods("POST")

	return router
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
		if mockClient != nil && reflect.TypeOf(mockClient).Implements(reflect.TypeOf((*clusterprovider.CertClient)(nil)).Elem()) {
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
