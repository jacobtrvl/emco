// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package distribution

import (
	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
)

type DistributionContext struct {
	AppContext          appcontext.AppContext
	AppHandle           interface{}
	CaCert              module.Cert // CA
	Project             string
	ContextID           string
	ResOrder            []string
	EnrollmentContextID string
	CertificateRequests []cmv1.CertificateRequest // Holds the retrived CSR(s) from the issuing cluster
	Resources           DistributionResource
	ClusterGroups       []module.ClusterGroup
	ClusterGroup        module.ClusterGroup
	Namespace           string
	ClientName          string
	Cluster             string
	ClusterHandle       interface{}
}

type DistributionResource struct {
	ClusterIssuer []cmv1.ClusterIssuer
}
