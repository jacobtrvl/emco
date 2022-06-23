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

type cpCertHandler struct {
	manager clusterprovider.CaCertManager
}

// handleCertificateCreates handles the route for creating a new CA cert intent
func (h *cpCertHandler) handleCertificateCreate(w http.ResponseWriter, r *http.Request) {
	h.createOrUpdateCertificate(w, r)
}

// handleCertificateDelete handles the route for deleting a CA cert intent
func (h *cpCertHandler) handleCertificateDelete(w http.ResponseWriter, r *http.Request) {
	// get the route variables
	vars := _cpVars(mux.Vars(r))
	if err := h.manager.DeleteCert(vars.cert, vars.clusterProvider); err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleCertificateGets handles the route for retrieving a CA cert intent
func (h *cpCertHandler) handleCertificateGet(w http.ResponseWriter, r *http.Request) {
	var (
		certs interface{}
		err   error
	)

	// get the route variables
	vars := _cpVars(mux.Vars(r))
	if len(vars.cert) == 0 {
		certs, err = h.manager.GetAllCert(vars.clusterProvider)
	} else {
		certs, err = h.manager.GetCert(vars.cert, vars.clusterProvider)
	}

	if err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	sendResponse(w, certs, http.StatusOK)
}

// handleCertificateUpdate handles the route for updating a CA cert intent
func (h *cpCertHandler) handleCertificateUpdate(w http.ResponseWriter, r *http.Request) {
	h.createOrUpdateCertificate(w, r)
}

// createOrUpdateCertificate create/update the CA cert intent based on the request method
func (h *cpCertHandler) createOrUpdateCertificate(w http.ResponseWriter, r *http.Request) {
	var cert module.CaCert

	// validate the request body before storing it in the database
	if code, err := validateRequestBody(r.Body, &cert, CertificateSchemaJson); err != nil {
		http.Error(w, err.Error(), code)
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
		if cert.MetaData.Name != vars.cert {
			logutils.Error("The cert name is not matching with the name in the request",
				logutils.Fields{"Cert": cert,
					"CertName": vars.cert})
			http.Error(w, "the cert name is not matching with the name in the request",
				http.StatusBadRequest)
			return
		}
	}

	crt, certExists, err := h.manager.CreateCert(cert, vars.clusterProvider, methodPost)
	if err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, cert, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	code := http.StatusCreated
	if certExists {
		// cert does have a current representation and that representation is successfully modified
		code = http.StatusOK
	}

	sendResponse(w, crt, code)
}
