// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strconv"
	"strings"

	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext/subresources"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/installappclient"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/controller"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"
	rsync "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/db"
	"gopkg.in/yaml.v2"
	k8scertsv1 "k8s.io/api/certificates/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// rsyncName denotes the name of the rsync controller
const rsyncName = "rsync"

type Resource struct {
	ApiVersion    string         `yaml:"apiVersion"`
	Kind          string         `yaml:"kind"`
	MetaData      MetaDatas      `yaml:"metadata"`
	Specification Specs          `yaml:"spec,omitempty"`
	Rules         []RoleRules    `yaml:"rules,omitempty"`
	Subjects      []RoleSubjects `yaml:"subjects,omitempty"`
	RoleRefs      RoleRef        `yaml:"roleRef,omitempty"`
}

type MetaDatas struct {
	Name      string            `yaml:"name"`
	Namespace string            `yaml:"namespace,omitempty"`
	Labels    map[string]string `yaml:"labels,omitempty"`
}

type Specs struct {
	Request    string   `yaml:"request,omitempty"`    // for CSR
	Usages     []string `yaml:"usages,omitempty"`     // for CSR
	SignerName string   `yaml:"signerName,omitempty"` // for CSR
	// TODO: validate quota keys
	// //Hard           logicalcloud.QSpec    `yaml:"hard,omitempty"`
	// Hard QSpec `yaml:"hard,omitempty"`
	Hard map[string]string `yaml:"hard,omitempty"` // for Quotas
}

type RoleRules struct {
	ApiGroups []string `yaml:"apiGroups"`
	Resources []string `yaml:"resources"`
	Verbs     []string `yaml:"verbs"`
}

type RoleSubjects struct {
	Kind     string `yaml:"kind"`
	Name     string `yaml:"name"`
	ApiGroup string `yaml:"apiGroup"`
}

type RoleRef struct {
	Kind     string `yaml:"kind"`
	Name     string `yaml:"name"`
	ApiGroup string `yaml:"apiGroup"`
}

func cleanupCompositeApp(context appcontext.AppContext, err error, reason string, details []string) error {
	if err == nil {
		// create an error object to avoid wrap failures
		err = pkgerrors.New("Composite App cleanup.")
	}

	cleanuperr := context.DeleteCompositeApp()
	newerr := pkgerrors.Wrap(err, reason)
	if cleanuperr != nil {
		log.Warn("Error cleaning AppContext, ", log.Fields{
			"Related details": details,
		})
		// this would be useful: https://godoc.org/go.uber.org/multierr
		return pkgerrors.Wrap(err, "After previous error, cleaning the AppContext also failed.")
	}
	return newerr
}

func createNamespace(logicalcloud LogicalCloud) (string, string, error) {

	name := logicalcloud.Specification.NameSpace
	labels := logicalcloud.Specification.Labels

	namespace := Resource{
		ApiVersion: "v1",
		Kind:       "Namespace",
		MetaData: MetaDatas{
			Name:   name,
			Labels: labels,
		},
	}

	nsData, err := yaml.Marshal(&namespace)
	if err != nil {
		return "", "", err
	}

	return string(nsData), strings.Join([]string{name, "+Namespace"}, ""), nil
}

func createRoles(logicalcloud LogicalCloud, userpermissions []UserPermission) ([]string, []string, error) {
	var name string
	var kind string
	var datas []string
	var names []string

	roleCount := len(userpermissions)
	datas = make([]string, roleCount, roleCount)
	names = make([]string, roleCount, roleCount)

	for i, up := range userpermissions {
		if up.Specification.Namespace == "" {
			name = strings.Join([]string{logicalcloud.MetaData.LogicalCloudName, "-clusterRole", strconv.Itoa(i)}, "")
			kind = "ClusterRole"
		} else {
			name = strings.Join([]string{logicalcloud.MetaData.LogicalCloudName, "-role", strconv.Itoa(i)}, "")
			kind = "Role"
		}

		role := Resource{
			ApiVersion: "rbac.authorization.k8s.io/v1",
			Kind:       kind,
			MetaData: MetaDatas{
				Name: name,
				// Namespace: logicalcloud.Specification.NameSpace,
			},
			Rules: []RoleRules{RoleRules{
				ApiGroups: up.Specification.APIGroups,
				Resources: up.Specification.Resources,
				Verbs:     up.Specification.Verbs,
			},
			},
		}
		if up.Specification.Namespace != "" {
			role.MetaData.Namespace = up.Specification.Namespace
		}

		roleData, err := yaml.Marshal(&role)
		if err != nil {
			return []string{}, []string{}, err
		}

		datas[i] = string(roleData)
		names[i] = strings.Join([]string{name, "+", kind}, "")
	}

	return datas, names, nil
}

func createRoleBindings(logicalcloud LogicalCloud, userpermissions []UserPermission) ([]string, []string, error) {
	var name string
	var kind string
	var kindbinding string
	var datas []string
	var names []string

	roleCount := len(userpermissions)
	datas = make([]string, roleCount, roleCount)
	names = make([]string, roleCount, roleCount)

	for i, up := range userpermissions {
		if up.Specification.Namespace == "" {
			name = strings.Join([]string{logicalcloud.MetaData.LogicalCloudName, "-clusterRoleBinding", strconv.Itoa(i)}, "")
			kind = "ClusterRole"
			kindbinding = "ClusterRoleBinding"
		} else {
			name = strings.Join([]string{logicalcloud.MetaData.LogicalCloudName, "-roleBinding", strconv.Itoa(i)}, "")
			kind = "Role"
			kindbinding = "RoleBinding"
		}

		roleBinding := Resource{
			ApiVersion: "rbac.authorization.k8s.io/v1",
			Kind:       kindbinding,
			MetaData: MetaDatas{
				Name: name,
			},
			Subjects: []RoleSubjects{RoleSubjects{
				Kind:     "User",
				Name:     logicalcloud.Specification.User.UserName,
				ApiGroup: "",
			},
			},

			RoleRefs: RoleRef{
				Kind:     kind,
				ApiGroup: "",
			},
		}
		if up.Specification.Namespace != "" {
			roleBinding.MetaData.Namespace = up.Specification.Namespace
			roleBinding.RoleRefs.Name = strings.Join([]string{logicalcloud.MetaData.LogicalCloudName, "-role", strconv.Itoa(i)}, "")
		} else {
			roleBinding.RoleRefs.Name = strings.Join([]string{logicalcloud.MetaData.LogicalCloudName, "-clusterRole", strconv.Itoa(i)}, "")
		}

		rBData, err := yaml.Marshal(&roleBinding)
		if err != nil {
			return []string{}, []string{}, err
		}
		datas[i] = string(rBData)
		names[i] = strings.Join([]string{name, "+", kindbinding}, "")
	}

	return datas, names, nil
}

func createQuotas(quotaList []Quota, namespace string) ([]string, []string, error) {

	var name string
	var datas []string
	var names []string

	quotaCount := len(quotaList)
	datas = make([]string, quotaCount, quotaCount)
	names = make([]string, quotaCount, quotaCount)

	for i, lcQuota := range quotaList {

		name = lcQuota.MetaData.QuotaName

		q := Resource{
			ApiVersion: "v1",
			Kind:       "ResourceQuota",
			MetaData: MetaDatas{
				Name:      name,
				Namespace: namespace,
			},
			Specification: Specs{
				Hard: lcQuota.Specification,
			},
		}

		qData, err := yaml.Marshal(&q)
		if err != nil {
			return []string{}, []string{}, err
		}
		datas[i] = string(qData)
		names[i] = strings.Join([]string{name, "+ResourceQuota"}, "")
	}

	return datas, names, nil
}

func createUserCSR(logicalcloud LogicalCloud) (string, string, string, error) {

	KEYSIZE := 4096
	userName := logicalcloud.Specification.User.UserName
	name := strings.Join([]string{logicalcloud.MetaData.LogicalCloudName, "-user-csr"}, "")

	key, err := rsa.GenerateKey(rand.Reader, KEYSIZE)
	if err != nil {
		return "", "", "", err
	}

	csrTemplate := x509.CertificateRequest{Subject: pkix.Name{CommonName: userName}}

	csrCert, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, key)
	if err != nil {
		return "", "", "", err
	}

	//Encode csr
	csr := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csrCert,
	})

	csrObj := Resource{
		ApiVersion: "certificates.k8s.io/v1",
		Kind:       "CertificateSigningRequest",
		MetaData: MetaDatas{
			Name: name,
		},
		Specification: Specs{
			Request:    base64.StdEncoding.EncodeToString(csr),
			Usages:     []string{"digital signature", "key encipherment", "client auth"},
			SignerName: "kubernetes.io/kube-apiserver-client",
		},
	}

	csrData, err := yaml.Marshal(&csrObj)
	if err != nil {
		return "", "", "", err
	}

	keyData := base64.StdEncoding.EncodeToString(pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		},
	))
	if err != nil {
		return "", "", "", err
	}

	return string(csrData), string(keyData), strings.Join([]string{name, "+CertificateSigningRequest"}, ""), nil
}

func createApprovalSubresource(logicalcloud LogicalCloud) (string, error) {
	subresource := subresources.ApprovalSubresource{
		Message:        "Approved for Logical Cloud authentication",
		Reason:         "LogicalCloud",
		Type:           string(k8scertsv1.CertificateApproved),
		LastUpdateTime: metav1.Now().Format("2006-01-02T15:04:05Z"),
		Status:         "True",
	}
	csrData, err := json.Marshal(subresource)
	return string(csrData), err
}

/*
queryDBAndSetRsyncInfo queries the MCO db to find the record the sync controller
and then sets the RsyncInfo global variable.
*/
func queryDBAndSetRsyncInfo() (installappclient.RsyncInfo, error) {
	client := controller.NewControllerClient("resources", "data", "orchestrator")
	vals, _ := client.GetControllers()
	for _, v := range vals {
		if v.Metadata.Name == rsyncName {
			log.Info("Initializing RPC connection to resource synchronizer", log.Fields{
				"Controller": v.Metadata.Name,
			})
			rsyncInfo := installappclient.NewRsyncInfo(v.Metadata.Name, v.Spec.Host, v.Spec.Port)
			return rsyncInfo, nil
		}
	}
	return installappclient.RsyncInfo{}, pkgerrors.Errorf("queryRsyncInfoInMCODB Failed - Could not get find rsync by name : %v", rsyncName)
}

// callRsyncInstall method shall take in the app context id and invoke the rsync service via grpc
func callRsyncInstall(contextid interface{}) error {
	rsyncInfo, err := queryDBAndSetRsyncInfo()
	log.Info("Calling rsync", log.Fields{
		"RsyncName": rsyncInfo.RsyncName,
	})
	if err != nil {
		log.Error("", log.Fields{"err": err})
		return err
	}

	appContextID := fmt.Sprintf("%v", contextid)
	err = installappclient.InvokeInstallApp(appContextID)
	if err != nil {
		log.Error("", log.Fields{"err": err})
		return err
	}
	return nil
}

// callRsyncReadyNotify method shall take in the app context id and invoke the rsync ready-notify grpc api
func callRsyncReadyNotify(contextid interface{}) error {
	rsyncInfo, err := queryDBAndSetRsyncInfo()
	log.Info("Calling rsync", log.Fields{
		"RsyncName": rsyncInfo.RsyncName,
	})
	if err != nil {
		log.Error("", log.Fields{"err": err})
		return err
	}

	appContextID := fmt.Sprintf("%v", contextid)
	return InvokeReadyNotify(appContextID) // see dcm/pkg/module/client.go
}

// callRsyncUninstall method shall take in the app context id and invoke the rsync service via grpc
func callRsyncUninstall(contextid interface{}) error {
	rsyncInfo, err := queryDBAndSetRsyncInfo()
	log.Info("Calling rsync", log.Fields{
		"RsyncName": rsyncInfo.RsyncName,
	})
	if err != nil {
		log.Error("", log.Fields{"err": err})
		return err
	}

	appContextID := fmt.Sprintf("%v", contextid)
	err = installappclient.InvokeUninstallApp(appContextID)
	if err != nil {
		log.Error("", log.Fields{"err": err})
		return err
	}
	return nil
}

func prepL1ClusterAppContext(logicalcloud LogicalCloud, cluster Cluster, quotaList []Quota, userPermissionList []UserPermission, lcclient *LogicalCloudClient, lckey LogicalCloudKey, context appcontext.AppContext, cid string) error {
	logicalCloudName := logicalcloud.MetaData.LogicalCloudName
	clusterName := strings.Join([]string{cluster.Specification.ClusterProvider, "+", cluster.Specification.ClusterName}, "")
	appHandle, err := context.GetAppHandle("logical-cloud")          // caution: ignoring error
	clusterHandle, err := context.AddCluster(appHandle, clusterName) // caution: ignoring error
	// pre-build array to pass to cleanupCompositeApp() [for performance]
	details := []string{logicalCloudName, clusterName, cid}

	if err != nil {
		return cleanupCompositeApp(context, err, "Error adding Cluster to AppContext", details)
	}

	// Get resources to be added
	namespace, namespaceName, err := createNamespace(logicalcloud)
	if err != nil {
		return pkgerrors.Wrap(err, "Error Creating Namespace YAML for logical cloud")
	}

	roles, roleNames, err := createRoles(logicalcloud, userPermissionList)
	if err != nil {
		return pkgerrors.Wrap(err, "Error Creating Roles/ClusterRoles YAMLs for logical cloud")
	}

	roleBindings, roleBindingNames, err := createRoleBindings(logicalcloud, userPermissionList)
	if err != nil {
		return pkgerrors.Wrap(err, "Error Creating RoleBindings/ClusterRoleBindings YAMLs for logical cloud")
	}

	quotas, quotaNames, err := createQuotas(quotaList, logicalcloud.Specification.NameSpace)
	if err != nil {
		return pkgerrors.Wrap(err, "Error Creating Quota YAMLs for logical cloud")
	}

	csr, key, csrName, err := createUserCSR(logicalcloud)
	if err != nil {
		return pkgerrors.Wrap(err, "Error Creating User CSR and Key for logical cloud")
	}

	approval, err := createApprovalSubresource(logicalcloud)
	if err != nil {
		return pkgerrors.Wrap(err, "Error Creating approval subresource for logical cloud")
	}

	// Add namespace resource to each cluster
	_, err = context.AddResource(clusterHandle, namespaceName, namespace)
	if err != nil {
		return cleanupCompositeApp(context, err, "Error adding Namespace Resource to AppContext", details)
	}

	// Add csr resource to each cluster
	csrHandle, err := context.AddResource(clusterHandle, csrName, csr)
	if err != nil {
		return cleanupCompositeApp(context, err, "Error adding CSR Resource to AppContext", details)
	}

	// Add csr approval as a subresource of csr:
	_, err = context.AddLevelValue(csrHandle, "subresource/approval", approval)
	if err != nil {
		return cleanupCompositeApp(context, err, "Error approving CSR via AppContext", details)
	}

	// Add private key to MongoDB
	err = db.DBconn.Insert(lcclient.storeName, lckey, nil, "privatekey", key)
	if err != nil {
		return cleanupCompositeApp(context, err, "Error adding private key to DB", details)
	}

	// Add [Cluster]Role resources to each cluster
	for i, roleName := range roleNames {
		_, err = context.AddResource(clusterHandle, roleName, roles[i])
		if err != nil {
			return cleanupCompositeApp(context, err, "Error adding [Cluster]Role Resource to AppContext", details)
		}
	}

	// Add [Cluster]RoleBinding resource to each cluster
	for i, roleBindingName := range roleBindingNames {
		_, err = context.AddResource(clusterHandle, roleBindingName, roleBindings[i])
		if err != nil {
			return cleanupCompositeApp(context, err, "Error adding [Cluster]RoleBinding Resource to AppContext", details)
		}
	}

	// Add quota resources to each cluster
	for i, quotaName := range quotaNames {
		_, err = context.AddResource(clusterHandle, quotaName, quotas[i])
		if err != nil {
			return cleanupCompositeApp(context, err, "Error adding quota Resource to AppContext", details)
		}
	}

	// Add Subresource Order and Subresource Dependency
	subresOrder, err := json.Marshal(map[string][]string{"subresorder": []string{"approval"}})
	if err != nil {
		return pkgerrors.Wrap(err, "Error creating subresource order JSON")
	}

	// Add Resource Order
	resorderList := []string{namespaceName, csrName}
	resorderList = append(resorderList, quotaNames...)
	resorderList = append(resorderList, roleNames...)
	resorderList = append(resorderList, roleBindingNames...)
	resOrder, err := json.Marshal(map[string][]string{"resorder": resorderList})
	if err != nil {
		return pkgerrors.Wrap(err, "Error creating resource order JSON")
	}

	// Add Resource-level Order
	_, err = context.AddInstruction(clusterHandle, "resource", "order", string(resOrder))
	if err != nil {
		return cleanupCompositeApp(context, err, "Error adding instruction order to AppContext", details)
	}
	_, err = context.AddInstruction(csrHandle, "subresource", "order", string(subresOrder))
	if err != nil {
		return cleanupCompositeApp(context, err, "Error adding instruction order to AppContext", details)
	}

	return nil
}

// Instantiate prepares all yaml resources to be given to the clusters via rsync,
// then creates an appcontext with such resources and asks rsync to instantiate the logical cloud
func Instantiate(project string, logicalcloud LogicalCloud, clusterList []Cluster,
	quotaList []Quota, userPermissionList []UserPermission) error {
	APP := "logical-cloud"

	logicalCloudName := logicalcloud.MetaData.LogicalCloudName
	level := logicalcloud.Specification.Level

	lcclient := NewLogicalCloudClient()
	lckey := LogicalCloudKey{
		LogicalCloudName: logicalCloudName,
		Project:          project,
	}

	// Check if there was a previous context for this logical cloud
	s, err := lcclient.GetState(project, logicalCloudName)
	if err != nil {
		return err
	}
	cid := state.GetLastContextIdFromStateInfo(s)
	if cid != "" {
		ac, err := state.GetAppContextFromId(cid)
		if err != nil {
			return err
		}
		acStatus, err := GetAppContextStatus(ac)
		if err != nil {
			return err
		}

		// If we're trying to instantiate a stopped termination, first clear the stop flag
		stateVal, err := state.GetCurrentStateFromStateInfo(s)
		if err != nil {
			return err
		}
		if stateVal == state.StateEnum.TerminateStopped {
			err = state.UpdateAppContextStopFlag(cid, false)
			if err != nil {
				return err
			}
		}

		// Make sure rsync status for this logical cloud is Terminated,
		// otherwise we can't re-instantiate logical cloud yet
		switch acStatus.Status {
		case appcontext.AppContextStatusEnum.Terminated:
			// Fully delete the old AppContext
			err := ac.DeleteCompositeApp()
			if err != nil {
				log.Error("Error deleting AppContext CompositeApp Logical Cloud", log.Fields{"logicalcloud": logicalCloudName})
				return pkgerrors.Wrap(err, "Error deleting AppContext CompositeApp Logical Cloud")
			}
		case appcontext.AppContextStatusEnum.Terminating:
			log.Error("The Logical Cloud can't be re-instantiated yet, it is being terminated", log.Fields{"logicalcloud": logicalCloudName})
			return pkgerrors.New("The Logical Cloud can't be re-instantiated yet, it is being terminated")
		case appcontext.AppContextStatusEnum.Instantiated:
			log.Error("The Logical Cloud is already instantiated", log.Fields{"logicalcloud": logicalCloudName})
			return pkgerrors.New("The Logical Cloud is already instantiated")
		case appcontext.AppContextStatusEnum.Instantiating:
			log.Error("The Logical Cloud is already instantiating", log.Fields{"logicalcloud": logicalCloudName})
			return pkgerrors.New("The Logical Cloud is already instantiating")
		case appcontext.AppContextStatusEnum.InstantiateFailed:
			log.Error("The Logical Cloud has failed instantiating before, please terminate and try again", log.Fields{"logicalcloud": logicalCloudName})
			return pkgerrors.New("The Logical Cloud has failed instantiating before, please terminate and try again")
		case appcontext.AppContextStatusEnum.TerminateFailed:
			log.Error("The Logical Cloud has failed terminating, please delete the Logical Cloud", log.Fields{"logicalcloud": logicalCloudName})
			return pkgerrors.New("The Logical Cloud has failed terminating, please delete the Logical Cloud")
		default:
			log.Error("The Logical Cloud isn't in an expected status so not taking any action", log.Fields{"logicalcloud": logicalCloudName, "status": acStatus.Status})
			return pkgerrors.New("The Logical Cloud isn't in an expected status so not taking any action")
		}
	}

	// if this is an L0 logical cloud, only the following will be done as part of instantiate
	// ================================================================================
	if level == "0" {
		l0ns := ""
		// cycle through all clusters to obtain and validate the single level-0 namespace to use
		// the namespace of each cluster is retrieved from CloudConfig in rsync
		for _, cluster := range clusterList {

			ccc := rsync.NewCloudConfigClient()
			log.Info("Asking rsync's CloudConfig for this cluster's namespace at level-0", log.Fields{"cluster": cluster.Specification.ClusterName})
			ns, err := ccc.GetNamespace(
				cluster.Specification.ClusterProvider,
				cluster.Specification.ClusterName,
			)
			if err != nil {
				if err.Error() == "No CloudConfig was returned" {
					return pkgerrors.New("It looks like the cluster provided as reference does not exist")
				}
				return pkgerrors.Wrap(err, "Couldn't determine namespace for L0 logical cloud")
			}
			// we're checking here if any of the clusters have a differently-named namespace at level 0 and, if so,
			// we abort the instantiate operation because a single namespace name for this logical cloud cannot be inferred
			if len(l0ns) > 0 && ns != l0ns {
				log.Error("The clusters associated to this L0 logical cloud don't all share the same namespace name", log.Fields{"logicalcloud": logicalCloudName})
				return pkgerrors.New("The clusters associated to this L0 logical cloud don't all share the same namespace name")
			}
			l0ns = ns
		}
		// if l0ns is still empty, something definitely went wrong so we can't let this pass
		if len(l0ns) == 0 {
			log.Error("Something went wrong as no cluster namespaces got checked", log.Fields{"logicalcloud": logicalCloudName})
			return pkgerrors.New("Something went wrong as no cluster namespaces got checked")
		}
		// at this point we know what namespace name to give to the logical cloud
		logicalcloud.Specification.NameSpace = l0ns
		// the following is an update operation:
		err = db.DBconn.Insert(lcclient.storeName, lckey, nil, lcclient.tagMeta, logicalcloud)
		if err != nil {
			log.Error("Failed to update L0 logical cloud with a namespace name", log.Fields{"logicalcloud": logicalCloudName, "namespace": l0ns})
			return pkgerrors.Wrap(err, "Failed to update L0 logical cloud with a namespace name")
		}
		log.Info("The L0 logical cloud has been updated with a namespace name", log.Fields{"logicalcloud": logicalCloudName, "namespace": l0ns})

		// prepare empty-shell appcontext for the L0 LC in order to officially set it as Instantiated
		context := appcontext.AppContext{}
		ctxVal, err := context.InitAppContext()
		if err != nil {
			return pkgerrors.Wrap(err, "Error creating L0 LC AppContext")
		}
		cid = ctxVal.(string)

		handle, err := context.CreateCompositeApp()
		if err != nil {
			return pkgerrors.Wrap(err, "Error creating L0 LC AppContext CompositeApp")
		}

		appHandle, err := context.AddApp(handle, APP)
		if err != nil {
			return cleanupCompositeApp(context, err, "Error adding App to L0 LC AppContext", []string{logicalCloudName, cid})
		}

		// iterate through cluster list and add all the clusters (as empty-shells)
		for _, cluster := range clusterList {
			clusterName := strings.Join([]string{cluster.Specification.ClusterProvider, "+", cluster.Specification.ClusterName}, "")
			clusterHandle, err := context.AddCluster(appHandle, clusterName)
			// pre-build array to pass to cleanupCompositeApp() [for performance]
			details := []string{logicalCloudName, clusterName, cid}

			if err != nil {
				return cleanupCompositeApp(context, err, "Error adding Cluster to L0 LC AppContext", details)
			}

			// resource-level order is mandatory too for an empty-shell appcontext
			resOrder, err := json.Marshal(map[string][]string{"resorder": []string{}})
			if err != nil {
				return pkgerrors.Wrap(err, "Error creating resource order JSON")
			}
			_, err = context.AddInstruction(clusterHandle, "resource", "order", string(resOrder))
			if err != nil {
				return cleanupCompositeApp(context, err, "Error adding resource-level order to L0 LC AppContext", details)
			}
			// TODO add resource-level dependency as well
			// app-level order is mandatory too for an empty-shell appcontext
			appOrder, err := json.Marshal(map[string][]string{"apporder": []string{APP}})
			if err != nil {
				return pkgerrors.Wrap(err, "Error creating app order JSON")
			}
			_, err = context.AddInstruction(handle, "app", "order", string(appOrder))
			if err != nil {
				return cleanupCompositeApp(context, err, "Error adding app-level order to L0 LC AppContext", details)
			}
			// TODO add app-level dependency as well
			// TODO move app-level order/dependency out of loop
		}

		// call resource synchronizer to instantiate the CRs in the cluster
		err = callRsyncInstall(ctxVal)
		if err != nil {
			log.Error("Failed calling rsync install-app", log.Fields{"err": err})
			return pkgerrors.Wrap(err, "Failed calling rsync install-app")
		}

		// Update state with switch to Instantiated state, along with storing the AppContext ID for future retrieval
		err = addState(lcclient, project, logicalCloudName, cid, state.StateEnum.Instantiated)
		if err != nil {
			return cleanupCompositeApp(context, err, "Error adding L0 LC AppContext to DB", []string{logicalCloudName, cid})
		}

		log.Info("The L0 logical cloud is now associated with an empty-shell appcontext and is ready to be used", log.Fields{"logicalcloud": logicalCloudName, "namespace": l0ns})
		return nil
	}

	if len(userPermissionList) == 0 {
		return pkgerrors.New("Level-1 Logical Clouds require at least a User Permission assigned to its primary namespace")
	}

	primaryUP := false
	for _, up := range userPermissionList {
		if up.Specification.Namespace == logicalcloud.Specification.NameSpace {
			primaryUP = true
			break
		}
	}
	if !primaryUP {
		return pkgerrors.New("Level-1 Logical Clouds require a User Permission assigned to its primary namespace")
	}

	// From this point on, we are dealing with a new AppContext
	context := appcontext.AppContext{}
	ctxVal, err := context.InitAppContext()
	if err != nil {
		return pkgerrors.Wrap(err, "Error creating AppContext")
	}
	cid = ctxVal.(string)

	handle, err := context.CreateCompositeApp()
	if err != nil {
		return pkgerrors.Wrap(err, "Error creating AppContext CompositeApp")
	}

	_, err = context.AddApp(handle, APP)
	if err != nil {
		return cleanupCompositeApp(context, err, "Error adding App to AppContext", []string{logicalCloudName, cid})
	}

	// Create a Logical Cloud Meta that will store the project and the logical cloud name in the AppContext:
	err = context.AddCompositeAppMeta(appcontext.CompositeAppMeta{Project: project, LogicalCloud: logicalCloudName})
	if err != nil {
		return pkgerrors.Wrap(err, "Error Adding Logical Cloud Meta to AppContext")
	}

	// Create a Logical Cloud Meta with all data needed for a successful L1 (standard/privileged) instantiation:
	// project name, logical cloud name, level=0 and namespace=default (for rsync cluster access - may get modularized in the future)
	err = context.AddCompositeAppMeta(
		appcontext.CompositeAppMeta{
			Project:      project,
			LogicalCloud: logicalCloudName,
			Level:        "0",
			Namespace:    "default"})
	if err != nil {
		return cleanupCompositeApp(context, err, "Error Adding Logical Cloud Meta to AppContext", []string{logicalCloudName, cid})
	}

	// Add App Order and App Dependency
	appOrder, err := json.Marshal(map[string][]string{"apporder": []string{APP}})
	if err != nil {
		return pkgerrors.Wrap(err, "Error creating app order JSON")
	}
	appDependency, err := json.Marshal(map[string]map[string]string{"appdependency": map[string]string{APP: "go"}})
	if err != nil {
		return pkgerrors.Wrap(err, "Error creating app dependency JSON")
	}

	// Add App-level Order and Dependency
	_, err = context.AddInstruction(handle, "app", "order", string(appOrder))
	if err != nil {
		return cleanupCompositeApp(context, err, "Error adding app-level order to AppContext", []string{logicalCloudName})
	}
	_, err = context.AddInstruction(handle, "app", "dependency", string(appDependency))
	if err != nil {
		return cleanupCompositeApp(context, err, "Error adding app-level dependency to AppContext", []string{logicalCloudName})
	}

	// Iterate through cluster list and add all the clusters
	for _, cluster := range clusterList {
		prepL1ClusterAppContext(logicalcloud, cluster, quotaList, userPermissionList, lcclient, lckey, context, cid)
	}

	// call resource synchronizer to instantiate the CRs in the cluster
	err = callRsyncInstall(ctxVal)
	if err != nil {
		return err
	}

	// Update state with switch to Instantiated state, along with storing the AppContext ID for future retrieval
	err = addState(lcclient, project, logicalCloudName, cid, state.StateEnum.Instantiated)
	if err != nil {
		return cleanupCompositeApp(context, err, "Error adding L1 LC AppContext to DB", []string{logicalCloudName, cid})
	}

	// call grpc streaming api in rsync, which launches a goroutine to wait for the response of
	// every cluster (function should know how many clusters are expected and only finish when
	// all respective certificates have been obtained and all kubeconfigs stored in CloudConfig)
	err = callRsyncReadyNotify(ctxVal)
	if err != nil {
		log.Error("Failed calling rsync ready-notify", log.Fields{"err": err})
		return pkgerrors.Wrap(err, "Failed calling rsync ready-notify")
	}

	return nil

}

// Terminate asks rsync to terminate the logical cloud
func Terminate(project string, logicalcloud LogicalCloud, clusterList []Cluster,
	quotaList []Quota) error {

	logicalCloudName := logicalcloud.MetaData.LogicalCloudName
	level := logicalcloud.Specification.Level
	namespace := logicalcloud.Specification.NameSpace

	lcclient := NewLogicalCloudClient()

	// Check if there was a previous context for this logical cloud
	s, err := lcclient.GetState(project, logicalCloudName)
	if err != nil {
		return err
	}
	cid := state.GetLastContextIdFromStateInfo(s)
	if cid != "" {
		ac, err := state.GetAppContextFromId(cid)
		if err != nil {
			return err
		}

		// If we're trying to terminate a stopped instantiation, first clear the stop flag
		stateVal, err := state.GetCurrentStateFromStateInfo(s)
		if err != nil {
			return err
		}
		if stateVal == state.StateEnum.InstantiateStopped {
			err = state.UpdateAppContextStopFlag(cid, false)
			if err != nil {
				return err
			}
		}

		// Make sure rsync status for this logical cloud is Terminated,
		// otherwise we can't re-instantiate logical cloud yet
		acStatus, err := GetAppContextStatus(ac)
		if err != nil {
			return err
		}
		switch acStatus.Status {
		case appcontext.AppContextStatusEnum.Terminated:
			log.Error("The Logical Cloud has already been terminated", log.Fields{"logicalcloud": logicalCloudName})
			return pkgerrors.New("The Logical Cloud has already been terminated")
		case appcontext.AppContextStatusEnum.Terminating:
			log.Error("The Logical Cloud is already being terminated", log.Fields{"logicalcloud": logicalCloudName})
			return pkgerrors.New("The Logical Cloud is already being terminated")
		case appcontext.AppContextStatusEnum.Instantiating:
			log.Error("The Logical Cloud is still instantiating", log.Fields{"logicalcloud": logicalCloudName})
			return pkgerrors.New("The Logical Cloud is still instantiating")
		case appcontext.AppContextStatusEnum.TerminateFailed:
			log.Error("The Logical Cloud has failed terminating, please delete the Logical Cloud", log.Fields{"logicalcloud": logicalCloudName})
			return pkgerrors.New("The Logical Cloud has failed terminating, please delete the Logical Cloud")
		case appcontext.AppContextStatusEnum.InstantiateFailed:
			// try to terminate anyway
			fallthrough
		case appcontext.AppContextStatusEnum.Instantiated:
			// call resource synchronizer to delete the CRs from every cluster of the logical cloud
			err = callRsyncUninstall(cid)
			if err != nil {
				return err
			}
			// destroy kubeconfigs from CloudConfig if this is an L1 logical cloud
			if level == "1" {

				ccc := rsync.NewCloudConfigClient()
				for _, cluster := range clusterList {
					log.Info("Destroying CloudConfig of logicalcloud/cluster pair via rsync", log.Fields{"cluster": cluster.Specification.ClusterName, "logicalcloud": logicalCloudName, "level": level})
					err = ccc.DeleteCloudConfig(
						cluster.Specification.ClusterProvider,
						cluster.Specification.ClusterName,
						level,
						namespace,
					)

					if err != nil {
						log.Error("Failed destroying at least one CloudConfig of L1 LC", log.Fields{"cluster": cluster, "err": err})
						// continue terminating and removing any remaining CloudConfigs
						// (this happens when terminating a Logical Cloud before all kubeconfigs had a chance to be generated, such as after a stopped instantiation)
					}
				}
			}

			// Set State as Terminated
			err = addState(lcclient, project, logicalCloudName, cid, state.StateEnum.Terminated)
			if err != nil {
				return err // error already logged
			}

		default:
			log.Error("The Logical Cloud isn't in an expected status so not taking any action", log.Fields{"logicalcloud": logicalCloudName, "status": acStatus.Status})
			return pkgerrors.New("The Logical Cloud isn't in an expected status so not taking any action")
		}
	}
	return nil
}

// Stop asks rsync to stop the instantiation or termination of the logical cloud
func Stop(project string, logicalcloud LogicalCloud) error {

	logicalCloudName := logicalcloud.MetaData.LogicalCloudName
	lcclient := NewLogicalCloudClient()

	// Find and deal with state
	s, err := lcclient.GetState(project, logicalCloudName)
	if err != nil {
		return err
	}

	stateVal, err := state.GetCurrentStateFromStateInfo(s)
	if err != nil {
		return pkgerrors.Wrap(err, "Error getting current state from LogicalCloud stateInfo: "+logicalCloudName)
	}

	cid := state.GetLastContextIdFromStateInfo(s)
	// Uncomment to use appcontext status if additional info is needed to inform the actual state transition that should take place:
	// ac, err := state.GetAppContextFromId(cid)
	// if err != nil {
	// 	return err
	// }
	// acStatus, err := GetAppContextStatus(ac)

	stopState := state.StateEnum.Undefined
	switch stateVal {
	case state.StateEnum.Created:
		return pkgerrors.New("LogicalCloud has not been asked to instantiate: " + logicalCloudName)
	case state.StateEnum.Instantiated:
		stopState = state.StateEnum.InstantiateStopped
	case state.StateEnum.Terminated:
		stopState = state.StateEnum.TerminateStopped
	case state.StateEnum.TerminateStopped:
		return pkgerrors.New("LogicalCloud termination already stopped: " + logicalCloudName)
	case state.StateEnum.InstantiateStopped:
		return pkgerrors.New("LogicalCloud instantiation already stopped: " + logicalCloudName)
	default:
		return pkgerrors.New("LogicalCloud is in an invalid state: " + logicalCloudName + " " + stateVal)
	}

	err = state.UpdateAppContextStopFlag(cid, true)
	if err != nil {
		return err
	}

	err = addState(lcclient, project, logicalCloudName, cid, stopState)
	if err != nil {
		return pkgerrors.Wrap(err, "Error updating the stateInfo of the LogicalCloud: "+logicalCloudName)
	}

	return nil
}
