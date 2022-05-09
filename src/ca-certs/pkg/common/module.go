// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package common

// DbInfo holds the MongoDB collection and attributes info
type DbInfo struct {
	storeName string // name of the mongodb collection to use for client documents
	tagState  string // attribute key name for the state of a client document
}
