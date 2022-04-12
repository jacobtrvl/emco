package module

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

import (
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

// Client for using the services in the ncm
type Client struct {

	// Add Clients for API's here

	BaseResType       *ResTypeClient
	BaseAppConfigType *AppConfigClient
}

// NewClient creates a new client for using the services
func NewClient() *Client {
	c := &Client{}
	c.BaseResType = NewResTypeClient()
	c.BaseAppConfigType = NewAppConfigClient()
	log.Info("Setting the client!", log.Fields{})
	// Add Client API handlers here
	return c
}
