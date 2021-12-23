// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package utils

import (
	pkgerrors "github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

// DecodeYAMLData reads a string to extract the Kubernetes object definition
func DecodeYAMLData(data string, into runtime.Object) (runtime.Object, error) {
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode([]byte(data), nil, into)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Deserialize YAML error")
	}

	return obj, nil
}
