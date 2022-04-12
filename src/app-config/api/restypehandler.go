package api

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	_ "strings"

	"github.com/gorilla/mux"
	moduleLib "gitlab.com/project-emco/core/emco-base/src/app-config/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"

	_ "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/validation"
)

//var brJSONFile string = "json-schemas/resource.json"

// Used to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type restypeHandler struct {
	// Interface that implements resource operations
	// We will set this variable with a mock interface for testing
	client moduleLib.ResTypeManager
}

func (h restypeHandler) createResTypeHandler(w http.ResponseWriter, r *http.Request) {
	var br moduleLib.ResType
	var brc moduleLib.ResTypeFileContent
	vars := mux.Vars(r)

	// Implemenation using multipart form and set maxSize 16MB
	err := r.ParseMultipartForm(16777216)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	jsn := bytes.NewBuffer([]byte(r.FormValue("metadata")))
	err = json.NewDecoder(jsn).Decode(&br)
	switch {
	case err == io.EOF:
		log.Error(":: Empty resource POST body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding resource POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

//	err, httpError := validation.ValidateJsonSchemaData(brJSONFile, br)
//	if err != nil {
//		log.Error(":: JSON validation failed ::", log.Fields{"Error": err})
//		http.Error(w, err.Error(), httpError)
//		return
//	}

	// if newobject is true, and its neither a configmap nor a secret and then contentFile should be there, or else throw exception
	// if the newobject is false the contentfile if any is ignored
	file, _, err := r.FormFile("file")

	if err != nil {
		log.Error(":: Unable to process file, check if file is present ::", log.Fields{"Error": err})
		http.Error(w, "Unable to process file", http.StatusUnprocessableEntity)
		return
	}
	defer file.Close()

		//Convert the file content to base64 for storage
	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Error(":: File read failed ::", log.Fields{"Error": err})
		http.Error(w, "Unable to read file", http.StatusUnprocessableEntity)
		return
	}
	brc.TemplateContent = base64.StdEncoding.EncodeToString(content)


	if br.Metadata.Name == "" {
		log.Error(":: Missing name in POST request ::", log.Fields{"Error": err})
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	p := vars["project"]
	ca := vars["compositeApp"]
	cv := vars["compositeAppVersion"]

	ret, err := h.client.CreateResType(br, brc, p, ca, cv, false)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, br, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Encoding error ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h restypeHandler) putResTypeHandler(w http.ResponseWriter, r *http.Request) {
	var br moduleLib.ResType
	var brc moduleLib.ResTypeFileContent
	vars := mux.Vars(r)

	// Implemenation using multipart form and set maxSize 16MB
	err := r.ParseMultipartForm(16777216)
	if err != nil {
		log.Error(":: Parsing form failure::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	jsn := bytes.NewBuffer([]byte(r.FormValue("resource")))
	err = json.NewDecoder(jsn).Decode(&br)
	switch {
	case err == io.EOF:
		log.Error(":: Empty resource body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding resource body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}


	// if newobject is true, and its neither a configmap nor a secret and then contentFile should be there, or else throw exception
	// if the newobject is false or kind is configmap or secret, there should not be any file.
	file, _, err := r.FormFile("file")
	if err != nil {
		log.Error(":: Unable to process file, check if file is present ::", log.Fields{"Error": err})
		http.Error(w, "Unable to process file, Check if file is present", http.StatusUnprocessableEntity)
		return
	}
	defer file.Close()

		//Convert the file content to base64 for storage
	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Error(":: File read failed ::", log.Fields{"Error": err})
		http.Error(w, "Unable to read file", http.StatusUnprocessableEntity)
		return
	}
	brc.TemplateContent = base64.StdEncoding.EncodeToString(content)

	if br.Metadata.Name == "" {
		log.Error(":: Missing name in POST request ::", log.Fields{"Error": err})
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	p := vars["project"]
	ca := vars["compositeApp"]
	cv := vars["compositeAppVersion"]

	ret, err := h.client.CreateResType(br, brc, p, ca, cv, true)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, br, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Encoding error ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h restypeHandler) getResTypeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["restype"]
	p := vars["project"]
	ca := vars["compositeApp"]
	cv := vars["compositeAppVersion"]

	if len(name) == 0 {
		var brList []moduleLib.ResType

		ret, err := h.client.GetAllResTypes(p, ca, cv)
		if err != nil {
			apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
			http.Error(w, apiErr.Message, apiErr.Status)
			return
		}

		for _, br := range ret {
			brList = append(brList, moduleLib.ResType{Metadata: br.Metadata, Spec: br.Spec})
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(brList)
		if err != nil {
			log.Error(":: Encoding resource failure::", log.Fields{"Error": err})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	accepted, _, err := mime.ParseMediaType(r.Header.Get("Accept"))
	if err != nil {
		log.Error(":: Mime parser failure::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusNotAcceptable)
		return
	}

	var retBr moduleLib.ResType
	var retBrContent moduleLib.ResTypeFileContent

	retBr, err = h.client.GetResType(name, p, ca, cv)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	retBrContent, err = h.client.GetResTypeContent(name, p, ca, cv)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	switch accepted {
	case "multipart/form-data":
		mpw := multipart.NewWriter(w)
		w.Header().Set("Content-Type", mpw.FormDataContentType())
		w.WriteHeader(http.StatusOK)
		pw, err := mpw.CreatePart(textproto.MIMEHeader{"Content-Type": {"application/json"}, "Content-Disposition": {"form-data; name=resource"}})
		if err != nil {
			log.Error(":: multipart/form-data :: application/json failure::", log.Fields{"Error": err})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := json.NewEncoder(pw).Encode(retBr); err != nil {
			log.Error(":: multipart/form-data :: encoding failure", log.Fields{"Error": err})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		pw, err = mpw.CreatePart(textproto.MIMEHeader{"Content-Type": {"application/octet-stream"}, "Content-Disposition": {"form-data; name=file; filename=resourceTemplate"}})
		if err != nil {
			log.Error(":: multipart/form-data :: application/octet-stream failure ::", log.Fields{"Error": err})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		brBytes, err := base64.StdEncoding.DecodeString(retBrContent.TemplateContent)
		if err != nil {
			log.Error(":: multipart/form-data :: application/octet-stream decode failure ::", log.Fields{"Error": err})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = pw.Write(brBytes)
		if err != nil {
			log.Error(":: FileWriter failure ::", log.Fields{"Error": err})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case "application/json":
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(retBr)
		if err != nil {
			log.Error(":: application/json encoding failure::", log.Fields{"Error": err})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case "application/octet-stream":
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		brBytes, err := base64.StdEncoding.DecodeString(retBrContent.TemplateContent)
		if err != nil {
			log.Error(":: application/octet-stream failure::", log.Fields{"Error": err})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = w.Write(brBytes)
		if err != nil {
			log.Error(":: FileWriter failure ::", log.Fields{"Error": err})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "set Accept: multipart/form-data, application/json or application/octet-stream", http.StatusMultipleChoices)
		return
	}
}

func (h restypeHandler) deleteResTypeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["restype"]
	p := vars["project"]
	ca := vars["compositeApp"]
	cv := vars["compositeAppVersion"]

	err := h.client.DeleteResType(name, p, ca, cv)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
