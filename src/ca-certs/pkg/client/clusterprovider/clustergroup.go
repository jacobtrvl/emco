// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package clusterprovider

import (
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
)

// ClusterGroupManager
type ClusterGroupManager interface {
	CreateClusterGroup(cluster module.ClusterGroup, cert, clusterProvider string, failIfExists bool) (module.ClusterGroup, bool, error)
	DeleteClusterGroup(cert, cluster, clusterProvider string) error
	GetAllClusterGroups(cert, clusterProvider string) ([]module.ClusterGroup, error)
	GetClusterGroup(cert, cluster, clusterProvider string) (module.ClusterGroup, error)
}

// ClusterGroupKey
type ClusterGroupKey struct {
	Cert            string `json:"caCertCp"`
	ClusterGroup    string `json:"caCertClusterGroupCp"`
	ClusterProvider string `json:"clusterProvider"`
}

// ClusterGroupClient
type ClusterGroupClient struct {
}

// NewClusterGroupClient
func NewClusterGroupClient() *ClusterGroupClient {
	return &ClusterGroupClient{}
}

// CreateClusterGroup
func (c *ClusterGroupClient) CreateClusterGroup(group module.ClusterGroup, cert, clusterProvider string, failIfExists bool) (module.ClusterGroup, bool, error) {
	ck := ClusterGroupKey{
		Cert:            cert,
		ClusterGroup:    group.MetaData.Name,
		ClusterProvider: clusterProvider}

	return module.NewClusterGroupClient(ck).CreateClusterGroup(group, failIfExists)
}

// DeleteClusterGroup
func (c *ClusterGroupClient) DeleteClusterGroup(cert, group, clusterProvider string) error {
	ck := ClusterGroupKey{
		Cert:            cert,
		ClusterGroup:    group,
		ClusterProvider: clusterProvider}

	return module.NewClusterGroupClient(ck).DeleteClusterGroup()
}

// GetAllClusterGroups
func (c *ClusterGroupClient) GetAllClusterGroups(cert, clusterProvider string) ([]module.ClusterGroup, error) {
	ck := ClusterGroupKey{
		Cert:            cert,
		ClusterProvider: clusterProvider}

	return module.NewClusterGroupClient(ck).GetAllClusterGroups()
}

// GetClusterGroup
func (c *ClusterGroupClient) GetClusterGroup(cert, clusterGroup, clusterProvider string) (module.ClusterGroup, error) {
	ck := ClusterGroupKey{
		Cert:            cert,
		ClusterGroup:    clusterGroup,
		ClusterProvider: clusterProvider}

	return module.NewClusterGroupClient(ck).GetClusterGroup()
}
