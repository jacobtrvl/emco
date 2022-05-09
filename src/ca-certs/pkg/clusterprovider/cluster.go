// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package clusterprovider

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/clm/pkg/cluster"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

// ClusterManager
type ClusterManager interface {
	CreateClusterGroup(cluster ClusterGroup, cert, clusterProvider string, failIfExists bool) (ClusterGroup, bool, error)
	DeleteClusterGroup(cert, cluster, clusterProvider string) error
	GetAllClusterGroups(cert, clusterProvider string) ([]ClusterGroup, error)
	GetClusterGroup(cert, cluster, clusterProvider string) (ClusterGroup, error)
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
	db DbInfo
}

// NewClusterClient
func NewClusterClient() *ClusterClient {
	return &ClusterClient{
		db: DbInfo{
			storeName: "resources",
			tagMeta:   "data"}}
}

// CreateClusterGroup
func (c *ClusterClient) CreateClusterGroup(group ClusterGroup, cert, clusterProvider string, failIfExists bool) (ClusterGroup, bool, error) {
	cExists := false
	key := ClusterGroupKey{
		Cert:            cert,
		ClusterGroup:    group.MetaData.Name,
		ClusterProvider: clusterProvider}

	// TODO:- Confirm if we need to check it exists or directly update
	if clr, err := c.GetClusterGroup(cert, group.MetaData.Name, clusterProvider); err == nil &&
		!reflect.DeepEqual(clr, ClusterGroup{}) {
		cExists = true
	}

	if cExists &&
		failIfExists {
		return ClusterGroup{}, cExists, errors.New("ClusterGroup already exists")
	}

	if err := db.DBconn.Insert(c.db.storeName, key, nil, c.db.tagMeta, group); err != nil {
		return ClusterGroup{}, cExists, err
	}

	// create a cluster resource for all the clusters in the group
	switch strings.ToLower(group.Spec.Scope) {
	case "name":
		if err := c.createCluster(group.Spec.Name, group.MetaData.Name, clusterProvider, cert); err != nil {
			return ClusterGroup{}, cExists, err // TODO - revisti thsi logic
		}

	case "label":
		// get clusters by label TODO: Confirm if we need to veirfy the cluster exists or not
		list, err := cluster.NewClusterClient().GetClustersWithLabel(clusterProvider, group.Spec.Label)
		if err != nil {
			fmt.Println("Failed to get clusters by label", cert, clusterProvider, group.Spec.Label, "", err)
			return ClusterGroup{}, cExists, err
		}

		for _, name := range list { // TODO - revisti thsi logic, this is not required if we don't have to save the cert details into the db
			if err := c.createCluster(name, group.MetaData.Name, clusterProvider, cert); err != nil {
				return ClusterGroup{}, cExists, err
			}
		}
	}

	return group, cExists, nil
}

// DeleteClustersToTheCertificate
func (c *ClusterClient) DeleteClusterGroup(cert, group, clusterProvider string) error {
	key := ClusterGroupKey{
		Cert:            cert,
		ClusterGroup:    group,
		ClusterProvider: clusterProvider}

	clusters, err := c.getCluster("", cert, group, clusterProvider)
	if err != nil {
		fmt.Println(err.Error())
	}
	for _, cluster := range clusters {
		m := MetaData{}
		err = db.DBconn.Unmarshal(cluster, &m)
		if err != nil {
			fmt.Println(err.Error())
		}
		if er := c.deleteCluster(m.Name, group, clusterProvider, cert); er != nil {
			fmt.Println(er.Error())
		}

	}

	return db.DBconn.Remove(c.db.storeName, key)
}

// GetAllClusterGroups
func (c *ClusterClient) GetAllClusterGroups(cert, clusterProvider string) ([]ClusterGroup, error) {
	key := ClusterGroupKey{
		Cert:            cert,
		ClusterProvider: clusterProvider}

	values, err := db.DBconn.Find(c.db.storeName, key, c.db.tagMeta)
	if err != nil {
		return []ClusterGroup{}, err
	}

	var clusters []ClusterGroup
	for _, value := range values {
		clr := ClusterGroup{}
		if err = db.DBconn.Unmarshal(value, &clr); err != nil {
			return []ClusterGroup{}, err
		}
		clusters = append(clusters, clr)
	}

	return clusters, nil

}

// GetClusterGroup
func (c *ClusterClient) GetClusterGroup(cert, cluster, clusterProvider string) (ClusterGroup, error) {
	key := ClusterGroupKey{
		Cert:            cert,
		ClusterGroup:    cluster,
		ClusterProvider: clusterProvider}

	value, err := db.DBconn.Find(c.db.storeName, key, c.db.tagMeta)
	if err != nil {
		return ClusterGroup{}, err
	}

	if len(value) == 0 {
		return ClusterGroup{}, errors.New("ClusterGroup not found")
	}

	if value != nil {
		c := ClusterGroup{}
		if err = db.DBconn.Unmarshal(value[0], &c); err != nil {
			return ClusterGroup{}, err
		}
		return c, nil
	}

	return ClusterGroup{}, errors.New("Unknown Error")

}

// createCluster creates cluster resource for each cluster in the group
func (c *ClusterClient) createCluster(cluster, group, clusterProvider, cert string) error {
	key := ClusterKey{
		Cert:            cert,
		Cluster:         cluster,
		ClusterGroup:    group,
		ClusterProvider: clusterProvider}

	m := MetaData{
		Name: cluster,
	}

	return db.DBconn.Insert(c.db.storeName, key, nil, c.db.tagMeta, m)
}

// deleteCluster deletes cluster resource for each cluster in the group
func (c *ClusterClient) deleteCluster(cluster, group, clusterProvider, cert string) error {
	key := ClusterKey{
		Cert:            cert,
		Cluster:         cluster,
		ClusterGroup:    group,
		ClusterProvider: clusterProvider}

	return db.DBconn.Remove(c.db.storeName, key)
}

// getCluster retrieves cluster resource for each cluster in the group
func (c *ClusterClient) getCluster(cluster, cert, group, clusterProvider string) ([][]byte, error) {
	key := ClusterKey{
		Cert:            cert,
		ClusterGroup:    group,
		ClusterProvider: clusterProvider}

	return db.DBconn.Find(c.db.storeName, key, c.db.tagMeta)
}
