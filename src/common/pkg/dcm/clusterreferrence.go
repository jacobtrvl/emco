// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020-2022 Intel Corporation

package dcm

// Cluster contains the parameters needed for a Cluster
type ClusterReference struct {
	MetaData      ClusterMeta `json:"metadata"`
	Specification ClusterSpec `json:"spec"`
}

type ClusterMeta struct {
	ClusterReference string `json:"name"`
	Description      string `json:"description"`
	UserData1        string `json:"userData1"`
	UserData2        string `json:"userData2"`
}

type ClusterSpec struct {
	ClusterProvider string `json:"clusterProvider"`
	ClusterName     string `json:"cluster"`
	LoadBalancerIP  string `json:"loadBalancerIP"`
	Certificate     string `json:"certificate"`
}
