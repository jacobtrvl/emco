// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"os"
	"reflect"
	"strings"

	pkgerrors "github.com/pkg/errors"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

type Encryptor interface {
	Encrypt(data []byte) ([]byte, error)
	Decrypt(data []byte) ([]byte, error)
}

type AESGCMEncryptor struct {
	block cipher.Block
}

func NewAESGCMEncryptor(key []byte) (*AESGCMEncryptor, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return &AESGCMEncryptor{
		block,
	}, nil
}

func (e *AESGCMEncryptor) Encrypt(data []byte) ([]byte, error) {
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
	return aead.Seal(dst[nonceSize:nonceSize], dst[:nonceSize], data, nil), nil
}

func (e *AESGCMEncryptor) Decrypt(data []byte) ([]byte, error) {
	aead, err := cipher.NewGCM(e.block)
	if err != nil {
		return nil, err
	}
	// The nonce is contained at the beginning of the data
	nonceSize := aead.NonceSize()
	if len(data) < nonceSize {
		return nil, pkgerrors.Errorf("Data must include nonce")
	}
	return aead.Open(nil, data[:nonceSize], data[nonceSize:], nil)
}

type IObjectEncryptor interface {
	EncryptObject(o interface{}) (interface{}, error)
	EncryptString(message string) (string, error)
	DecryptObject(o interface{}) (interface{}, error)
	DecryptString(ciphermessage string) (string, error)
}

type MyObjectEncryptor struct {
	provider Encryptor
}

var gobjencs = make(map[string]IObjectEncryptor)

func GetObjectEncryptor(provider string) IObjectEncryptor {
	if gobjencs[provider] == nil {
		envkey := strings.ToUpper(provider) + "_DATA_KEY"
		if len(os.Getenv(envkey)) > 0 {
			// Format key
			key := []byte(os.Getenv(envkey))
			nkey := make([]byte, 32)
			for i := 0; i < 32; i++ {
				if i < len(key) {
					nkey[i] = key[i]
				} else {
					nkey[i] = 10
				}
			}
			aesgcm, err := NewAESGCMEncryptor(nkey)
			if err != nil {
				log.Error("Create encryptor error :: ", log.Fields{"Error": err})
				return nil
			}
			gobjencs[provider] = &MyObjectEncryptor{
				aesgcm,
			}
		} else {
			return nil
		}
	}

	return gobjencs[provider]
}

func (c *MyObjectEncryptor) EncryptObject(o interface{}) (interface{}, error) {
	return c.processObject(o, false, c.EncryptString)
}

func (c *MyObjectEncryptor) DecryptObject(o interface{}) (interface{}, error) {
	return c.processObject(o, false, c.DecryptString)
}

func (c *MyObjectEncryptor) EncryptString(message string) (string, error) {
	ciphermessage, err := c.provider.Encrypt([]byte(message))
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
	message, err := c.provider.Decrypt(cm)
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
