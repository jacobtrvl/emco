// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package resources

import (
	"context"
	"fmt"
	"strings"

	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/status"

	"github.com/fluxcd/go-git-providers/gitprovider"
	pkgerrors "github.com/pkg/errors"
	clm "gitlab.com/project-emco/core/emco-base/src/clm/pkg/cluster"
	emcogit "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/gitops/emcogit"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/internal/utils"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type GitProvider struct {
	Cid       string
	Cluster   string
	App       string
	Namespace string
	Level     string
	GitType   string
	GitToken  string
	UserName  string
	Branch    string
	RepoName  string
	Url       string
	Client    gitprovider.Client
}

func NewGitProvider(cid, app, cluster, level, namespace string) (*GitProvider, error) {
	result := strings.SplitN(cluster, "+", 2)
	cc := clm.NewClusterClient()
	c, err := cc.GetCluster(result[0], result[1])
	if err != nil {
		return nil, err
	}

	refObject, err := cc.GetClusterSyncObjects(result[0], c.Spec.Props.GitOpsReferenceObject)
	if err != nil {
		log.Error("Invalid refObject :", log.Fields{"refObj": c.Spec.Props.GitOpsReferenceObject})
		return nil, pkgerrors.Errorf("Invalid refObject: " + c.Spec.Props.GitOpsReferenceObject)
	}

	kvRef := refObject.Spec.Kv

	var gitType, gitToken, branch, userName, repoName string

	for _, kvpair := range kvRef {
		log.Info("kvpair", log.Fields{"kvpair": kvpair})
		v, ok := kvpair["gitType"]
		if ok {
			gitType = fmt.Sprintf("%v", v)
			continue
		}
		v, ok = kvpair["gitToken"]
		if ok {
			gitToken = fmt.Sprintf("%v", v)
			continue
		}
		v, ok = kvpair["repoName"]
		if ok {
			repoName = fmt.Sprintf("%v", v)
			continue
		}
		v, ok = kvpair["userName"]
		if ok {
			userName = fmt.Sprintf("%v", v)
			continue
		}
		v, ok = kvpair["branch"]
		if ok {
			branch = fmt.Sprintf("%v", v)
			continue
		}
	}
	if len(gitType) <= 0 || len(gitToken) <= 0 || len(branch) <= 0 || len(userName) <= 0 || len(repoName) <= 0 {
		log.Error("Missing information for Github", log.Fields{"gitType": gitType, "token": gitToken, "branch": branch, "userName": userName, "repoName": repoName})
		return nil, pkgerrors.Errorf("Missing Information for Github")
	}

	p := GitProvider{
		Cid:       cid,
		App:       app,
		Cluster:   cluster,
		Level:     level,
		Namespace: namespace,
		GitType:   gitType,
		GitToken:  gitToken,
		Branch:    branch,
		UserName:  userName,
		RepoName:  repoName,
		Url:       "https://" + gitType + ".com/" + userName + "/" + repoName,
	}
	client, err := emcogit.CreateClient(gitToken, gitType)
	if err != nil {
		log.Error("Error getting github client", log.Fields{"err": err})
		return nil, err
	}
	p.Client = client.(gitprovider.Client)
	return &p, nil
}

func (p *GitProvider) GetPath() string {
	return "clusters/" + p.Cluster + "/context/" + p.Cid + "/app/" + p.App + "/"
}

// func convertToCommitFile(ref interface{}) []gitprovider.CommitFile {
// 	var exists bool
// 	switch ref.(type) {
// 	case []gitprovider.CommitFile:
// 		exists = true
// 	default:
// 		exists = false
// 	}
// 	var rf []gitprovider.CommitFile
// 	// Create rf is doesn't exist
// 	if !exists {
// 		rf = []gitprovider.CommitFile{}
// 	} else {
// 		rf = ref.([]gitprovider.CommitFile)
// 	}
// 	return rf
// }

// Creates a new resource if the not already existing
func (p *GitProvider) Create(name string, ref interface{}, content []byte) (interface{}, error) {
	// Add the label based on the Status Appcontext ID
	label := p.Cid + "-" + p.App
	b, err := status.TagResource(content, label)
	if err != nil {
		log.Error("Error Tag Resoruce with label:", log.Fields{"err": err, "label": label, "resource": name})
		return nil, err
	}
	path := p.GetPath() + name + ".yaml"

	ref = emcogit.Add(path, string(b), ref, p.GitType)
	return ref, nil
}

// Apply resource to the cluster
func (p *GitProvider) Apply(name string, ref interface{}, content []byte) (interface{}, error) {
	// Add the label based on the Status Appcontext ID
	label := p.Cid + "-" + p.App

	//Decode the yaml to create a runtime.Object
	unstruct := &unstructured.Unstructured{}
	//Ignore the returned obj as we expect the data in unstruct
	_, err := utils.DecodeYAMLData(string(content), unstruct)
	if err != nil {
		return nil, err
	}
	//Add the tracking label to all resources created here
	labels := unstruct.GetLabels()
	//Check if labels exist for this object
	if labels == nil {
		labels = map[string]string{}
	}
	// Set label
	labels["emco/deployment-id"] = label
	unstruct.SetLabels(labels)
	// Set Namespace
	unstruct.SetNamespace(p.Namespace)

	// This checks if the resource we are creating has a podSpec in it
	// Eg: Deployment, StatefulSet, Job etc..
	// If a PodSpec is found, the label will be added to it too.
	//connector.TagPodsIfPresent(unstruct, client.GetInstanceID())
	status.TagPodsIfPresent(unstruct, label)
	b, err := unstruct.MarshalJSON()
	if err != nil {
		return nil, err
	}

	path := p.GetPath() + name + ".yaml"

	ref = emcogit.Add(path, string(b), ref, p.GitType)
	return ref, nil

}

// Delete resource from the cluster
func (p *GitProvider) Delete(name string, ref interface{}, content []byte) (interface{}, error) {
	path := p.GetPath() + name + ".yaml"
	ref = emcogit.Delete(path, ref, p.GitType)
	return ref, nil

}

// Get resource from the cluster
func (p *GitProvider) Get(name string, gvkRes []byte) ([]byte, error) {

	return []byte{}, nil
}

// Commit resources to the cluster
func (p *GitProvider) Commit(ctx context.Context, ref interface{}) error {
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
	err := emcogit.CommitFiles(ctx, p.Client, p.UserName, p.RepoName, p.Branch, "Commit for "+p.GetPath(), ref.([]gitprovider.CommitFile), p.GitType)

	return err
}

// IsReachable cluster reachablity test
func (p *GitProvider) IsReachable() error {
	return nil
}
