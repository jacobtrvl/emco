// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package clusterprovider

import (
	"fmt"
	"strings"

	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certificate"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

//
type StateOptions struct {
	Key interface{}
	Collection,
	ContextID,
	State,
	Tag string
	CreateIfNotExists bool
}

func getCertificate(cert, clusterProvider string) (certificate.Cert, error) {
	// verify the ca cert
	caCert, err := NewCertClient().GetCert(cert, clusterProvider)
	if err != nil {
		logutils.Error("Failed to instantiate the enrollment", logutils.Fields{
			"Cert":            cert,
			"ClusterProvider": clusterProvider,
			"Error":           err.Error()})
		return certificate.Cert{}, err
	}
	return caCert, nil
}

func getAllClusterGroup(cert, clusterProvider string) ([]ClusterGroup, error) {
	// //verify the cluster provider
	// cClient := cluster.NewClusterClient()
	// if _, err := cClient.GetClusterProvider(clusterProvider); err != nil {
	// 	logutils.Error("Failed to instantiate the enrollment", logutils.Fields{
	// 		"Cert":            cert,
	// 		"ClusterProvider": clusterProvider,
	// 		"Error":           err.Error()})
	// 	return []Cluster{}, err
	// }
	// get all the clusters within the ca cert and cluster provider
	clusters, err := NewClusterClient().GetAllClusterGroups(cert, clusterProvider)
	if err != nil {
		logutils.Error("Failed to instantiate the enrollment", logutils.Fields{
			"Cert":            cert,
			"ClusterProvider": clusterProvider,
			"Error":           err.Error()})
		return []ClusterGroup{}, err
	}

	return clusters, nil
}

func certificateRequestName(cert, cluster, clusterProvider, contextID string) string {
	return strings.ToLower(strings.Join([]string{cert, clusterProvider, cluster, contextID, "cr"}, "-"))
}
func certificateRequestResourceName(name, kind string) string {
	return strings.ToLower(strings.Join([]string{name, kind}, "+"))
}

func secretName(cert, cluster, clusterProvider, contextID string) string {
	return strings.ToLower(strings.Join([]string{cert, clusterProvider, cluster, contextID, "ca"}, "-"))
}

func proxyConfigName(cert, cluster, clusterProvider, contextID string) string {
	return strings.ToLower(strings.Join([]string{cert, clusterProvider, cluster, contextID, "pc"}, "-"))
}

func clusterIssuerName(cert, cluster, clusterProvider, contextID string) string {
	return strings.ToLower(strings.Join([]string{cert, clusterProvider, cluster, contextID, "issuer"}, "-"))
}

func addResource(context appcontext.AppContext, handle interface{}, name string, value interface{}) error {
	h, err := context.AddResource(handle, name, value)
	if err != nil {
		logutils.Error("Failed to add resource(s) to the app",
			logutils.Fields{
				"Handle":   handle,
				"Resource": name,
				"Value":    value, // TODO: remove this from the erroe logs
				"Error":    err.Error()})
		// deleteCompositeApp(context) // TODO: COnfirm if we need to delete this for any single resource add
		return err
	}

	fmt.Println("addResource: ", h)

	return nil
}

func deleteCompositeApp(context appcontext.AppContext) error {
	if err := context.DeleteCompositeApp(); err != nil {
		logutils.Error("Failed to delete the compositeApp",
			logutils.Fields{
				"Error": err.Error()})
		return err
	}

	return nil
}
