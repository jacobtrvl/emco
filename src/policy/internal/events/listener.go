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
	eventpb "emcopolicy/pkg/grpc"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ListenOne connects with the agent, and listens for any new event updates
// from the agent. This acts as a rpc client for the agents.
func ListenOne(ctx context.Context, addr string, markForRecover func(), eventStream chan *eventpb.Event) {
	defer markForRecover()
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	//conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatal("Failed to connect %v", log.Fields{"err": err})
	}
	defer conn.Close()
	c := eventpb.NewEventsClient(conn)
	client, err := c.EventUpdate(ctx, &eventpb.ServerSpec{ServerId: 1003})
	if err != nil || client == nil {
		log.Error("Couldn't connect to client: ", log.Fields{"Err": err})
		return
	}
	// Listen for events from agents
	for {
		m, err := client.Recv()
		if err != nil {
			log.Error("Agent Receiver error \n", log.Fields{"Err": err})
			break
		}
		log.Debug("New event received", log.Fields{"mesg": m})
		// Put the event to eventStream.
		// Controller consumes the event from the stream for further processing.
		eventStream <- m
	}
}
