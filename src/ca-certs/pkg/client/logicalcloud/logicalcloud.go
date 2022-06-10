// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package logicalcloud

import (
	"reflect"

	"github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

// LogicalCloudManager
type LogicalCloudManager interface {
	CreateLogicalCloud(logicalCloud LogicalCloud, cert, project string, failIfExists bool) (LogicalCloud, bool, error)
	DeleteLogicalCloud(logicalCloud, cert, project string) error
	GetAllLogicalClouds(cert, project string) ([]LogicalCloud, error)
	GetLogicalCloud(logicalCloud, cert, project string) (LogicalCloud, error)
}

// LogicalCloudKey
type LogicalCloudKey struct {
	Cert         string `json:"caCertLc"`
	LogicalCloud string `json:"caCertLogicalCloud"`
	Project      string `json:"project"`
}

// LogicalCloudClient
type LogicalCloudClient struct {
	dbInfo db.DbInfo
}

// NewLogicalCloudClient
func NewLogicalCloudClient() *LogicalCloudClient {
	return &LogicalCloudClient{
		dbInfo: db.DbInfo{
			StoreName: "resources",
			TagMeta:   "data"}}
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

	if err := db.DBconn.Insert(c.dbInfo.StoreName, key, nil, c.dbInfo.TagMeta, logicalCloud); err != nil {
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

	return db.DBconn.Remove(c.dbInfo.StoreName, key)
}

// GetAllLogicalClouds
func (c *LogicalCloudClient) GetAllLogicalClouds(cert, project string) ([]LogicalCloud, error) {
	key := LogicalCloudKey{
		Cert:    cert,
		Project: project}

	values, err := db.DBconn.Find(c.dbInfo.StoreName, key, c.dbInfo.TagMeta)
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

	value, err := db.DBconn.Find(c.dbInfo.StoreName, key, c.dbInfo.TagMeta)
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

// // Convert the key to string to preserve the underlying structure
// func (k LogicalCloudKey) String() string {
// 	out, err := json.Marshal(k)
// 	if err != nil {
// 		return ""
// 	}
// 	return string(out)
// }
