// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package api

import (
	"net/http"

	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/logicalcloud"
)

type logicalCloudCertDistributionHandler struct {
	manager logicalcloud.CertDistributionManager
}

// handleInstantiate
func (h *logicalCloudCertDistributionHandler) handleInstantiate(w http.ResponseWriter, r *http.Request) {

}

// handleStatus
func (h *logicalCloudCertDistributionHandler) handleStatus(w http.ResponseWriter, r *http.Request) {

}

// handleTerminate
func (h *logicalCloudCertDistributionHandler) handleTerminate(w http.ResponseWriter, r *http.Request) {

}

// handleUpdate
func (h *logicalCloudCertDistributionHandler) handleUpdate(w http.ResponseWriter, r *http.Request) {

}
