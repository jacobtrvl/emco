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

package controller

import (
	"context"
	event "emcopolicy/internal/events"
	"emcopolicy/internal/intent"
	events "emcopolicy/pkg/grpc"
	"emcopolicy/pkg/plugins"
	"github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

func Init(_ ...string) (*Controller, error) {
	err := db.InitializeDatabaseConnection("emco")
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, errors.Errorf("Unable to initialize mongo database connection: %s", err)
	}

	// DB connection is a package level variable (db.DBconn) in orchestrator db package.
	// Scoping this to the client context for better readability
	c := &Controller{
		db:        db.DBconn,
		tag:       "data",
		storeName: "resources",
		reverseMap: &ReverseMap{
			eventMap: make(map[intent.Event][]intent.Intent),
		},
		agentMap: &AgentMap{
			runtime: make(map[AgentID]*AgentRuntime),
		},
		updateStream:    make(chan intent.StreamData),
		eventStream:     make(chan event.Event),
		agentStream:     make(chan event.StreamAgentData),
		requireRecovery: make(chan interface{}),
		eventsQueue:     make(chan *events.Event, 1),
		actors:          make(map[string]event.Actor),
	}
	c.policyClient = intent.NewClient(intent.Config{
		Db:           c.db,
		Tag:          c.tag,
		StoreName:    c.storeName,
		UpdateStream: c.updateStream,
	})
	c.eventClient = event.NewClient(event.Config{
		Db:          c.db,
		Tag:         c.tag,
		StoreName:   c.storeName,
		AgentStream: c.agentStream,
	})
	c.actors["temporal"] = new(plugins.TemporalActor)

	key := Module{"Agent"}
	err = c.db.Insert(c.storeName, key, nil, c.tag, key)
	if err != nil {
		return nil, errors.Errorf("Error while Initializing DB %s", err)
	}
	return c, nil
}

func (c *Controller) StartScheduler(ctx context.Context) error {
	err := c.BuildReverseMap(ctx)
	if err != nil {
		return errors.Wrap(err, "Starting OperationalScheduler failed")
	}
	// Starting AgentManager before the call BuildAgentMap to avoid
	// requireRecovery channel getting blocked
	go c.AgentManager(ctx)
	err = c.BuildAgentMap(ctx)
	if err != nil {
		return errors.Wrap(err, "Starting OperationalScheduler failed")
	}
	go c.OperationalScheduler(ctx)
	go c.EventsManager(ctx)
	return nil
}

func (c Controller) PolicyClient() *intent.Client {
	return c.policyClient
}

func (c Controller) EventClient() *event.Client {
	return c.eventClient
}
