// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package action_test

import (
	"encoding/base64"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	// "gopkg.in/yaml.v2"
	"gitlab.com/project-emco/core/emco-base/src/genericactioncontroller/internal/action"
	v1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/yaml"
)

var (
	o action.UpdateOptions
)

var _ = Describe("Test Strategic Merge Patch",
	func() {
		BeforeEach(func() {
			o = action.UpdateOptions{}
		})
		Context("add a new container to an existing Deployment container list", func() {
			It("returns the Deployment with the new container list", func() {
				o.Customization = mockCustomization("deploy-web-customization")
				o.CustomizationContent = mockDeploymentCustomizationContent()
				o.Resource.Spec.ResourceGVK.APIVersion = "apps/v1"
				o.Resource.Spec.ResourceGVK.Kind = "Deployment"
				data, err := base64.StdEncoding.DecodeString(mockDeploymentResourceContent().Content)
				Expect(err).To(BeNil())
				do := v1.Deployment{}
				err = yaml.Unmarshal(data, &do)
				Expect(err).To(BeNil())
				Expect(len(do.Spec.Template.Spec.Containers)).To(Equal(1))
				result, err := o.MergePatch(data)
				Expect(err).To(BeNil())
				dn := v1.Deployment{}
				err = yaml.Unmarshal(result, &dn)
				Expect(err).To(BeNil())
				Expect(len(dn.Spec.Template.Spec.Containers)).To(Equal(2))
			})
		})
		Context("add hostAliases to an existing statefulset", func() {
			It("returns the statefulset with the hostAliases details ", func() {
				o.Customization = mockCustomization("sts-etcd-customization")
				o.CustomizationContent = mockStatefulSetCustomizationContent()
				o.Resource.Spec.ResourceGVK.APIVersion = "apps/v1"
				o.Resource.Spec.ResourceGVK.Kind = "StatefulSet"
				data, err := base64.StdEncoding.DecodeString(mockStatefulSetResourceContent().Content)
				fmt.Println(string(data))
				Expect(err).To(BeNil())
				sso := v1.StatefulSet{}
				err = yaml.Unmarshal(data, &sso)
				Expect(err).To(BeNil())
				Expect(len(sso.Spec.Template.Spec.HostAliases)).To(Equal(0))
				result, err := o.MergePatch(data)
				Expect(err).To(BeNil())
				// j, err := yaml.YAMLToJSON(result)
				// Expect(err).To(BeNil())
				ssn := v1.StatefulSet{}
				err = yaml.Unmarshal(result, &ssn)
				Expect(err).To(BeNil())
				Expect(len(ssn.Spec.Template.Spec.HostAliases)).To(Equal(1))
				Expect(len(ssn.Spec.Template.Spec.HostAliases[0].Hostnames)).To(Equal(1))
				Expect(ssn.Spec.Template.Spec.HostAliases[0].Hostnames[0]).To(Equal("host1"))
			})
		})
	},
)
