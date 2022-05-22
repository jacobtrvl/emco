// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package certificate

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
)

// CreateCertificateSigningRequest
func CreateCertificateSigningRequest(info CertificateRequestInfo, pk interface{}) ([]byte, error) {
	csrInfo := x509.CertificateRequest{
		Version:            info.Version,
		SignatureAlgorithm: signatureAlgorithm(info.SignatureAlgorithm),
		PublicKeyAlgorithm: publicKeyAlgorithm(info.PublicKeyAlgorithm),
		Subject: pkix.Name{
			Country:            info.Subject.Country,
			Locality:           info.Subject.Locality,
			PostalCode:         info.Subject.PostalCode,
			Province:           info.Subject.Province,
			StreetAddress:      info.Subject.StreetAddress,
			CommonName:         info.Subject.CommonName,
			Organization:       info.Subject.Organization,
			OrganizationalUnit: info.Subject.OrganizationalUnit},
		DNSNames:       info.DNSNames,
		EmailAddresses: info.EmailAddresses}

	csr, err := x509.CreateCertificateRequest(rand.Reader, &csrInfo, pk)
	if err != nil {
		return []byte{}, err
	}

	//Encode csr
	request := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csr,
	})

	return request, nil
}

// GeneratePrivateKey
func GeneratePrivateKey(keySize int) (*pem.Block, error) {
	key, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, errors.New("Failed to generate the certitifcate signing key")
	}

	pemBlock, _ := pem.Decode([]byte(pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key)})))
	if pemBlock == nil {
		return nil, errors.New("Failed to decode the private key")
	}

	return pemBlock, nil
}

// ParsePrivateKey
func ParsePrivateKey(der []byte) (*rsa.PrivateKey, error) {
	return x509.ParsePKCS1PrivateKey(der)
}
