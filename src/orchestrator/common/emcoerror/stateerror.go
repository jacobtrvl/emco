// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package emcoerror

import (
	"fmt"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
)

// StateError defines the resource state errors
type StateError struct {
	Resource string // Resource Type e.g: LogicalCloud, DeploymentIntentGroup, CaCert etc.
	Event    string // Life Cycle Event e.g: Instantiate, Terminate etc.
	Status   appcontext.StatusValue
}

// Error implements the error interface
func (e *StateError) Error() string {
	switch e.Status {
	case appcontext.AppContextStatusEnum.Terminating:
		return fmt.Sprintf("Failed to %s. The %s is being terminated", e.Event, e.Resource)
	case appcontext.AppContextStatusEnum.Instantiating:
		return fmt.Sprintf("Failed to %s. The %s is in instantiating status", e.Event, e.Resource)
	case appcontext.AppContextStatusEnum.TerminateFailed:
		return fmt.Sprintf("Failed to %s. The %s has failed terminating, please delete the %s", e.Event, e.Resource, e.Resource)
	case appcontext.AppContextStatusEnum.Terminated:
		// handle events specific use cases
		switch e.Event {
		case "Terminate":
			return fmt.Sprintf("The %s is already terminated", e.Resource)
		}
	case appcontext.AppContextStatusEnum.Instantiated:
		switch e.Event {
		case "Instantiate":
			return fmt.Sprintf("The %s is already instantiated", e.Resource)
		}
	case appcontext.AppContextStatusEnum.InstantiateFailed:
		switch e.Event {
		case "Instantiate":
			return fmt.Sprintf("The %s has failed instantiating before, please terminate and try again", e.Resource)
		}
	}

	return fmt.Sprintf("The %s isn't in an expected status so not taking any action", e.Resource)
}
