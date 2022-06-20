package api

import (
	"emcopolicy/internal/controller"
	"emcopolicy/internal/intent"
	"github.com/gorilla/mux"
	"net/http"
)

const (
	Version             = "v2" //API Version
	policyIndentBaseUrl = "/projects/{project}/composite-apps/{compositeApp}/" +
		"{compositeAppVersion}/deployment-intent-groups/{deploymentIntentGroup}/" +
		"policy-intents/{policyIntentId}"
	eventBaseUrl  = ""
	policyBaseUrl = ""
)

type HandleFunc func(string, func(http.ResponseWriter, *http.Request)) *mux.Route

func NewRouter(c *controller.Controller) *mux.Router {
	r := mux.NewRouter().PathPrefix("/" + Version).Subrouter()
	r.HandleFunc("/health", c.Health).Methods(http.MethodGet)
	registerPolicyIntentHandlers(r.HandleFunc, c.PolicyClient())
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
