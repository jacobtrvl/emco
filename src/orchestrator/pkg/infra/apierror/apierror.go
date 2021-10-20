// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package apierror

import (
	"net/http"
	"strings"

	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

type APIError struct {
	ID      string
	Message string
	Status  int
}

var dbErrors = []APIError{
	{ID: "db Find error", Message: "Error finding referencing resources", Status: http.StatusInternalServerError},
	{ID: "db Remove error", Message: "Error removing referencing resources", Status: http.StatusInternalServerError},
	{ID: "db Remove resource not found", Message: "The requested resource not found", Status: http.StatusNotFound},
	{ID: "db Remove parent child constraint", Message: "Cannot delete parent without deleting child references first", Status: http.StatusConflict},
	{ID: "db Remove referential constraint", Message: "Cannot delete without deleting or updating referencing resources first", Status: http.StatusConflict},
	{ID: "db Insert error", Message: "Error adding or updating referencing resources", Status: http.StatusInternalServerError},
	{ID: "db Insert parent resource not found", Message: "Cannot perform requested operation. Parent resource not found", Status: http.StatusConflict},
	{ID: "db Insert referential schema missing", Message: "Cannot perform requested operation. The requested resource is not defined in the referential schema", Status: http.StatusConflict},
}

// HandleErrors handles api resources add/update/create errors
// Returns APIError with the ID, message and the http status based on the error
func HandleErrors(params map[string]string, err error, mod interface{}, apiErr []APIError) APIError {
	log.Error("Error :: ", log.Fields{"Parameters": params, "Error": err, "Module": mod})

	// db errors
	for _, e := range dbErrors {
		if strings.Contains(err.Error(), e.ID) {
			return e
		}
	}

	// api specific errors
	for _, e := range apiErr {
		if strings.Contains(err.Error(), e.ID) {
			return e
		}
	}

	// Default
	return APIError{ID: "Internal server error", Message: "The server encountered an internal error and was unable to complete your request.", Status: http.StatusInternalServerError}
}

// HandleLogicalCloudErrors handles logical cloud errors
// Returns APIError with the ID, message and the http status based on the error
func HandleLogicalCloudErrors(params map[string]string, err error, lcErrors []APIError) APIError {
	for _, e := range lcErrors {
		if strings.Contains(err.Error(), e.ID) {
			log.Error("Logical cloud error :: ", log.Fields{"Parameters": params, "Error": err})
			return e
		}
	}
	return APIError{}
}
