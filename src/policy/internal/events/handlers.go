package event

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"net/http"
)

func (c Client) CreateEventHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	eventData := new(Event)
	if err := json.NewDecoder(r.Body).Decode(eventData); err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	request := &Request{
		Dummy: v["Event"],
	}
	response := c.CreateEvent(ctx, request)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error(":: Error encoding create event response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	log.Info("Create Event processed successfully", log.Fields{"IntentID": request})
}

func (c Client) DeleteEventHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	eventData := new(Event)
	if err := json.NewDecoder(r.Body).Decode(eventData); err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	request := &Request{
		Dummy: v["dummy"],
	}
	c.DeleteEvent(ctx, request)
	//TODO Error Handling
	w.WriteHeader(http.StatusOK)
	log.Info("Delete Event processed successfully", log.Fields{"IntentID": request})
}

func (c Client) GetEventHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	eventData := new(Event)
	if err := json.NewDecoder(r.Body).Decode(eventData); err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	request := &Request{
		Dummy: v["Ddd"],
	}
	response := c.GetEvent(ctx, request)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error(":: Error encoding Get Event response ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	log.Info("Get Event processed successfully", log.Fields{"IntentID": request})
}
