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

	"github.com/gorilla/mux"
	moduleLib "gitlab.com/project-emco/core/emco-base/src/app-config/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/apierror"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

//var brJSONFile string = "json-schemas/appConfig.json"

// Used to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type AppConfigHandler struct {
	// Interface that implements appConfig operations
	// We will set this variable with a mock interface for testing
	client moduleLib.AppConfigManager
}

func (h AppConfigHandler) createAppConfigHandler(w http.ResponseWriter, r *http.Request) {
	var br moduleLib.AppConfig
	//	var brc moduleLib.SpecFileContent
	var contentArray []string
	var fileNameArray []string

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
		log.Error(":: Empty appConfig POST body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding appConfig POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	//	err, httpError := validation.ValidateJsonSchemaData(brJSONFile, br)
	//	if err != nil {
	//		log.Error(":: JSON validation failed ::", log.Fields{"Error": err})
	//		http.Error(w, err.Error(), httpError)
	//		return
	//	}

	// BEGIN: AppConfig file processing
	formData := r.MultipartForm

	//get the *fileheaders
	files := formData.File["files"]

	for i := range files {
		file, err := files[i].Open()
		defer file.Close()
		if err != nil {
			logutils.Info("Unable to open file", log.Fields{"FileName": files[i].Filename})
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		content, err := ioutil.ReadAll(file)
		if err != nil {
			log.Error(":: File read failed ::", log.Fields{"Error": err})
			http.Error(w, "Unable to read file", http.StatusUnprocessableEntity)
			return
		}
		contStr := base64.StdEncoding.EncodeToString(content)
		contentArray = append(contentArray, contStr)
		fileNameArray = append(fileNameArray, files[i].Filename)
		logutils.Info("Appended file", log.Fields{"FileName": files[i].Filename})

	}

	specFC := moduleLib.SpecFileContent{FileContents: contentArray, FileNames: fileNameArray}
	// END: Customization file processing

	if br.Metadata.Name == "" {
		log.Error(":: Missing name in POST request ::", log.Fields{"Error": err})
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	p := vars["project"]
	ca := vars["compositeApp"]
	cv := vars["compositeAppVersion"]
	dig := vars["deploymentIntentGroup"]

	ret, err := h.client.CreateAppConfig(br, specFC, p, ca, cv, dig, false)
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

func (h AppConfigHandler) getAppConfigHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["appConfig"]
	p := vars["project"]
	ca := vars["compositeApp"]
	cv := vars["compositeAppVersion"]
	dig := vars["deploymentIntentGroup"]

	if len(name) == 0 {
		var brList []moduleLib.AppConfig

		ret, err := h.client.GetAllAppConfig(p, ca, cv, dig)
		if err != nil {
			apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
			http.Error(w, apiErr.Message, apiErr.Status)
			return
		}

		for _, br := range ret {
			brList = append(brList, moduleLib.AppConfig{Metadata: br.Metadata, Spec: br.Spec})
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(brList)
		if err != nil {
			log.Error(":: Encoding appConfig failure::", log.Fields{"Error": err})
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

	var retBr moduleLib.AppConfig
	var specFC moduleLib.SpecFileContent

	retBr, err = h.client.GetAppConfig(name, p, ca, cv, dig)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	specFC, err = h.client.GetAppConfigContent(name, p, ca, cv, dig)
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
		pw, err := mpw.CreatePart(textproto.MIMEHeader{"Content-Type": {"application/json"}, "Content-Disposition": {"form-data; name=customization"}})
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
		pw, err = mpw.CreatePart(textproto.MIMEHeader{"Content-Type": {"application/octet-stream"}, "Content-Disposition": {"form-data; name=files; filename=customizationFile"}})
		if err != nil {
			log.Error(":: multipart/form-data :: application/octet-stream failure ::", log.Fields{"Error": err})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, content := range specFC.FileContents {
			brBytes, err := base64.StdEncoding.DecodeString(content)
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

		for _, content := range specFC.FileContents {
			acBytes, err := base64.StdEncoding.DecodeString(content)
			if err != nil {
				log.Error(":: application/octet-stream failure::", log.Fields{"Error": err})
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			_, err = w.Write(acBytes)
			if err != nil {
				log.Error(":: FileWriter failure ::", log.Fields{"Error": err})
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

	default:
		http.Error(w, "set Accept: multipart/form-data, application/json or application/octet-stream", http.StatusMultipleChoices)
		return

	}
}

func (h AppConfigHandler) deleteAppConfigHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["appConfig"]
	p := vars["project"]
	ca := vars["compositeApp"]
	cv := vars["compositeAppVersion"]
	dig := vars["deploymentIntentGroup"]

	err := h.client.DeleteAppConfig(name, p, ca, cv, dig)
	if err != nil {
		apiErr := apierror.HandleErrors(vars, err, nil, apiErrors)
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
