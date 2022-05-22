// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/client/clusterprovider"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

type cpClusterHandler struct {
	manager clusterprovider.ClusterManager
}

// handleClusterCreate handles the route for creating a new cluster group
func (h *cpClusterHandler) handleClusterCreate(w http.ResponseWriter, r *http.Request) {
	h.createOrUpdateCluster(w, r)
}

// handleClusterDeletes handles the route for deleting a cluster group
func (h *cpClusterHandler) handleClusterDelete(w http.ResponseWriter, r *http.Request) {
	// get the route variables
	vars := _cpVars(mux.Vars(r))
	if err := h.manager.DeleteClusterGroup(vars.cert, vars.cluster, vars.clusterProvider); err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleClusterGet handles the route for retrieving a cluster group
func (h *cpClusterHandler) handleClusterGet(w http.ResponseWriter, r *http.Request) {
	var (
		clusters interface{}
		err      error
	)

	// get the route variables
	vars := _cpVars(mux.Vars(r))
	if len(vars.cluster) == 0 {
		clusters, err = h.manager.GetAllClusterGroups(vars.cert, vars.clusterProvider)
	} else {
		clusters, err = h.manager.GetClusterGroup(vars.cert, vars.cluster, vars.clusterProvider)
	}

	if err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	sendResponse(w, clusters, http.StatusOK)
}

// handleClusterUpdate handles the route for updating a cluster group
func (h *cpClusterHandler) handleClusterUpdate(w http.ResponseWriter, r *http.Request) {
	h.createOrUpdateCluster(w, r)
}

// createOrUpdateCluster create/update the cluster group based on the request method
func (h *cpClusterHandler) createOrUpdateCluster(w http.ResponseWriter, r *http.Request) {
	var cluster module.ClusterGroup

	// validate the request body before storing it in the database
	if code, err := validateRequestBody(r.Body, &cluster, ClusterSchemaJson); err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	if len(cluster.Spec.Label) == 0 && len(cluster.Spec.Name) == 0 {
		logutils.Error("The cluster label or name should be provided",
			logutils.Fields{"Cluster": cluster})
		http.Error(w, "the cluster label or name should be provided",
			http.StatusBadRequest)
		return

	}

	// get the route variables
	vars := _cpVars(mux.Vars(r))

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

	clr, clusterExists, err := h.manager.CreateClusterGroup(cluster, vars.cert, vars.clusterProvider, methodPost)
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
