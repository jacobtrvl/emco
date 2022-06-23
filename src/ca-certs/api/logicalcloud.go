// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/client/logicalcloud"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

type lcHandler struct {
	manager logicalcloud.CaCertLogicalCloudManager
}

// handleLogicalCloudCreate handles the route for creating a new logical cloud
func (h *lcHandler) handleLogicalCloudCreate(w http.ResponseWriter, r *http.Request) {
	h.createOrUpdateLogicalCloud(w, r)
}

// handleLogicalCloudDelete handles the route for deleting a logical cloud
func (h *lcHandler) handleLogicalCloudDelete(w http.ResponseWriter, r *http.Request) {
	// get the route variables
	vars := _lcVars(mux.Vars(r))
	if err := h.manager.DeleteLogicalCloud(vars.logicalCloud, vars.cert, vars.project); err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleLogicalCloudGet handles the route for retrieving a logical cloud
func (h *lcHandler) handleLogicalCloudGet(w http.ResponseWriter, r *http.Request) {
	var (
		logicalClouds interface{}
		err           error
	)

	// get the route variables
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

// handleLogicalCloudUpdate handles the route for updating a logical cloud
func (h *lcHandler) handleLogicalCloudUpdate(w http.ResponseWriter, r *http.Request) {
	h.createOrUpdateLogicalCloud(w, r)
}

// createOrUpdateLogicalCloud create/update the logical cloud based on the request method
func (h *lcHandler) createOrUpdateLogicalCloud(w http.ResponseWriter, r *http.Request) {
	var logicalCloud logicalcloud.CaCertLogicalCloud
	if code, err := validateRequestBody(r.Body, &logicalCloud, LogicalCloudSchemaJson); err != nil {
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
		if logicalCloud.MetaData.Name != vars.logicalCloud {
			logutils.Error("The logical-cloud name is not matching with the name in the request",
				logutils.Fields{"LogicalCloud": logicalCloud,
					"Name": vars.logicalCloud})
			http.Error(w, "the logical-cloud name is not matching with the name in the request",
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
		// logical-cloud does have a current representation and that representation is successfully modified
		code = http.StatusOK
	}

	sendResponse(w, clr, code)
}

// validateLogicalCloudData validate the logical cloud payload for the required values
func validateLogicalCloudData(lc logicalcloud.CaCertLogicalCloud) error {
	var err []string
	if len(lc.MetaData.Name) == 0 {
		logutils.Error("LogicalCloud name may not be empty",
			logutils.Fields{})
		err = append(err, "logicalCloud name may not be empty")
	}

	if len(lc.Spec.LogicalCloud) == 0 {
		logutils.Error("LogicalCloud may not be empty",
			logutils.Fields{})
		err = append(err, "logicalCloud may not be empty")
	}

	if len(err) > 0 {
		return errors.New(strings.Join(err, "\n"))
	}

	return nil
}
