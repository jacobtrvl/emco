// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package distribution

import (
	"encoding/base64"
	"strings"

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certissuer/certmanagerissuer"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/service/istioservice"
	v1 "k8s.io/api/core/v1"
)

// createSecret creates a secret to store the certificate
func (ctx *DistributionContext) createSecret(cr cmv1.CertificateRequest, name, namespace string) error {
	// retrieve the Private Key from mongo
	key, err := ctx.retrievePrivateKey()
	if err != nil {
		return err
	}

	data := map[string][]byte{}
	data[v1.TLSCertKey] = cr.Status.Certificate
	data[v1.TLSPrivateKeyKey] = key

	s := certmanagerissuer.CreateSecret(name, namespace, data)
	if err := module.AddResource(ctx.AppContext, s, ctx.ClusterHandle, module.ResourceName(s.ObjectMeta.Name, s.Kind)); err != nil {
		return err
	}

	ctx.ResOrder = append(ctx.ResOrder, module.ResourceName(s.ObjectMeta.Name, s.Kind))

	return nil
}

// createClusterIssuer creates a ClusterIssuer to issue the certificates
func (ctx *DistributionContext) createClusterIssuer(secretName string) error {
	ns := ""
	if len(ctx.Namespace) > 0 &&
		strings.ToLower(ctx.Namespace) != "default" {
		ns = ctx.Namespace
	}

	iName := certmanagerissuer.ClusterIssuerName(ctx.ContextID, ctx.CaCert.MetaData.Name, ctx.ClusterGroup.Spec.Provider, ctx.Cluster)
	i := certmanagerissuer.CreateClusterIssuer(iName, ns, secretName)
	if err := module.AddResource(ctx.AppContext, i, ctx.ClusterHandle, module.ResourceName(i.ObjectMeta.Name, i.Kind)); err != nil {
		return err
	}

	ctx.ResOrder = append(ctx.ResOrder, module.ResourceName(i.ObjectMeta.Name, i.Kind))
	ctx.Resources.ClusterIssuer = append(ctx.Resources.ClusterIssuer, cmv1.ClusterIssuer{
		ObjectMeta: i.ObjectMeta,
		Spec:       i.Spec}) // this is needed to create the proxyconfig

	return nil
}

// createProxyConfig creates a ProxyConfig to control the traffic between workloads
func (ctx *DistributionContext) createProxyConfig(issuer cmv1.ClusterIssuer) error {
	ns := ""
	if len(ctx.Namespace) > 0 &&
		strings.ToLower(ctx.Namespace) != "default" {
		ns = ctx.Namespace
	}

	environmentVariables := map[string]string{}
	environmentVariables["ISTIO_META_CERT_SIGNER"] = issuer.ObjectMeta.Name
	pcName := istioservice.ProxyConfigName(ctx.ContextID, ctx.CaCert.MetaData.Name, ctx.ClusterGroup.Spec.Provider, ctx.Cluster)
	pc := istioservice.CreateProxyConfig(pcName, ns, environmentVariables)

	if err := module.AddResource(ctx.AppContext, pc, ctx.ClusterHandle, module.ResourceName(pc.MetaData.Name, pc.Kind)); err != nil {
		return err
	}

	ctx.ResOrder = append(ctx.ResOrder, module.ResourceName(pc.MetaData.Name, pc.Kind))

	return nil
}

// retrievePrivateKey
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
		return []byte{}, errors.New("PrivateKey not found")
	}

	return base64.StdEncoding.DecodeString(key.Val)
}

// retrieveClusterIssuer
func (ctx *DistributionContext) retrieveClusterIssuer(cluster string) cmv1.ClusterIssuer {
	var iName string
	for _, issuer := range ctx.Resources.ClusterIssuer {
		iName = certmanagerissuer.ClusterIssuerName(ctx.ContextID, ctx.CaCert.MetaData.Name, ctx.ClusterGroup.Spec.Provider, cluster)
		if issuer.ObjectMeta.Name == iName {
			return issuer
		}
	}

	return cmv1.ClusterIssuer{}
}