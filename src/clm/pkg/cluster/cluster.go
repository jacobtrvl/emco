// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package cluster

import (
	"strings"
	"time"

	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	mtypes "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/types"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"
	rsync "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/db"

	clmController "gitlab.com/project-emco/core/emco-base/src/clm/pkg/controller"
	clmcontrollerpb "gitlab.com/project-emco/core/emco-base/src/clm/pkg/grpc/controller-eventchannel"

	clmcontrollereventchannelclient "gitlab.com/project-emco/core/emco-base/src/clm/pkg/grpc/controllereventchannelclient"
)

type clientDbInfo struct {
	storeName string // name of the mongodb collection to use for client documents
	tagMeta   string // attribute key name for the json data of a client document
	tagState  string // attribute key name for StateInfo object in the cluster
}

// ClusterProvider contains the parameters needed for ClusterProviders
type ClusterProvider struct {
	Metadata mtypes.Metadata `json:"metadata"`
}

type Cluster struct {
	Metadata mtypes.Metadata `json:"metadata"`
	Spec     GitOpsSpec      `json:"spec,omitempty"`
}

type ClusterWithLabels struct {
	Metadata mtypes.Metadata `json:"metadata"`
	Labels   []ClusterLabel  `json:"labels"`
}

type ClusterContent struct {
	Kubeconfig string `json:"kubeconfig"`
}

type ClusterLabel struct {
	LabelName string `json:"clusterLabel"`
}

type ClusterKvPairs struct {
	Metadata mtypes.Metadata `json:"metadata"`
	Spec     ClusterKvSpec   `json:"spec"`
}

type ClusterKvSpec struct {
	Kv []map[string]interface{} `json:"kv"`
}

type GitOpsSpec struct {
	Props GitOpsProps `json:"gitOps"`
}

type GitOpsProps struct {
	GitOpsType            string `json:"gitOpsType"`
	GitOpsReferenceObject string `json:"gitOpsReferenceObject"`
	GitOpsResourceObject  string `json:"gitOpsResourceObject"`
}

type ClusterSyncObjects struct {
	Metadata mtypes.Metadata       `json:"metadata"`
	Spec     ClusterSyncObjectSpec `json:"spec"`
}

type ClusterSyncObjectSpec struct {
	Kv []map[string]interface{} `json:"kv"`
}

// ClusterProviderKey is the key structure that is used in the database
type ClusterProviderKey struct {
	ClusterProviderName string `json:"clusterProvider"`
}

// ClusterKey is the key structure that is used in the database
type ClusterKey struct {
	ClusterProviderName string `json:"clusterProvider"`
	ClusterName         string `json:"cluster"`
}

// ClusterLabelKey is the key structure that is used in the database
type ClusterLabelKey struct {
	ClusterProviderName string `json:"clusterProvider"`
	ClusterName         string `json:"cluster"`
	ClusterLabelName    string `json:"clusterLabel"`
}

// LabelKey is the key structure that is used in the database
type LabelKey struct {
	ClusterProviderName string `json:"clusterProvider"`
	ClusterLabelName    string `json:"clusterLabel"`
}

// ClusterKvPairsKey is the key structure that is used in the database
type ClusterKvPairsKey struct {
	ClusterProviderName string `json:"clusterProvider"`
	ClusterName         string `json:"cluster"`
	ClusterKvPairsName  string `json:"clusterKv"`
}

// ClusterSyncObjectKey is the key structure that is used in the database
type ClusterSyncObjectsKey struct {
	ClusterProviderName    string `json:"clusterProvider"`
	ClusterSyncObjectsName string `json:"clusterSyncObject"`
}

const SEPARATOR = "+"
const CONTEXT_CLUSTER_APP = "network-intents"
const CONTEXT_CLUSTER_RESOURCE = "network-intents"

// ClusterManager is an interface exposes the Cluster functionality
type ClusterManager interface {
	CreateClusterProvider(pr ClusterProvider, exists bool) (ClusterProvider, error)
	GetClusterProvider(name string) (ClusterProvider, error)
	GetClusterProviders() ([]ClusterProvider, error)
	DeleteClusterProvider(name string) error
	CreateCluster(provider string, pr Cluster, qr ClusterContent) (Cluster, error)
	GetCluster(provider, name string) (Cluster, error)
	GetClusterContent(provider, name string) (ClusterContent, error)
	GetClusterState(provider, name string) (state.StateInfo, error)
	GetClusters(provider string) ([]Cluster, error)
	GetClustersWithLabel(provider, label string) ([]string, error)
	GetAllClustersAndLabels(provider string) ([]ClusterWithLabels, error)
	DeleteCluster(provider, name string) error
	CreateClusterLabel(provider, cluster string, pr ClusterLabel, exists bool) (ClusterLabel, error)
	GetClusterLabel(provider, cluster, label string) (ClusterLabel, error)
	GetClusterLabels(provider, cluster string) ([]ClusterLabel, error)
	DeleteClusterLabel(provider, cluster, label string) error
	CreateClusterKvPairs(provider, cluster string, pr ClusterKvPairs, exists bool) (ClusterKvPairs, error)
	GetClusterKvPairs(provider, cluster, kvpair string) (ClusterKvPairs, error)
	GetClusterKvPairsValue(provider, cluster, kvpair, kvkey string) (interface{}, error)
	GetAllClusterKvPairs(provider, cluster string) ([]ClusterKvPairs, error)
	DeleteClusterKvPairs(provider, cluster, kvpair string) error
	CreateClusterSyncObjects(provider string, pr ClusterSyncObjects, exists bool) (ClusterSyncObjects, error)
	GetClusterSyncObjects(provider, syncobject string) (ClusterSyncObjects, error)
	GetClusterSyncObjectsValue(provider, syncobject, syncobjectkey string) (interface{}, error)
	GetAllClusterSyncObjects(provider string) ([]ClusterSyncObjects, error)
	DeleteClusterSyncObjects(provider, syncobject string) error
}

// ClusterClient implements the Manager
// It will also be used to maintain some localized state
type ClusterClient struct {
	db clientDbInfo
}

// NewClusterClient returns an instance of the ClusterClient
// which implements the Manager
func NewClusterClient() *ClusterClient {
	return &ClusterClient{
		db: clientDbInfo{
			storeName: "resources",
			tagMeta:   "data",
			tagState:  "stateInfo",
		},
	}
}

// CreateClusterProvider - create a new Cluster Provider
func (v *ClusterClient) CreateClusterProvider(p ClusterProvider, exists bool) (ClusterProvider, error) {

	//Construct key and tag to select the entry
	key := ClusterProviderKey{
		ClusterProviderName: p.Metadata.Name,
	}

	//Check if this ClusterProvider already exists
	_, err := v.GetClusterProvider(p.Metadata.Name)
	if err == nil && !exists {
		return ClusterProvider{}, pkgerrors.New("ClusterProvider already exists")
	}

	err = db.DBconn.Insert(v.db.storeName, key, nil, v.db.tagMeta, p)
	if err != nil {
		return ClusterProvider{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return p, nil
}

// GetClusterProvider returns the ClusterProvider for corresponding name
func (v *ClusterClient) GetClusterProvider(name string) (ClusterProvider, error) {

	//Construct key and tag to select the entry
	key := ClusterProviderKey{
		ClusterProviderName: name,
	}

	value, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return ClusterProvider{}, err
	} else if len(value) == 0 {
		return ClusterProvider{}, pkgerrors.New("Cluster provider not found")
	}

	//value is a byte array
	if value != nil {
		cp := ClusterProvider{}
		err = db.DBconn.Unmarshal(value[0], &cp)
		if err != nil {
			return ClusterProvider{}, err
		}
		return cp, nil
	}

	return ClusterProvider{}, pkgerrors.New("Unknown Error")
}

// GetClusterProviderList returns all of the ClusterProvider for corresponding name
func (v *ClusterClient) GetClusterProviders() ([]ClusterProvider, error) {

	//Construct key and tag to select the entry
	key := ClusterProviderKey{
		ClusterProviderName: "",
	}

	values, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return []ClusterProvider{}, err
	}

	resp := make([]ClusterProvider, 0)
	for _, value := range values {
		cp := ClusterProvider{}
		err = db.DBconn.Unmarshal(value, &cp)
		if err != nil {
			return []ClusterProvider{}, err
		}
		resp = append(resp, cp)
	}

	return resp, nil
}

// DeleteClusterProvider the  ClusterProvider from database
func (v *ClusterClient) DeleteClusterProvider(name string) error {

	//Construct key and tag to select the entry
	key := ClusterProviderKey{
		ClusterProviderName: name,
	}

	err := db.DBconn.Remove(v.db.storeName, key)
	return err
}

// CreateCluster - create a new Cluster for a cluster-provider
func (v *ClusterClient) CreateCluster(provider string, p Cluster, q ClusterContent) (Cluster, error) {

	//Construct key and tag to select the entry
	key := ClusterKey{
		ClusterProviderName: provider,
		ClusterName:         p.Metadata.Name,
	}

	//Check if this Cluster already exists
	_, err := v.GetCluster(provider, p.Metadata.Name)
	if err == nil {
		return Cluster{}, pkgerrors.New("Cluster already exists")
	}

	err = db.DBconn.Insert(v.db.storeName, key, nil, v.db.tagMeta, p)
	if err != nil {
		return Cluster{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	// Add the stateInfo record
	s := state.StateInfo{}
	a := state.ActionEntry{
		State:     state.StateEnum.Created,
		ContextId: "",
		TimeStamp: time.Now(),
	}
	s.Actions = append(s.Actions, a)

	err = db.DBconn.Insert(v.db.storeName, key, nil, v.db.tagState, s)
	if err != nil {
		return Cluster{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	ccc := rsync.NewCloudConfigClient()

	_, err = ccc.CreateCloudConfig(provider, p.Metadata.Name, "0", "default", q.Kubeconfig)
	if err != nil {
		return Cluster{}, pkgerrors.Wrap(err, "Error creating cloud config")
	}

	// Loop through CLM controllers and publish CLUSTER_CREATE event
	client := clmController.NewControllerClient()
	ctrls, _ := client.GetControllers()
	for _, c := range ctrls {
		log.Info("CLM CreateController .. controller info.", log.Fields{"clusterProvider": provider, "cluster": p.Metadata.Name, "Controller": c})
		err = clmcontrollereventchannelclient.SendControllerEvent(provider, p.Metadata.Name, clmcontrollerpb.ClmControllerEventType_CLUSTER_CREATED, c)
		if err != nil {
			log.Error("CLM CreateController .. Failed publishing event to clmController.", log.Fields{"clusterProvider": provider, "cluster": p.Metadata.Name, "Controller": c})
			return Cluster{}, pkgerrors.Wrapf(err, "CLM failed publishing event to clm-controller[%v]", c.Metadata.Name)
		}
	}

	return p, nil
}

// GetCluster returns the Cluster for corresponding provider and name
func (v *ClusterClient) GetCluster(provider, name string) (Cluster, error) {
	//Construct key and tag to select the entry
	key := ClusterKey{
		ClusterProviderName: provider,
		ClusterName:         name,
	}

	value, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return Cluster{}, err
	} else if len(value) == 0 {
		return Cluster{}, pkgerrors.New("Cluster not found")
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

// GetClusterContent returns the ClusterContent for corresponding provider and name
func (v *ClusterClient) GetClusterContent(provider, name string) (ClusterContent, error) {

	// Fetch the kubeconfig from rsync according to new workflow
	ccc := rsync.NewCloudConfigClient()

	cconfig, err := ccc.GetCloudConfig(provider, name, "0", "")
	if err != nil {
		if strings.Contains(err.Error(), "Finding CloudConfig failed") {
			return ClusterContent{}, pkgerrors.Wrap(err, "GetCloudConfig error - not found")
		} else {
			return ClusterContent{}, pkgerrors.Wrap(err, "GetCloudConfig error - general")
		}
	}

	ccontent := ClusterContent{}
	ccontent.Kubeconfig = cconfig.Config

	return ccontent, nil
}

// GetClusterState returns the StateInfo structure for corresponding cluster provider and cluster
func (v *ClusterClient) GetClusterState(provider, name string) (state.StateInfo, error) {
	//Construct key and tag to select the entry
	key := ClusterKey{
		ClusterProviderName: provider,
		ClusterName:         name,
	}

	result, err := db.DBconn.Find(v.db.storeName, key, v.db.tagState)
	if err != nil {
		return state.StateInfo{}, err
	} else if len(result) == 0 {
		return state.StateInfo{}, pkgerrors.New("Cluster StateInfo not found")
	}

	if result != nil {
		s := state.StateInfo{}
		err = db.DBconn.Unmarshal(result[0], &s)
		if err != nil {
			return state.StateInfo{}, err
		}
		return s, nil
	}

	return state.StateInfo{}, pkgerrors.New("Unknown Error")
}

// GetClusters returns all the Clusters for corresponding provider
func (v *ClusterClient) GetClusters(provider string) ([]Cluster, error) {
	//Construct key and tag to select the entry
	key := ClusterKey{
		ClusterProviderName: provider,
		ClusterName:         "",
	}

	//Verify Cluster provider exists
	_, err := v.GetClusterProvider(provider)
	if err != nil {
		return []Cluster{}, err
	}

	values, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return []Cluster{}, err
	}

	resp := make([]Cluster, 0)
	for _, value := range values {
		cp := Cluster{}
		err = db.DBconn.Unmarshal(value, &cp)
		if err != nil {
			return []Cluster{}, err
		}
		resp = append(resp, cp)
	}

	return resp, nil
}

// GetAllClustersAndLabels returns all the the clusters and their labels
func (v *ClusterClient) GetAllClustersAndLabels(provider string) ([]ClusterWithLabels, error) {

	// Get All clusters
	cl, err := v.GetClusters(provider)
	if err != nil {
		return []ClusterWithLabels{}, err
	}

	resp := make([]ClusterWithLabels, len(cl))

	// Get all cluster labels
	for k, value := range cl {
		resp[k].Metadata = value.Metadata
		resp[k].Labels, err = v.GetClusterLabels(provider, value.Metadata.Name)
		if err != nil {
			return []ClusterWithLabels{}, err
		}
	}
	return resp, nil
}

// GetClustersWithLabel returns all the Clusters with Labels for provider
// Support Query like /cluster-providers/{Provider}/clusters?label={label}
func (v *ClusterClient) GetClustersWithLabel(provider, label string) ([]string, error) {
	//Construct key and tag to select the entry
	key := LabelKey{
		ClusterProviderName: provider,
		ClusterLabelName:    label,
	}

	//Verify Cluster provider exists
	_, err := v.GetClusterProvider(provider)
	if err != nil {
		return []string{}, err
	}

	values, err := db.DBconn.Find(v.db.storeName, key, "cluster")
	if err != nil {
		return []string{}, err
	}

	resp := make([]string, 0)
	for _, value := range values {
		cp := string(value)
		resp = append(resp, cp)
	}

	return resp, nil
}

// DeleteCluster the  Cluster from database
func (v *ClusterClient) DeleteCluster(provider, name string) error {
	//Construct key and tag to select the entry
	key := ClusterKey{
		ClusterProviderName: provider,
		ClusterName:         name,
	}

	s, err := v.GetClusterState(provider, name)
	if err != nil {
		// If the StateInfo cannot be found, then a proper cluster record is not present.
		// Call the DB delete to clean up any errant record without a StateInfo element that may exist.
		err = db.DBconn.Remove(v.db.storeName, key)
		return err
	}

	stateVal, err := state.GetCurrentStateFromStateInfo(s)
	if err != nil {
		return pkgerrors.Wrap(err, "Error getting current state from Cluster stateInfo: "+name)
	}

	if stateVal == state.StateEnum.Applied || stateVal == state.StateEnum.InstantiateStopped {
		return pkgerrors.Errorf("Cluster network intents must be terminated before it can be deleted " + name)
	}

	// remove the app contexts associated with this cluster
	if stateVal == state.StateEnum.Terminated || stateVal == state.StateEnum.TerminateStopped {
		// Verify that the appcontext has completed terminating
		ctxid := state.GetLastContextIdFromStateInfo(s)
		acStatus, err := state.GetAppContextStatus(ctxid)
		if err == nil &&
			!(acStatus.Status == appcontext.AppContextStatusEnum.Terminated || acStatus.Status == appcontext.AppContextStatusEnum.TerminateFailed) {
			return pkgerrors.Errorf("Network intents for cluster have not completed terminating " + name)
		}

		for _, id := range state.GetContextIdsFromStateInfo(s) {
			context, err := state.GetAppContextFromId(id)
			if err != nil {
				log.Info("Delete Cluster .. appcontext not found", log.Fields{"clusterProvider": provider, "cluster": name, "appContext": id})
				continue
			}
			err = context.DeleteCompositeApp()
			if err != nil {
				return pkgerrors.Wrap(err, "Error deleting appcontext for Cluster")
			}
		}
	}

	// Remove the cluster resource
	err = db.DBconn.Remove(v.db.storeName, key)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete Cluster Entry;")
	}

	// Loop through CLM controllers and publish CLUSTER_DELETE event
	client := clmController.NewControllerClient()
	vals, _ := client.GetControllers()
	for _, v := range vals {
		log.Info("DeleteCluster .. controller info.", log.Fields{"clusterProvider": provider, "cluster": name, "Controller": v})
		err = clmcontrollereventchannelclient.SendControllerEvent(provider, name, clmcontrollerpb.ClmControllerEventType_CLUSTER_DELETED, v)
		if err != nil {
			log.Error("DeleteCluster .. Failed publishing event to controller.", log.Fields{"clusterProvider": provider, "cluster": name, "Controller": v})
		}
	}

	// Delete the Cloud Config resource associated with this cluster
	ccc := rsync.NewCloudConfigClient()

	err = ccc.DeleteCloudConfig(provider, name, "0", "default")
	if err != nil {
		log.Warn("DeleteCluster .. error deleting cloud config", log.Fields{"clusterProvider": provider, "cluster": name, "error": err})
	}

	return nil
}

// CreateClusterLabel - create a new Cluster Label mongo document for a cluster-provider/cluster
func (v *ClusterClient) CreateClusterLabel(provider string, cluster string, p ClusterLabel, exists bool) (ClusterLabel, error) {
	//Construct key and tag to select the entry
	key := ClusterLabelKey{
		ClusterProviderName: provider,
		ClusterName:         cluster,
		ClusterLabelName:    p.LabelName,
	}

	//Check if this ClusterLabel already exists
	_, err := v.GetClusterLabel(provider, cluster, p.LabelName)
	if err == nil && !exists {
		return ClusterLabel{}, pkgerrors.New("Cluster Label already exists")
	}

	err = db.DBconn.Insert(v.db.storeName, key, nil, v.db.tagMeta, p)
	if err != nil {
		return ClusterLabel{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return p, nil
}

// GetClusterLabel returns the Cluster for corresponding provider, cluster and label
func (v *ClusterClient) GetClusterLabel(provider, cluster, label string) (ClusterLabel, error) {
	//Construct key and tag to select the entry
	key := ClusterLabelKey{
		ClusterProviderName: provider,
		ClusterName:         cluster,
		ClusterLabelName:    label,
	}

	value, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return ClusterLabel{}, err
	} else if len(value) == 0 {
		return ClusterLabel{}, pkgerrors.New("Cluster label not found")
	}

	//value is a byte array
	if value != nil {
		cl := ClusterLabel{}
		err = db.DBconn.Unmarshal(value[0], &cl)
		if err != nil {
			return ClusterLabel{}, err
		}
		return cl, nil
	}

	return ClusterLabel{}, pkgerrors.New("Unknown Error")
}

// GetClusterLabels returns the Cluster Labels for corresponding provider and cluster
func (v *ClusterClient) GetClusterLabels(provider, cluster string) ([]ClusterLabel, error) {
	// Construct key and tag to select the entry
	key := ClusterLabelKey{
		ClusterProviderName: provider,
		ClusterName:         cluster,
		ClusterLabelName:    "",
	}

	// Verify Cluster already exists
	_, err := v.GetCluster(provider, cluster)
	if err != nil {
		return []ClusterLabel{}, err
	}

	values, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return []ClusterLabel{}, err
	}

	resp := make([]ClusterLabel, 0)
	for _, value := range values {
		cp := ClusterLabel{}
		err = db.DBconn.Unmarshal(value, &cp)
		if err != nil {
			return []ClusterLabel{}, err
		}
		resp = append(resp, cp)
	}

	return resp, nil
}

// DeleteClusterLabel ... Delete the Cluster Label from database
func (v *ClusterClient) DeleteClusterLabel(provider, cluster, label string) error {
	//Construct key and tag to select the entry
	key := ClusterLabelKey{
		ClusterProviderName: provider,
		ClusterName:         cluster,
		ClusterLabelName:    label,
	}

	err := db.DBconn.Remove(v.db.storeName, key)
	return err
}

// CreateClusterKvPairs - Create a New Cluster KV pairs document
func (v *ClusterClient) CreateClusterKvPairs(provider string, cluster string, p ClusterKvPairs, exists bool) (ClusterKvPairs, error) {
	key := ClusterKvPairsKey{
		ClusterProviderName: provider,
		ClusterName:         cluster,
		ClusterKvPairsName:  p.Metadata.Name,
	}

	//Check if this ClusterKvPairs already exists
	_, err := v.GetClusterKvPairs(provider, cluster, p.Metadata.Name)
	if err == nil && !exists {
		return ClusterKvPairs{}, pkgerrors.New("Cluster KV Pair already exists")
	}

	err = db.DBconn.Insert(v.db.storeName, key, nil, v.db.tagMeta, p)
	if err != nil {
		return ClusterKvPairs{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return p, nil
}

// GetClusterKvPairs returns the Cluster KeyValue pair for corresponding provider, cluster and KV pair name
func (v *ClusterClient) GetClusterKvPairs(provider, cluster, kvpair string) (ClusterKvPairs, error) {
	//Construct key and tag to select entry
	key := ClusterKvPairsKey{
		ClusterProviderName: provider,
		ClusterName:         cluster,
		ClusterKvPairsName:  kvpair,
	}

	value, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return ClusterKvPairs{}, err
	} else if len(value) == 0 {
		return ClusterKvPairs{}, pkgerrors.New("Cluster key value pair not found")
	}

	//value is a byte array
	if value != nil {
		ckvp := ClusterKvPairs{}
		err = db.DBconn.Unmarshal(value[0], &ckvp)
		if err != nil {
			return ClusterKvPairs{}, err
		}
		return ckvp, nil
	}

	return ClusterKvPairs{}, pkgerrors.New("Unknown Error")
}

// GetClusterKvPairsValue returns the value of the key from the corresponding provider, cluster and KV pair name
func (v *ClusterClient) GetClusterKvPairsValue(provider, cluster, kvpair, kvkey string) (interface{}, error) {
	//Construct key and tag to select entry
	key := ClusterKvPairsKey{
		ClusterProviderName: provider,
		ClusterName:         cluster,
		ClusterKvPairsName:  kvpair,
	}

	value, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return ClusterKvPairs{}, err
	} else if len(value) == 0 {
		return Cluster{}, pkgerrors.New("Cluster key value pair not found")
	}

	//value is a byte array
	if value != nil {
		ckvp := ClusterKvPairs{}
		err = db.DBconn.Unmarshal(value[0], &ckvp)
		if err != nil {
			return nil, err
		}

		for _, kvmap := range ckvp.Spec.Kv {
			if val, ok := kvmap[kvkey]; ok {
				return struct {
					Value interface{} `json:"value"`
				}{Value: val}, nil
			}
		}
		return nil, pkgerrors.New("Cluster KV pair key value not found")
	}

	return nil, pkgerrors.New("Unknown Error")
}

// GetAllClusterKvPairs returns the Cluster Kv Pairs for corresponding provider and cluster
func (v *ClusterClient) GetAllClusterKvPairs(provider, cluster string) ([]ClusterKvPairs, error) {
	//Construct key and tag to select the entry
	key := ClusterKvPairsKey{
		ClusterProviderName: provider,
		ClusterName:         cluster,
		ClusterKvPairsName:  "",
	}

	// Verify Cluster exists
	_, err := v.GetCluster(provider, cluster)
	if err != nil {
		return []ClusterKvPairs{}, err
	}

	values, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return []ClusterKvPairs{}, err
	}

	resp := make([]ClusterKvPairs, 0)
	for _, value := range values {
		cp := ClusterKvPairs{}
		err = db.DBconn.Unmarshal(value, &cp)
		if err != nil {
			return []ClusterKvPairs{}, err
		}
		resp = append(resp, cp)
	}

	return resp, nil
}

// DeleteClusterKvPairs the  ClusterKvPairs from database
func (v *ClusterClient) DeleteClusterKvPairs(provider, cluster, kvpair string) error {
	//Construct key and tag to select entry
	key := ClusterKvPairsKey{
		ClusterProviderName: provider,
		ClusterName:         cluster,
		ClusterKvPairsName:  kvpair,
	}

	err := db.DBconn.Remove(v.db.storeName, key)
	return err
}

// CreateClusterSyncObjects - Create a New Cluster sync objects document
func (v *ClusterClient) CreateClusterSyncObjects(provider string, p ClusterSyncObjects, exists bool) (ClusterSyncObjects, error) {
	key := ClusterSyncObjectsKey{
		ClusterProviderName:    provider,
		ClusterSyncObjectsName: p.Metadata.Name,
	}

	//Check if this ClusterSyncObjects already exists
	_, err := v.GetClusterSyncObjects(provider, p.Metadata.Name)
	if err == nil && !exists {
		return ClusterSyncObjects{}, pkgerrors.New("Cluster Sync Objects already exists")
	}

	err = db.DBconn.Insert(v.db.storeName, key, nil, v.db.tagMeta, p)
	if err != nil {
		return ClusterSyncObjects{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return p, nil
}

// GetClusterSyncObjects returns the Cluster Sync objects for corresponding provider and sync object name
func (v *ClusterClient) GetClusterSyncObjects(provider, syncobject string) (ClusterSyncObjects, error) {
	//Construct key and tag to select entry
	key := ClusterSyncObjectsKey{
		ClusterProviderName:    provider,
		ClusterSyncObjectsName: syncobject,
	}

	value, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return ClusterSyncObjects{}, err
	} else if len(value) == 0 {
		return ClusterSyncObjects{}, pkgerrors.New("Cluster sync object not found")
	}

	//value is a byte array
	if value != nil {
		ckvp := ClusterSyncObjects{}
		err = db.DBconn.Unmarshal(value[0], &ckvp)
		if err != nil {
			return ClusterSyncObjects{}, err
		}
		return ckvp, nil
	}

	return ClusterSyncObjects{}, pkgerrors.New("Unknown Error")
}

// DeleteClusterSyncObjects the  ClusterSyncObjects from database
func (v *ClusterClient) DeleteClusterSyncObjects(provider, syncobject string) error {
	//Construct key and tag to select entry
	key := ClusterSyncObjectsKey{
		ClusterProviderName:    provider,
		ClusterSyncObjectsName: syncobject,
	}

	err := db.DBconn.Remove(v.db.storeName, key)
	return err
}

// GetClusterSyncObjectsValue returns the value of the key from the corresponding provider and Sync Object name
func (v *ClusterClient) GetClusterSyncObjectsValue(provider, syncobject, syncobjectkey string) (interface{}, error) {
	//Construct key and tag to select entry
	key := ClusterSyncObjectsKey{
		ClusterProviderName:    provider,
		ClusterSyncObjectsName: syncobject,
	}

	value, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return ClusterSyncObjects{}, err
	} else if len(value) == 0 {
		return Cluster{}, pkgerrors.New("Cluster sync object not found")
	}

	//value is a byte array
	if value != nil {
		ckvp := ClusterSyncObjects{}
		err = db.DBconn.Unmarshal(value[0], &ckvp)
		if err != nil {
			return nil, err
		}

		for _, kvmap := range ckvp.Spec.Kv {
			if val, ok := kvmap[syncobjectkey]; ok {
				return struct {
					Value interface{} `json:"value"`
				}{Value: val}, nil
			}
		}
		return nil, pkgerrors.New("Cluster Sync Object key value not found")
	}

	return nil, pkgerrors.New("Unknown Error")
}

// GetAllClusterSyncObjects returns the Cluster Sync Objects for corresponding provider
func (v *ClusterClient) GetAllClusterSyncObjects(provider string) ([]ClusterSyncObjects, error) {
	//Construct key and tag to select the entry
	key := ClusterSyncObjectsKey{
		ClusterProviderName:    provider,
		ClusterSyncObjectsName: "",
	}

	// // Verify Cluster Provider exists
	_, err := v.GetClusterProvider(provider)
	if err != nil {
		return []ClusterSyncObjects{}, err
	}

	values, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return []ClusterSyncObjects{}, err
	}

	resp := make([]ClusterSyncObjects, 0)
	for _, value := range values {
		cp := ClusterSyncObjects{}
		err = db.DBconn.Unmarshal(value, &cp)
		if err != nil {
			return []ClusterSyncObjects{}, err
		}
		resp = append(resp, cp)
	}

	return resp, nil
}
