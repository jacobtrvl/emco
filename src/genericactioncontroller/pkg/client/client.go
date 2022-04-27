// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package client

import "gitlab.com/project-emco/core/emco-base/src/genericactioncontroller/pkg/module"

// Client for using the services
type Client struct {
	// Add Clients for API's here
	Customization    *module.CustomizationClient
	GenericK8sIntent *module.GenericK8sIntentClient
	Resource         *module.ResourceClient
}

// NewClient creates a new client for using the services
func NewClient() *Client {
	c := &Client{}
	c.GenericK8sIntent = module.NewGenericK8sIntentClient()
	c.Resource = module.NewResourceClient()
	c.Customization = module.NewCustomizationClient()
	return c
}
