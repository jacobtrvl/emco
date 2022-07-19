// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/client/clusterprovider"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/common/emcoerror"
)

// cpCertDistributionHandler implements the clusterProvider caCert distribution handler functions
type cpCertDistributionHandler struct {
	manager clusterprovider.CaCertDistributionManager
}

// handleInstantiate handles the route for instantiating the caCert distribution
func (h *cpCertDistributionHandler) handleInstantiate(w http.ResponseWriter, r *http.Request) {
	// get the route variables
	vars := _cpVars(mux.Vars(r))
	if err := h.manager.Instantiate(vars.cert, vars.clusterProvider); err != nil {
		apiErr := emcoerror.HandleAPIError(err)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// handleStatus handles the route for getting the status of the caCert distribution
func (h *cpCertDistributionHandler) handleStatus(w http.ResponseWriter, r *http.Request) {
	// get the route variables
	vars := _cpVars(mux.Vars(r))

	qParams, err := _statusQueryParams(r)
	if err != nil {
		apiErr := emcoerror.HandleAPIError(err)
		http.Error(w, apiErr.Message, apiErr.Status)
	}

	stat, err := h.manager.Status(vars.cert, vars.clusterProvider,
		qParams.qInstance,
		qParams.qType,
		qParams.qOutput,
		qParams.fApps,
		qParams.fClusters,
		qParams.fResources)
	if err != nil {
		apiErr := emcoerror.HandleAPIError(err)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	sendResponse(w, stat, http.StatusOK)
}

// handleTerminate handles the route for terminating the caCert distribution
func (h *cpCertDistributionHandler) handleTerminate(w http.ResponseWriter, r *http.Request) {
	// get the route variables
	vars := _cpVars(mux.Vars(r))
	if err := h.manager.Terminate(vars.cert, vars.clusterProvider); err != nil {
		apiErr := emcoerror.HandleAPIError(err)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// handleUpdate handles the route for updating the caCert distribution
func (h *cpCertDistributionHandler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	// get the route variables
	vars := _cpVars(mux.Vars(r))
	if err := h.manager.Update(vars.cert, vars.clusterProvider); err != nil {
		apiErr := emcoerror.HandleAPIError(err)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
