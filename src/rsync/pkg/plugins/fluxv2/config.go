// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package fluxv2

import (
	"context"
	"github.com/fluxcd/go-git-providers/gitprovider"
	kustomize "github.com/fluxcd/kustomize-controller/api/v1beta2"
	fluxsc "github.com/fluxcd/source-controller/api/v1beta1"
	yaml "github.com/ghodss/yaml"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	emcogithub "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/gitops/emcogithub"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

// Create GitRepository and Kustomization CR's for Flux
func (p *Fluxv2Provider) ApplyConfig(ctx context.Context, config interface{}) error {

	// Create Source CR and Kcustomize CR
	gr := fluxsc.GitRepository{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "source.toolkit.fluxcd.io/v1beta1",
			Kind:       "GitRepository",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.cid,
			Namespace: p.namespace,
		},
		Spec: fluxsc.GitRepositorySpec{
			URL:       p.url,
			Interval:  metav1.Duration{Duration: time.Second * 30},
			Reference: &fluxsc.GitRepositoryRef{Branch: p.branch},
		},
	}
	x, err := yaml.Marshal(&gr)
	if err != nil {
		log.Error("ApplyConfig:: Marshal err", log.Fields{"err": err, "gr": gr})
		return err
	}
	path := "clusters/" + p.cluster + "/" + gr.Name + ".yaml"
	// Add to the commit
	gp := emcogithub.Add(path, string(x), []gitprovider.CommitFile{})

	kc := kustomize.Kustomization{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kustomize.toolkit.fluxcd.io/v1beta2",
			Kind:       "Kustomization",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kcust" + p.cid,
			Namespace: p.namespace,
		},
		Spec: kustomize.KustomizationSpec{
			Interval: metav1.Duration{Duration: time.Second * 300},
			Path:     "clusters/" + p.cluster + "/context/" + p.cid,
			Prune:    true,
			SourceRef: kustomize.CrossNamespaceSourceReference{
				Kind: "GitRepository",
				Name: gr.Name,
			},
			TargetNamespace: p.namespace,
		},
	}
	y, err := yaml.Marshal(&kc)
	if err != nil {
		log.Error("ApplyConfig:: Marshal err", log.Fields{"err": err, "kc": kc})
		return err
	}
	path = "clusters/" + p.cluster + "/" + kc.Name + ".yaml"
	// Add to the commit
	gp = emcogithub.Add(path, string(y), gp)
	// Commit
	err = emcogithub.CommitFiles(ctx, p.client, p.userName, p.repoName, p.branch, "Commit for "+p.getPath("context"), gp)
	if err != nil {
		log.Error("ApplyConfig:: Commit files err", log.Fields{"err": err, "gp": gp})
	}
	return err
}

// Delete GitRepository and Kustomization CR's for Flux
func (p *Fluxv2Provider) DeleteConfig(ctx context.Context, config interface{}) error {
	path := "clusters/" + p.cluster + "/" + p.cid + ".yaml"
	gp := emcogithub.Delete(path, []gitprovider.CommitFile{})
	path = "clusters/" + p.cluster + "/" + "kcust" + p.cid + ".yaml"
	gp = emcogithub.Delete(path, gp)
	err := emcogithub.CommitFiles(ctx, p.client, p.userName, p.repoName, p.branch, "Commit for "+p.getPath("context"), gp)
	if err != nil {
		log.Error("ApplyConfig:: Commit files err", log.Fields{"err": err, "gp": gp})
	}
	return err
}
