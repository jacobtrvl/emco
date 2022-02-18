// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package fluxv2

import (
	"context"
)

// func convertToCommitFile(ref interface{}) []gitprovider.CommitFile {
// 	var exists bool
// 	switch ref.(type) {
// 	case []gitprovider.CommitFile:
// 		exists = true
// 	default:
// 		exists = false
// 	}
// 	var rf []gitprovider.CommitFile
// 	// Create rf is doesn't exist
// 	if !exists {
// 		rf = []gitprovider.CommitFile{}
// 	} else {
// 		rf = ref.([]gitprovider.CommitFile)
// 	}
// 	return rf
// }

// Creates a new resource if the not already existing
func (p *Fluxv2Provider) Create(name string, ref interface{}, content []byte) (interface{}, error) {

	res, err := p.gitProvider.Create(name, ref, content)
	return res, err
}

// Apply resource to the cluster
func (p *Fluxv2Provider) Apply(name string, ref interface{}, content []byte) (interface{}, error) {

	res, err := p.gitProvider.Apply(name, ref, content)
	return res, err
}

// Delete resource from the cluster
func (p *Fluxv2Provider) Delete(name string, ref interface{}, content []byte) (interface{}, error) {

	res, err := p.gitProvider.Delete(name, ref, content)
	return res, err
}

// Get resource from the cluster
func (p *Fluxv2Provider) Get(name string, gvkRes []byte) ([]byte, error) {

	return []byte{}, nil
}

// Commit resources to the cluster
func (p *Fluxv2Provider) Commit(ctx context.Context, ref interface{}) error {
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
func (p *Fluxv2Provider) IsReachable() error {
	return nil
}
