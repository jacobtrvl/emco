package policy

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"net/http"
)

func (c Client) CreatePolicyHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	policyData := new(Policy)
	if err := json.NewDecoder(r.Body).Decode(policyData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	request := &PolicyRequest{
		PolicyId: v["policyId"],
		Policy:   policyData,
	}
	response, err := c.CreatePolicy(ctx, request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Info("Create Policy processed successfully", log.Fields{"IntentID": request.PolicyId})
}

func (c Client) DeletePolicyHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	request := &PolicyRequest{
		PolicyId: v["policyId"],
	}
	if err := c.DeletePolicy(ctx, request); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	log.Info("Delete Policy processed successfully", log.Fields{"IntentID": request.PolicyId})
}

func (c Client) GetPolicyHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	request := &PolicyRequest{
		PolicyId: v["policyId"],
	}
	response, err := c.GetPolicy(ctx, request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info("Get Policy processed successfully", log.Fields{"IntentID": request.PolicyId})

}

func (c Client) CreatePolicyIntentHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	intentData := new(Intent)
	if err := json.NewDecoder(r.Body).Decode(intentData); err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	request := &IntentRequest{
		Project:               v["project"],
		CompositeApp:          v["compositeApp"],
		CompositeAppVersion:   v["compositeAppVersion"],
		DeploymentIntentGroup: v["deploymentIntentGroup"],
		PolicyIntentId:        v["policyIntentId"],
		IntentData:            intentData,
	}
	response, err := c.CreateIntent(ctx, request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info("Create Policy intent processed successfully", log.Fields{"IntentID": request.PolicyIntentId})
}

func (c Client) GetPolicyIntentHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	request := &IntentRequest{
		Project:               v["project"],
		CompositeApp:          v["compositeApp"],
		CompositeAppVersion:   v["compositeAppVersion"],
		DeploymentIntentGroup: v["deploymentIntentGroup"],
		PolicyIntentId:        v["policyIntentId"],
	}
	response, err := c.GetIntent(ctx, request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if response == nil {
		http.Error(w, "404 Policy Intent not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Info("Get Policy intent processed successfully", log.Fields{"IntentID": request.PolicyIntentId})
}

func (c Client) DeletePolicyIntentHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	request := &IntentRequest{
		Project:               v["project"],
		CompositeApp:          v["compositeApp"],
		CompositeAppVersion:   v["compositeAppVersion"],
		DeploymentIntentGroup: v["deploymentIntentGroup"],
		PolicyIntentId:        v["policyIntentId"],
	}

	if err := c.DeleteIntent(ctx, request); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	log.Info("Delete Policy intent processed successfully", log.Fields{"IntentID": request.PolicyIntentId})
}
