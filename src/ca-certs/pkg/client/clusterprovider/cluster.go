// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package clusterprovider

import (
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
)

// ClusterManager
type ClusterManager interface {
	CreateClusterGroup(cluster module.ClusterGroup, cert, clusterProvider string, failIfExists bool) (module.ClusterGroup, bool, error)
	DeleteClusterGroup(cert, cluster, clusterProvider string) error
	GetAllClusterGroups(cert, clusterProvider string) ([]module.ClusterGroup, error)
	GetClusterGroup(cert, cluster, clusterProvider string) (module.ClusterGroup, error)
}

// ClusterGroupKey
type ClusterGroupKey struct {
	Cert            string `json:"clusterProviderCert"`
	ClusterGroup    string `json:"clusterProviderClusterGroup"`
	ClusterProvider string `json:"clusterProvider"`
}

type ClusterKey struct {
	Cert            string `json:"clusterProviderCert"`
	Cluster         string `json:"clusterProviderCluster"`
	ClusterGroup    string `json:"clusterProviderClusterGroup"`
	ClusterProvider string `json:"clusterProvider"`
}

// ClusterClient
type ClusterClient struct {
}

// NewClusterClient
func NewClusterClient() *ClusterClient {
	return &ClusterClient{}
}

// CreateClusterGroup
func (c *ClusterClient) CreateClusterGroup(group module.ClusterGroup, cert, clusterProvider string, failIfExists bool) (module.ClusterGroup, bool, error) {
	ck := ClusterGroupKey{
		Cert:            cert,
		ClusterGroup:    group.MetaData.Name,
		ClusterProvider: clusterProvider}

	return module.NewClusterClient(ck).CreateClusterGroup(group, failIfExists)
}

// DeleteClustersToTheCertificate
func (c *ClusterClient) DeleteClusterGroup(cert, group, clusterProvider string) error {
	ck := ClusterGroupKey{
		Cert:            cert,
		ClusterGroup:    group,
		ClusterProvider: clusterProvider}

	return module.NewClusterClient(ck).DeleteClusterGroup()
}

// GetAllClusterGroups
func (c *ClusterClient) GetAllClusterGroups(cert, clusterProvider string) ([]module.ClusterGroup, error) {
	ck := ClusterGroupKey{
		Cert:            cert,
		ClusterProvider: clusterProvider}

	return module.NewClusterClient(ck).GetAllClusterGroups()
}

// GetClusterGroup
func (c *ClusterClient) GetClusterGroup(cert, clusterGroup, clusterProvider string) (module.ClusterGroup, error) {
	ck := ClusterGroupKey{
		Cert:            cert,
		ClusterGroup:    clusterGroup,
		ClusterProvider: clusterProvider}

	return module.NewClusterClient(ck).GetClusterGroup()
}
