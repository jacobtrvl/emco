// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"

	"gitlab.com/project-emco/core/emco-base/src/genericactioncontroller/pkg/module"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/validation"
	k8s "k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// file stores the uploaded resource/customization file info
type file struct {
	Name    string
	Content string
}

// validateRequestBody validate the request body before storing it in the database
func validateRequestBody(r io.Reader, v interface{}, jsonSchema string) (int, error) {
	err := json.NewDecoder(r).Decode(&v)
	switch {
	case err == io.EOF:
		log.Error("Empty request body",
			log.Fields{
				"Error": err.Error()})
		return http.StatusBadRequest, errors.New("empty request body")
	case err != nil:
		log.Error("Error decoding the request body",
			log.Fields{
				"Error": err.Error()})
		return http.StatusUnprocessableEntity, errors.New("error decoding the request body")
	}

	// validate the payload for the required values
	if err = validateData(v); err != nil {
		return http.StatusBadRequest, err
	}

	// ensure that the request body matches the schema defined in the JSON file
	err, code := validation.ValidateJsonSchemaData(jsonSchema, v)
	if err != nil {
		log.Error("Json schema validation failed",
			log.Fields{
				"JsonSchema": jsonSchema,
				"Error":      err.Error()})
		return code, err
	}

	return 0, nil
}

// validateData validate the payload for the required values
func validateData(i interface{}) error {
	switch p := i.(type) {
	case *module.Customization:
		return validateCustomizationData(*p)
	case *module.GenericK8sIntent:
		return validateGenericK8sIntentData(*p)
	case *module.Resource:
		return validateResourceData(*p)
	default:
		log.Error("Invalid payload",
			log.Fields{
				"Type": p})
		return errors.New("invalid payload")
	}
}

// parseFile read the content from each file and returns a base64 encoded value
func parseFile(fileHeader []*multipart.FileHeader) ([]file, int, error) {
	var files []file
	for _, fh := range fileHeader {
		mpf, err := fh.Open()
		if err != nil {
			log.Error("Failed to open the file",
				log.Fields{
					"Name":  fh.Filename,
					"Error": err})
			return nil, http.StatusUnprocessableEntity, fmt.Errorf("failed to open the file %s", fh.Filename)
		}

		defer mpf.Close()

		data, err := ioutil.ReadAll(mpf)
		if err != nil {
			log.Error("Failed to read the file",
				log.Fields{
					"Name":  fh.Filename,
					"Error": err})
			return nil, http.StatusUnprocessableEntity, fmt.Errorf("failed to read the file %s", fh.Filename)
		}
		if len(data) > 0 {
			f := file{
				Name:    fh.Filename,
				Content: base64.StdEncoding.EncodeToString(data),
			}
			files = append(files, f)
		}
	}

	return files, 0, nil
}

// sendResponse sends an application/json response to the client
func sendResponse(w http.ResponseWriter, v interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Error("Failed to encode the response",
			log.Fields{
				"Error":    err,
				"Response": v})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// sendMultipartResponse sends a multipart/form-data response to the client
func sendMultipartResponse(w http.ResponseWriter, v interface{}, files []file, name string) {
	mpw := multipart.NewWriter(w)

	w.Header().Set("Content-Type", mpw.FormDataContentType())
	w.WriteHeader(http.StatusOK)

	// create a new multipart section with the provided header
	pw, err := mpw.CreatePart(textproto.MIMEHeader{"Content-Type": {"application/json"},
		"Content-Disposition": {"form-data; name=" + name}})
	if err != nil {
		log.Error("Failed to create a new multipart section with the provided header",
			log.Fields{
				"Headers": "Content-Type: {application/json}, Content-Disposition: {form-data; name=" + name + "}",
				"Error":   err.Error()})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(pw).Encode(v); err != nil {
		log.Error("Failed to encode the response",
			log.Fields{
				"Error":    err,
				"Response": v})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	name = "file"
	if len(files) > 1 {
		name = "files"
	}

	for _, f := range files {
		// create a new multipart section with the provided header
		pw, err = mpw.CreatePart(textproto.MIMEHeader{"Content-Type": {"application/octet-stream"},
			"Content-Disposition": {"form-data; name=" + name + "; filename=" + f.Name}})
		if err != nil {
			log.Error("Failed to create a new multipart section with the provided header",
				log.Fields{
					"Headers": "Content-Type: {application/octet-stream}, Content-Disposition: {form-data; name=files; filename=" + f.Name + "}",
					"Error":   err.Error()})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data, err := base64.StdEncoding.DecodeString(f.Content)
		if err != nil {
			log.Error("Failed to decode the base64 data",
				log.Fields{
					"MediaType": "multipart/form-data",
					"Error":     err.Error()})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err = pw.Write(data); err != nil {
			log.Error("Failed to write the content to the underlying data stream",
				log.Fields{
					"MediaType": "multipart/form-data",
					"Error":     err.Error()})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// sendOctetStreamResponse sends an application/octet-stream response to the client
func sendOctetStreamResponse(w http.ResponseWriter, files []file) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)

	for _, f := range files {
		data, err := base64.StdEncoding.DecodeString(f.Content)
		if err != nil {
			log.Error("Failed to decode the base64 data",
				log.Fields{
					"MediaType": "application/octet-stream",
					"Error":     err.Error()})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err = w.Write(data); err != nil {
			log.Error("Failed to write the content to the underlying data stream",
				log.Fields{
					"MediaType": "application/octet-stream",
					"Error":     err.Error()})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// validateContent validate the Resource template content serialized to a byte array
func validateContent(data []byte) error {
	out, err := yaml.ToJSON(data)
	if err != nil {
		log.Error("Failed to convert the resource template content into a json document",
			log.Fields{
				"Error": err.Error()})
		return err
	}

	var object interface{}
	if err := json.Unmarshal(out, &object); err != nil {
		log.Error("Failed to unmarshal the resource template content",
			log.Fields{
				"Error": err.Error()})
		return err
	}

	fields, ok := object.(map[string]interface{})
	if !ok || fields == nil {
		log.Error("Invalid object to validate",
			log.Fields{
				"Object": object})
		return errors.New("invalid object to validate")
	}

	if err = validateObjectVersionKind(fields); err != nil {
		return err
	}

	if err = validateObjectMetadata(fields); err != nil {
		return err
	}

	return nil
}

// validateObjectVersionKind validate the Resource template content for the required version/kind fields
func validateObjectVersionKind(fields map[string]interface{}) error {
	apiVersion := fields["apiVersion"]
	if apiVersion == nil {
		log.Error("apiVersion not set",
			log.Fields{})
		return errors.New("apiVersion not set")
	}
	gv, ok := apiVersion.(string)
	if !ok {
		log.Error("apiVersion is not string type",
			log.Fields{
				"apiVersion": apiVersion})
		return errors.New("apiVersion is not string type")
	}
	if len(gv) == 0 {
		log.Error("apiVersion may not be empty",
			log.Fields{})
		return errors.New("apiVersion may not be empty")
	}

	kind := fields["kind"]
	if kind == nil {
		log.Error("kind not set",
			log.Fields{})
		return errors.New("kind not set")
	}
	if _, ok := kind.(string); !ok {
		log.Error("kind is not string type",
			log.Fields{
				"kind": kind})
		return errors.New("kind is not string type")
	}

	return nil
}

// validateObjectMetadata validate the Resource template content for the required metadata fields
func validateObjectMetadata(fields map[string]interface{}) error {
	metadata := fields["metadata"]
	if metadata == nil {
		log.Error("metadata not set",
			log.Fields{})
		return errors.New("metadata not set")
	}

	data, ok := metadata.(map[string]interface{})
	if !ok || data == nil {
		log.Error("Invalid metadata",
			log.Fields{
				"metadata": metadata})
		return errors.New("invalid metadata")
	}

	name := data["name"]
	if name == nil {
		log.Error("Resource name not set",
			log.Fields{})
		return errors.New("resource name not set")
	}

	n, ok := name.(string)
	if !ok {
		log.Error("Resource name is not string type",
			log.Fields{
				"Name": name})
		return errors.New("resource name is not string type")
	}
	if len(n) == 0 {
		log.Error("Resource name may not be empty",
			log.Fields{})
		return errors.New("resource name may not be empty")
	}

	return nil
}

// customizeDataKey maps the given data key with the specific ConfigMap/Secret Data
func customizeDataKey(customizationContent module.CustomizationContent, keyOptions []module.KeyOptions) error {
	keys := map[string]string{}
	for i, content := range customizationContent.Content {
		for _, key := range keyOptions {
			if content.FileName == key.FileName {
				// validate the string is a valid key for the ConfigMap/Secret Data
				if err := validateKey(key.KeyName, keys); err != nil {
					return err
				}
				// update the key for the ConfigMap/Secret
				content.KeyName = key.KeyName
				keys[key.KeyName] = key.FileName // this is to track the keys
				break
			}
		}
		// update the customization content
		customizationContent.Content[i] = content
	}

	return nil
}

// validateKey validate the string is a valid key for a ConfigMap/Secret
func validateKey(key string, keys map[string]string) error {
	// validate this is a valid key for a ConfigMap/Secret
	if errs := k8s.IsConfigMapKey(key); len(errs) > 0 {
		log.Error("Invalid key",
			log.Fields{
				"Key":   key,
				"Error": strings.Join(errs, "\n")})
		return fmt.Errorf("%s is not a valid key name for a ConfigMap or Secret",
			key)
	}
	// check for duplicate key
	if _, exists := keys[key]; exists {
		log.Error("Duplicate key",
			log.Fields{
				"Key":   key,
				"Error": "A key with the name already exists in Data"})
		return fmt.Errorf("a key with the name %s already exists in Data", key)
	}

	return nil
}
