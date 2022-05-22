// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package enrollment

import (
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
)

type EnrollmentContext struct {
	AppContext    appcontext.AppContext
	AppHandle     interface{}
	CaCert        module.Cert // CA
	ContextID     string
	ResOrder      []string
	ClientName    string
	ClusterGroups []module.ClusterGroup
	ClusterGroup  module.ClusterGroup
	IssuerHandle  interface{}
	Cluster       string
}
