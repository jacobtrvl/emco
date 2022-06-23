// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package logicalcloud

import (
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
)

// ClusterGroupManager
type ClusterGroupManager interface {
	CreateClusterGroup(cluster module.ClusterGroup, logicalCloud, cert, project string, failIfExists bool) (module.ClusterGroup, bool, error)
	DeleteClusterGroup(cluster, logicalCloud, cert, project string) error
	GetAllClusterGroups(logicalCloud, cert, project string) ([]module.ClusterGroup, error)
	GetClusterGroup(cluster, logicalCloud, cert, project string) (module.ClusterGroup, error)
}

// ClusterGroupKey
type ClusterGroupKey struct {
	Cert               string `json:"caCertLc"`
	ClusterGroup       string `json:"caCertClusterGroupLc"`
	CaCertLogicalCloud string `json:"caCertLogicalCloud"`
	Project            string `json:"project"`
}

// ClusterGroupClient
type ClusterGroupClient struct {
}

// NewClusterGroupClient
func NewClusterGroupClient() *ClusterGroupClient {
	return &ClusterGroupClient{}
}

// CreateClusterGroup
func (c *ClusterGroupClient) CreateClusterGroup(group module.ClusterGroup, caCertLogicalCloud, cert, project string, failIfExists bool) (module.ClusterGroup, bool, error) {
	ck := ClusterGroupKey{
		Cert:               cert,
		ClusterGroup:       group.MetaData.Name,
		CaCertLogicalCloud: caCertLogicalCloud,
		Project:            project}

	return module.NewClusterGroupClient(ck).CreateClusterGroup(group, failIfExists)
}

// DeleteClusterGroup
func (c *ClusterGroupClient) DeleteClusterGroup(clusterGroup, caCertLogicalCloud, cert, project string) error {
	ck := ClusterGroupKey{
		Cert:               cert,
		ClusterGroup:       clusterGroup,
		CaCertLogicalCloud: caCertLogicalCloud,
		Project:            project}

	return module.NewClusterGroupClient(ck).DeleteClusterGroup()
}

// GetAllClusterGroups
func (c *ClusterGroupClient) GetAllClusterGroups(caCertLogicalCloud, cert, project string) ([]module.ClusterGroup, error) {
	ck := ClusterGroupKey{
		Cert:               cert,
		CaCertLogicalCloud: caCertLogicalCloud,
		Project:            project}

	return module.NewClusterGroupClient(ck).GetAllClusterGroups()
}

// GetClusterGroup
func (c *ClusterGroupClient) GetClusterGroup(clusterGroup, caCertLogicalCloud, cert, project string) (module.ClusterGroup, error) {
	ck := ClusterGroupKey{
		Cert:               cert,
		ClusterGroup:       clusterGroup,
		CaCertLogicalCloud: caCertLogicalCloud,
		Project:            project}

	return module.NewClusterGroupClient(ck).GetClusterGroup()
}
