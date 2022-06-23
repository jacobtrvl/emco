// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package enrollment

import (
	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
)

// EnrollmentContext
type EnrollmentContext struct {
	AppContext    appcontext.AppContext
	AppHandle     interface{}
	CaCert        module.CaCert // CA
	ContextID     string
	ResOrder      []string
	ClientName    string
	ClusterGroups []module.ClusterGroup
	ClusterGroup  module.ClusterGroup
	IssuerHandle  interface{}
	Cluster       string
	Resources     EnrollmentResource
}

// EnrollmentResource
type EnrollmentResource struct {
	CertificateRequest map[string]*cmv1.CertificateRequest
}
