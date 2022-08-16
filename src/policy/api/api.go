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
	"emcopolicy/internal/controller"
	event "emcopolicy/internal/events"
	"emcopolicy/internal/intent"
	"github.com/gorilla/mux"
	"net/http"
)

const (
	Version             = "v2" //API Version
	policyIndentBaseUrl = "/projects/{project}/composite-apps/{compositeApp}/" +
		"{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/" +
		"policy-intents/{policyIntentId}"
	agentBaseUrl = "/policy/agents"
)

type HandleFunc func(string, func(http.ResponseWriter, *http.Request)) *mux.Route

func NewRouter(c *controller.Controller) *mux.Router {
	r := mux.NewRouter().PathPrefix("/" + Version).Subrouter()
	r.HandleFunc("/health", c.Health).Methods(http.MethodGet)
	registerPolicyIntentHandlers(r.HandleFunc, c.PolicyClient())
	registerEventHandlers(r.HandleFunc, c.EventClient())
	return r
}

func registerPolicyIntentHandlers(handle HandleFunc, client *intent.Client) {
	handle(policyIndentBaseUrl, func(w http.ResponseWriter, r *http.Request) {
		client.CreatePolicyIntentHandler(r.Context(), w, r)
	}).Methods(http.MethodPost)
	handle(policyIndentBaseUrl, func(w http.ResponseWriter, r *http.Request) {
		client.GetPolicyIntentHandler(r.Context(), w, r)
	}).Methods(http.MethodGet)
	handle(policyIndentBaseUrl, func(w http.ResponseWriter, r *http.Request) {
		client.DeletePolicyIntentHandler(r.Context(), w, r)
	}).Methods(http.MethodDelete)
}

func registerEventHandlers(handle HandleFunc, client *event.Client) {
	handle(agentBaseUrl+"/{id}", func(w http.ResponseWriter, r *http.Request) {
		client.RegisterAgentHandler(r.Context(), w, r)
	}).Methods(http.MethodPost)
	handle(agentBaseUrl+"/{id}", func(w http.ResponseWriter, r *http.Request) {
		client.GetAgentHandler(r.Context(), w, r)
	}).Methods(http.MethodGet)
	handle(agentBaseUrl, func(w http.ResponseWriter, r *http.Request) {
		client.GetAllAgentHandler(r.Context(), w, r)
	}).Methods(http.MethodGet)
	handle(agentBaseUrl+"/{id}", func(w http.ResponseWriter, r *http.Request) {
		client.DeleteAgentHandler(r.Context(), w, r)
	})
}
