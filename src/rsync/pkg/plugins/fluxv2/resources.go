// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package fluxv2

import (
	"context"

	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/status"

	"github.com/fluxcd/go-git-providers/gitprovider"
	emcogithub "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/gitops/emcogithub"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/internal/utils"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func (p *Fluxv2Provider) getPath(t string) string {
	return "clusters/" + p.cluster + "/" + t + "/" + p.cid + "/app/" + p.app + "/"
}

func convertToCommitFile(ref interface{}) []gitprovider.CommitFile {
	var exists bool
	switch ref.(type) {
	case []gitprovider.CommitFile:
		exists = true
	default:
		exists = false
	}
	var rf []gitprovider.CommitFile
	// Create rf is doesn't exist
	if !exists {
		rf = []gitprovider.CommitFile{}
	} else {
		rf = ref.([]gitprovider.CommitFile)
	}
	return rf
}

// Creates a new resource if the not already existing
func (p *Fluxv2Provider) Create(name string, ref interface{}, content []byte) (interface{}, error) {
	path := p.getPath("context") + name + ".yaml"
	rf := convertToCommitFile(ref)
	ref = emcogithub.Add(path, string(content), rf)
	return ref, nil
}

// Apply resource to the cluster
func (p *Fluxv2Provider) Apply(name string, ref interface{}, content []byte) (interface{}, error) {
	//Decode the yaml to create a runtime.Object
	unstruct := &unstructured.Unstructured{}
	//Ignore the returned obj as we expect the data in unstruct
	_, err := utils.DecodeYAMLData(string(content), unstruct)
	if err != nil {
		return nil, err
	}
	// Set Namespace
	unstruct.SetNamespace(p.namespace)
	b, err := unstruct.MarshalJSON()
	if err != nil {
		return nil, err
	}

	path := p.getPath("context") + name + ".yaml"
	rf := convertToCommitFile(ref)
	ref = emcogithub.Add(path, string(b), rf)
	return ref, nil

}

// Delete resource from the cluster
func (p *Fluxv2Provider) Delete(name string, ref interface{}, content []byte) (interface{}, error) {
	path := p.getPath("context") + name + ".yaml"
	rf := convertToCommitFile(ref)
	ref = emcogithub.Delete(path, rf)
	return ref, nil

}

// Get resource from the cluster
func (p *Fluxv2Provider) Get(name string, gvkRes []byte) ([]byte, error) {

	return []byte{}, nil
}

// Commit resources to the cluster
func (p *Fluxv2Provider) Commit(ctx context.Context, ref interface{}) error {
	var exists bool
	switch ref.(type) {
	case []gitprovider.CommitFile:
		exists = true
	default:
		exists = false

	}
	// Check for rf
	if !exists {
		log.Error("Commit: No ref found", log.Fields{})
		return nil
	}
	err := emcogithub.CommitFiles(ctx, p.client, p.userName, p.repoName, p.branch, "Commit for "+p.getPath("context"), ref.([]gitprovider.CommitFile))

	return err
}

// IsReachable cluster reachablity test
func (p *Fluxv2Provider) IsReachable() error {
	return nil
}

func (m *Fluxv2Provider) TagResource(res []byte, label string) ([]byte, error) {
	b, err := status.TagResource(res, label)
	if err != nil {
		log.Error("Error Tag Resoruce with label:", log.Fields{"err": err, "label": label, "resource": res})
		return nil, err
	}
	return b, nil
}
