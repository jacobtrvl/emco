package module

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

import (
	"strings"

	"gitlab.com/project-emco/core/emco-base/src/clm/pkg/cluster"
)

// isValidClusterToApplyByLabel checks if the cluster being authenticated for(acName) falls under the given label(cLabel) and provider(cProvider)
func isValidClusterToApplyByLabel(cProvider, acName, cLabel string) (bool, error) {

	clusterNamesList, err := cluster.NewClusterClient().GetClustersWithLabel(cProvider, cLabel)
	if err != nil {
		return false, err
	}
	acName = strings.Split(acName, SEPARATOR)[1]
	for _, cn := range clusterNamesList {

		if cn == acName {
			return true, nil
		}
	}
	return false, nil
}

// isValidClusterToApplyByName checks if a given cluster(gcName) under a provider(cProvider) matches with the cluster which is authenticated for(acName).
func isValidClusterToApplyByName(cProvider, acName, gcName string) (bool, error) {

	clusterNamesList, err := cluster.NewClusterClient().GetClusters(cProvider)
	if err != nil {
		return false, err
	}
	acName = strings.Split(acName, SEPARATOR)[1]
	for _, cn := range clusterNamesList {
		if cn.Metadata.Name == acName && acName == gcName {
			return true, nil
		}
	}
	return false, nil
}
