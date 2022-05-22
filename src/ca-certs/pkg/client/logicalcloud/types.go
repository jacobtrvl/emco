// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package logicalcloud

type EnrollmentKey struct {
	Cert       string `json:"cert"`
	Project    string `json:"project"`
	Enrollment string `json:"enrollment"`
}

// StateKey
type DistributionKey struct {
	Cert         string `json:"cert"`
	Project      string `json:"project"`
	Distribution string `json:"distribution"`
}
