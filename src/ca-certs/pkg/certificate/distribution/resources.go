// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package distribution

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	tcsv1 "github.com/intel/trusted-certificate-issuer/api/v1alpha1"
	"github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certissuer/certmanagerissuer"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certissuer/tcsissuer"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/service/istioservice"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/service/knccservice"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	v1 "k8s.io/api/core/v1"
)

// createSecret creates a secret to store the certificate
func (ctx *DistributionContext) createSecret(cr cmv1.CertificateRequest, name, namespace string) error {
	if s, exists := ctx.Resources.Secret[name]; exists {
		// a secret already exists, update the context resource order
		ctx.ResOrder = append(ctx.ResOrder, module.ResourceName(s.ObjectMeta.Name, s.Kind))
		return nil
	}
	// retrieve the Private Key from mongo
	key, err := ctx.retrievePrivateKey()
	if err != nil {
		return err
	}

	data := map[string][]byte{}
	data[v1.TLSCertKey] = bytes.Join([][]byte{cr.Status.Certificate, cr.Status.CA}, []byte{}) // create the caCert chain
	data[v1.TLSPrivateKeyKey] = key

	s := certmanagerissuer.CreateSecret(name, namespace, data)
	if err := module.AddResource(ctx.AppContext, s, ctx.ClusterHandle, module.ResourceName(s.ObjectMeta.Name, s.Kind)); err != nil {
		return err
	}

	ctx.ResOrder = append(ctx.ResOrder, module.ResourceName(s.ObjectMeta.Name, s.Kind))

	ctx.Resources.Secret[name] = s // this is needed to create the kncc config

	return nil
}

// createClusterIssuer creates a clusterIssuer to issue the certificates
func (ctx *DistributionContext) createClusterIssuer(secretName string) error {
	iName := certmanagerissuer.ClusterIssuerName(ctx.ContextID, ctx.CaCert.MetaData.Name, ctx.ClusterGroup.Spec.Provider, ctx.Cluster)
	if i, exists := ctx.Resources.ClusterIssuer[iName]; exists {
		// a clusterIssuer already exists, update the context resource order
		ctx.ResOrder = append(ctx.ResOrder, module.ResourceName(i.ObjectMeta.Name, i.Kind))
		return nil
	}

	i := certmanagerissuer.CreateClusterIssuer(iName, secretName)
	if err := module.AddResource(ctx.AppContext, i, ctx.ClusterHandle, module.ResourceName(i.ObjectMeta.Name, i.Kind)); err != nil {
		return err
	}

	ctx.ResOrder = append(ctx.ResOrder, module.ResourceName(i.ObjectMeta.Name, i.Kind))
	ctx.Resources.ClusterIssuer[iName] = i // this is needed to create the proxyconfig

	return nil
}

// createProxyConfig creates a proxyConfig to control the traffic between workloads
func (ctx *DistributionContext) createProxyConfig(issuer string) error {
	pcName := istioservice.ProxyConfigName(ctx.ContextID, ctx.CaCert.MetaData.Name, ctx.ClusterGroup.Spec.Provider, ctx.Cluster, ctx.Namespace)
	if pc, exists := ctx.Resources.ProxyConfig[pcName]; exists {
		// a proxyConfig already exists, update the context resource order
		ctx.ResOrder = append(ctx.ResOrder, module.ResourceName(pc.MetaData.Name, pc.Kind))
		return nil
	}

	environmentVariables := map[string]string{}
	environmentVariables["ISTIO_META_CERT_SIGNER"] = issuer
	pc := istioservice.CreateProxyConfig(pcName, ctx.Namespace, environmentVariables)

	if err := module.AddResource(ctx.AppContext, pc, ctx.ClusterHandle, module.ResourceName(pc.MetaData.Name, pc.Kind)); err != nil {
		return err
	}

	ctx.Resources.ProxyConfig[pcName] = pc

	ctx.ResOrder = append(ctx.ResOrder, module.ResourceName(pc.MetaData.Name, pc.Kind))

	for _, p := range ctx.Resources.ProxyConfig {
		// update the context resource order with any other proxyConfig created for the same appContext, caCert, clusterProvider and cluster
		if strings.Contains(p.MetaData.Name, fmt.Sprintf("%s-%s-%s-%s", ctx.ContextID, ctx.CaCert.MetaData.Name, ctx.ClusterGroup.Spec.Provider, ctx.Cluster)) {
			exists := false
			for _, o := range ctx.ResOrder {
				if o == module.ResourceName(p.MetaData.Name, p.Kind) {
					exists = true
					break
				}
			}

			if !exists {
				ctx.ResOrder = append(ctx.ResOrder, module.ResourceName(p.MetaData.Name, p.Kind))
			}
		}
	}

	return nil
}

// createKnccConfig creates a kncc config patch to update the configMap
func (ctx *DistributionContext) createKnccConfig(namespace, resourceName, resourceNamespace string,
	patch []map[string]string) error {
	cName := knccservice.KnccConfigName(ctx.ContextID, ctx.CaCert.MetaData.Name, ctx.ClusterGroup.Spec.Provider, ctx.Cluster)
	if c, exists := ctx.Resources.KnccConfig[cName]; exists {
		// a KnccConfig already exists, update the config
		return ctx.updateKnccConfig(patch, c)
	}

	c := knccservice.CreateKnccConfig(cName, namespace, resourceName, resourceNamespace, patch)
	if err := module.AddResource(ctx.AppContext, c, ctx.ClusterHandle, module.ResourceName(c.ObjectMeta.Name, c.TypeMeta.Kind)); err != nil {
		return err
	}

	ctx.Resources.KnccConfig[cName] = c

	ctx.ResOrder = append(ctx.ResOrder, module.ResourceName(c.ObjectMeta.Name, c.TypeMeta.Kind))

	return nil
}

// retrievePrivateKey retrieves the rsa private key from mongo
func (ctx *DistributionContext) retrievePrivateKey() ([]byte, error) {
	dbKey := module.DBKey{
		Cert:            ctx.CaCert.MetaData.Name,
		Cluster:         ctx.Cluster,
		ClusterProvider: ctx.ClusterGroup.Spec.Provider,
		ContextID:       ctx.EnrollmentContextID}

	key, err := module.NewKeyClient(dbKey).Get()
	if err != nil {
		return []byte{}, err
	}

	if key.Name != certmanagerissuer.CertificateRequestName(ctx.EnrollmentContextID, ctx.CaCert.MetaData.Name, ctx.ClusterGroup.Spec.Provider, ctx.Cluster) {
		err := errors.New("PrivateKey not found")
		logutils.Error("",
			logutils.Fields{
				"CaCert": ctx.CaCert.MetaData.Name,
				"Error":  err.Error()})
		return []byte{}, err
	}

	return base64.StdEncoding.DecodeString(key.Val)
}

// retrieveClusterIssuer returns the specific issuer from the distribution resources list
func (ctx *DistributionContext) retrieveClusterIssuer(cluster string) *cmv1.ClusterIssuer {
	var iName string
	for _, issuer := range ctx.Resources.ClusterIssuer {
		iName = certmanagerissuer.ClusterIssuerName(ctx.ContextID, ctx.CaCert.MetaData.Name, ctx.ClusterGroup.Spec.Provider, cluster)
		if issuer.ObjectMeta.Name == iName {
			return issuer
		}
	}

	return &cmv1.ClusterIssuer{}
}

// retrieveSecret returns the specific secret from the distribution resources list
func (ctx *DistributionContext) retrieveSecret(cluster string) *v1.Secret {
	var sName string
	for _, s := range ctx.Resources.Secret {
		sName = certmanagerissuer.SecretName(ctx.ContextID, ctx.CaCert.MetaData.Name, ctx.ClusterGroup.Spec.Provider, cluster)
		if s.ObjectMeta.Name == sName {
			return s
		}
	}

	return &v1.Secret{}
}

// updateKnccConfig updates the kncc resource patch
func (ctx *DistributionContext) updateKnccConfig(patch []map[string]string, c *knccservice.Config) error {
	// update kncc config
	knccservice.UpdateKnccConfig(patch, c)
	// update resource appContext
	if err := module.AddResource(ctx.AppContext, c, ctx.ClusterHandle, module.ResourceName(c.ObjectMeta.Name, c.TypeMeta.Kind)); err != nil {
		return err
	}
	// store the new config
	ctx.Resources.KnccConfig[c.ObjectMeta.Name] = c
	// update the resource order
	ctx.ResOrder = append(ctx.ResOrder, module.ResourceName(c.ObjectMeta.Name, c.TypeMeta.Kind))

	return nil
}

// createTCSIssuer creates a TCSIssuer to issue the certificates
func (ctx *DistributionContext) createTCSIssuer(secret string, selfSign bool) error {
	iName := tcsissuer.TCSIssuerName(ctx.EnrollmentContextID, ctx.CaCert.MetaData.Name, ctx.ClusterGroup.Spec.Provider, ctx.Cluster)
	if i, exists := ctx.Resources.TCSIssuer[iName]; exists {
		// a tcsIssuer already exists, update the context resource order
		ctx.ResOrder = append(ctx.ResOrder, module.ResourceName(i.ObjectMeta.Name, i.Kind))
		return nil
	}

	i := tcsissuer.CreateTCSIssuer(iName, secret, selfSign)
	if err := module.AddResource(ctx.AppContext, i, ctx.ClusterHandle, module.ResourceName(i.ObjectMeta.Name, i.Kind)); err != nil {
		return err
	}

	ctx.ResOrder = append(ctx.ResOrder, module.ResourceName(i.ObjectMeta.Name, i.Kind))
	ctx.Resources.TCSIssuer[iName] = i // this is needed to create the proxyconfig

	return nil
}

// retrieveTCSIssuer returns the specific tcsissuer from the distribution resources list
func (ctx *DistributionContext) retrieveTCSIssuer(cluster string) *tcsv1.TCSIssuer {
	var iName string
	for _, issuer := range ctx.Resources.TCSIssuer {
		iName = tcsissuer.TCSIssuerName(ctx.EnrollmentContextID, ctx.CaCert.MetaData.Name, ctx.ClusterGroup.Spec.Provider, cluster)
		if issuer.ObjectMeta.Name == iName {
			return issuer
		}
	}

	return &tcsv1.TCSIssuer{}
}
