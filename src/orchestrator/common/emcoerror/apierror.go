// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package emcoerror

import "net/http"

// APIError defines the API error message and status code
type APIError struct {
	Message string
	Status  int
}

// getStatusCode returns the HTTP status code based on the error type
func getStatusCode(errorType ErrorType) int {
	switch errorType {
	case BadRequest:
		return http.StatusBadRequest
	case NotFound:
		return http.StatusNotFound
	case Conflict:
		return http.StatusConflict
	case Unknown:
		return http.StatusInternalServerError
	}

	return http.StatusInternalServerError
}

// HandleAPIError returns the HTTP error message and status code
func HandleAPIError(err interface{}) APIError {
	switch e := err.(type) {
	case *Error:
		return APIError{Message: e.Error(), Status: getStatusCode(e.Type)}
	case error:
		return APIError{Message: e.Error(), Status: http.StatusInternalServerError}
	}

	return APIError{Message: InternalServerErrorMessage, Status: http.StatusInternalServerError}
}
