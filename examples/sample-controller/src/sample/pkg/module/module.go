// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

// Package module implements all the business logic.
// It is a middleware/facade between the handler and the database.
package module

// Client combines different clients into a single type.
// Every handler is associated with a client.
// The handler then uses its associated client to perform the requested operation.
// You can have different clients based on the requirement and its implementation.
// In this example, we only have one client.
type Client struct {
	SampleIntent *SampleIntentClient
	// Add other required clients here.
	// ref: https://gitlab.com/project-emco/core/emco-base/-/blob/main/src/ncm/pkg/module/module.go

}

// NewClient returns a new client instance
func NewClient() *Client {
	c := &Client{}
	c.SampleIntent = NewIntentClient()
	// If you have multiple clients, init them here.
	// ref: https://gitlab.com/project-emco/core/emco-base/-/blob/main/src/ncm/pkg/module/module.go
	return c
}

// DBInfo represents the mongoDB properties
type DBInfo struct {
	collection string // name of the mongodb collection to use for client documents
	tag        string // attribute key name for the json data of a client document
}
