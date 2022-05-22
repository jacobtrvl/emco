// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package certmanagerissuer

import (
	"fmt"
)

// newSecret returns an instance of the Secret
func newSecret() *Secret {
	return &Secret{
		APIVersion: "v1",
		Kind:       "Secret",
		Type:       "Opaque",
		Data:       map[string]string{}}
}

// SecretName retun the Secret name
func SecretName(contextID, cert, clusterProvider, cluster string) string {
	return fmt.Sprintf("%s-%s-%s-%s-%s", contextID, cert, clusterProvider, cluster, "ca")
}

// ResourceName returns the Secret resource name, used by the rsync
func (s *Secret) ResourceName() string {
	return fmt.Sprintf("%s+%s", s.MetaData.Name, "secret")
}

// CreateSecret retun the Secret object
func CreateSecret(name, namespace string, data map[string]string) *Secret {
	s := newSecret()

	s.MetaData.Name = name

	if len(namespace) > 0 {
		s.MetaData.Namespace = namespace
	}

	for key, val := range data {
		s.Data[key] = val
	}

	return s
}
