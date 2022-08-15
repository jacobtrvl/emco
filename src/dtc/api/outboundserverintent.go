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

var outServerIntJSONFile string = "json-schemas/outbound-server.json"

type outboundserverintentHandler struct {
	client module.OutboundServerIntentManager
}

// Check for valid format of input parameters
func validateOutboundServerIntentInputs(osi module.OutboundServerIntent) error {
	// validate metadata
	err := module.IsValidMetadata(osi.Metadata)
	if err != nil {
		return pkgerrors.Wrap(err, "Invalid outbound server intent metadata")
	}
	return nil
}

func (h outboundserverintentHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	var osi module.OutboundServerIntent
	vars := mux.Vars(r)
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deploymentIntentGroupName := vars["deploymentIntentGroup"]
	trafficIntentGroupName := vars["trafficGroupIntent"]
	outboundIntentName := vars["outboundClientIntent"]
	// check if the deploymentIntentGrpName exists
	_, err := orcmod.NewDeploymentIntentGroupClient().GetDeploymentIntentGroup(deploymentIntentGroupName, project, compositeApp, compositeAppVersion)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}
	err = json.NewDecoder(r.Body).Decode(&osi)

	switch {
	case err == io.EOF:
		log.Error(":: Empty outbound server POST body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding outbound server POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err, httpError := validation.ValidateJsonSchemaData(outServerIntJSONFile, osi)
	if err != nil {
		log.Error(":: Error validating outbound server POST data ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	// Name is required.
	if osi.Metadata.Name == "" {
		log.Error(":: Missing name in outbound server POST request ::", log.Fields{})
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	err = validateOutboundServerIntentInputs(osi)
	if err != nil {
		log.Error(":: Invalid create outbound server body inputs ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateServerOutboundIntent(osi, project, compositeApp, compositeAppVersion, deploymentIntentGroupName, trafficIntentGroupName, outboundIntentName, false)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, osi, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding create outbound server response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}
func (h outboundserverintentHandler) putHandler(w http.ResponseWriter, r *http.Request) {
	var osi module.OutboundServerIntent
	vars := mux.Vars(r)
	name := vars["outboundServerIntent"]
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deploymentIntentGroupName := vars["deploymentIntentGroup"]
	trafficIntentGroupName := vars["trafficGroupIntent"]
	outboundIntentName := vars["outboundClientIntent"]

	// check if the deploymentIntentGrpName exists
	_, err := orcmod.NewDeploymentIntentGroupClient().GetDeploymentIntentGroup(deploymentIntentGroupName, project, compositeApp, compositeAppVersion)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}
	err = json.NewDecoder(r.Body).Decode(&osi)

	switch {
	case err == io.EOF:
		log.Error(":: Empty outbound server PUT body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding outbound server PUT body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Name is required.
	if osi.Metadata.Name == "" {
		log.Error(":: Missing name in outbound server PUT request ::", log.Fields{})
		http.Error(w, "Missing name in PUT request", http.StatusBadRequest)
		return
	}

	// Name in URL should match name in body
	if osi.Metadata.Name != name {
		log.Error(":: Mismatched name in outbound server PUT request ::", log.Fields{})
		http.Error(w, "Mismatched name in PUT request", http.StatusBadRequest)
		return
	}

	err = validateOutboundServerIntentInputs(osi)
	if err != nil {
		log.Error(":: Invalid outbound server PUT inputs ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateServerOutboundIntent(osi, project, compositeApp, compositeAppVersion, deploymentIntentGroupName, trafficIntentGroupName, outboundIntentName, true)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, osi, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Error encoding outbound server update response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h outboundserverintentHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["outboundServerIntent"]
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deploymentIntentGroupName := vars["deploymentIntentGroup"]
	trafficIntentGroupName := vars["trafficGroupIntent"]
	outboundIntentName := vars["outboundClientIntent"]

	var ret interface{}
	var err error

	if len(name) == 0 {
		ret, err = h.client.GetServerOutboundIntents(project, compositeApp, compositeAppVersion, deploymentIntentGroupName, trafficIntentGroupName, outboundIntentName)
	} else {
		ret, err = h.client.GetServerOutboundIntent(name, project, compositeApp, compositeAppVersion, deploymentIntentGroupName, trafficIntentGroupName, outboundIntentName)
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
		log.Error(":: Error encoding get outbound server response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func (h outboundserverintentHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["outboundServerIntent"]
	project := vars["project"]
	compositeApp := vars["compositeApp"]
	compositeAppVersion := vars["compositeAppVersion"]
	deploymentIntentGroupName := vars["deploymentIntentGroup"]
	trafficIntentGroupName := vars["trafficGroupIntent"]
	outboundIntentName := vars["outboundClientIntent"]

	err := h.client.DeleteServerOutboundIntent(name, project, compositeApp, compositeAppVersion, deploymentIntentGroupName, trafficIntentGroupName, outboundIntentName)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
