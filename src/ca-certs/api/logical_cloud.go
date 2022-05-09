// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/logicalcloud"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

type logicalCloudHandler struct {
	manager logicalcloud.LogicalCloudManager
}

// handleLogicalCloudCreate
func (h *logicalCloudHandler) handleLogicalCloudCreate(w http.ResponseWriter, r *http.Request) {
	h.createOrUpdateLogicalCloud(w, r)
}

// handleLogicalCloudDelete
func (h *logicalCloudHandler) handleLogicalCloudDelete(w http.ResponseWriter, r *http.Request) {
	vars := _lcVars(mux.Vars(r))
	if err := h.manager.DeleteLogicalCloud(vars.logicalCloud, vars.cert, vars.project); err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleLogicalCloudGet
func (h *logicalCloudHandler) handleLogicalCloudGet(w http.ResponseWriter, r *http.Request) {
	var (
		logicalClouds interface{}
		err           error
	)

	vars := _lcVars(mux.Vars(r))
	if len(vars.logicalCloud) == 0 {
		logicalClouds, err = h.manager.GetAllLogicalClouds(vars.cert, vars.project)
	} else {
		logicalClouds, err = h.manager.GetLogicalCloud(vars.logicalCloud, vars.cert, vars.project)
	}

	if err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	sendResponse(w, logicalClouds, http.StatusOK)
}

// handleLogicalCloudUpdate
func (h *logicalCloudHandler) handleLogicalCloudUpdate(w http.ResponseWriter, r *http.Request) {
	h.createOrUpdateLogicalCloud(w, r)
}

// createOrUpdateLogicalCloud create/update the CA Cert based on the request method
func (h *logicalCloudHandler) createOrUpdateLogicalCloud(w http.ResponseWriter, r *http.Request) {
	var logicalCloud logicalcloud.LogicalCloud
	if code, err := validateRequestBody(r.Body, &logicalCloud, LogicalCloudSchemaJson); err != nil {
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
		if logicalCloud.MetaData.Name != vars.logicalCloud {
			logutils.Error("The logical-cloud name is not matching with the name in the request",
				logutils.Fields{"LogicalCloud": logicalCloud,
					"Name": vars.logicalCloud})
			http.Error(w, "the intent name is not matching with the name in the request",
				http.StatusBadRequest)
			return
		}
	}

	clr, clrExists, err := h.manager.CreateLogicalCloud(logicalCloud, vars.cert, vars.project, methodPost)
	if err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, logicalCloud, apiErrors)
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
