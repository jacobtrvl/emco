//=======================================================================
// Copyright (c) 2022 Aarna Networks, Inc.
// All rights reserved.
// ======================================================================
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//           http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// ========================================================================

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
