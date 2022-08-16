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
	"github.com/pkg/errors"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

func DoAction(actorName string, evaluationResult []byte, intentSpec []byte, agentSpec []byte, actors map[string]Actor) error {
	var (
		actor Actor
	)
	switch actorName {
	case "temporal":
		actor = actors["temporal"]
	default:
		log.Error("DoAction: No Actor Plugin matched", log.Fields{"actor": actorName})
		return errors.Errorf("DoAction: No Actor Plugin matched with provided actor name: %s", actorName)
	}
	return actor.Execute(evaluationResult, intentSpec, agentSpec)
}
