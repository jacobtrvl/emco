// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package azurearcv2

import (
	"context"
)

// StartClusterWatcher watches for CR
// Same as K8s
func (p *AzureArcV2Provider) StartClusterWatcher() error {
	return p.gitProvider.StartClusterWatcher()
}

// ApplyStatusCR applies status CR
func (p *AzureArcV2Provider) ApplyStatusCR(name string, content []byte) error {
	ref, err := p.gitProvider.Apply(name, nil, content)
	if err != nil {
		return err
	}
	return p.gitProvider.Commit(context.Background(), ref)
}

// DeleteStatusCR deletes status CR
func (p *AzureArcV2Provider) DeleteStatusCR(name string, content []byte) error {

	ref, err := p.gitProvider.Delete(name, nil, content)
	if err != nil {
		return err
	}
	return p.gitProvider.Commit(context.Background(), ref)
}
