// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package enrollment

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"reflect"
	"strings"

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certificate"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certissuer/certmanagerissuer"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/certissuer/tcsissuer"
	"gitlab.com/project-emco/core/emco-base/src/ca-certs/pkg/module"
	clm "gitlab.com/project-emco/core/emco-base/src/clm/pkg/cluster"
)

// createCertManagerCertificateRequest creates a certificaterequest to generate the certificate
func (ctx *EnrollmentContext) createCertManagerCertificateRequest() error {
	name := certmanagerissuer.CertificateRequestName(ctx.ContextID, ctx.CaCert.MetaData.Name, ctx.ClusterGroup.Spec.Provider, ctx.Cluster)
	if _, exists := ctx.Resources.CertificateRequest[name]; exists {
		// a certificaterequest already exists
		return nil
	}

	commonName := strings.Join([]string{ctx.CaCert.MetaData.Name, ctx.ClusterGroup.Spec.Provider, ctx.Cluster, "ca"}, "-")
	if len(ctx.CaCert.Spec.CertificateSigningInfo.Subject.Names.CommonNamePrefix) > 0 {
		commonName = strings.Join([]string{commonName, ctx.CaCert.Spec.CertificateSigningInfo.Subject.Names.CommonNamePrefix}, "-")
	}

	// This needs to be a unique name for each cluster
	ctx.CaCert.Spec.CertificateSigningInfo.Subject.Names.CommonName = commonName

	// check if a cluster specific commonName is available
	if val, err := clm.NewClusterClient().GetClusterKvPairsValue(context.Background(), ctx.ClusterGroup.Spec.Provider, ctx.Cluster, "csrData", "commonName"); err == nil {
		v, e := module.GetValue(val)
		if e != nil {
			return e
		}

		ctx.CaCert.Spec.CertificateSigningInfo.Subject.Names.CommonName = v
	}

	// generate the private key for the csr
	pemBlock, err := certificate.GeneratePrivateKey(ctx.CaCert.Spec.CertificateSigningInfo.KeySize)
	if err != nil {
		return err
	}

	// parse the RSA key in PKCS #1, ASN.1 DER form
	pk, err := certificate.ParsePrivateKey(pemBlock.Bytes)
	if err != nil {
		return err
	}

	// create a certificate signing request
	request, err := ctx.createCertificateSigningRequest(pk)
	if err != nil {
		return err
	}

	// create the cert-manager CertificateRequest resource
	cr, err := certmanagerissuer.CreateCertificateRequest(ctx.CaCert, name, request)
	if err != nil {
		return err
	}

	// add the CertificateRequest resource in to the appContext
	if err := module.AddResource(ctx.AppContext, cr, ctx.IssuerHandle, module.ResourceName(cr.ObjectMeta.Name, cr.TypeMeta.Kind)); err != nil {
		return err
	}

	ctx.Resources.CertificateRequest[name] = cr

	ctx.ResOrder = append(ctx.ResOrder, module.ResourceName(cr.ObjectMeta.Name, cr.TypeMeta.Kind))

	// save the PK in mongo
	return ctx.savePrivateKey(name, base64.StdEncoding.EncodeToString(pem.EncodeToMemory(pemBlock)))

}

// createCertificateSigningRequest creates a certificate signing request
func (ctx *EnrollmentContext) createCertificateSigningRequest(pk *rsa.PrivateKey) ([]byte, error) {
	return certificate.CreateCertificateSigningRequest(x509.CertificateRequest{
		Version:            ctx.CaCert.Spec.CertificateSigningInfo.Version,
		SignatureAlgorithm: certificate.SignatureAlgorithm(ctx.CaCert.Spec.CertificateSigningInfo.Algorithm.SignatureAlgorithm),
		PublicKeyAlgorithm: certificate.PublicKeyAlgorithm(ctx.CaCert.Spec.CertificateSigningInfo.Algorithm.PublicKeyAlgorithm),
		Subject: pkix.Name{
			Country:            ctx.CaCert.Spec.CertificateSigningInfo.Subject.Locale.Country,
			Locality:           ctx.CaCert.Spec.CertificateSigningInfo.Subject.Locale.Locality,
			PostalCode:         ctx.CaCert.Spec.CertificateSigningInfo.Subject.Locale.PostalCode,
			Province:           ctx.CaCert.Spec.CertificateSigningInfo.Subject.Locale.Province,
			StreetAddress:      ctx.CaCert.Spec.CertificateSigningInfo.Subject.Locale.StreetAddress,
			CommonName:         ctx.CaCert.Spec.CertificateSigningInfo.Subject.Names.CommonName,
			Organization:       ctx.CaCert.Spec.CertificateSigningInfo.Subject.Organization.Names,
			OrganizationalUnit: ctx.CaCert.Spec.CertificateSigningInfo.Subject.Organization.Units},
		DNSNames:       ctx.CaCert.Spec.CertificateSigningInfo.DNSNames,
		EmailAddresses: ctx.CaCert.Spec.CertificateSigningInfo.EmailAddresses}, pk)
}

// savePrivateKey saves the rsa private key in mongo
func (ctx *EnrollmentContext) savePrivateKey(name, val string) error {
	dbKey := module.DBKey{
		Cert:            ctx.CaCert.MetaData.Name,
		Cluster:         ctx.Cluster,
		ClusterProvider: ctx.ClusterGroup.Spec.Provider,
		ContextID:       ctx.ContextID}
	key := module.Key{
		Name: name,
		Val:  val}

	return module.NewKeyClient(dbKey).Save(key)
}

// deletePrivateKey delete the rsa private key from mongo
func (ctx *EnrollmentContext) deletePrivateKey() error {
	dbKey := module.DBKey{
		Cert:            ctx.CaCert.MetaData.Name,
		Cluster:         ctx.Cluster,
		ClusterProvider: ctx.ClusterGroup.Spec.Provider,
		ContextID:       ctx.ContextID}

	return module.NewKeyClient(dbKey).Delete()
}

// privateKeyExists verifies the rsa private key exists in mongo
func (ctx *EnrollmentContext) privateKeyExists() bool {
	dbKey := module.DBKey{
		Cert:            ctx.CaCert.MetaData.Name,
		Cluster:         ctx.Cluster,
		ClusterProvider: ctx.ClusterGroup.Spec.Provider,
		ContextID:       ctx.ContextID}

	if k, err := module.NewKeyClient(dbKey).Get(); err == nil &&
		!reflect.DeepEqual(k, module.Key{}) {
		return true
	}

	return false
}

// createCertManagerCertificate creates a cert-manager certificate resource to generate the key and the cert
func (ctx *EnrollmentContext) createCertManagerCertificate() error {
	name := certmanagerissuer.CertificateName(ctx.ContextID, ctx.CaCert.MetaData.Name, ctx.ClusterGroup.Spec.Provider, ctx.Cluster)
	if _, exists := ctx.Resources.Certificate[name]; exists {
		// a certificate already exists
		return nil
	}

	// create a Secret to store the certificate
	sName := certmanagerissuer.SecretName(ctx.ContextID, ctx.CaCert.MetaData.Name, ctx.ClusterGroup.Spec.Provider, ctx.Cluster)
	// specify the issuer which uses this secret
	iName := tcsissuer.TCSIssuerName(ctx.ContextID, ctx.CaCert.MetaData.Name, ctx.ClusterGroup.Spec.Provider, ctx.Cluster)

	ns := "default"
	if len(ctx.Namespace) > 0 {
		ns = ctx.Namespace
	}

	secretTemplate := cmv1.CertificateSecretTemplate{
		Annotations: map[string]string{
			"issuer-name":      iName,
			"issuer-namespace": ns,
		},
		Labels: map[string]string{
			"emco/deployment-id": fmt.Sprintf("%s-%s", ctx.ContextID, AppName),
		},
	}

	cert, err := certmanagerissuer.CreateCertificate(ctx.CaCert, name, sName, secretTemplate)
	if err != nil {
		return err
	}

	// add the Certificate resource in to the appContext
	if err := module.AddResource(ctx.AppContext, cert, ctx.IssuerHandle, module.ResourceName(cert.ObjectMeta.Name, cert.TypeMeta.Kind)); err != nil {
		return err
	}

	ctx.Resources.Certificate[name] = cert

	ctx.ResOrder = append(ctx.ResOrder, module.ResourceName(cert.ObjectMeta.Name, cert.TypeMeta.Kind))

	return nil
}
