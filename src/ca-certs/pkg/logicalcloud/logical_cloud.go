// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package logicalcloud

import (
	"reflect"

	"github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

type LogicalCloud struct {
	MetaData MetaData         `json:"metadata"`
	Spec     LogicalCloudSpec `json:"spec"`
}

type LogicalCloudSpec struct {
	Name string `json:"name"` // name of the logical-cloud
}

// MetaData holds the data
type MetaData struct {
	Name        string `json:"name"` // name of the Logical-Cloud intent
	Description string `json:"description"`
	UserData1   string `json:"userData1"`
	UserData2   string `json:"userData2"`
}

// DbInfo holds the MongoDB collection and attributes info
type DbInfo struct {
	storeName string // name of the mongodb collection to use for client documents
	tagMeta   string // attribute key name for the json data of a client document
}

// LogicalCloudClient
type LogicalCloudClient struct {
	db DbInfo
}

// LogicalCloudKey
type LogicalCloudKey struct {
	Cert         string `json:"logicalCloudCert"`
	LogicalCloud string `json:"caLogicalCloud"`
	Project      string `json:"project"`
}

type LogicalCloudManager interface {
	CreateLogicalCloud(logicalCloud LogicalCloud, cert, project string, failIfExists bool) (LogicalCloud, bool, error)
	DeleteLogicalCloud(logicalCloud, cert, project string) error
	GetAllLogicalClouds(cert, project string) ([]LogicalCloud, error)
	GetLogicalCloud(logicalCloud, cert, project string) (LogicalCloud, error)
}

// CreateLogicalCloud
func (c *LogicalCloudClient) CreateLogicalCloud(logicalCloud LogicalCloud, cert, project string, failIfExists bool) (LogicalCloud, bool, error) {
	lcExists := false
	key := LogicalCloudKey{
		Cert:         cert,
		Project:      project,
		LogicalCloud: logicalCloud.MetaData.Name}

	if lc, err := c.GetLogicalCloud(logicalCloud.MetaData.Name, cert, project); err == nil &&
		!reflect.DeepEqual(lc, LogicalCloud{}) {
		lcExists = true
	}

	if lcExists &&
		failIfExists {
		return LogicalCloud{}, lcExists, errors.New("LogicalCloud already exists")
	}

	if err := db.DBconn.Insert(c.db.storeName, key, nil, c.db.tagMeta, logicalCloud); err != nil {
		return LogicalCloud{}, lcExists, err
	}

	return logicalCloud, lcExists, nil
}

// DeleteLogicalCloud
func (c *LogicalCloudClient) DeleteLogicalCloud(logicalCloud, cert, project string) error {
	key := LogicalCloudKey{
		Cert:         cert,
		LogicalCloud: logicalCloud,
		Project:      project}

	return db.DBconn.Remove(c.db.storeName, key)
}

// GetAllLogicalClouds
func (c *LogicalCloudClient) GetAllLogicalClouds(cert, project string) ([]LogicalCloud, error) {
	key := LogicalCloudKey{
		Project: project}

	values, err := db.DBconn.Find(c.db.storeName, key, c.db.tagMeta)
	if err != nil {
		return []LogicalCloud{}, err
	}

	var logicalClouds []LogicalCloud
	for _, value := range values {
		lc := LogicalCloud{}
		if err = db.DBconn.Unmarshal(value, &lc); err != nil {
			return []LogicalCloud{}, err
		}
		logicalClouds = append(logicalClouds, lc)
	}

	return logicalClouds, nil
}

// GetLogicalCloud
func (c *LogicalCloudClient) GetLogicalCloud(logicalCloud, cert, project string) (LogicalCloud, error) {
	key := LogicalCloudKey{
		Cert:         cert,
		LogicalCloud: logicalCloud,
		Project:      project}

	value, err := db.DBconn.Find(c.db.storeName, key, c.db.tagMeta)
	if err != nil {
		return LogicalCloud{}, err
	}

	if len(value) == 0 {
		return LogicalCloud{}, errors.New("LogicalCloud not found")
	}

	if value != nil {
		lc := LogicalCloud{}
		if err = db.DBconn.Unmarshal(value[0], &lc); err != nil {
			return LogicalCloud{}, err
		}
		return lc, nil
	}

	return LogicalCloud{}, errors.New("Unknown Error")
}

// NewLogicalCloudClient
func NewLogicalCloudClient() *LogicalCloudClient {
	return &LogicalCloudClient{
		db: DbInfo{
			storeName: "resources",
			tagMeta:   "data"}}
}

// // Convert the key to string to preserve the underlying structure
// func (k LogicalCloudKey) String() string {
// 	out, err := json.Marshal(k)
// 	if err != nil {
// 		return ""
// 	}
// 	return string(out)
// }
