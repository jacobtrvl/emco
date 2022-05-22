// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package distribution

import (
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certissuer/certmanagerissuer"
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
	CertificateRequests []certmanagerissuer.CertificateRequest // Holds the retrived CSR(s) from the issuing cluster
	Resources           DistributionResource
	ClusterGroups       []module.ClusterGroup
	ClusterGroup        module.ClusterGroup
	Namespace           string
	ClientName          string
	Cluster             string
	ClusterHandle       interface{}
}

type DistributionResource struct {
	ClusterIssuer []certmanagerissuer.ClusterIssuer
}
