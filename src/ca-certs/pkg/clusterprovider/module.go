// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package clusterprovider

// DbInfo holds the MongoDB collection and attributes info
type DbInfo struct {
	storeName string // name of the mongodb collection to use for client documents
	tagMeta   string // attribute key name for the json data of a client document
	tagState  string //
}

type ClusterGroup struct {
	MetaData MetaData         `json:"metadata"`
	Spec     ClusterGroupSpec `json:"spec"`
}

type ClusterGroupSpec struct {
	Label string `json:"label"` // select all the clusters with the specific label within the cluster-provider
	Name  string `json:"name"`  // select the specific cluster within the cluster-provider
	Scope string `json:"scope"` // indicates label or name should be used to select the cluster from cluster-provider
}

// MetaData holds the data
type MetaData struct {
	Name        string `json:"name"` // name of the cluster provider intent
	Description string `json:"description"`
	UserData1   string `json:"userData1"`
	UserData2   string `json:"userData2"`
}

// StateKey
type StateKey struct {
	Cert            string `json:"clusterProviderCert"`
	ClusterProvider string `json:"clusterProvider"`
	AppName         string `json:"appName"`
}
