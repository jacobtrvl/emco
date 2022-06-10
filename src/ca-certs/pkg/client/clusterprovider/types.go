// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package clusterprovider

type EnrollmentKey struct {
	Cert            string `json:"caCert"`
	ClusterProvider string `json:"clusterProvider"`
	Enrollment      string `json:"caCertEnrollment"`
}

type DistributionKey struct {
	Cert            string `json:"caCert"`
	ClusterProvider string `json:"clusterProvider"`
	Distribution    string `json:"caCertDistribution"`
}
