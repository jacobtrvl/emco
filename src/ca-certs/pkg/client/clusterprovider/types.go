// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package clusterprovider

// EnrollmentKey
type EnrollmentKey struct {
	Cert            string `json:"caCert"`
	ClusterProvider string `json:"clusterProvider"`
	Enrollment      string `json:"caCertEnrollment"`
}

// DistributionKey
type DistributionKey struct {
	Cert            string `json:"caCert"`
	ClusterProvider string `json:"clusterProvider"`
	Distribution    string `json:"caCertDistribution"`
}
