package api

import (
	event "emcopolicy/internal/events"
	"emcopolicy/internal/policy"
	"emcopolicy/internal/sacontroller"
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

func NewRouter(c *sacontroller.Controller) *mux.Router {
	r := mux.NewRouter().PathPrefix("/" + Version).Subrouter()
	r.HandleFunc("/health", c.Health).Methods(http.MethodGet)
	registerEventHandlers(r.HandleFunc, c.EventClient())
	registerPolicyHandlers(r.HandleFunc, c.PolicyClient())
	registerPolicyIntentHandlers(r.HandleFunc, c.PolicyClient())
	return r
}

func registerEventHandlers(handle HandleFunc, client *event.Client) {
	handle(eventBaseUrl, func(w http.ResponseWriter, r *http.Request) {
		client.CreateEventHandler(r.Context(), w, r)
	}).Methods(http.MethodPost)
	handle(eventBaseUrl, func(w http.ResponseWriter, r *http.Request) {
		client.GetEventHandler(r.Context(), w, r)
	}).Methods(http.MethodGet)
	handle(eventBaseUrl, func(w http.ResponseWriter, r *http.Request) {
		client.DeleteEventHandler(r.Context(), w, r)
	}).Methods(http.MethodDelete)
}

func registerPolicyHandlers(handle HandleFunc, client *policy.Client) {
	handle(policyBaseUrl, func(w http.ResponseWriter, r *http.Request) {
		client.CreatePolicyHandler(r.Context(), w, r)
	}).Methods(http.MethodPost)
	handle(policyBaseUrl, func(w http.ResponseWriter, r *http.Request) {
		client.GetPolicyHandler(r.Context(), w, r)
	}).Methods(http.MethodGet)
	handle(policyBaseUrl, func(w http.ResponseWriter, r *http.Request) {
		client.DeletePolicyHandler(r.Context(), w, r)
	}).Methods(http.MethodDelete)
}

func registerPolicyIntentHandlers(handle HandleFunc, client *policy.Client) {
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
