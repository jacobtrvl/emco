// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package fluxv2

import (
	"context"

	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/internal/utils"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/status"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Creates a new resource if the not already existing
func (p *Fluxv2Provider) Create(name string, ref interface{}, content []byte) (interface{}, error) {

	res, err := p.gitProvider.Create(name, ref, content)
	return res, err
}

// Apply resource to the cluster
func (p *Fluxv2Provider) Apply(ctx context.Context, name string, ref interface{}, content []byte) (interface{}, error) {

	//Decode the yaml to create a runtime.Object
	unstruct := &unstructured.Unstructured{}
	//Ignore the returned obj as we expect the data in unstruct
	_, err := utils.DecodeYAMLData(string(content), unstruct)
	if err != nil {
		return nil, err
	}
	if unstruct.GetNamespace() == "" {
		if unstruct.GetKind() != "Namespace" {
			// Set Namespace
			unstruct.SetNamespace(p.gitProvider.Namespace)
		}
	}
	b, err := unstruct.MarshalJSON()
	if err != nil {
		return nil, err
	}
	path := p.gitProvider.GetPath("context") + name + ".yaml"
	res, err := p.gitProvider.Apply(path, ref, b)
	return res, err

}

// Delete resource from the cluster
func (p *Fluxv2Provider) Delete(name string, ref interface{}, content []byte) (interface{}, error) {

	path := p.gitProvider.GetPath("context") + name + ".yaml"
	res, err := p.gitProvider.Delete(path, ref, content)
	return res, err

}

// Get resource from the cluster
func (p *Fluxv2Provider) Get(ctx context.Context, name string, gvkRes []byte) ([]byte, error) {

	return []byte{}, nil
}

// Commit resources to the cluster
func (p *Fluxv2Provider) Commit(ctx context.Context, ref interface{}) error {

	err := p.gitProvider.Commit(ctx, ref)
	return err
}

// IsReachable cluster reachablity test
func (p *Fluxv2Provider) IsReachable() error {
	return nil
}

func (m *Fluxv2Provider) TagResource(res []byte, label map[string]string) ([]byte, error) {
	b, err := status.TagResource(res, label)
	if err != nil {
		log.Error("Error Tag Resoruce with label:", log.Fields{"err": err, "label": label, "resource": res})
		return nil, err
	}
	return b, nil
}
