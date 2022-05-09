// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package action

import (
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

// UpdateAppContext creates/updates the k8s resources in the given appContext and intent
func UpdateAppContext(intent, appContextID string) error {
	log.Info("Begin app context update",
		log.Fields{
			"AppContext": appContextID,
			"Intent":     intent})

	return nil
}
