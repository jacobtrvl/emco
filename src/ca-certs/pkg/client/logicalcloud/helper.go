// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package logicalcloud

import (
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

func getCertificate(cert, project string) (module.Cert, error) {
	// verify the ca cert
	caCert, err := NewCertClient().GetCert(cert, project)
	if err != nil {
		logutils.Error("Failed to instantiate the enrollment", logutils.Fields{
			"Cert":    cert,
			"Project": project,
			"Error":   err.Error()})
		return module.Cert{}, err
	}
	return caCert, nil
}

func getAllLogicalClouds(cert, project string) ([]LogicalCloud, error) {
	// verify the ca cert
	lcs, err := NewLogicalCloudClient().GetAllLogicalClouds(cert, project)
	if err != nil {
		logutils.Error("Failed to instantiate the enrollment", logutils.Fields{
			"Cert":    cert,
			"Project": project,
			"Error":   err.Error()})
		return []LogicalCloud{}, err
	}
	return lcs, nil
}

func getAllClusterGroup(logicalCloud, cert, project string) ([]module.ClusterGroup, error) {
	// get all the clusters within the ca cert and cluster provider
	clusters, err := NewClusterClient().GetAllClusterGroups(logicalCloud, cert, project)
	if err != nil {
		logutils.Error("Failed to instantiate the enrollment", logutils.Fields{
			"Cert":         cert,
			"LogicalCloud": logicalCloud,
			"Project":      project,
			"Error":        err.Error()})
		return []module.ClusterGroup{}, err
	}

	return clusters, nil
}
