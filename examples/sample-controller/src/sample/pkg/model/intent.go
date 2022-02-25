// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

// Package model contains the data model necessary for the implementations.
// In this example, SampleIntent
package model

// SampleIntent defines the high level structure of a network chain document
type SampleIntent struct {
	Metadata Metadata         `json:"metadata" yaml:"metadata"`
	Spec     SampleIntentSpec `json:"spec" yaml:"spec"`
}

type Metadata struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"-"`
	UserData1   string `json:"userData1" yaml:"-"`
	UserData2   string `json:"userData2" yaml:"-"`
}

// SampleIntentSpec contains the specification of a network chain
type SampleIntentSpec struct {
	App              string `json:"app"`
	SampleIntentData string `json:"sampleIntentData"`
}

// SampleIntentKey is the key structure that is used in the database
type SampleIntentKey struct {
	Project               string `json:"project"`
	CompositeApp          string `json:"compositeApp"`
	CompositeAppVersion   string `json:"compositeAppVersion"`
	DeploymentIntentGroup string `json:"deploymentIntentGroup"`
	SampleIntent          string `json:"sampleIntent"`
}
