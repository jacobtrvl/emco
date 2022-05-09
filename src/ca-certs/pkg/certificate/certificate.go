// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package certificate

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"fmt"

	"github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

// Cert

type ClusterProviderCertRequestKey struct {
	Cert            string `json:"clusterProviderCert"`
	Cluster         string `json:"clusterProviderCluster"`
	ClusterGroup    string `json:"clusterProviderClusterGroup"`
	ClusterProvider string `json:"clusterProvider"`
	CertRequest     string `json:"clusterProviderCertRequest"`
}

// CertificateClient
type CertificateClient struct {
	db DbInfo
}

// CertificateManager
type CertificateManager interface {
	// Get() error
	Delete() error
	Save() error
	Update()
}

// createCertificateSigningRequest
func createCertificateSigningRequest(info CertificateSigningInfo) ([]byte, error) {
	key, err := rsa.GenerateKey(rand.Reader, info.KeySize)
	if err != nil {
		return []byte{}, errors.New("Failed to generate the certitifcate signing key")
	}

	pemBlock, _ := pem.Decode([]byte(pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key)})))
	if pemBlock == nil {
		return []byte{}, errors.New("Failed to decode the private key")
	}

	pk, err := x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
	if err != nil {
		return []byte{}, err
	}

	template := x509.CertificateRequest{
		Version:            info.Version,
		SignatureAlgorithm: signatureAlgorithm(info.Algorithm.SignatureAlgorithm),
		PublicKeyAlgorithm: publicKeyAlgorithm(info.Algorithm.PublicKeyAlgorithm),
		Subject: pkix.Name{
			Country:            info.Subject.Locale.Country,
			Locality:           info.Subject.Locale.Locality,
			PostalCode:         info.Subject.Locale.PostalCode,
			Province:           info.Subject.Locale.Province,
			StreetAddress:      info.Subject.Locale.StreetAddress,
			CommonName:         info.Subject.Names.CommonName,
			Organization:       info.Subject.Organization.Names,
			OrganizationalUnit: info.Subject.Organization.Units},
		DNSNames:       info.DNSNames,
		EmailAddresses: info.EmailAddresses}

	return x509.CreateCertificateRequest(rand.Reader, &template, pk)
}

// CreateCertificateRequest
func CreateCertificateRequest(caCert Cert, name string) (CertificateRequest, error) {
	csr, err := createCertificateSigningRequest(caCert.Spec.CertificateSigningInfo)
	if err != nil {
		return CertificateRequest{}, err
	}

	//Encode csr
	request := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csr,
	})

	return CertificateRequest{
		ApiVersion: "cert-manager.io/v1",
		Kind:       "CertificateRequest",
		MetaData: CertificateRequestMeta{
			Name: name,
		},
		Spec: CertificateRequestSpec{
			Request:              base64.StdEncoding.EncodeToString(request),
			CertificateAuthority: caCert.Spec.CertificateAuthority,
			Duration:             caCert.Spec.Duration,
			IssuerRef: IssuerRef{
				Name:  caCert.Spec.IssuerRef.Name,
				Kind:  caCert.Spec.IssuerRef.Kind,
				Group: caCert.Spec.IssuerRef.Group}}}, nil

}

// validateCertificates
func ValidateCertificateRequest(cr CertificateRequest) error {
	// if len(cr.Status.Certificate) == 0 {
	// 	return errors.New("generated certificate is invalid")
	// }

	// if len(cr.Status.CertificateAuthority) == 0 {
	// 	return errors.New("certificate is generated by an Invalid certifcate authority")
	// }

	// approved := false
	// for _, state := range cr.Status.States {
	// 	if strings.ToLower(state.Status) == "true" &&
	// 		strings.ToLower(state.Type) == "approved" {
	// 		approved = true
	// 		break
	// 	}
	// }

	// if !approved {
	// 	return errors.New("the certificate is not yet approved by the CA")
	// }
	return nil
}

func (c *CertificateClient) SaveClusterProviderCertRequest(cert, cluster, clusterGroup, clusterProvider string, cr CertificateRequest) error {
	if err := ValidateCertificateRequest(cr); err != nil {
		return err
	}

	key := ClusterProviderCertRequestKey{
		Cert:            cert,
		Cluster:         cluster,
		ClusterGroup:    clusterGroup,
		ClusterProvider: clusterProvider,
		CertRequest:     cr.MetaData.Name}

	// TODO : Confirm for exisiting certificates
	if err := db.DBconn.Insert(c.db.storeName, key, nil, c.db.tagMeta, cr); err != nil {
		fmt.Println(err.Error())
		return err
	}

	return nil
}

func (c *CertificateClient) GetClusterProviderCertRequest(cert, cluster, clusterGroup, clusterProvider string) (CertificateRequest, error) {
	key := ClusterProviderCertRequestKey{
		Cert:            cert,
		Cluster:         cluster,
		ClusterGroup:    clusterGroup,
		ClusterProvider: clusterProvider}

	value, err := db.DBconn.Find(c.db.storeName, key, c.db.tagMeta)
	if err != nil {
		return CertificateRequest{}, err
	}

	if len(value) == 0 {
		return CertificateRequest{}, errors.New("CertificateRequest not found")
	}

	if value != nil {
		cr := CertificateRequest{}
		if err = db.DBconn.Unmarshal(value[0], &cr); err != nil {
			return CertificateRequest{}, err
		}
		return cr, nil
	}

	return CertificateRequest{}, errors.New("Unknown Error")
}

func (c *CertificateClient) DeletelusterProviderCertRequest(cert, cluster, clusterGroup, clusterProvider, certRequest string) error {
	key := ClusterProviderCertRequestKey{
		Cert:            cert,
		Cluster:         cluster,
		ClusterGroup:    clusterGroup,
		ClusterProvider: clusterProvider,
		CertRequest:     certRequest}

	// TODO : Confirm for exisiting certificates
	return db.DBconn.Remove(c.db.storeName, key)

}

// NewCertificateClient
func NewCertificateClient() *CertificateClient {
	return &CertificateClient{
		db: DbInfo{
			storeName: "resources",
			tagMeta:   "data"}}
}

func (cr *CertificateRequest) ResourceName() string {
	return fmt.Sprintf("%s+%s", cr.MetaData.Name, "certificaterequest")
}
