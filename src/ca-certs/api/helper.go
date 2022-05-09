// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/validation"
)

var CertificateSchemaJson string = "json-schemas/certificate.json"
var ClusterSchemaJson string = "json-schemas/cluster.json"
var LogicalCloudSchemaJson string = "json-schemas/logicalCloud.json"

type cpVars struct {
	cert,
	cluster,
	clusterProvider string
}

type lcVars struct {
	cert,
	cluster,
	logicalCloud,
	project string
}

// validateRequestBody validate the request body before storing it in the database
func validateRequestBody(r io.Reader, v interface{}, jsonSchema string) (int, error) {
	err := json.NewDecoder(r).Decode(&v)
	switch {
	case err == io.EOF:
		logutils.Error("Empty request body",
			logutils.Fields{
				"Error": err.Error()})
		return http.StatusBadRequest, errors.New("empty request body")
	case err != nil:
		logutils.Error("Error decoding the request body",
			logutils.Fields{
				"Error": err.Error()})
		return http.StatusUnprocessableEntity, errors.New("error decoding the request body")
	}

	// // validate the payload for the required values
	// if err = validateData(v); err != nil {
	// 	return http.StatusBadRequest, err
	// }

	// ensure that the request body matches the schema defined in the JSON file
	err, code := validation.ValidateJsonSchemaData(jsonSchema, v)
	if err != nil {
		logutils.Error("Json schema validation failed",
			logutils.Fields{
				"JsonSchema": jsonSchema,
				"Error":      err.Error()})
		return code, err
	}

	return 0, nil
}

// // validateData validate the payload for the required values
// func validateData(i interface{}) error {
// 	switch p := i.(type) {
// 	case *module.Customization:
// 		return validateCustomizationData(*p)
// 	case *module.GenericK8sIntent:
// 		return validateGenericK8sIntentData(*p)
// 	case *module.Resource:
// 		return validateResourceData(*p)
// 	default:
// 		logutils.Error("Invalid payload",
// 			logutils.Fields{
// 				"Type": p})
// 		return errors.New("invalid payload")
// 	}

// 	return nil
// }

// sendResponse sends an application/json response to the client
func sendResponse(w http.ResponseWriter, v interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		logutils.Error("Failed to encode the response",
			logutils.Fields{
				"Error":    err,
				"Response": v})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// cpVars returns the route variables for the current request
func _cpVars(vars map[string]string) cpVars {
	return cpVars{
		cert:            vars["caCert"],
		cluster:         vars["cluster"],
		clusterProvider: vars["clusterProvider"]}
}

// _cVars returns the route variables for the current request
func _lcVars(vars map[string]string) lcVars {
	return lcVars{
		cert:         vars["caCert"],
		cluster:      vars["cluster"],
		logicalCloud: vars["logicalCloud"],
		project:      vars["project"]}
}
