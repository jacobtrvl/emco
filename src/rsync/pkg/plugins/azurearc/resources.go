// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package azurearc

import (
	"context"
)

// Creates a new resource if the not already existing
func (p *AzureArcProvider) Create(name string, ref interface{}, content []byte) (interface{}, error) {

	res, err := p.gitProvider.Create(name, ref, content)
	return res, err
}

// Apply resource to the cluster
func (p *AzureArcProvider) Apply(name string, ref interface{}, content []byte) (interface{}, error) {

	res, err := p.gitProvider.Apply(name, ref, content)
	return res, err
}

// Delete resource from the cluster
func (p *AzureArcProvider) Delete(name string, ref interface{}, content []byte) (interface{}, error) {

	res, err := p.gitProvider.Delete(name, ref, content)
	return res, err
}

// Get resource from the cluster
func (p *AzureArcProvider) Get(name string, gvkRes []byte) ([]byte, error) {

	return []byte{}, nil
}

// Commit resources to the cluster
func (p *AzureArcProvider) Commit(ctx context.Context, ref interface{}) error {
	// var exists bool
	// switch ref.(type) {
	// case []gitprovider.CommitFile:
	// 	exists = true
	// default:
	// 	exists = false

	// }
	// // Check for rf
	// if !exists {
	// 	log.Error("Commit: No ref found", log.Fields{})
	// 	return nil
	// }
	err := p.gitProvider.Commit(ctx, ref)
	return err
}

// IsReachable cluster reachablity test
func (p *AzureArcProvider) IsReachable() error {
	return nil
}
