// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/clusterprovider"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
)

type clusterProviderCertDistributionHandler struct {
	manager clusterprovider.CertDistributionManager
}

// handleInstantiate
func (h *clusterProviderCertDistributionHandler) handleInstantiate(w http.ResponseWriter, r *http.Request) {
	vars := _cpVars(mux.Vars(r))
	if err := h.manager.Instantiate(vars.cert, vars.clusterProvider); err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// handleStatus
func (h *clusterProviderCertDistributionHandler) handleStatus(w http.ResponseWriter, r *http.Request) {
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
func (h *clusterProviderCertDistributionHandler) handleTerminate(w http.ResponseWriter, r *http.Request) {
	vars := _cpVars(mux.Vars(r))
	if err := h.manager.Terminate(vars.cert, vars.clusterProvider); err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// handleUpdate
func (h *clusterProviderCertDistributionHandler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	vars := _cpVars(mux.Vars(r))
	if err := h.manager.Update(vars.cert, vars.clusterProvider); err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
