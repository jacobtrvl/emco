// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package distribution

import (
	"strings"

	"github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certissuer/certmanagerissuer"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/service/istioservice"
)

// createSecret creates a secret to store the certificate
func (ctx *DistributionContext) createSecret(cr certmanagerissuer.CertificateRequest, name, namespace string) error {
	// retrive the Private Key from mongo
	key, err := ctx.retrivePrivateKey()
	if err != nil {
		return err
	}

	data := map[string]string{}
	data["tls.crt"] = cr.Status.Certificate
	data["tls.key"] = key

	s := certmanagerissuer.CreateSecret(name, namespace, data)
	if err := module.AddResource(ctx.AppContext, s, ctx.ClusterHandle, s.ResourceName()); err != nil {
		return err
	}

	ctx.ResOrder = append(ctx.ResOrder, s.ResourceName())

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
	if err := module.AddResource(ctx.AppContext, i, ctx.ClusterHandle, i.ResourceName()); err != nil {
		return err
	}

	ctx.ResOrder = append(ctx.ResOrder, i.ResourceName())
	ctx.Resources.ClusterIssuer = append(ctx.Resources.ClusterIssuer, certmanagerissuer.ClusterIssuer{
		MetaData: i.MetaData,
		Spec:     i.Spec}) // this is needed to create the proxyconfig

	return nil
}

// createProxyConfig creates a ProxyConfig to control the traffic between workloads
func (ctx *DistributionContext) createProxyConfig(issuer certmanagerissuer.ClusterIssuer) error {
	ns := ""
	if len(ctx.Namespace) > 0 &&
		strings.ToLower(ctx.Namespace) != "default" {
		ns = ctx.Namespace
	}

	environmentVariables := map[string]string{}
	environmentVariables["ISTIO_META_CERT_SIGNER"] = issuer.MetaData.Name
	pcName := istioservice.ProxyConfigName(ctx.ContextID, ctx.CaCert.MetaData.Name, ctx.ClusterGroup.Spec.Provider, ctx.Cluster)
	pc := istioservice.CreateProxyConfig(pcName, ns, environmentVariables)

	if err := module.AddResource(ctx.AppContext, pc, ctx.ClusterHandle, pc.ResourceName()); err != nil {
		return err
	}

	ctx.ResOrder = append(ctx.ResOrder, pc.ResourceName())

	return nil
}

// retrivePrivateKey
func (ctx *DistributionContext) retrivePrivateKey() (string, error) {
	dbKey := module.DBKey{
		Cert:            ctx.CaCert.MetaData.Name,
		Cluster:         ctx.Cluster,
		ClusterProvider: ctx.ClusterGroup.Spec.Provider,
		ContextID:       ctx.EnrollmentContextID}

	key, err := module.NewKeyClient(dbKey).Get()
	if err != nil {
		return "", err
	}

	if key.Name != certmanagerissuer.CertificateRequestName(ctx.EnrollmentContextID, ctx.CaCert.MetaData.Name, ctx.ClusterGroup.Spec.Provider, ctx.Cluster) {
		return "", errors.New("PrivateKey not found")
	}

	return key.Val, nil
}

// retrieveClusterIssuer
func (ctx *DistributionContext) retrieveClusterIssuer(cluster string) certmanagerissuer.ClusterIssuer {
	var iName string
	for _, issuer := range ctx.Resources.ClusterIssuer {
		iName = certmanagerissuer.ClusterIssuerName(ctx.ContextID, ctx.CaCert.MetaData.Name, ctx.ClusterGroup.Spec.Provider, cluster)
		if issuer.MetaData.Name == iName {
			return issuer
		}
	}

	return certmanagerissuer.ClusterIssuer{}
}
