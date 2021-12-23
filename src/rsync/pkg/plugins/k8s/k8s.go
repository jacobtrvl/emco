// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package k8s

import (
	//"fmt"
	"io/ioutil"
	"os"

	pkgerrors "github.com/pkg/errors"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	//. "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/types"

	"encoding/base64"
	"strings"

	kubeclient "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/client"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/db"
)

// Connection is for a cluster
type K8sProvider struct {
	cid       string
	cluster   string
	app       string
	namespace string
	level     string
	fileName  string
	client    *kubeclient.Client
}

var GetKubeConfig = func(clustername string, level string, namespace string) ([]byte, error) {
	if !strings.Contains(clustername, "+") {
		return nil, pkgerrors.New("Not a valid cluster name")
	}
	strs := strings.Split(clustername, "+")
	if len(strs) != 2 {
		return nil, pkgerrors.New("Not a valid cluster name")
	}

	ccc := db.NewCloudConfigClient()

	// log.Info("Querying CloudConfig", log.Fields{"strs": strs, "level": level, "namespace": namespace})
	cconfig, err := ccc.GetCloudConfig(strs[0], strs[1], level, namespace)
	if err != nil {
		return nil, pkgerrors.New("Get kubeconfig failed")
	}
	log.Info("CloudConfig found", log.Fields{".Provider": cconfig.Provider, ".Cluster": cconfig.Cluster, ".Level": cconfig.Level, ".Namespace": cconfig.Namespace})

	dec, err := base64.StdEncoding.DecodeString(cconfig.Config)
	if err != nil {
		return nil, err
	}
	return dec, nil
}

func NewK8sProvider(cid, app, cluster, level, namespace string) (*K8sProvider, error) {
	p := K8sProvider{
		cid:       cid,
		app:       app,
		cluster:   cluster,
		level:     level,
		namespace: namespace,
	}
	// Get file from DB
	dec, err := GetKubeConfig(cluster, level, namespace)
	if err != nil {
		return nil, err
	}
	//
	f, err := ioutil.TempFile("/tmp", "rsync-config-"+cluster+"-")
	if err != nil {
		log.Error("Unable to create temp file in tmp directory", log.Fields{"err": err})
		return nil, err
	}
	fileName := f.Name()
	log.Info("Temp file for Kubeconfig", log.Fields{"fileName": fileName})
	_, err = f.Write(dec)
	if err != nil {
		log.Error("Unable to write tmp directory", log.Fields{"err": err, "filename": fileName})
		return nil, err
	}

	client := kubeclient.New("", fileName, namespace)
	if client == nil {
		return nil, pkgerrors.New("failed to connect with the cluster")
	}
	p.fileName = fileName
	p.client = client
	return &p, nil
}

// If file exists delete it
func (p *K8sProvider) CleanClientProvider() error {
	if _, err := os.Stat(p.fileName); err == nil {
		os.Remove(p.fileName)
	}
	return nil

}
