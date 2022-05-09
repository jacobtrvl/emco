// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/clusterprovider"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
)

// type cpeVars struct {
// 	cert,
// 	clusterProvider string
// }

type clusterProviderCertEnrollmentHandler struct {
	manager clusterprovider.CertEnrollmentManager
}

// handleInstantiate
func (h *clusterProviderCertEnrollmentHandler) handleInstantiate(w http.ResponseWriter, r *http.Request) {
	vars := _cpVars(mux.Vars(r))
	if err := h.manager.Instantiate(vars.cert, vars.clusterProvider); err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// handleStatus
func (h *clusterProviderCertEnrollmentHandler) handleStatus(w http.ResponseWriter, r *http.Request) {
	vars := _cpVars(mux.Vars(r))
	stat, err := h.manager.Status(vars.cert, vars.clusterProvider)
	if err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	sendResponse(w, stat, http.StatusOK)
}

// handleTerminate
func (h *clusterProviderCertEnrollmentHandler) handleTerminate(w http.ResponseWriter, r *http.Request) {
	vars := _cpVars(mux.Vars(r))
	if err := h.manager.Terminate(vars.cert, vars.clusterProvider); err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// handleUpdate
func (h *clusterProviderCertEnrollmentHandler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	vars := _cpVars(mux.Vars(r))
	if err := h.manager.Update(vars.cert, vars.clusterProvider); err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusAccepted)

}
