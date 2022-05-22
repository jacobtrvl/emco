// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package certmanagerissuer

import (
	"fmt"
)

// newClusterIssuer returns an instance of the ClusterIssuer
func newClusterIssuer() *ClusterIssuer {
	return &ClusterIssuer{
		APIVersion: "cert-manager.io/v1",
		Kind:       "ClusterIssuer"}
}

// ClusterIssuerName retun the ClusterIssuer name
func ClusterIssuerName(contextID, cert, clusterProvider, cluster string) string {
	return fmt.Sprintf("%s-%s-%s-%s-%s", contextID, cert, clusterProvider, cluster, "issuer")
}

// ResourceName returns the ClusterIssuer resource name, used by the rsync
func (i *ClusterIssuer) ResourceName() string {
	return fmt.Sprintf("%s+%s", i.MetaData.Name, "clusterissuer")
}

// CreateClusterIssuer retun the cert-manager ClusterIssuer object
func CreateClusterIssuer(name, namespace, secret string) *ClusterIssuer {
	i := newClusterIssuer()

	i.MetaData.Name = name

	if len(namespace) > 0 {
		i.MetaData.Namespace = namespace
	}

	i.Spec.CA.SecretName = secret

	return i
}
