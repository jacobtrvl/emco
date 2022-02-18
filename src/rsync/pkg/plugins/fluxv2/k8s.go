// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package fluxv2

import (
	"strings"

	resources "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/plugins/resources"

	pkgerrors "github.com/pkg/errors"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"

	clm "gitlab.com/project-emco/core/emco-base/src/clm/pkg/cluster"
)

// Connection is for a cluster
type Fluxv2Provider struct {
	gitProvider resources.GitProvider
}

func NewFluxv2Provider(cid, app, cluster, level, namespace string) (*Fluxv2Provider, error) {

	result := strings.SplitN(cluster, "+", 2)
	cc := clm.NewClusterClient()
	c, err := cc.GetCluster(result[0], result[1])
	if err != nil {
		return nil, err
	}
	if c.Spec.Props.GitOpsType != "fluxcd" {
		log.Error("Invalid GitOps type:", log.Fields{})
		return nil, pkgerrors.Errorf("Invalid GitOps type: " + c.Spec.Props.GitOpsType)
	}

	gitProvider, err := resources.NewGitProvider(cid, app, cluster, level, namespace)
	if err != nil {
		return nil, err
	}
	p := Fluxv2Provider{
		gitProvider: *gitProvider,
	}
	return &p, nil
}

func (p *Fluxv2Provider) CleanClientProvider() error {
	return nil
}
