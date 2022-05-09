// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package logicalcloud

import (
	"reflect"

	"github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/clusterprovider"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

// ClusterClient
type ClusterClient struct {
	db DbInfo
}

// ClusterKey
type ClusterKey struct {
	Cert         string `json:"logicalCloudCert"`
	Cluster      string `json:"logicalCloudCertCluster"`
	LogicalCloud string `json:"caLogicalCloud"`
	Project      string `json:"project"`
}

type ClusterManager interface {
	CreateClusterGroup(cluster clusterprovider.ClusterGroup, cert, logicalCloud, project string, failIfExists bool) (clusterprovider.ClusterGroup, bool, error)
	DeleteClusterGroup(cluster, cert, logicalCloud, project string) error
	GetAllClusterGroups(cert, logicalCloud, project string) ([]clusterprovider.ClusterGroup, error)
	GetClusterGroup(cluster, cert, logicalCloud, project string) (clusterprovider.ClusterGroup, error)
}

// CreateClusterGroup
func (c *ClusterClient) CreateClusterGroup(cluster clusterprovider.ClusterGroup, cert, logicalCloud, project string, failIfExists bool) (clusterprovider.ClusterGroup, bool, error) {
	cExists := false
	key := ClusterKey{
		Cert:         cert,
		Cluster:      cluster.MetaData.Name,
		LogicalCloud: logicalCloud,
		Project:      project}

	// TODO:- Confirm if we need to check it exists or directly update
	if clr, err := c.GetClusterGroup(cluster.MetaData.Name, cert, logicalCloud, project); err == nil &&
		!reflect.DeepEqual(clr, clusterprovider.ClusterGroup{}) {
		cExists = true
	}

	if cExists &&
		failIfExists {
		return clusterprovider.ClusterGroup{}, cExists, errors.New("certificate.Cert already exists")
	}

	if err := db.DBconn.Insert(c.db.storeName, key, nil, c.db.tagMeta, cluster); err != nil {
		return clusterprovider.ClusterGroup{}, cExists, err
	}

	return cluster, cExists, nil
}

// DeleteClusterGroup
func (c *ClusterClient) DeleteClusterGroup(cluster, cert, logicalCloud, project string) error {
	key := ClusterKey{
		Cert:         cert,
		Cluster:      cluster,
		LogicalCloud: logicalCloud,
		Project:      project}
	return db.DBconn.Remove(c.db.storeName, key)
}

// GetAllClusterGroups
func (c *ClusterClient) GetAllClusterGroups(cert, logicalCloud, project string) ([]clusterprovider.ClusterGroup, error) {
	key := ClusterKey{
		Cert:         cert,
		LogicalCloud: logicalCloud,
		Project:      project}

	values, err := db.DBconn.Find(c.db.storeName, key, c.db.tagMeta)
	if err != nil {
		return []clusterprovider.ClusterGroup{}, err
	}

	var clusters []clusterprovider.ClusterGroup
	for _, value := range values {
		clr := clusterprovider.ClusterGroup{}
		if err = db.DBconn.Unmarshal(value, &clr); err != nil {
			return []clusterprovider.ClusterGroup{}, err
		}
		clusters = append(clusters, clr)
	}

	return clusters, nil
}

// GetClusterGroup
func (c *ClusterClient) GetClusterGroup(cluster, cert, logicalCloud, project string) (clusterprovider.ClusterGroup, error) {
	key := ClusterKey{
		Cert:         cert,
		Cluster:      cluster,
		LogicalCloud: logicalCloud,
		Project:      project}

	value, err := db.DBconn.Find(c.db.storeName, key, c.db.tagMeta)
	if err != nil {
		return clusterprovider.ClusterGroup{}, err
	}

	if len(value) == 0 {
		return clusterprovider.ClusterGroup{}, errors.New("clusterprovider.ClusterGroup not found")
	}

	if value != nil {
		c := clusterprovider.ClusterGroup{}
		if err = db.DBconn.Unmarshal(value[0], &c); err != nil {
			return clusterprovider.ClusterGroup{}, err
		}
		return c, nil
	}

	return clusterprovider.ClusterGroup{}, errors.New("Unknown Error")
}

// NewClusterClient
func NewClusterClient() *ClusterClient {
	return &ClusterClient{
		db: DbInfo{
			storeName: "resources",
			tagMeta:   "data"}}
}

// // Convert the key to string to preserve the underlying structure
// func (k ClusterKey) String() string {
// 	out, err := json.Marshal(k)
// 	if err != nil {
// 		return ""
// 	}
// 	return string(out)
// }
