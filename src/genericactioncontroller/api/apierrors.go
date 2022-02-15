package api

import (
	"net/http"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
)

var apiErrors = []apierror.APIError{
	{ID: "Customization already exists", Message: "customization already exists", Status: http.StatusConflict},
	{ID: "Customization not found", Message: "customization not found", Status: http.StatusNotFound},
	{ID: "GenericK8sIntent already exists", Message: "genericK8sIntent already exists", Status: http.StatusConflict},
	{ID: "GenericK8sIntent not found", Message: "genericK8sIntent not found", Status: http.StatusNotFound},
	{ID: "Resource already exists", Message: "resource already exists", Status: http.StatusConflict},
	{ID: "Resource not found", Message: "resource not found", Status: http.StatusNotFound},
}

// HandleErrors exposes the generic action controller API errors
func HandleErrors(params map[string]string, err error, mod interface{}) apierror.APIError {
	return apierror.HandleErrors(params, err, mod, apiErrors)
}
