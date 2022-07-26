// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package module

import (
	"reflect"
	"strings"

	clm "gitlab.com/project-emco/core/emco-base/src/clm/pkg/cluster"
	dcm "gitlab.com/project-emco/core/emco-base/src/dcm/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/common/emcoerror"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

// ClusterGroupManager exposes all the clusterGroup functionalities
type ClusterGroupManager interface {
	CreateClusterGroup(cluster ClusterGroup, failIfExists bool) (ClusterGroup, bool, error)
	DeleteClusterGroup() error
	GetAllClusterGroups() ([]ClusterGroup, error)
	GetClusterGroup() (ClusterGroup, error)
}

// ClusterGroupClient holds the client properties
type ClusterGroupClient struct {
	dbInfo db.DbInfo
	dbKey  interface{}
}

// NewClusterGroupClient returns an instance of the ClusterGroupClient which implements the Manager
func NewClusterGroupClient(dbKey interface{}) *ClusterGroupClient {
	return &ClusterGroupClient{
		dbInfo: db.DbInfo{
			StoreName: "resources",
			TagMeta:   "data"},
		dbKey: dbKey}
}

// CreateClusterGroup creates a clusterGroup
func (c *ClusterGroupClient) CreateClusterGroup(group ClusterGroup, failIfExists bool) (ClusterGroup, bool, error) {
	cExists := false

	if clr, err := c.GetClusterGroup(); err == nil &&
		!reflect.DeepEqual(clr, ClusterGroup{}) {
		cExists = true
	}

	if cExists &&
		failIfExists {
		return ClusterGroup{}, cExists, &emcoerror.Error{
			Kind: emcoerror.CaCertClusterGroupAlreadyExists,
		}
	}

	if err := db.DBconn.Insert(c.dbInfo.StoreName, c.dbKey, nil, c.dbInfo.TagMeta, group); err != nil {
		return ClusterGroup{}, cExists, err
	}

	return group, cExists, nil
}

// DeleteClusterGroup deletes a clusterGroup
func (c *ClusterGroupClient) DeleteClusterGroup() error {
	return db.DBconn.Remove(c.dbInfo.StoreName, c.dbKey)
}

// GetAllClusterGroups returns  all the clusterGroup
func (c *ClusterGroupClient) GetAllClusterGroups() ([]ClusterGroup, error) {
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

// GetClusterGroup returns the clusterGroup
func (c *ClusterGroupClient) GetClusterGroup() (ClusterGroup, error) {
	value, err := db.DBconn.Find(c.dbInfo.StoreName, c.dbKey, c.dbInfo.TagMeta)
	if err != nil {
		return ClusterGroup{}, err
	}

	if len(value) == 0 {
		return ClusterGroup{}, &emcoerror.Error{
			Kind: emcoerror.CaCertClusterGroupNotFound,
		}
	}

	if value != nil {
		c := ClusterGroup{}
		if err = db.DBconn.Unmarshal(value[0], &c); err != nil {
			return ClusterGroup{}, err
		}
		return c, nil
	}

	return ClusterGroup{}, &emcoerror.Error{
		Kind: emcoerror.UnknownError,
	}
}

// GetClusters returns the list of clusters based on the logicalcloud and scope
func GetClusters(group ClusterGroup, project, logicalcloud string) (clusters []string, err error) {
	if len(logicalcloud) > 0 {
		return getLogicalCloudReferencedClusters(group, project, logicalcloud)
	}

	return getClusters(group)
}

// getClusters returns the list of clusters based on the scope
func getClusters(group ClusterGroup) (clusters []string, err error) {
	clusters = []string{}
	switch strings.ToLower(group.Spec.Scope) {
	case "name":
		// get cluster by provider and the name
		if _, err = clm.NewClusterClient().GetCluster(group.Spec.Provider, group.Spec.Cluster); err != nil {
			return clusters, err
		}

		clusters = append(clusters, group.Spec.Cluster)
	case "label":
		// get clusters by label
		list, err := clm.NewClusterClient().GetClustersWithLabel(group.Spec.Provider, group.Spec.Label)
		if err != nil {
			return clusters, err
		}

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

// getLogicalCloudReferencedClusters returns the list of clusters part of the logicalCloud
func getLogicalCloudReferencedClusters(group ClusterGroup, project, logicalCloud string) ([]string, error) {
	cList, err := getClusters(group)
	if err != nil {
		return []string{}, err
	}

	// get all the clusters referenced by the project and logicalCloud
	cListLc, err := dcm.NewClusterClient().GetAllClusters(project, logicalCloud)
	if err != nil {
		return []string{}, err
	}

	// filter clusters which are referenced by the clusterGroup, project and logicalCloud
	var clusters []string = []string{}
	for _, c := range cList {
		// check if the cluster is referenced by the logicalCloud
		for _, cLc := range cListLc {
			if c == cLc.Specification.ClusterName {
				clusters = append(clusters, c)
				break
			}
		}
	}

	return clusters, nil
}
