// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	pkgerrors "github.com/pkg/errors"
	rb "gitlab.com/project-emco/core/emco-base/src/monitor/pkg/apis/k8splugin/v1alpha1"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"
	rsync "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/db"
	"gopkg.in/yaml.v2"
)

// Cluster contains the parameters needed for a Cluster
type Cluster struct {
	MetaData      ClusterMeta `json:"metadata"`
	Specification ClusterSpec `json:"spec"`
}

type ClusterMeta struct {
	ClusterReference string `json:"name"`
	Description      string `json:"description"`
	UserData1        string `json:"userData1"`
	UserData2        string `json:"userData2"`
}

type ClusterSpec struct {
	ClusterProvider string `json:"clusterProvider"`
	ClusterName     string `json:"cluster"`
	LoadBalancerIP  string `json:"loadBalancerIP"`
	Certificate     string `json:"certificate"`
}

type ClusterKey struct {
	Project          string `json:"project"`
	LogicalCloudName string `json:"logicalCloud"`
	ClusterReference string `json:"clusterReference"`
}

type KubeConfig struct {
	ApiVersion     string            `yaml:"apiVersion"`
	Kind           string            `yaml:"kind"`
	Clusters       []KubeCluster     `yaml:"clusters"`
	Contexts       []KubeContext     `yaml:"contexts"`
	CurrentContext string            `yaml:"current-context"`
	Preferences    map[string]string `yaml:"preferences"`
	Users          []KubeUser        `yaml:"users"`
}

type KubeCluster struct {
	ClusterDef  KubeClusterDef `yaml:"cluster"`
	ClusterName string         `yaml:"name"`
}

type KubeClusterDef struct {
	CertificateAuthorityData string `yaml:"certificate-authority-data"`
	Server                   string `yaml:"server"`
}

type KubeContext struct {
	ContextDef  KubeContextDef `yaml:"context"`
	ContextName string         `yaml:"name"`
}

type KubeContextDef struct {
	Cluster   string `yaml:"cluster"`
	Namespace string `yaml:"namespace,omitempty"`
	User      string `yaml:"user"`
}

type KubeUser struct {
	UserName string      `yaml:"name"`
	UserDef  KubeUserDef `yaml:"user"`
}

type KubeUserDef struct {
	ClientCertificateData string `yaml:"client-certificate-data"`
	ClientKeyData         string `yaml:"client-key-data"`
	// client-certificate and client-key are NOT implemented
}

// ClusterManager is an interface that exposes the connection
// functionality
type ClusterManager interface {
	CreateCluster(project, logicalCloud string, c Cluster) (Cluster, error)
	GetCluster(project, logicalCloud, name string) (Cluster, error)
	GetAllClusters(project, logicalCloud string) ([]Cluster, error)
	DeleteCluster(project, logicalCloud, name string) error
	UpdateCluster(project, logicalCloud, name string, c Cluster) (Cluster, error)
	GetClusterConfig(project, logicalcloud, name string) (string, error)
}

// ClusterClient implements the ClusterManager
// It will also be used to maintain some localized state
type ClusterClient struct {
	storeName string
	tagMeta   string
}

// ClusterClient returns an instance of the ClusterClient
// which implements the ClusterManager
func NewClusterClient() *ClusterClient {
	return &ClusterClient{
		storeName: "resources",
		tagMeta:   "data",
	}
}

// Create entry for the cluster reference resource in the database
func (v *ClusterClient) CreateCluster(project, logicalCloud string, c Cluster) (Cluster, error) {

	//Construct key consisting of name
	key := ClusterKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		ClusterReference: c.MetaData.ClusterReference,
	}
	lcClient := NewLogicalCloudClient()

	s, err := lcClient.GetState(project, logicalCloud)
	if err != nil {
		return Cluster{}, err
	}
	cid := state.GetLastContextIdFromStateInfo(s)

	if cid != "" {
		ac, err := state.GetAppContextFromId(cid)
		if err != nil {
			return Cluster{}, err
		}

		// since there's a context associated, if the logical cloud isn't fully Terminated then prevent
		// clusters from being added since this is a functional scenario not currently supported
		acStatus, err := GetAppContextStatus(ac)
		if err != nil {
			return Cluster{}, err
		}
		switch acStatus.Status {
		case appcontext.AppContextStatusEnum.Terminated:
			break
		default:
			return Cluster{}, pkgerrors.New("Cluster References cannot be added/removed unless the Logical Cloud is not instantiated")
		}
	}

	//Check if this Cluster Reference already exists
	_, err = v.GetCluster(project, logicalCloud, c.MetaData.ClusterReference)
	if err == nil {
		return Cluster{}, pkgerrors.New("Cluster reference already exists")
	}

	err = db.DBconn.Insert(v.storeName, key, nil, v.tagMeta, c)
	if err != nil {
		return Cluster{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return c, nil
}

// Get returns  Cluster for corresponding cluster reference
func (v *ClusterClient) GetCluster(project, logicalCloud, clusterReference string) (Cluster, error) {

	//Construct the composite key to select the entry
	key := ClusterKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		ClusterReference: clusterReference,
	}

	value, err := db.DBconn.Find(v.storeName, key, v.tagMeta)
	if err != nil {
		return Cluster{}, err
	}

	if len(value) == 0 {
		return Cluster{}, pkgerrors.New("Cluster reference not found")
	}

	//value is a byte array
	if value != nil {
		cl := Cluster{}
		err = db.DBconn.Unmarshal(value[0], &cl)
		if err != nil {
			return Cluster{}, err
		}
		return cl, nil
	}

	return Cluster{}, pkgerrors.New("Unknown Error")
}

// GetAll returns all cluster references in the logical cloud
func (v *ClusterClient) GetAllClusters(project, logicalCloud string) ([]Cluster, error) {
	//Construct the composite key to select clusters
	key := ClusterKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		ClusterReference: "",
	}
	var resp []Cluster
	values, err := db.DBconn.Find(v.storeName, key, v.tagMeta)
	if err != nil {
		return []Cluster{}, err
	}
	if len(values) == 0 {
		return []Cluster{}, pkgerrors.New("No Cluster References associated")
	}

	for _, value := range values {
		cl := Cluster{}
		err = db.DBconn.Unmarshal(value, &cl)
		if err != nil {
			return []Cluster{}, err
		}
		resp = append(resp, cl)
	}

	return resp, nil
}

// Delete the Cluster reference entry from database
func (v *ClusterClient) DeleteCluster(project, logicalCloud, clusterReference string) error {
	//Construct the composite key to select the entry
	key := ClusterKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		ClusterReference: clusterReference,
	}

	lcClient := NewLogicalCloudClient()

	s, err := lcClient.GetState(project, logicalCloud)
	if err != nil {
		return err
	}
	cid := state.GetLastContextIdFromStateInfo(s)

	if cid == "" {
		// Just go ahead and delete the reference if there is no logical cloud context yet
		err := db.DBconn.Remove(v.storeName, key)
		if err != nil {
			return pkgerrors.Wrap(err, "Failed deleting Cluster Reference")
		}
		return nil
	}

	// Make sure rsync status for this logical cloud is Terminated,
	// otherwise prevent the clusters from being removed
	ac, err := state.GetAppContextFromId(cid)
	if err != nil {
		return err
	}
	acStatus, err := GetAppContextStatus(ac)
	if err != nil {
		return pkgerrors.Wrap(err, "Error getting app context status")
	}
	switch acStatus.Status {
	case appcontext.AppContextStatusEnum.Terminating:
		log.Error("Can't remove Cluster Reference yet: the Logical Cloud is being terminated.", log.Fields{"clusterreference": clusterReference, "logicalcloud": logicalCloud})
		return pkgerrors.New("Can't remove Cluster Reference: the Logical Cloud is being terminated.")
	case appcontext.AppContextStatusEnum.Instantiated:
		log.Error("Can't remove Cluster Reference: the Logical Cloud is instantiated, please terminate first.", log.Fields{"clusterreference": clusterReference, "logicalcloud": logicalCloud})
		return pkgerrors.New("Can't remove Cluster Reference: the Logical Cloud is instantiated, please terminate first.")
	case appcontext.AppContextStatusEnum.Instantiating:
		log.Error("Can't remove Cluster Reference: the Logical Cloud is instantiating, please wait and then terminate.", log.Fields{"clusterreference": clusterReference, "logicalcloud": logicalCloud})
		return pkgerrors.New("Can't remove Cluster Reference: the Logical Cloud is instantiating, please wait and then terminate.")
	case appcontext.AppContextStatusEnum.InstantiateFailed:
		log.Error("Can't remove Cluster Reference: the Logical Cloud has failed instantiating, for safety please terminate and try again.", log.Fields{"clusterreference": clusterReference, "logicalcloud": logicalCloud})
		return pkgerrors.New("Can't remove Cluster Reference: the Logical Cloud has failed instantiating, for safety please terminate and try again.")
	case appcontext.AppContextStatusEnum.TerminateFailed:
		log.Info("The Logical Cloud has failed terminating, proceeding with the delete operation.", log.Fields{"clusterreference": clusterReference, "logicalcloud": logicalCloud})
		// try to delete anyway since termination failed
		fallthrough
	case appcontext.AppContextStatusEnum.Terminated:
		err := db.DBconn.Remove(v.storeName, key)
		if err != nil {
			return pkgerrors.Wrap(err, "Error deleting Cluster Reference")
		}

		log.Info("Removed cluster reference from Logical Cloud.", log.Fields{"logicalcloud": logicalCloud})
		return nil
	default:
		log.Error("Failure removing Cluster Reference: the Logical Cloud isn't in an expected status so not taking any action.", log.Fields{"clusterreference": clusterReference, "logicalcloud": logicalCloud, "status": acStatus.Status})
		return pkgerrors.New("Failure removing Cluster Reference: the Logical Cloud isn't in an expected status so not taking any action.")
	}
}

// Update an entry for the Cluster reference in the database
func (v *ClusterClient) UpdateCluster(project, logicalCloud, clusterReference string, c Cluster) (Cluster, error) {

	key := ClusterKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		ClusterReference: clusterReference,
	}

	//Check for name mismatch in cluster reference
	if c.MetaData.ClusterReference != clusterReference {
		return Cluster{}, pkgerrors.New("Cluster Reference mismatch")
	}
	//Check if this Cluster reference exists
	_, err := v.GetCluster(project, logicalCloud, clusterReference)
	if err != nil {
		return Cluster{}, err
	}
	err = db.DBconn.Insert(v.storeName, key, nil, v.tagMeta, c)
	if err != nil {
		return Cluster{}, pkgerrors.Wrap(err, "Updating DB Entry")
	}
	return c, nil
}

// Get returns Cluster's kubeconfig for corresponding cluster reference
func (v *ClusterClient) GetClusterConfig(project, logicalCloud, clusterReference string) (string, error) {
	lcClient := NewLogicalCloudClient()
	lckey := LogicalCloudKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
	}
	s, err := lcClient.GetState(project, logicalCloud)
	if err != nil {
		return "", err
	}
	cid := state.GetLastContextIdFromStateInfo(s)

	if cid == "" {
		return "", pkgerrors.New("Logical Cloud hasn't been instantiated yet")
	}

	ac, err := state.GetAppContextFromId(cid)
	if err != nil {
		return "", err
	}

	// get logical cloud resource
	lc, err := lcClient.Get(project, logicalCloud)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Error getting logical cloud")
	}
	// get user's private key
	privateKeyData, err := db.DBconn.Find(v.storeName, lckey, "privatekey")
	if err != nil {
		return "", pkgerrors.Wrap(err, "Error getting private key from logical cloud")
	}

	// get cluster from dcm (need provider/name)
	cluster, err := v.GetCluster(project, logicalCloud, clusterReference)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Error getting cluster")
	}

	// before attempting to generate a kubeconfig,
	// check if certificate has been issued and copy it from etcd to mongodb
	if cluster.Specification.Certificate == "" {
		log.Info("Certificate not yet in MongoDB, checking etcd.", log.Fields{})

		// access etcd
		clusterName := strings.Join([]string{cluster.Specification.ClusterProvider, "+", cluster.Specification.ClusterName}, "")

		// get the app context handle for the status of this cluster (which should contain the certificate inside, if already issued)
		statusHandle, err := ac.GetClusterStatusHandle("logical-cloud", clusterName)

		if err != nil {
			return "", pkgerrors.Wrap(err, "The cluster doesn't contain status, please check if all services are up and running")
		}
		statusRaw, err := ac.GetValue(statusHandle)
		if err != nil {
			return "", pkgerrors.Wrap(err, "An error occurred while reading the cluster status")
		}

		var rbstatus rb.ResourceBundleStateStatus
		err = json.Unmarshal([]byte(statusRaw.(string)), &rbstatus)
		if err != nil {
			return "", pkgerrors.Wrap(err, "An error occurred while parsing the cluster status")
		}

		if len(rbstatus.CsrStatuses) == 0 {
			return "", pkgerrors.New("A status for the CSR hasn't been returned yet")
		}

		// validate that we indeed obtained a certificate before persisting it in the database:
		approved := false
		for _, c := range rbstatus.CsrStatuses[0].Status.Conditions {
			if c.Type == "Denied" {
				return "", pkgerrors.New("Certificate was denied!")
			}
			if c.Type == "Failed" {
				return "", pkgerrors.New("Certificate issue failed")
			}
			if c.Type == "Approved" {
				approved = true
			}
		}
		if !approved {
			return "", pkgerrors.New("The CSR hasn't been approved yet or the certificate hasn't been issued yet")
		}

		//just double-check certificate field contents aren't empty:
		cert := rbstatus.CsrStatuses[0].Status.Certificate
		if len(cert) > 0 {
			cluster.Specification.Certificate = base64.StdEncoding.EncodeToString([]byte(cert))
		} else {
			return "", pkgerrors.New("Certificate issued was invalid")
		}

		// copy key to MongoDB
		// func (v *ClusterClient)
		// UpdateCluster(project, logicalCloud, clusterReference string, c Cluster) (Cluster, error) {
		_, err = v.UpdateCluster(project, logicalCloud, clusterReference, cluster)
		if err != nil {
			return "", pkgerrors.Wrap(err, "An error occurred while storing the certificate")
		}
	} else {
		// certificate is already in MongoDB so just hand it over to create the API response
		log.Info("Certificate already in MongoDB, pass it to API", log.Fields{})
	}

	// sanity check for cluster-issued certificate
	if cluster.Specification.Certificate == "" {
		return "", pkgerrors.New("Failed creating kubeconfig due to unexpected empty certificate")
	}

	// get kubeconfig from L0 cloudconfig respective to the cluster referenced by this logical cloud
	ccc := rsync.NewCloudConfigClient()
	cconfig, err := ccc.GetCloudConfig(cluster.Specification.ClusterProvider, cluster.Specification.ClusterName, "0", "")
	if err != nil {
		return "", pkgerrors.New("Failed fetching kubeconfig from rsync's CloudConfig")
	}
	adminConfig, err := base64.StdEncoding.DecodeString(cconfig.Config)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Failed decoding CloudConfig's kubeconfig from base64")
	}

	// unmarshall CloudConfig's kubeconfig into struct
	adminKubeConfig := KubeConfig{}
	err = yaml.Unmarshal(adminConfig, &adminKubeConfig)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Failed parsing CloudConfig's kubeconfig yaml")
	}

	// all data needed for final kubeconfig:
	privateKey := string(privateKeyData[0])
	signedCert := cluster.Specification.Certificate
	clusterCert := adminKubeConfig.Clusters[0].ClusterDef.CertificateAuthorityData
	clusterAddr := adminKubeConfig.Clusters[0].ClusterDef.Server
	namespace := lc.Specification.NameSpace
	userName := lc.Specification.User.UserName
	contextName := userName + "@" + clusterReference

	kubeconfig := KubeConfig{
		ApiVersion: "v1",
		Kind:       "Config",
		Clusters: []KubeCluster{
			KubeCluster{
				ClusterName: clusterReference,
				ClusterDef: KubeClusterDef{
					CertificateAuthorityData: clusterCert,
					Server:                   clusterAddr,
				},
			},
		},
		Contexts: []KubeContext{
			KubeContext{
				ContextName: contextName,
				ContextDef: KubeContextDef{
					Cluster:   clusterReference,
					Namespace: namespace,
					User:      userName,
				},
			},
		},
		CurrentContext: contextName,
		Preferences:    map[string]string{},
		Users: []KubeUser{
			KubeUser{
				UserName: userName,
				UserDef: KubeUserDef{
					ClientCertificateData: signedCert,
					ClientKeyData:         privateKey,
				},
			},
		},
	}

	yaml, err := yaml.Marshal(&kubeconfig)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Failed marshalling user kubeconfig into yaml")
	}

	// now that we have the L1 kubeconfig for this L1 logical cloud,
	// let's give it to rsync so it can get stored in the right place
	_, err = ccc.CreateCloudConfig(
		cluster.Specification.ClusterProvider,
		cluster.Specification.ClusterName,
		lc.Specification.Level,
		lc.Specification.NameSpace,
		base64.StdEncoding.EncodeToString(yaml))

	if err != nil {
		if err.Error() != "CloudConfig already exists" {
			return "", pkgerrors.Wrap(err, "Failed creating a new kubeconfig in rsync's CloudConfig")
		}
	}

	return string(yaml), nil
}
