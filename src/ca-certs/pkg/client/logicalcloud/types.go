// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package logicalcloud

import "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/types"

// EnrollmentKey
type EnrollmentKey struct {
	Cert       string `json:"caCert"`
	Project    string `json:"project"`
	Enrollment string `json:"caCertEnrollment"`
}

// DistributionKey
type DistributionKey struct {
	Cert         string `json:"caCert"`
	Project      string `json:"project"`
	Distribution string `json:"caCertDistribution"`
}

// CaCertLogicalCloud
type CaCertLogicalCloud struct {
	MetaData types.Metadata         `json:"metadata"`
	Spec     CaCertLogicalCloudSpec `json:"spec"`
}

// CaCertLogicalCloudSpec
type CaCertLogicalCloudSpec struct {
	LogicalCloud string `json:"logicalCloud"` // name of the logicalCloud
}
