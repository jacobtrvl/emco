// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package action_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gitlab.com/project-emco/core/emco-base/src/genericactioncontroller/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/types"
)

func TestApi(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Action Suite")
}

// mockCustomization
func mockCustomization(name string) module.Customization {
	return module.Customization{
		Metadata: types.Metadata{
			Name:        name,
			Description: "test customization",
			UserData1:   "some user data 1",
			UserData2:   "some user data 2",
		},
		Spec: module.CustomizationSpec{
			ClusterSpecific: "true",
			ClusterInfo: module.ClusterInfo{
				Scope:           "label",
				ClusterProvider: "provider-1",
				ClusterName:     "cluster-1",
				ClusterLabel:    "edge-cluster-1",
				Mode:            "allow",
			},
			PatchType: "merge",
		},
	}
}

// mockDeploymentCustomizationContent
func mockDeploymentCustomizationContent() module.CustomizationContent {
	return module.CustomizationContent{
		Content: []module.Content{
			{
				FileName: "container-patch.yaml",
				Content:  "c3BlYzoNCiAgdGVtcGxhdGU6DQogICAgc3BlYzoNCiAgICAgIGNvbnRhaW5lcnM6DQogICAgICAtIG5hbWU6IHJlZGlzLWN0cg0KICAgICAgICBpbWFnZTogcmVkaXMNCg==",
				KeyName:  "container-patch.yaml"},
		},
	}
}

// mockDeploymentResourceContent
func mockDeploymentResourceContent() module.ResourceContent {
	return module.ResourceContent{
		Content: "YXBpVmVyc2lvbjogYXBwcy92MQ0Ka2luZDogRGVwbG95bWVudA0KbWV0YWRhdGE6DQogIG5hbWU6IGRlcGxveS13ZWINCnNwZWM6DQog" +
			"IHNlbGVjdG9yOg0KICAgIG1hdGNoTGFiZWxzOg0KICAgICAgYXBwOiBuZ2lueA0KICB0ZW1wbGF0ZToNCiAgICBtZXRhZGF0YToNCiAgICAgIGxhYmVsczo" +
			"NCiAgICAgICAgYXBwOiBuZ2lueA0KICAgIHNwZWM6DQogICAgICBjb250YWluZXJzOg0KICAgICAgLSBuYW1lOiBuZ2lueC1jdHINCiAgICAgICAgaW1hZ2U6IG5naW54DQo="}
}

// mockStatefulSetCustomizationContent
func mockStatefulSetCustomizationContent() module.CustomizationContent {
	return module.CustomizationContent{
		Content: []module.Content{
			{
				FileName: "hostalias-patch.yaml",
				Content:  "c3BlYzoKICB0ZW1wbGF0ZToKICAgIHNwZWM6CiAgICAgIGhvc3RBbGlhc2VzOgogICAgICAtIGlwOiAxLjIuMy40CiAgICAgICAgaG9zdG5hbWVzOiAKICAgICAgICAtICJob3N0MSI=",
				KeyName:  "hostalias-patch.yaml"},
		},
	}
}

// mockStatefulSetResourceContent
func mockStatefulSetResourceContent() module.ResourceContent {
	return module.ResourceContent{
		Content: "YXBpVmVyc2lvbjogYXBwcy92MQpraW5kOiBTdGF0ZWZ1bFNldAptZXRhZGF0YToKICBuYW1lOiBldGNkCg=="}
}
