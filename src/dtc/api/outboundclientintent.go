// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"gitlab.com/project-emco/core/emco-base/src/dtc/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/validation"
	orcmod "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
	pkgerrors "github.com/pkg/errors"
)

var outClientIntJSONFile string = "json-schemas/outbound-client.json"

type outboundclientintentHandler struct {
	client module.OutboundClientIntentManager
}

// Check for valid format of input parameters
func validateOutboundClientIntentInputs(oci module.OutboundClientIntent) error {
	// validate metadata
	err := module.IsValidMetadata(oci.Metadata)
	if err != nil {
		return pkgerrors.Wrap(err, "Invalid outbound client intent metadata")
	}
	return nil
}

func (h outboundclientintentHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	var oci module.OutboundClientIntent
	vars := mux.Vars(r)
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deploymentIntentGroupName := vars["deploymentIntentGroup"]
	trafficIntentGroupName := vars["trafficGroupIntent"]
	// check if the deploymentIntentGrpName exists
	_, err := orcmod.NewDeploymentIntentGroupClient().GetDeploymentIntentGroup(deploymentIntentGroupName, project, compositeApp, compositeAppVersion)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}
	err = json.NewDecoder(r.Body).Decode(&oci)
	switch {
	case err == io.EOF:
		log.Error(":: Empty outbound client POST body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding outbound client POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	err, httpError := validation.ValidateJsonSchemaData(outClientIntJSONFile, oci)
	if err != nil {
		log.Error(":: Error validating outbound client POST data ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	// Name is required.
	if oci.Metadata.Name == "" {
		log.Error(":: Missing name in outbound client POST request ::", log.Fields{})
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	err = validateOutboundClientIntentInputs(oci)
	if err != nil {
		log.Error(":: Invalid create outbound client body inputs ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateClientOutboundIntent(oci, project, compositeApp, compositeAppVersion, deploymentIntentGroupName, trafficIntentGroupName, false)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, oci, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding create outbound client response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}
func (h outboundclientintentHandler) putHandler(w http.ResponseWriter, r *http.Request) {
	var oci module.OutboundClientIntent
	vars := mux.Vars(r)
	name := vars["outboundClientIntent"]
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deploymentIntentGroupName := vars["deploymentIntentGroup"]
	trafficIntentGroupName := vars["trafficGroupIntent"]
	// check if the deploymentIntentGrpName exists
	_, err := orcmod.NewDeploymentIntentGroupClient().GetDeploymentIntentGroup(deploymentIntentGroupName, project, compositeApp, compositeAppVersion)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	err = json.NewDecoder(r.Body).Decode(&oci)

	switch {
	case err == io.EOF:
		log.Error(":: Empty outbound client PUT body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding outbound client PUT body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Name is required.
	if oci.Metadata.Name == "" {
		log.Error(":: Missing name in outbound client PUT request ::", log.Fields{})
		http.Error(w, "Missing name in PUT request", http.StatusBadRequest)
		return
	}

	// Name in URL should match name in body
	if oci.Metadata.Name != name {
		log.Error(":: Mismatched name in outbound client PUT request ::", log.Fields{})
		http.Error(w, "Mismatched name in PUT request", http.StatusBadRequest)
		return
	}

	err = validateOutboundClientIntentInputs(oci)
	if err != nil {
		log.Error(":: Invalid outbound client PUT inputs ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateClientOutboundIntent(oci, project, compositeApp, compositeAppVersion, deploymentIntentGroupName, trafficIntentGroupName, true)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, oci, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding outbound client update response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h outboundclientintentHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["outboundClientIntent"]
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deploymentIntentGroupName := vars["deploymentIntentGroup"]
	trafficIntentGroupName := vars["trafficGroupIntent"]

	var ret interface{}
	var err error

	if len(name) == 0 {
		ret, err = h.client.GetClientOutboundIntents(project, compositeApp, compositeAppVersion, deploymentIntentGroupName, trafficIntentGroupName)
	} else {
		ret, err = h.client.GetClientOutboundIntent(name, project, compositeApp, compositeAppVersion, deploymentIntentGroupName, trafficIntentGroupName)
	}

	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding get outbound client response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func (h outboundclientintentHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["outboundClientIntent"]
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deploymentIntentGroupName := vars["deploymentIntentGroup"]
	trafficIntentGroupName := vars["trafficGroupIntent"]

	err := h.client.DeleteClientOutboundIntent(name, project, compositeApp, compositeAppVersion, deploymentIntentGroupName, trafficIntentGroupName)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
