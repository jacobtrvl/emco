// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package connector

import (
	"fmt"

	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/plugins/k8s"
	. "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/types"
)

// Connection is for a cluster
type Provider struct {
	cid string
}

func NewProvider(id interface{}) Provider {
	return Provider{
		cid: fmt.Sprintf("%v", id),
	}
}

func (p *Provider) GetClientProviders(app, cluster, level, namespace string) (ClientProvider, error) {

	// Only supported type at this time
	var providerType string = "k8s"

	switch providerType {
	case "k8s":
		cl, err := k8s.NewK8sProvider(p.cid, app, cluster, level, namespace)
		if err != nil {
			return nil, err
		}
		return cl, nil
		//Add other types like Azure Arc, Fluxcd etc here
	}
	return nil, nil
}
