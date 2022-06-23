// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package clusterprovider

import (
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

// getCertificate retrieves the caCert from db
func getCertificate(cert, clusterProvider string) (module.CaCert, error) {
	// verify the caCert
	caCert, err := NewCaCertClient().GetCert(cert, clusterProvider)
	if err != nil {
		logutils.Error("Failed to retrieve the caCert", logutils.Fields{
			"Cert":            cert,
			"ClusterProvider": clusterProvider,
			"Error":           err.Error()})
		return module.CaCert{}, err
	}
	return caCert, nil
}

// getAllClusterGroup retrieves the clusterGroup(s) from db
func getAllClusterGroup(cert, clusterProvider string) ([]module.ClusterGroup, error) {
	// get all the clusters within the caCert and clusterProvider
	clusters, err := NewClusterGroupClient().GetAllClusterGroups(cert, clusterProvider)
	if err != nil {
		logutils.Error("Failed to retrieve the clusterGroup(s)", logutils.Fields{
			"Cert":            cert,
			"ClusterProvider": clusterProvider,
			"Error":           err.Error()})
		return []module.ClusterGroup{}, err
	}

	return clusters, nil
}
