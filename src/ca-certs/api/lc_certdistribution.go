// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/client/logicalcloud"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/common/emcoerror"
)

// lcCertDistributionHandler implements the logicalCloud caCert distribution handler functions
type lcCertDistributionHandler struct {
	manager logicalcloud.CaCertDistributionManager
}

// handleInstantiate handles the route for instantiating the caCert distribution
func (h *lcCertDistributionHandler) handleInstantiate(w http.ResponseWriter, r *http.Request) {
	// get the route variables
	vars := _lcVars(mux.Vars(r))
	if err := h.manager.Instantiate(vars.cert, vars.project); err != nil {
		apiErr := emcoerror.HandleAPIError(err)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// handleStatus handles the route for getting the status of the caCert distribution
func (h *lcCertDistributionHandler) handleStatus(w http.ResponseWriter, r *http.Request) {
	// get the route variables
	vars := _lcVars(mux.Vars(r))

	qParams, err := _statusQueryParams(r)
	if err != nil {
		apiErr := emcoerror.HandleAPIError(err)
		http.Error(w, apiErr.Message, apiErr.Status)
	}

	stat, err := h.manager.Status(vars.cert, vars.project,
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
func (h *lcCertDistributionHandler) handleTerminate(w http.ResponseWriter, r *http.Request) {
	// get the route variables
	vars := _lcVars(mux.Vars(r))
	if err := h.manager.Terminate(vars.cert, vars.project); err != nil {
		apiErr := emcoerror.HandleAPIError(err)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// handleUpdate handles the route for updating the caCert distribution
func (h *lcCertDistributionHandler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	// get the route variables
	vars := _lcVars(mux.Vars(r))
	if err := h.manager.Update(vars.cert, vars.project); err != nil {
		apiErr := emcoerror.HandleAPIError(err)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
