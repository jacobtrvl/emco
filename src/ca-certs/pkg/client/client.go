// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package client

import (
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/client/clusterprovider"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/client/logicalcloud"
)

// Client for using the services
type Client struct {
	ClusterProviderCert             *clusterprovider.CertClient
	ClusterProviderCluster          *clusterprovider.ClusterClient
	ClusterProviderCertDistribution *clusterprovider.CertDistributionClient
	ClusterProviderCertEnrollment   *clusterprovider.CertEnrollmentClient
	LogicalCloud                    *logicalcloud.LogicalCloudClient
	LogicalCloudCert                *logicalcloud.CertClient
	LogicalCloudCluster             *logicalcloud.ClusterClient
	LogicalCloudCertDistribution    *logicalcloud.CertDistributionClient
	LogicalCloudCertEnrollment      *logicalcloud.CertEnrollmentClient
}

// NewClient creates a new client for using the services
func NewClient() *Client {
	c := &Client{}
	c.ClusterProviderCert = clusterprovider.NewCertClient()
	c.ClusterProviderCluster = clusterprovider.NewClusterClient()
	c.ClusterProviderCertDistribution = clusterprovider.NewCertDistributionClient()
	c.ClusterProviderCertEnrollment = clusterprovider.NewCertEnrollmentClient()
	c.LogicalCloud = logicalcloud.NewLogicalCloudClient()
	c.LogicalCloudCert = logicalcloud.NewCertClient()
	c.LogicalCloudCluster = logicalcloud.NewClusterClient()
	c.LogicalCloudCertDistribution = logicalcloud.NewCertDistributionClient()
	c.LogicalCloudCertEnrollment = logicalcloud.NewCertEnrollmentClient()
	return c
}
