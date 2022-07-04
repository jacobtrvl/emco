package controller

import (
	"bytes"
	event "emcopolicy/internal/events"
	"emcopolicy/internal/intent"
	"encoding/json"
	"github.com/pkg/errors"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"io/ioutil"
	"net/http"
)

func (c *Controller) ExecuteEvent(intentSpecJson []byte, agentSpec []byte, eventMessage []byte) {
	intentSpec := new(intent.Spec)
	if err := json.Unmarshal(intentSpecJson, intentSpec); err != nil {
		log.Error("ExecuteEvent failed", log.Fields{"err": err})
		return
	}
	type inputSpec struct {
		IntentSpec json.RawMessage `json:"intentSpec"`
		AgentSpec  json.RawMessage `json:"agentSpec"`
		Event      json.RawMessage `json:"event"`
	}
	policyInput := struct {
		Input inputSpec `json:"input"`
	}{inputSpec{
		IntentSpec: intentSpecJson,
		AgentSpec:  agentSpec,
		Event:      eventMessage,
	},
	}
	input, err := json.Marshal(policyInput)
	if err != nil {
		log.Error("ExecuteEvent failed", log.Fields{"err": err})
		return
	}
	policyUrl := "http://" + intentSpec.Policy.EngineUrl + "/" + intentSpec.Policy.PolicyName
	response, err := EvaluatePolicy(policyUrl, input)
	if err != nil {
		log.Error("ExecuteEvent failed", log.Fields{"err": err})
		return
	}
	if err := event.DoAction(intentSpec.Actor, response); err != nil {
		log.Error("ExecuteEvent failed", log.Fields{"err": err})
		return
	}
}

func EvaluatePolicy(url string, input []byte) ([]byte, error) {
	response, err := http.Post(url, "application/json", bytes.NewBuffer(input))
	if err != nil {
		return []byte{}, err
	}
	if response.StatusCode != http.StatusOK {
		log.Error("EvaluatePolicy failed due to http error", log.Fields{"status code": response.StatusCode, "url": url})
		return []byte{}, errors.Errorf("EvaluatePolicy failed due to http error. http Status code: %d", response.StatusCode)
	}
	defer response.Body.Close()
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return []byte{}, err
	}
	return responseData, nil
}
