// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package azurearc

import (
	"fmt"
	"strings"

	pkgerrors "github.com/pkg/errors"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"

	clm "gitlab.com/project-emco/core/emco-base/src/clm/pkg/cluster"
	resources "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/plugins/resources"
)

// Connection is for a cluster
type AzureArcProvider struct {
	gitProvider      resources.GitProvider
	clientID         string
	tenantID         string
	clientSecret     string
	subscriptionID   string
	arcCluster       string
	arcResourceGroup string
}

func NewAzureArcProvider(cid, app, cluster, level, namespace string) (*AzureArcProvider, error) {

	result := strings.SplitN(cluster, "+", 2)
	cc := clm.NewClusterClient()
	c, err := cc.GetCluster(result[0], result[1])
	if err != nil {
		return nil, err
	}
	if c.Spec.Props.GitOpsType != "azureArc" {
		log.Error("Invalid GitOps type:", log.Fields{})
		return nil, pkgerrors.Errorf("Invalid GitOps type: " + c.Spec.Props.GitOpsType)
	}

	gitProvider, err := resources.NewGitProvider(cid, app, cluster, level, namespace)
	if err != nil {
		return nil, err
	}

	resObject, err := cc.GetClusterSyncObjects(result[0], c.Spec.Props.GitOpsResourceObject)
	if err != nil {
		log.Error("Invalid resObject :", log.Fields{"resObj": c.Spec.Props.GitOpsResourceObject})
		return nil, pkgerrors.Errorf("Invalid resObject: " + c.Spec.Props.GitOpsResourceObject)
	}
	kvRes := resObject.Spec.Kv

	var clientID, tenantID, clientSecret, subscriptionID, arcCluster, arcResourceGroup string

	for _, kvpair := range kvRes {
		log.Info("kvpair", log.Fields{"kvpair": kvpair})
		v, ok := kvpair["clientID"]
		if ok {
			clientID = fmt.Sprintf("%v", v)
			continue
		}
		v, ok = kvpair["tenantID"]
		if ok {
			tenantID = fmt.Sprintf("%v", v)
			continue
		}
		v, ok = kvpair["clientSecret"]
		if ok {
			clientSecret = fmt.Sprintf("%v", v)
			continue
		}
		v, ok = kvpair["subscriptionID"]
		if ok {
			subscriptionID = fmt.Sprintf("%v", v)
			continue
		}
		v, ok = kvpair["arcCluster"]
		if ok {
			arcCluster = fmt.Sprintf("%v", v)
			continue
		}
		v, ok = kvpair["arcResourceGroup"]
		if ok {
			arcResourceGroup = fmt.Sprintf("%v", v)
			continue
		}
	}
	if len(clientID) <= 0 || len(tenantID) <= 0 || len(clientSecret) <= 0 || len(subscriptionID) <= 0 || len(arcCluster) <= 0 || len(arcResourceGroup) <= 0 {
		log.Error("Missing information for Github", log.Fields{"clientID": clientID, "tenantID": tenantID, "clientSecret": clientSecret, "subscriptionID": subscriptionID,
			"arcCluster": arcCluster, "arcResourceGroup": arcResourceGroup})
		return nil, pkgerrors.Errorf("Missing Information for Azure Arc")
	}
	p := AzureArcProvider{

		gitProvider:      *gitProvider,
		clientID:         clientID,
		tenantID:         tenantID,
		clientSecret:     clientSecret,
		subscriptionID:   subscriptionID,
		arcCluster:       arcCluster,
		arcResourceGroup: arcResourceGroup,
	}
	return &p, nil
}

func (p *AzureArcProvider) CleanClientProvider() error {
	return nil
}
