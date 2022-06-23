// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package distribution

import (
	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/service/istioservice"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/service/knccservice"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	v1 "k8s.io/api/core/v1"
)

// DistributionContext
type DistributionContext struct {
	AppContext          appcontext.AppContext
	AppHandle           interface{}
	CaCert              module.CaCert // CA
	Project             string
	ContextID           string
	ResOrder            []string
	EnrollmentContextID string
	CertificateRequests []cmv1.CertificateRequest
	Resources           DistributionResource
	ClusterGroups       []module.ClusterGroup
	ClusterGroup        module.ClusterGroup
	Namespace           string
	ClientName          string
	Cluster             string
	ClusterHandle       interface{}
}

// DistributionResource
type DistributionResource struct {
	ClusterIssuer map[string]*cmv1.ClusterIssuer
	ProxyConfig   map[string]*istioservice.ProxyConfig
	Secret        map[string]*v1.Secret
	KnccConfig    map[string]*knccservice.Config
}
