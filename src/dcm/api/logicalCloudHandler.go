// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
	dcm "gitlab.com/project-emco/core/emco-base/src/dcm/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/validation"
	orch "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
)

var logicalCloudJSONValidation string = "json-schemas/logical-cloud.json"

// logicalCloudHandler is used to store backend implementations objects
type logicalCloudHandler struct {
	client               dcm.LogicalCloudManager
	clusterClient        dcm.ClusterManager
	quotaClient          dcm.QuotaManager
	userPermissionClient dcm.UserPermissionManager
}

// CreateHandler handles the creation of a logical cloud
func (h logicalCloudHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project"]
	var v dcm.LogicalCloud
	var err error

	err = json.NewDecoder(r.Body).Decode(&v)
	switch {
	case err == io.EOF:
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case err != nil:
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err, httpError := validation.ValidateJsonSchemaData(logicalCloudJSONValidation, v)
	if err != nil {
		log.Error(":: Invalid Logical Cloud JSON ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	// Logical Cloud Name is required.
	if v.MetaData.LogicalCloudName == "" {
		msg := "Missing name in POST request"
		log.Error(msg, log.Fields{})
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	// Validate that the specified Project exists
	// before associating a Logical Cloud with it
	p := orch.NewProjectClient()
	_, err = p.GetProject(project)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, v, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	ret, err := h.client.Create(project, v)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, v, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// getAllHandler handles GET operations over logical clouds
// Returns a list of Logical Clouds
func (h logicalCloudHandler) getAllHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project"]
	var ret interface{}
	var err error

	ret, err = h.client.GetAll(project)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// getHandler handles GET operations on a particular name
// Returns a Logical Cloud
func (h logicalCloudHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project"]
	name := vars["logicalCloud"]
	var ret interface{}
	var err error

	ret, err = h.client.Get(project, name)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// updateHandler handles Update operations on a particular logical cloud
func (h logicalCloudHandler) updateHandler(w http.ResponseWriter, r *http.Request) {
	var v dcm.LogicalCloud
	vars := mux.Vars(r)
	project := vars["project"]
	name := vars["logicalCloud"]
	var err error

	// Get logical cloud
	_, err = h.client.Get(project, name)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	ret, err := h.client.UpdateInstantiation(project, name, v)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, v, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		http.Error(w, err.Error(),
			http.StatusInternalServerError)
		return
	}
}

// putHandler handles PUT API update operations on a particular logical cloud
func (h logicalCloudHandler) putHandler(w http.ResponseWriter, r *http.Request) {
	var v dcm.LogicalCloud
	vars := mux.Vars(r)
	project := vars["project"]
	name := vars["logicalCloud"]
	var err error

	// Get logical cloud
	_, err = h.client.Get(project, name)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	err = json.NewDecoder(r.Body).Decode(&v)
	switch {
	case err == io.EOF:
		log.Error(err.Error(), log.Fields{})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err, httpError := validation.ValidateJsonSchemaData(logicalCloudJSONValidation, v)
	if err != nil {
		log.Error(":: Invalid Logical Cloud JSON ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	if v.MetaData.LogicalCloudName == "" {
		log.Error("API: Missing name in PUT request", log.Fields{})
		http.Error(w, "Missing name in PUT request", http.StatusBadRequest)
		return
	}

	ret, err := h.client.UpdateLogicalCloud(project, name, v)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, v, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		http.Error(w, err.Error(),
			http.StatusInternalServerError)
		return
	}
}

// deleteHandler handles Delete operations on a particular logical cloud
func (h logicalCloudHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project"]
	name := vars["logicalCloud"]
	var err error

	// call to Delete also takes care of checking whether Logical Cloud exists
	err = h.client.Delete(project, name)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// instantiateHandler handles instantiateing a particular logical cloud
func (h logicalCloudHandler) instantiateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project"]
	name := vars["logicalCloud"]
	var err error

	// Get logical cloud
	lc, err := h.client.Get(project, name)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	// Get Clusters
	clusters, err := h.clusterClient.GetAllClusters(project, name)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	// Get Quotas
	quotas, err := h.quotaClient.GetAllQuotas(project, name)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	userPermissions, err := h.userPermissionClient.GetAllUserPerms(project, name)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	// Instantiate the Logical Cloud
	err = dcm.Instantiate(project, lc, clusters, quotas, userPermissions)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// terminateHandler handles terminating a particular logical cloud
func (h logicalCloudHandler) terminateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project"]
	name := vars["logicalCloud"]
	var err error

	// Get logical cloud
	lc, err := h.client.Get(project, name)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	// Check if any DIGs are associated to this logical cloud before attempting any termination
	digClient := orch.NewDeploymentIntentGroupClient()
	digs, _ := digClient.GetAllDeploymentIntentGroups(project, "", "")

	// filter DIGs for given Logical Cloud
	for _, dig := range digs {
		if dig.Spec.LogicalCloud == name {
			// found at least 1 logical cloud:
			log.Error("Found at least one DIG associated with Logical Cloud attempted to be terminated", log.Fields{})
			http.Error(w, "Found at least one DIG associated with Logical Cloud attempted to be terminated", http.StatusBadRequest)
			return
		}
	}

	// Get Clusters
	clusters, err := h.clusterClient.GetAllClusters(project, name)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	// Get Quotas
	quotas, err := h.quotaClient.GetAllQuotas(project, name)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	// Terminate the Logical Cloud
	err = dcm.Terminate(project, lc, clusters, quotas)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// stopHandler handles aborting the pending instantiation or termination of a logical cloud
func (h logicalCloudHandler) stopHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project"]
	name := vars["logicalCloud"]
	var err error

	// Get logical cloud
	lc, err := h.client.Get(project, name)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	// Attempt to stop instantiating/terminating
	err = dcm.Stop(project, lc)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h logicalCloudHandler) statusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	p := vars["project"]
	lc := vars["logicalCloud"]
	var status interface{}
	var err error
	// variables prepended with "q" are for queries, and with "f" for filters

	qParams, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// /status?instance = "<instance id>"
	var qInstance string
	if o, found := qParams["instance"]; found {
		qInstance = o[0]
		if qInstance == "" {
			log.Error("Invalid query instance", log.Fields{})
			http.Error(w, "Invalid query instance", http.StatusBadRequest)
			return
		}
	} else {
		qInstance = "" // default instance value
	}

	// /status?status = "ready" or "deployed"
	var qType string
	if t, found := qParams["status"]; found {
		qType = t[0]
		if qType != "ready" && qType != "deployed" {
			log.Error("Invalid query status", log.Fields{})
			http.Error(w, "Invalid query status", http.StatusBadRequest)
			return
		}
	} else {
		// default ?status="ready" if not specified by the API request
		qType = "ready"
	}

	// /status?output = "summary" or "all" or "detail"
	var qOutput string
	if o, found := qParams["output"]; found {
		qOutput = o[0]
		if qOutput != "summary" && qOutput != "all" && qOutput != "detail" {
			log.Error("Invalid query output", log.Fields{})
			http.Error(w, "Invalid query output", http.StatusBadRequest)
			return
		}
	} else {
		qOutput = "all" // default output format
	}

	// /status?clusters
	var qClusters bool
	if _, found := qParams["clusters"]; found {
		qClusters = true
	} else {
		qClusters = false
	}

	// /status?resources
	var qResources bool
	if _, found := qParams["resources"]; found {
		qResources = true
	} else {
		qResources = false
	}

	// /status?cluster =
	var fClusters []string
	if c, found := qParams["cluster"]; found {
		fClusters = c
		for _, cl := range fClusters {
			parts := strings.Split(cl, "+")
			if len(parts) != 2 {
				log.Error("Invalid cluster query", log.Fields{})
				http.Error(w, "Invalid cluster query", http.StatusBadRequest)
				return
			}
			for _, p := range parts {
				errs := validation.IsValidName(p)
				if len(errs) > 0 {
					log.Error(errs[len(errs)-1], log.Fields{}) // log the most recently appended msg
					http.Error(w, "Invalid cluster query", http.StatusBadRequest)
					return
				}
			}
		}
	} else {
		fClusters = make([]string, 0)
	}

	// /status?resource =
	var fResources []string
	if r, found := qParams["resource"]; found {
		fResources = r
		for _, res := range fResources {
			errs := validation.IsValidName(res)
			if len(errs) > 0 {
				log.Error(errs[len(errs)-1], log.Fields{}) // log the most recently appended msg
				http.Error(w, "Invalid resources query", http.StatusBadRequest)
				return
			}
		}
	} else {
		fResources = make([]string, 0)
	}

	// Different backend status functions are invoked based on which query parameters have been provided.
	// The query parameters will be handled with the following precedence to determine which status query is
	// invoked: i. clusters, ii. resources, iii. default.
	// Supplied query parameters which are not appropriate for the select function call are simply ignored.
	if qClusters {
		status, err = h.client.StatusClusters(p, lc, qInstance)
	} else if qResources {
		status, err = h.client.StatusResources(p, lc, qInstance, qType, fClusters)
	} else {
		status, err = h.client.Status(p, lc, qInstance, qType, qOutput, fClusters, fResources)
	}
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		log.Error(apiErr.Message, log.Fields{})
		http.Error(w, apiErr.Message, apiErr.Status)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(status)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		log.Error(apiErr.Message, log.Fields{})
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}
}
