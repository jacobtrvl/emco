package api

import (
	"net/http"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
)

var apiErrors = []apierror.APIError{
	{ID: "Resource already exists", Message: "Resource already exists", Status: http.StatusConflict},
	{ID: "Resource not found", Message: "Resource not found", Status: http.StatusNotFound},
	{ID: "Resource File Content not found", Message: "Resource File Content not found", Status: http.StatusNotFound},
}
