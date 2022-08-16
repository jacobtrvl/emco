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

package event

import (
	"context"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
)

type Provider interface {
	listen()
}

type AgentSpec struct {
	Id       string `json:"id,omitempty"`
	EndPoint string `json:"addr"`
}

type AgentKey struct {
	PolicyModule string `json:"policyModule"`
	AgentID      string `json:"agent"`
}

func (c Client) RegisterAgent(_ context.Context, id string, spec AgentSpec) (*AgentSpec, error) {
	key := AgentKey{
		"Agent",
		id,
	}
	err := c.db.Insert(c.storeName, key, nil, c.tag, spec)
	if err != nil {
		return nil, errors.Wrap(err, "Agent Registration failed")
	}
	c.agentStream <- StreamAgentData{
		Spec:      spec,
		Operation: "APPEND",
	}
	return &spec, nil
}

func (c Client) GetAllAgents(_ context.Context) ([]AgentSpec, error) {
	var (
		agents []AgentSpec
	)
	key := struct {
		PolicyModule string `json:"policyModule"`
		AgentID      bson.M `json:"agent"`
	}{"Agent", bson.M{"$exists": true}}
	value, err := c.db.Find(c.storeName, key, c.tag)
	if err != nil {
		return nil, errors.Wrap(err, "GetAgents failed")
	}
	if value == nil || len(value) == 0 {
		return nil, nil
	}
	for _, v := range value {
		agent := new(AgentSpec)
		if err := c.db.Unmarshal(v, agent); err != nil {
			return agents, errors.Wrap(err, "GetAgents failed")
		}
		agents = append(agents, *agent)
	}
	return agents, nil
}

func (c Client) GetAgent(_ context.Context, id string) (*AgentSpec, error) {
	agent := new(AgentSpec)
	key := AgentKey{
		"Agent",
		id,
	}
	value, err := c.db.Find(c.storeName, key, c.tag)

	if err != nil {
		return nil, errors.Wrap(err, "GetAgents failed")
	}
	if value == nil || len(value) == 0 {
		return nil, errors.Errorf("Agent with id: %v not found", id)
	}
	if err := c.db.Unmarshal(value[0], agent); err != nil {
		return nil, err
	}
	return agent, nil
}

func (c Client) DeleteAgent(_ context.Context, id string) error {
	key := AgentKey{
		"Agent",
		id,
	}
	err := c.db.Remove(c.storeName, key)
	if err != nil {
		return errors.Wrap(err, "DeleteAgent failed")
	}
	c.agentStream <- StreamAgentData{
		Spec: AgentSpec{
			Id:       id,
			EndPoint: "",
		},
		Operation: "DELETE",
	}
	return nil
}
