// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package fluxv2

import (
	"context"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"time"
	emcogithub "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/gitops/emcogithub"
	v1alpha1 "gitlab.com/project-emco/core/emco-base/src/monitor/pkg/apis/k8splugin/v1alpha1"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/status"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/internal/utils"
	"github.com/fluxcd/go-git-providers/gitprovider"
)

// 30 seconds wait time
var waitTime int = 30
// StartClusterWatcher watches for CR
func (p *Fluxv2Provider) StartClusterWatcher() error {
	// Start thread to sync monitor CR
	go func() error {
		ctx := context.Background()
		for {
			select {
			case <-time.After(time.Duration(waitTime) * time.Second):
				if ctx.Err() != nil {
					return ctx.Err()
				}
				// Check if AppContext doesn't exist then exit the thread
				if _, err := utils.NewAppContextReference(p.cid); err != nil {
					// Delete the status CR
					path :=  p.getPath("status") + p.cid + "-" + p.app
					rf := []gitprovider.CommitFile{}		
					ref := emcogithub.Delete(path, rf)
					err = emcogithub.CommitFiles(context.Background(), p.client, p.userName, p.repoName, p.branch, 
					"Commit for Delete Status CR "+p.getPath("status"), ref)
					// Exit thread
					return err
				}
				path :=  p.getPath("status")
				// Read file
				cp, err := emcogithub.GetFiles(ctx, p.client, p.userName, p.repoName, p.branch, path)
				if err != nil {
					log.Error("", log.Fields{"error": err, "cluster": p.cluster, "resource": path})
					continue
				}
				log.Info("Fluxv2 read status", log.Fields{"id": p.cid, "cluster": p.cluster,  "length": len(cp), "content": cp})
				if len(cp) > 0 {
					// Only one file expected in the location
					content := &v1alpha1.ResourceBundleState{}	
					_, err := utils.DecodeYAMLData(*cp[0].Content, content)
					if err != nil {
						log.Error("", log.Fields{"error": err, "cluster": p.cluster, "resource": path})
						return err
					}
					log.Info("Fluxv2 read status", log.Fields{"id": p.cid, "cluster": p.cluster,  "content": content})
					status.HandleResourcesStatus(p.cid, p.app, p.cluster, content)
				}
			// Check if the context is canceled
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}()
	return nil
}

// ApplyStatusCR applies status CR
func (p *Fluxv2Provider) ApplyStatusCR(name string, content []byte) error {
	ref, err := p.Apply(name, nil, content)
	if err != nil {
		return err
	}
	err = emcogithub.CommitFiles(context.Background(), p.client, p.userName, p.repoName, p.branch, "Commit for Apply Status CR"+p.getPath("context"), ref.([]gitprovider.CommitFile))

	return err

}

// DeleteStatusCR deletes status CR
func (p *Fluxv2Provider) DeleteStatusCR(name string, content []byte) error {
	ref, err := p.Delete(name, nil, content)
	if err != nil {
		return err
	}
	err = emcogithub.CommitFiles(context.Background(), p.client, p.userName, p.repoName, p.branch, "Commit for Delete Status CR"+p.getPath("context"), ref.([]gitprovider.CommitFile))

	return err
}
