// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package api

import (
	"net/http"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
)

var apiErrors = []apierror.APIError{
	{ID: "Certificate already exists", Message: "certificate already exists", Status: http.StatusConflict},
	{ID: "Certificate not found", Message: "certificate not found", Status: http.StatusNotFound},
	{ID: "ClusterGroup already exists", Message: "cluster group already exists", Status: http.StatusConflict},
	{ID: "ClusterGroup not found", Message: "cluster group not found", Status: http.StatusNotFound},
	{ID: "LogicalCloud already exists", Message: "logical cloud already exists", Status: http.StatusConflict},
	{ID: "LogicalCloud not found", Message: "logical cloud not found", Status: http.StatusNotFound},
}

// HandleErrors exposes the ca cert controller API errors
func HandleErrors(params map[string]string, err error, mod interface{}) apierror.APIError {
	return apierror.HandleErrors(params, err, mod, apiErrors)
}
