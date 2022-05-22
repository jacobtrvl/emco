// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package api

import (
	"github.com/gorilla/mux"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/client"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/client/clusterprovider"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/client/logicalcloud"
)

type route struct {
	router *mux.Router
	client *client.Client
	mock   interface{}
}

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

func (r *route) setClusterProviderRoutes() {
	// route to create the cluster-provider CA cert intent
	cpCert := cpCertHandler{
		manager: setClient(r.client.ClusterProviderCert, r.mock).(clusterprovider.CertManager)}
	r.router.HandleFunc(clusterProviderCertURL, cpCert.handleCertificateGet).Methods("GET")
	r.router.HandleFunc(clusterProviderCertURL, cpCert.handleCertificateCreate).Methods("POST")
	r.router.HandleFunc(clusterProviderCertURL+"/{caCert}", cpCert.handleCertificateGet).Methods("GET")
	r.router.HandleFunc(clusterProviderCertURL+"/{caCert}", cpCert.handleCertificateDelete).Methods("DELETE")
	r.router.HandleFunc(clusterProviderCertURL+"/{caCert}", cpCert.handleCertificateUpdate).Methods("PUT")
	// route to add clusters to the cluster-provider CA cert intent
	cpCluster := cpClusterHandler{
		manager: setClient(r.client.ClusterProviderCluster, r.mock).(clusterprovider.ClusterManager)}
	r.router.HandleFunc(clusterProviderCertURL+"/{caCert}/clusters", cpCluster.handleClusterGet).Methods("GET")
	r.router.HandleFunc(clusterProviderCertURL+"/{caCert}/clusters", cpCluster.handleClusterCreate).Methods("POST")
	r.router.HandleFunc(clusterProviderCertURL+"/{caCert}/clusters/{cluster}", cpCluster.handleClusterGet).Methods("GET")
	r.router.HandleFunc(clusterProviderCertURL+"/{caCert}/clusters/{cluster}", cpCluster.handleClusterDelete).Methods("DELETE")
	r.router.HandleFunc(clusterProviderCertURL+"/{caCert}/clusters/{cluster}", cpCluster.handleClusterUpdate).Methods("PUT")

	// route to enroll the cluster-provider CA cert intent
	cpCertEnrollment := cpCertEnrollmentHandler{
		manager: setClient(r.client.ClusterProviderCertEnrollment, r.mock).(clusterprovider.CertEnrollmentManager)}
	r.router.HandleFunc(clusterProviderCertEnrollmentURL+"/status", cpCertEnrollment.handleStatus).Methods("GET")
	r.router.HandleFunc(clusterProviderCertEnrollmentURL+"/instantiate", cpCertEnrollment.handleInstantiate).Methods("POST")
	r.router.HandleFunc(clusterProviderCertEnrollmentURL+"/terminate", cpCertEnrollment.handleTerminate).Methods("POST")
	r.router.HandleFunc(clusterProviderCertEnrollmentURL+"/update", cpCertEnrollment.handleUpdate).Methods("POST")

	// route to distribute the cluster-provider CA cert intent
	cpCertDistribution := cpCertDistributionHandler{
		manager: setClient(r.client.ClusterProviderCertDistribution, r.mock).(clusterprovider.CertDistributionManager)}
	r.router.HandleFunc(clusterProviderCertDistributionURL+"/status", cpCertDistribution.handleStatus).Methods("GET")
	r.router.HandleFunc(clusterProviderCertDistributionURL+"/instantiate", cpCertDistribution.handleInstantiate).Methods("POST")
	r.router.HandleFunc(clusterProviderCertDistributionURL+"/terminate", cpCertDistribution.handleTerminate).Methods("POST")
	r.router.HandleFunc(clusterProviderCertDistributionURL+"/update", cpCertDistribution.handleUpdate).Methods("POST")

}

func (r *route) setLogicalCloudRoutes() {
	// route to create the logical-cloud CA cert intent
	lcCert := lcCertHandler{
		manager: setClient(r.client.LogicalCloudCert, r.mock).(logicalcloud.CertManager)}
	r.router.HandleFunc(logicalCloudCertURL, lcCert.handleCertificateGet).Methods("GET")
	r.router.HandleFunc(logicalCloudCertURL, lcCert.handleCertificateCreate).Methods("POST")
	r.router.HandleFunc(logicalCloudCertURL+"/{caCert}", lcCert.handleCertificateGet).Methods("GET")
	r.router.HandleFunc(logicalCloudCertURL+"/{caCert}", lcCert.handleCertificateDelete).Methods("DELETE")
	r.router.HandleFunc(logicalCloudCertURL+"/{caCert}", lcCert.handleCertificateUpdate).Methods("PUT")

	// route to add logical-cloud to the logical-cloud CA cert intent
	lc := lcHandler{
		manager: setClient(r.client.LogicalCloud, r.mock).(logicalcloud.LogicalCloudManager)}
	r.router.HandleFunc(logicalCloudCertURL+"/{caCert}/logical-clouds", lc.handleLogicalCloudGet).Methods("GET")
	r.router.HandleFunc(logicalCloudCertURL+"/{caCert}/logical-clouds", lc.handleLogicalCloudCreate).Methods("POST")
	r.router.HandleFunc(logicalCloudCertURL+"/{caCert}/logical-clouds/{logicalCloud}", lc.handleLogicalCloudGet).Methods("GET")
	r.router.HandleFunc(logicalCloudCertURL+"/{caCert}/logical-clouds/{logicalCloud}", lc.handleLogicalCloudDelete).Methods("DELETE")
	r.router.HandleFunc(logicalCloudCertURL+"/{caCert}/logical-clouds/{logicalCloud}", lc.handleLogicalCloudUpdate).Methods("PUT")

	// route to add clusters to the logical-cloud CA cert intent
	lcCluster := lcClusterHandler{
		manager: setClient(r.client.LogicalCloudCluster, r.mock).(logicalcloud.ClusterManager)}
	r.router.HandleFunc(logicalCloudCertURL+"/{caCert}/logical-clouds/{logicalCloud}/clusters", lcCluster.handleClusterGet).Methods("GET")
	r.router.HandleFunc(logicalCloudCertURL+"/{caCert}/logical-clouds/{logicalCloud}/clusters", lcCluster.handleClusterCreate).Methods("POST")
	r.router.HandleFunc(logicalCloudCertURL+"/{caCert}/logical-clouds/{logicalCloud}/clusters/{cluster}", lcCluster.handleClusterGet).Methods("GET")
	r.router.HandleFunc(logicalCloudCertURL+"/{caCert}/logical-clouds/{logicalCloud}/clusters/{cluster}", lcCluster.handleClusterDelete).Methods("DELETE")
	r.router.HandleFunc(logicalCloudCertURL+"/{caCert}/logical-clouds/{logicalCloud}/clusters/{cluster}", lcCluster.handleClusterUpdate).Methods("PUT")

	// route to enroll the logical-cloud CA cert intent
	lcCertEnrollment := lcCertEnrollmentHandler{
		manager: setClient(r.client.LogicalCloudCertEnrollment, r.mock).(logicalcloud.CertEnrollmentManager)}
	r.router.HandleFunc(logicalCloudCertEnrollmentURL+"/status", lcCertEnrollment.handleStatus).Methods("GET")
	r.router.HandleFunc(logicalCloudCertEnrollmentURL+"/instantiate", lcCertEnrollment.handleInstantiate).Methods("POST")
	r.router.HandleFunc(logicalCloudCertEnrollmentURL+"/terminate", lcCertEnrollment.handleTerminate).Methods("POST")
	r.router.HandleFunc(logicalCloudCertEnrollmentURL+"/update", lcCertEnrollment.handleUpdate).Methods("POST")

	// route to distribute the logical-cloud CA cert intent
	lcCertDistribution := lcCertDistributionHandler{
		manager: setClient(r.client.LogicalCloudCertDistribution, r.mock).(logicalcloud.CertDistributionManager)}
	r.router.HandleFunc(logicalCloudCertDistributionURL+"/status", lcCertDistribution.handleStatus).Methods("GET")
	r.router.HandleFunc(logicalCloudCertDistributionURL+"/instantiate", lcCertDistribution.handleInstantiate).Methods("POST")
	r.router.HandleFunc(logicalCloudCertDistributionURL+"/terminate", lcCertDistribution.handleTerminate).Methods("POST")
	r.router.HandleFunc(logicalCloudCertDistributionURL+"/update", lcCertDistribution.handleUpdate).Methods("POST")
}
