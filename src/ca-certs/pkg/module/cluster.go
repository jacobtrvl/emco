// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package module

import (
	"reflect"
	"strings"

	"github.com/pkg/errors"
	clm "gitlab.com/project-emco/core/emco-base/src/clm/pkg/cluster"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

// ClusterManager
type ClusterManager interface {
	CreateClusterGroup(cluster ClusterGroup, failIfExists bool) (ClusterGroup, bool, error)
	DeleteClusterGroup() error
	GetAllClusterGroups() ([]ClusterGroup, error)
	GetClusterGroup() (ClusterGroup, error)
}

// ClusterClient
type ClusterClient struct {
	dbInfo db.DbInfo
	dbKey  interface{}
}

// NewClusterClient
func NewClusterClient(dbKey interface{}) *ClusterClient {
	return &ClusterClient{
		dbInfo: db.DbInfo{
			StoreName: "resources",
			TagMeta:   "data"},
		dbKey: dbKey}
}

// CreateClusterGroup
func (c *ClusterClient) CreateClusterGroup(group ClusterGroup, failIfExists bool) (ClusterGroup, bool, error) {
	cExists := false
	// TODO:- Confirm if we need to check it exists or directly update
	if clr, err := c.GetClusterGroup(); err == nil &&
		!reflect.DeepEqual(clr, ClusterGroup{}) {
		cExists = true
	}

	if cExists &&
		failIfExists {
		return ClusterGroup{}, cExists, errors.New("ClusterGroup already exists")
	}

	if err := db.DBconn.Insert(c.dbInfo.StoreName, c.dbKey, nil, c.dbInfo.TagMeta, group); err != nil {
		return ClusterGroup{}, cExists, err
	}

	return group, cExists, nil
}

// DeleteClustersToTheCertificate
func (c *ClusterClient) DeleteClusterGroup() error {
	return db.DBconn.Remove(c.dbInfo.StoreName, c.dbKey)
}

// GetAllClusterGroups
func (c *ClusterClient) GetAllClusterGroups() ([]ClusterGroup, error) {
	values, err := db.DBconn.Find(c.dbInfo.StoreName, c.dbKey, c.dbInfo.TagMeta)
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
func (c *ClusterClient) GetClusterGroup() (ClusterGroup, error) {
	value, err := db.DBconn.Find(c.dbInfo.StoreName, c.dbKey, c.dbInfo.TagMeta)
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

// GetClusters returns the list of clusters based on the scope
func GetClusters(group ClusterGroup) (clusters []string, err error) {
	clusters = []string{}
	switch strings.ToLower(group.Spec.Scope) {
	case "name":
		// TODO: Confirm if we need to verify the cluster exists or not
		// get cluster by provider and the name
		if _, err = clm.NewClusterClient().GetCluster(group.Spec.Provider, group.Spec.Name); err != nil {
			return clusters, err
		}

		clusters = append(clusters, group.Spec.Name)
	case "label":
		// get clusters by label
		list, err := clm.NewClusterClient().GetClustersWithLabel(group.Spec.Provider, group.Spec.Label)
		if err != nil {
			return clusters, err
		}

		// TODO: Confirm if we need to veirfy the cluster exists or not
		for _, name := range list {
			// get cluster by provider and the name
			if _, err = clm.NewClusterClient().GetCluster(group.Spec.Provider, name); err != nil {
				return clusters, err
			}
		}

		clusters = append(clusters, list...)
	}

	return clusters, err
}
