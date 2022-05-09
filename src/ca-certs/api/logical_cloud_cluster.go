// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/clusterprovider"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/logicalcloud"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

type logicalCloudClusterHandler struct {
	manager logicalcloud.ClusterManager
}

// handleClusterCreate
func (h *logicalCloudClusterHandler) handleClusterCreate(w http.ResponseWriter, r *http.Request) {
	h.createOrUpdateCluster(w, r)
}

// handleClusterDelete
func (h *logicalCloudClusterHandler) handleClusterDelete(w http.ResponseWriter, r *http.Request) {
	vars := _lcVars(mux.Vars(r))
	if err := h.manager.DeleteClusterGroup(vars.cluster, vars.cert, vars.logicalCloud, vars.project); err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleClusterGet
func (h *logicalCloudClusterHandler) handleClusterGet(w http.ResponseWriter, r *http.Request) {
	var (
		clusters interface{}
		err      error
	)

	vars := _lcVars(mux.Vars(r))
	if len(vars.cluster) == 0 {
		clusters, err = h.manager.GetAllClusterGroups(vars.cert, vars.logicalCloud, vars.project)
	} else {
		clusters, err = h.manager.GetClusterGroup(vars.cluster, vars.cert, vars.logicalCloud, vars.project)
	}

	if err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	sendResponse(w, clusters, http.StatusOK)
}

// handleClusterUpdate
func (h *logicalCloudClusterHandler) handleClusterUpdate(w http.ResponseWriter, r *http.Request) {
	h.createOrUpdateCluster(w, r)
}

// createOrUpdateCluster create/update the CA Cert based on the request method
func (h *logicalCloudClusterHandler) createOrUpdateCluster(w http.ResponseWriter, r *http.Request) {
	var cluster clusterprovider.ClusterGroup
	if code, err := validateRequestBody(r.Body, &cluster, ClusterSchemaJson); err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	vars := _lcVars(mux.Vars(r))

	methodPost := false
	if r.Method == http.MethodPost {
		methodPost = true
	}

	if !methodPost {
		// name in the URL should match the name in the body
		if cluster.MetaData.Name != vars.cluster {
			logutils.Error("The cluster name is not matching with the name in the request",
				logutils.Fields{"Cluster": cluster,
					"ClusterName": vars.cluster})
			http.Error(w, "the intent name is not matching with the name in the request",
				http.StatusBadRequest)
			return
		}
	}

	clr, clrExists, err := h.manager.CreateClusterGroup(cluster, vars.cert, vars.logicalCloud, vars.project, methodPost)
	if err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, cluster, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	code := http.StatusCreated
	if clrExists {
		// cluster does have a current representation and that representation is successfully modified
		code = http.StatusOK
	}

	sendResponse(w, clr, code)
}
