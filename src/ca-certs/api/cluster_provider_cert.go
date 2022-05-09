// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certificate"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/clusterprovider"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

type clusterProviderCertHandler struct {
	manager clusterprovider.CertManager
}

// handleCertificateCreates
func (h *clusterProviderCertHandler) handleCertificateCreate(w http.ResponseWriter, r *http.Request) {
	h.createOrUpdateCertificate(w, r)
}

// handleCertificateDelete
func (h *clusterProviderCertHandler) handleCertificateDelete(w http.ResponseWriter, r *http.Request) {
	vars := _cpVars(mux.Vars(r))
	if err := h.manager.DeleteCert(vars.cert, vars.clusterProvider); err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleCertificateGets
func (h *clusterProviderCertHandler) handleCertificateGet(w http.ResponseWriter, r *http.Request) {
	var (
		certs interface{}
		err   error
	)

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

// handleCertificateUpdate
func (h *clusterProviderCertHandler) handleCertificateUpdate(w http.ResponseWriter, r *http.Request) {
	h.createOrUpdateCertificate(w, r)
}

// createOrUpdateCertificate create/update the CA Cert based on the request method
func (h *clusterProviderCertHandler) createOrUpdateCertificate(w http.ResponseWriter, r *http.Request) {
	var cert certificate.Cert
	if code, err := validateRequestBody(r.Body, &cert, CertificateSchemaJson); err != nil {
		http.Error(w, err.Error(), code)
		return
	}

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
			http.Error(w, "the intent name is not matching with the name in the request",
				http.StatusBadRequest)
			return
		}
	}

	crt, crtExists, err := h.manager.CreateCert(cert, vars.clusterProvider, methodPost)
	if err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, cert, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	code := http.StatusCreated
	if crtExists {
		// certificate does have a current representation and that representation is successfully modified
		code = http.StatusOK
	}

	sendResponse(w, crt, code)
}
