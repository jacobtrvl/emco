// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package logicalcloud

import (
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
)

type ClusterManager interface {
	CreateClusterGroup(cluster module.ClusterGroup, logicalCloud, cert, project string, failIfExists bool) (module.ClusterGroup, bool, error)
	DeleteClusterGroup(cluster, logicalCloud, cert, project string) error
	GetAllClusterGroups(logicalCloud, cert, project string) ([]module.ClusterGroup, error)
	GetClusterGroup(cluster, logicalCloud, cert, project string) (module.ClusterGroup, error)
}

// ClusterGroupKey
type ClusterGroupKey struct {
	Cert         string `json:"logicalCloudCert"`
	ClusterGroup string `json:"logicalCloudClusterGroup"`
	LogicalCloud string `json:"caLogicalCloud"`
	Project      string `json:"project"`
}

// ClusterClient
type ClusterClient struct {
}

// NewClusterClient
func NewClusterClient() *ClusterClient {
	return &ClusterClient{}
}

// CreateClusterGroup
func (c *ClusterClient) CreateClusterGroup(group module.ClusterGroup, logicalCloud, cert, project string, failIfExists bool) (module.ClusterGroup, bool, error) {
	ck := ClusterGroupKey{
		Cert:         cert,
		ClusterGroup: group.MetaData.Name,
		LogicalCloud: logicalCloud,
		Project:      project}

	return module.NewClusterClient(ck).CreateClusterGroup(group, failIfExists)
}

// DeleteClusterGroup
func (c *ClusterClient) DeleteClusterGroup(clusterGroup, logicalCloud, cert, project string) error {
	ck := ClusterGroupKey{
		Cert:         cert,
		ClusterGroup: clusterGroup,
		LogicalCloud: logicalCloud,
		Project:      project}

	return module.NewClusterClient(ck).DeleteClusterGroup()
}

// GetAllClusterGroups
func (c *ClusterClient) GetAllClusterGroups(logicalCloud, cert, project string) ([]module.ClusterGroup, error) {
	ck := ClusterGroupKey{
		Cert:         cert,
		LogicalCloud: logicalCloud,
		Project:      project}

	return module.NewClusterClient(ck).GetAllClusterGroups()
}

// GetClusterGroup
func (c *ClusterClient) GetClusterGroup(clusterGroup, logicalCloud, cert, project string) (module.ClusterGroup, error) {
	ck := ClusterGroupKey{
		Cert:         cert,
		ClusterGroup: clusterGroup,
		LogicalCloud: logicalCloud,
		Project:      project}

	return module.NewClusterClient(ck).GetClusterGroup()
}

// // Convert the key to string to preserve the underlying structure
// func (k ClusterGroupKey) String() string {
// 	out, err := json.Marshal(k)
// 	if err != nil {
// 		return ""
// 	}
// 	return string(out)
// }
