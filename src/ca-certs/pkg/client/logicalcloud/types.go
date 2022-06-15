// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package logicalcloud

import "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/types"

type EnrollmentKey struct {
	Cert       string `json:"caCert"`
	Project    string `json:"project"`
	Enrollment string `json:"caCertEnrollment"`
}

type DistributionKey struct {
	Cert         string `json:"caCert"`
	Project      string `json:"project"`
	Distribution string `json:"caCertDistribution"`
}

// LogicalCloud
type CaCertLogicalCloud struct {
	MetaData types.Metadata         `json:"metadata"`
	Spec     CaCertLogicalCloudSpec `json:"spec"`
}

// LogicalCloudSpec
type CaCertLogicalCloudSpec struct {
	LogicalCloud string `json:"logicalCloud"` // name of the logicalCloud
}
