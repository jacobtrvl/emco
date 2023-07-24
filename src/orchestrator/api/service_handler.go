// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/validation"
	moduleLib "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
)

const serviceApiJsonFile = "json-schemas/service.json"

type serviceHandler struct {
	client moduleLib.ServiceManager
}

func (h serviceHandler) createServiceHandler(w http.ResponseWriter, r *http.Request) {
	var serviceReq moduleLib.ServiceRequest

	err := json.NewDecoder(r.Body).Decode(&serviceReq)
	switch {
	case err == io.EOF:
		log.Error(err.Error(), log.Fields{})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(serviceApiJsonFile, serviceReq)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	ctx := r.Context()
	vars := mux.Vars(r)
	service := serviceReq.ToService()
	service.MetaData.Project = vars["project"]
	err = h.client.CreateService(ctx, service)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, err.Error(), apiErr.Status)
		log.Error(err.Error(), log.Fields{})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(service)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h serviceHandler) getAllServicesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	pList := []string{"project"}
	err := validation.IsValidParameterPresent(vars, pList)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	project := vars["project"]
	serviceList, err := h.client.GetAllServices(ctx, project)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, err.Error(), apiErr.Status)
		log.Error(err.Error(), log.Fields{})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(serviceList)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h serviceHandler) getServiceHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	vars := mux.Vars(r)

	project := vars["project"]
	if project == "" {
		log.Error("Missing projectName in DELETE request", log.Fields{})
		http.Error(w, "Missing projectName in DELETE request", http.StatusBadRequest)
		return
	}

	serviceName := vars["service"]
	if serviceName == "" {
		log.Error("Missing name of Service in DELETE request", log.Fields{})
		http.Error(w, "Missing name of Service DELETE GET request", http.StatusBadRequest)
		return
	}

	key := &moduleLib.ServiceKey{
		Name:    serviceName,
		Project: project,
	}
	service, err := h.client.GetService(ctx, key)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, err.Error(), apiErr.Status)
		log.Error(err.Error(), log.Fields{})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(service)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func (h serviceHandler) updateServiceDIGsHandler(w http.ResponseWriter, r *http.Request) {
	var sdu moduleLib.ServiceDigsUpdate

	err := json.NewDecoder(r.Body).Decode(&sdu)
	switch {
	case err == io.EOF:
		log.Error(err.Error(), log.Fields{})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ctx := r.Context()
	vars := mux.Vars(r)

	key := &moduleLib.ServiceKey{
		Name:    vars["service"],
		Project: vars["project"],
	}
	service, err := h.client.UpdateServiceDigs(ctx, key, &sdu)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, err.Error(), apiErr.Status)
		log.Error(err.Error(), log.Fields{})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	err = json.NewEncoder(w).Encode(service)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h serviceHandler) updateServiceHandler(w http.ResponseWriter, r *http.Request) {
	var serviceReq moduleLib.ServiceRequest

	err := json.NewDecoder(r.Body).Decode(&serviceReq)
	switch {
	case err == io.EOF:
		log.Error(err.Error(), log.Fields{})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(serviceApiJsonFile, serviceReq)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	ctx := r.Context()
	vars := mux.Vars(r)
	service := serviceReq.ToService()
	service.MetaData.Project = vars["project"]
	err = h.client.UpdateService(ctx, service)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, err.Error(), apiErr.Status)
		log.Error(err.Error(), log.Fields{})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	err = json.NewEncoder(w).Encode(service)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h serviceHandler) deleteServiceHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	project := vars["project"]
	if project == "" {
		log.Error("Missing projectName in GET request", log.Fields{})
		http.Error(w, "Missing projectName in GET request", http.StatusBadRequest)
		return
	}

	serviceName := vars["service"]
	if serviceName == "" {
		log.Error("Missing name of Service", log.Fields{})
		http.Error(w, "Missing name of Service in GET request", http.StatusBadRequest)
		return
	}

	key := &moduleLib.ServiceKey{
		Name:    serviceName,
		Project: project,
	}
	err := h.client.DeleteService(ctx, key)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, err.Error(), apiErr.Status)
		log.Error(err.Error(), log.Fields{})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h serviceHandler) instantiateServiceHandler(w http.ResponseWriter, r *http.Request) {
	var sda moduleLib.ServiceDigsAction

	err := json.NewDecoder(r.Body).Decode(&sda)
	if err != nil && err != io.EOF {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ctx := r.Context()
	vars := mux.Vars(r)
	key := &moduleLib.ServiceKey{
		Name:    vars["service"],
		Project: vars["project"],
	}
	err = h.client.InstantiateService(ctx, key, &sda)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, err.Error(), apiErr.Status)
		log.Error(err.Error(), log.Fields{})
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h serviceHandler) terminateServiceHandler(w http.ResponseWriter, r *http.Request) {
	var sda moduleLib.ServiceDigsAction

	err := json.NewDecoder(r.Body).Decode(&sda)
	if err != nil && err != io.EOF {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ctx := r.Context()
	vars := mux.Vars(r)
	key := &moduleLib.ServiceKey{
		Name:    vars["service"],
		Project: vars["project"],
	}
	err = h.client.TerminateService(ctx, key, &sda)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, err.Error(), apiErr.Status)
		log.Error(err.Error(), log.Fields{})
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h serviceHandler) instantiateServiceDIGsHandler(w http.ResponseWriter, r *http.Request) {
	var sda moduleLib.ServiceDigsAction

	err := json.NewDecoder(r.Body).Decode(&sda)
	if err != nil && err != io.EOF {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ctx := r.Context()
	vars := mux.Vars(r)
	key := &moduleLib.ServiceKey{
		Name:    vars["service"],
		Project: vars["project"],
	}
	err = h.client.InstantiateServiceDIGs(ctx, key, &sda)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, err.Error(), apiErr.Status)
		log.Error(err.Error(), log.Fields{})
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h serviceHandler) terminateServiceDIGsHandler(w http.ResponseWriter, r *http.Request) {
	var sda moduleLib.ServiceDigsAction

	err := json.NewDecoder(r.Body).Decode(&sda)
	if err != nil && err != io.EOF {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ctx := r.Context()
	vars := mux.Vars(r)
	key := &moduleLib.ServiceKey{
		Name:    vars["service"],
		Project: vars["project"],
	}
	err = h.client.TerminateServiceDIGs(ctx, key, &sda)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, err.Error(), apiErr.Status)
		log.Error(err.Error(), log.Fields{})
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h serviceHandler) serviceStatusHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	key := &moduleLib.ServiceKey{
		Name:    vars["service"],
		Project: vars["project"],
	}

	status, err := h.client.ServiceStatus(ctx, key)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, err.Error(), apiErr.Status)
		log.Error(err.Error(), log.Fields{})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(*status)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
