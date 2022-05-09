// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package api

import (
	"net/http"

	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/logicalcloud"
)

type logicalCloudCertEnrollmentHandler struct {
	manager logicalcloud.CertEnrollmentManager
}

// handleInstantiate
func (h *logicalCloudCertEnrollmentHandler) handleInstantiate(w http.ResponseWriter, r *http.Request) {
}

// handleStatus
func (h *logicalCloudCertEnrollmentHandler) handleStatus(w http.ResponseWriter, r *http.Request) {

}

// handleTerminate
func (h *logicalCloudCertEnrollmentHandler) handleTerminate(w http.ResponseWriter, r *http.Request) {

}

// handleUpdate
func (h *logicalCloudCertEnrollmentHandler) handleUpdate(w http.ResponseWriter, r *http.Request) {

}
