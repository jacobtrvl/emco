// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Aarna Networks, Inc.

package api

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/contextdb"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"net/http"
)

type HandleFunc func(string, func(http.ResponseWriter, *http.Request)) *mux.Route

const (
	Version = "v2"
)

func NewRouter(ctrl contextdb.ContextDb) *mux.Router {
	r := mux.NewRouter().PathPrefix("/" + Version).Subrouter()
	r.HandleFunc("/get/{contextId}", func(w http.ResponseWriter, r *http.Request) {
		v := mux.Vars(r)
		key := "/context/" + v["contextId"] + "/meta/"
		var value json.RawMessage
		err := ctrl.Get(context.TODO(), key, value)
		if err != nil {
			log.Error("Error while getting context db data", log.Fields{"err": err})
			return
		}
		jsonValue, err := json.Marshal(value)
		if err != nil {
			log.Error("Error while parsing context db data", log.Fields{"err": err})
			return
		}
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(jsonValue)
		if err != nil {
			log.Error("Error while writing context db data", log.Fields{"err": err})
			return
		}
	}).Methods(http.MethodGet)
	return r
}
