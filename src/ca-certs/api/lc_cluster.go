// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/client/logicalcloud"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

type lcClusterHandler struct {
	manager logicalcloud.ClusterGroupManager
}

// handleClusterCreate handles the route for creating a new cluster group
func (h *lcClusterHandler) handleClusterCreate(w http.ResponseWriter, r *http.Request) {
	h.createOrUpdateCluster(w, r)
}

// handleClusterDelete handles the route for deleting a cluster group
func (h *lcClusterHandler) handleClusterDelete(w http.ResponseWriter, r *http.Request) {
	// get the route variables
	vars := _lcVars(mux.Vars(r))
	if err := h.manager.DeleteClusterGroup(vars.cluster, vars.logicalCloud, vars.cert, vars.project); err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleClusterGet handles the route for retrieving a cluster group
func (h *lcClusterHandler) handleClusterGet(w http.ResponseWriter, r *http.Request) {
	var (
		clusters interface{}
		err      error
	)

	// get the route variables
	vars := _lcVars(mux.Vars(r))
	if len(vars.cluster) == 0 {
		clusters, err = h.manager.GetAllClusterGroups(vars.logicalCloud, vars.cert, vars.project)
	} else {
		clusters, err = h.manager.GetClusterGroup(vars.cluster, vars.logicalCloud, vars.cert, vars.project)
	}

	if err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	sendResponse(w, clusters, http.StatusOK)
}

// handleClusterUpdate handles the route for updating a cluster group
func (h *lcClusterHandler) handleClusterUpdate(w http.ResponseWriter, r *http.Request) {
	h.createOrUpdateCluster(w, r)
}

// createOrUpdateCluster create/update the cluster group based on the request method
func (h *lcClusterHandler) createOrUpdateCluster(w http.ResponseWriter, r *http.Request) {
	var cluster module.ClusterGroup
	if code, err := validateRequestBody(r.Body, &cluster, ClusterSchemaJson); err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	// get the route variables
	vars := _lcVars(mux.Vars(r))

	methodPost := false
	if r.Method == http.MethodPost {
		methodPost = true
	}

	if !methodPost {
		// name in the URL should match the name in the body
		if cluster.MetaData.Name != vars.cluster {
			logutils.Error("The cluster group name is not matching with the name in the request",
				logutils.Fields{"ClusterGroup": cluster})
			http.Error(w, "the cluster group name is not matching with the name in the request",
				http.StatusBadRequest)
			return
		}
	}

	clr, clusterExists, err := h.manager.CreateClusterGroup(cluster, vars.logicalCloud, vars.cert, vars.project, methodPost)
	if err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, cluster, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	code := http.StatusCreated
	if clusterExists {
		// cluster group does have a current representation and that representation is successfully modified
		code = http.StatusOK
	}

	sendResponse(w, clr, code)
}
