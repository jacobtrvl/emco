// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package fluxv2

import ()

// StartClusterWatcher watches for CR
// Same as K8s
func (c *Fluxv2Provider) StartClusterWatcher() error {
	return nil
}

// ApplyStatusCR applies status CR
func (p *Fluxv2Provider) ApplyStatusCR(content []byte) error {

	return nil

}

// DeleteStatusCR deletes status CR
func (p *Fluxv2Provider) DeleteStatusCR(content []byte) error {

	return nil
}
