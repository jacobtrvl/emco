// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	pkgerrors "github.com/pkg/errors"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

type EncryptorConfiguration struct {
	Providers []ProviderConfiguration `json:"providers"`
}

type ProviderConfiguration struct {
	AESGCM *AESGCMConfiguration `json:"aesgcm,omitempty"`
}

type AESGCMConfiguration struct {
	Keys []Key `json:"keys"`
}

type Key struct {
	Name   string `json:"name"`
	Secret string `json:"secret"`
}

func readConfigFile(name string) (*EncryptorConfiguration, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var config EncryptorConfiguration
	decoder := json.NewDecoder(f)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func BuildEncryptor(config *EncryptorConfiguration) (Encryptor, error) {
	var es []Encryptor
	for _, p := range config.Providers {
		switch {
		case p.AESGCM != nil:
			for _, k := range p.AESGCM.Keys {
				key, err := base64.StdEncoding.DecodeString(k.Secret)
				if err != nil {
					return nil, err
				}
				e, err := NewAESGCMEncryptor(k.Name, key)
				if err != nil {
					return nil, err
				}
				es = append(es, e)
			}
		}
	}
	return NewEncryptorList(es)
}

type Encryptor interface {
	Encrypt(data, additionalData []byte) ([]byte, error)
	Decrypt(data, additionalData []byte) ([]byte, error)
}

var encryptorPrefix = "emco:enc:"

func addPrefix(data, prefix []byte) []byte {
	dst := make([]byte, len(prefix)+len(data))
	copy(dst, prefix)
	copy(dst[len(prefix):], data)
	return dst
}

func trimPrefix(data, prefix []byte) ([]byte, error) {
	if bytes.HasPrefix(data, prefix) {
		return bytes.TrimPrefix(data, prefix), nil
	}
	return nil, pkgerrors.Errorf("Invalid prefix")
}

var aesgcmPrefix = encryptorPrefix + "aesgcm:v1:"

type AESGCMEncryptor struct {
	block  cipher.Block
	prefix []byte
}

func NewAESGCMEncryptor(name string, key []byte) (*AESGCMEncryptor, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return &AESGCMEncryptor{
		block:  block,
		prefix: []byte(aesgcmPrefix + name + ":"),
	}, nil
}

func (e *AESGCMEncryptor) Encrypt(data, additionalData []byte) ([]byte, error) {
	aead, err := cipher.NewGCM(e.block)
	if err != nil {
		return nil, err
	}
	// Prepend the nonce to the returned data as it is required during decryption
	nonceSize := aead.NonceSize()
	dst := make([]byte, nonceSize+len(data)+aead.Overhead())
	n, err := rand.Read(dst[:nonceSize])
	if err != nil {
		return nil, err
	}
	if n != nonceSize {
		return nil, pkgerrors.Errorf("Need %d random bytes for nonce but only got %d bytes", nonceSize, n)
	}
	// dst[nonceSize:nonceSize] is to reuse dst's storage for the encrypted output
	ciphertext := aead.Seal(dst[nonceSize:nonceSize], dst[:nonceSize], data, additionalData)
	return addPrefix(dst[:nonceSize+len(ciphertext)], e.prefix), nil
}

func (e *AESGCMEncryptor) Decrypt(data, additionalData []byte) ([]byte, error) {
	src, err := trimPrefix(data, e.prefix)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(e.block)
	if err != nil {
		return nil, err
	}
	// The nonce is contained at the beginning of the data
	nonceSize := aead.NonceSize()
	if len(src) < nonceSize {
		return nil, pkgerrors.Errorf("Data must include nonce")
	}
	return aead.Open(nil, src[:nonceSize], src[nonceSize:], additionalData)
}

type EncryptorList struct {
	encryptors []Encryptor
}

func NewEncryptorList(encryptors []Encryptor) (*EncryptorList, error) {
	if len(encryptors) == 0 {
		return nil, pkgerrors.Errorf("Must include at least one encryptor")
	}
	return &EncryptorList{
		encryptors,
	}, nil
}

func (e *EncryptorList) Encrypt(data, additionalData []byte) ([]byte, error) {
	return e.encryptors[0].Encrypt(data, additionalData)
}

func (e *EncryptorList) Decrypt(data, additionalData []byte) ([]byte, error) {
	var (
		bs  []byte
		err error
	)
	for _, e := range e.encryptors {
		bs, err = e.Decrypt(data, additionalData)
		if err == nil {
			return bs, err
		}
	}
	return bs, err
}

type IObjectEncryptor interface {
	EncryptObject(o interface{}) (interface{}, error)
	DecryptObject(o interface{}) (interface{}, error)
}

type IdentityObjectEncryptor struct{}

func (c *IdentityObjectEncryptor) EncryptObject(o interface{}) (interface{}, error) {
	return o, nil
}

func (c *IdentityObjectEncryptor) DecryptObject(o interface{}) (interface{}, error) {
	return o, nil
}

type MyObjectEncryptor struct {
	encryptor Encryptor
}

func (c *MyObjectEncryptor) EncryptObject(o interface{}) (interface{}, error) {
	return c.processObject(o, false, c.EncryptString)
}

func (c *MyObjectEncryptor) DecryptObject(o interface{}) (interface{}, error) {
	return c.processObject(o, false, c.DecryptString)
}

func (c *MyObjectEncryptor) EncryptString(message string) (string, error) {
	ciphermessage, err := c.encryptor.Encrypt([]byte(message), nil)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(ciphermessage), nil
}

func (c *MyObjectEncryptor) DecryptString(ciphermessage string) (string, error) {
	cm, err := hex.DecodeString(ciphermessage)
	if err != nil {
		return "", err
	}
	message, err := c.encryptor.Decrypt(cm, nil)
	if err != nil {
		return "", err
	}
	return string(message), nil
}

func (c *MyObjectEncryptor) processObject(o interface{}, encrypt bool, oper func(string) (string, error)) (interface{}, error) {
	t := reflect.TypeOf(o)
	switch t.Kind() {
	case reflect.String:
		// only support do encryption on string field
		if encrypt {
			val, err := oper(o.(string))
			if err != nil {
				return nil, err
			}

			return val, nil
		}
	case reflect.Ptr:
		v := reflect.ValueOf(o)
		newv, err := c.processObject(v.Elem().Interface(), encrypt, oper)
		if err != nil {
			return nil, err
		}
		v.Elem().Set(reflect.ValueOf(newv))
		return o, nil
	case reflect.Struct:
		v := reflect.ValueOf(&o).Elem()
		newv := reflect.New(v.Elem().Type()).Elem()
		newv.Set(v.Elem())
		for k := 0; k < t.NumField(); k++ {
			_, fieldEncrypt := t.Field(k).Tag.Lookup("encrypted")
			isEncrypt := fieldEncrypt || encrypt
			if t.Field(k).IsExported() {
				newf, err := c.processObject(newv.Field(k).Interface(), isEncrypt, oper)
				if err != nil {
					return nil, err
				}
				newv.Field(k).Set(reflect.ValueOf(newf))
			}
		}
		return newv.Interface(), nil
	case reflect.Array:
		v := reflect.ValueOf(o)
		newv := reflect.New(t).Elem()
		for k := 0; k < v.Len(); k++ {
			newf, err := c.processObject(v.Index(k).Interface(), encrypt, oper)
			if err != nil {
				return nil, err
			}
			newv.Index(k).Set(reflect.ValueOf(newf))
		}
		return newv.Interface(), nil
	case reflect.Slice:
		v := reflect.ValueOf(o)
		newv := reflect.MakeSlice(t, v.Len(), v.Len())
		for k := 0; k < v.Len(); k++ {
			newf, err := c.processObject(v.Index(k).Interface(), encrypt, oper)
			if err != nil {
				return nil, err
			}
			newv.Index(k).Set(reflect.ValueOf(newf))
		}
		return newv.Interface(), nil
	case reflect.Map:
		v := reflect.ValueOf(o)
		newv := reflect.MakeMap(t)
		for _, k := range v.MapKeys() {
			newf, err := c.processObject(v.MapIndex(k).Interface(), encrypt, oper)
			if err != nil {
				return nil, err
			}
			newv.SetMapIndex(k, reflect.ValueOf(newf))
		}
		return newv.Interface(), nil
	default:
	}

	return o, nil
}

var (
	identityObjectEncryptor IdentityObjectEncryptor
	gobjencs                = make(map[string]IObjectEncryptor)
)

func GetObjectEncryptor(provider string) IObjectEncryptor {
	if gobjencs[provider] == nil {
		// Default is no encryption
		var objenc IObjectEncryptor = &identityObjectEncryptor

		name := fmt.Sprintf("encryptor/%s.json", provider)
		config, err := readConfigFile(name)
		if err != nil {
			log.Warn("Read encryptor configuration error :: ", log.Fields{"Error": err})
		} else {
			encryptor, err := BuildEncryptor(config)
			if err != nil {
				log.Warn("Build encryptor error :: ", log.Fields{"Error": err})
			} else {
				objenc = &MyObjectEncryptor{
					encryptor,
				}
			}
		}

		gobjencs[provider] = objenc
	}

	return gobjencs[provider]
}
