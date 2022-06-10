// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package clusterprovider

import (
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

// getCertificate
func getCertificate(cert, clusterProvider string) (module.Cert, error) {
	// verify the ca cert
	caCert, err := NewCertClient().GetCert(cert, clusterProvider)
	if err != nil {
		logutils.Error("Failed to retrieve the caCert", logutils.Fields{
			"Cert":            cert,
			"ClusterProvider": clusterProvider,
			"Error":           err.Error()})
		return module.Cert{}, err
	}
	return caCert, nil
}

// getAllClusterGroup
func getAllClusterGroup(cert, clusterProvider string) ([]module.ClusterGroup, error) {
	// get all the clusters within the ca cert and cluster provider
	clusters, err := NewClusterClient().GetAllClusterGroups(cert, clusterProvider)
	if err != nil {
		logutils.Error("Failed to retrieve the clusterGroup(s)", logutils.Fields{
			"Cert":            cert,
			"ClusterProvider": clusterProvider,
			"Error":           err.Error()})
		return []module.ClusterGroup{}, err
	}

	return clusters, nil
}
