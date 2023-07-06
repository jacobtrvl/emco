// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package utils

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"strings"

	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	pkgerrors "github.com/pkg/errors"
	yaml "gopkg.in/yaml.v3"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/common"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

// ListYamlStruct is applied when the kind is list
type ListYamlStruct struct {
	APIVersion string       `yaml:"apiVersion,omitempty"`
	Kind       string       `yaml:"kind,omitempty"`
	items      []YamlStruct `yaml:"items,omitempty"`
}

// YamlStruct represents normal parameters in a manifest file.
// Over the course of time, Pls add more parameters as and when you require.
type YamlStruct struct {
	APIVersion string `yaml:"apiVersion,omitempty"`
	Kind       string `yaml:"kind,omitempty"`
	Metadata   struct {
		Name      string `yaml:"name,omitempty"`
		Namespace string `yaml:"namespace,omitempty"`
		Labels    struct {
			RouterDeisIoRoutable string `yaml:"router.deis.io/routable,omitempty"`
		} `yaml:"labels"`
		Annotations struct {
			RouterDeisIoDomains string `yaml:"router.deis.io/domains,omitempty"`
		} `yaml:"annotations,omitempty"`
	} `yaml:"metadata,omitempty"`
	Spec struct {
		Type     string `yaml:"type,omitempty"`
		Selector struct {
			App string `yaml:"app,omitempty"`
		} `yaml:"selector,omitempty"`
		Ports []struct {
			Name     string `yaml:"name,omitempty"`
			Port     int    `yaml:"port,omitempty"`
			NodePort int    `yaml:"nodePort,omitempty"`
		} `yaml:"ports"`
	} `yaml:"spec"`
}

func (y YamlStruct) isValid() bool {
	if y.APIVersion == "" {
		log.Info("apiVersion is missing in manifest file", log.Fields{})
		return false
	}
	if y.Kind == "" {
		log.Info("kind is missing in manifest file", log.Fields{})
		return false
	}
	if y.Metadata.Name == "" {
		log.Info("metadata.name is missing in manifest file", log.Fields{})
		return false
	}
	return true
}

// ExtractYamlParameters is a method which takes in the abolute path of a manifest file
// and returns a struct accordingly
func ExtractYamlParameters(f string) (YamlStruct, error) {
	filename, _ := filepath.Abs(f)
	yamlFile, err := ioutil.ReadFile(filename)

	var yamlStruct YamlStruct

	err = yaml.Unmarshal(yamlFile, &yamlStruct)
	if err != nil {
		return YamlStruct{}, pkgerrors.New("Cant unmarshal yaml file ..")
	}

	/* This is a special case handling when the kind is "List".
	When the kind is list and the metadata name is empty.
	We set the metadata name as the file name. For eg:
	if filename is "/tmp/helm-tmpl-240995533/prometheus/templates/serviceaccount.yaml-0".
	We set metadata name as "serviceaccount.yaml-0"
	Usually when the kind is list, the list might contains a list of
	*/
	if yamlStruct.Kind == "List" && yamlStruct.Metadata.Name == "" {
		li := strings.LastIndex(filename, "/")
		fn := string(filename[li+1:])
		yamlStruct.Metadata.Name = fn
		log.Info("Setting the metadata", log.Fields{"MetaDataName": fn})
	}
	if yamlStruct.isValid() {
		log.Info(":: YAML parameters ::", log.Fields{"fileName": f, "yamlStruct": yamlStruct})

		return yamlStruct, nil
	}
	log.Info("YAML file has errors", log.Fields{"fileName": f})
	return YamlStruct{}, pkgerrors.Errorf("Cant extract parameters from yaml file :: %s", filename)

}

// ExtractTarBall provides functionality to extract a tar.gz file
// into a temporary location for later use.
// It returns the path to the new location
func ExtractTarBall(r io.Reader) (string, error) {
	//Check if it is a valid gz
	gzf, err := gzip.NewReader(r)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Invalid gzip format")
	}

	//Check if it is a valid tar file
	//Unfortunately this can only be done by inspecting all the tar contents
	tarR := tar.NewReader(gzf)
	first := true

	outDir, _ := ioutil.TempDir("", "k8s-ext-")

	for true {
		header, err := tarR.Next()

		if err == io.EOF {
			//Check if we have just a gzip file without a tar archive inside
			if first {
				return "", pkgerrors.New("Empty or non-existant Tar file found")
			}
			//End of archive
			break
		}

		if err != nil {
			return "", pkgerrors.Wrap(err, "Error reading tar file")
		}

		target := filepath.Join(outDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				// Using 755 read, write, execute for owner
				// groups and others get read and execute permissions
				// on the folder.
				if err := os.MkdirAll(target, 0700); err != nil {
					return "", pkgerrors.Wrap(err, "Creating directory")
				}
			}
		case tar.TypeReg:
			if target == outDir { // Handle '.' substituted to '' entry
				continue
			}

			err = EnsureDirectory(target)
			if err != nil {
				return "", pkgerrors.Wrap(err, "Creating Directory")
			}

			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return "", pkgerrors.Wrap(err, "Creating file")
			}

			// copy over contents
			if _, err := io.Copy(f, tarR); err != nil {
				return "", pkgerrors.Wrap(err, "Copying file content")
			}

			// close for each file instead of a defer for all
			// at the end of the function
			f.Close()
		}

		first = false
	}

	return outDir, nil
}

// EnsureDirectory makes sure that the directories specified in the path exist
// If not, it will create them, if possible.
func EnsureDirectory(f string) error {
	base := path.Dir(f)
	_, err := os.Stat(base)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return os.MkdirAll(base, 0700)
}

func ConvertType(in, out interface{}) error {
	b, err := json.Marshal(in)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, out)
}

func ContainCluster(clusterProvider, clusterName string, lcClusterRefs []common.Cluster) bool {
	for _, lcRef := range lcClusterRefs {
		if lcRef.Specification.ClusterProvider == clusterProvider && lcRef.Specification.ClusterName == clusterName {
			return true
		}
	}

	return false
}

func StructToMap(s interface{}) (map[string]interface{}, error) {
	var m map[string]interface{}
	return m, ConvertType(s, &m)
}

func MapKeys(m map[string]interface{}) []string {
	var keys []string
	for key := range m {
		keys = append(keys, key)
	}

	return keys
}

func HasDuplication(list []string) bool {
	m := map[string]bool{}
	for _, item := range list {
		if _, ok := m[item]; ok {
			return true
		}

		m[item] = true
	}

	return false
}

func HasIntersection(list1 []string, list2 []string) bool {
	m := map[string]bool{}
	for _, item := range list1 {
		m[item] = true
	}

	for _, item := range list2 {
		if _, ok := m[item]; ok {
			return true
		}
	}

	return false
}

func ListDifference(list1 []string, list2 []string) []string {
	m := map[string]bool{}
	for _, item := range list2 {
		m[item] = true
	}

	difference := []string{}
	for _, item := range list1 {
		if _, ok := m[item]; !ok {
			difference = append(difference, item)
		}
	}

	return difference
}
