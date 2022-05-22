// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package client_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

var (
	Mockdb *db.NewMockDB
	V      vars
)

type vars struct {
	Project,
	CompositeApp,
	Version,
	DeploymentIntentGroup,
	Intent,
	Resource string
}

func init() {
	Mockdb = &db.NewMockDB{}
	V = newVars()
	db.DBconn = Mockdb
}

func TestApi(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Api Suite")
}

func newVars() vars {
	return vars{
		Project:               "test-project",
		CompositeApp:          "test-compositeapp",
		Version:               "v1",
		DeploymentIntentGroup: "test-dig",
		Intent:                "test-gki",
		Resource:              "test-resource",
	}
}

func validateError(err error, message string) {
	if len(message) == 0 {
		Expect(err).NotTo(HaveOccurred())
		Expect(err).To(BeNil())
		return
	}
	Expect(err.Error()).To(ContainSubstring(message))
}
