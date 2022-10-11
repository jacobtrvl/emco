// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"encoding/json"
	"io"
	"net/http"

	"context"
	"github.com/gorilla/mux"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/validation"
	controller "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/controller"
)

var controllerJSONFile string = "json-schemas/controller.json"

// Used to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type controllerHandler struct {
	// Interface that implements controller operations
	// We will set this variable with a mock interface for testing
	client controller.ControllerManager
}

// Create handles creation of the controller entry in the database
func (h controllerHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	var m controller.Controller

	err := json.NewDecoder(r.Body).Decode(&m)
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
	err, httpError := validation.ValidateJsonSchemaData(controllerJSONFile, m)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	ret, err := h.client.CreateController(context.Background(), m, false)
	if err != nil {
		apiErr := apierror.HandleErrors(mux.Vars(r), err, m, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Put handles creation or update of the controller entry in the database
func (h controllerHandler) putHandler(w http.ResponseWriter, r *http.Request) {
	var m controller.Controller
	vars := mux.Vars(r)
	name := vars["dtcController"]

	err := json.NewDecoder(r.Body).Decode(&m)
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
	err, httpError := validation.ValidateJsonSchemaData(controllerJSONFile, m)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	// name in URL should match name in body
	if m.Metadata.Name != name {
		log.Error("Mismatched name in PUT request", log.Fields{})
		http.Error(w, "Mismatched name in PUT request", http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateController(context.Background(), m, true)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, m, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Get handles GET operations on a particular controller Name
// Returns a controller
func (h controllerHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["dtcController"]
	var ret interface{}
	var err error

	// handle the get all controllers case
	if len(name) == 0 {
		ret, err = h.client.GetControllers(context.Background())
	} else {
		ret, err = h.client.GetController(context.Background(), name)
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
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Delete handles DELETE operations on a particular controller Name
func (h controllerHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["dtcController"]

	err := h.client.DeleteController(context.Background(), name)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
