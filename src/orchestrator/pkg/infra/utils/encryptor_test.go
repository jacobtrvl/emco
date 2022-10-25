// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package utils

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

func TestConfiguration(t *testing.T) {
	t.Run("Build encryptor from configuration file", func(t *testing.T) {
		f, err := os.Open("../../../tests/configs/mock_encryptor_config.json")
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()

		var config EncryptorConfiguration
		decoder := json.NewDecoder(f)
		err = decoder.Decode(&config)
		if err != nil {
			t.Fatal(err)
		}

		_, err = BuildEncryptor(&config)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("First configured provider will encrypt", func(t *testing.T) {
		bs := []byte(`{"providers":[{"aesgcm":{"keys":[{"name":"key1","secret":"c2VjcmV0IGlzIHNlY3VyZQ=="},{"name":"key2","secret":"dGhpcyBpcyBwYXNzd29yZA=="}]}}]}`)
		var config1 EncryptorConfiguration
		err := json.Unmarshal(bs, &config1)
		if err != nil {
			t.Fatal(err)
		}
		e1, err := BuildEncryptor(&config1)
		if err != nil {
			t.Fatal(err)
		}
		message := []byte("Hello")
		encrypted, err := e1.Encrypt(message, nil)
		if err != nil {
			t.Fatal(err)
		}
		decrypted, err := e1.Decrypt(encrypted, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(decrypted, message) {
			t.Errorf("Decrypted ciphertext != plaintext")
		}

		// Verify that first key was used by checking that
		// second key will not decrypt
		bs = []byte(`{"providers":[{"aesgcm":{"keys":[{"name":"key2","secret":"dGhpcyBpcyBwYXNzd29yZA=="}]}}]}`)
		var config2 EncryptorConfiguration
		err = json.Unmarshal(bs, &config2)
		if err != nil {
			t.Fatal(err)
		}
		e2, err := BuildEncryptor(&config2)
		if err != nil {
			t.Fatal(err)
		}
		decrypted, err = e2.Decrypt(encrypted, nil)
		if err == nil {
			t.Fatal(err)
		}
	})

	t.Run("Any configured provider listed may decrypt", func(t *testing.T) {
		// Encrypt message with key2
		bs := []byte(`{"providers":[{"aesgcm":{"keys":[{"name":"key2","secret":"dGhpcyBpcyBwYXNzd29yZA=="}]}}]}`)
		var config1 EncryptorConfiguration
		err := json.Unmarshal(bs, &config1)
		if err != nil {
			t.Fatal(err)
		}
		e, err := BuildEncryptor(&config1)
		if err != nil {
			t.Fatal(err)
		}
		message := []byte("Hello")
		encrypted, err := e.Encrypt(message, nil)
		if err != nil {
			t.Fatal(err)
		}

		// Now create a new encryptor with key2 as the last
		// provider
		bs = []byte(`{"providers":[{"aesgcm":{"keys":[{"name":"key1","secret":"c2VjcmV0IGlzIHNlY3VyZQ=="},{"name":"key2","secret":"dGhpcyBpcyBwYXNzd29yZA=="}]}}]}`)
		var config2 EncryptorConfiguration
		err = json.Unmarshal(bs, &config2)
		if err != nil {
			t.Fatal(err)
		}
		e, err = BuildEncryptor(&config2)
		if err != nil {
			t.Fatal(err)
		}

		// Now decrypt
		decrypted, err := e.Decrypt(encrypted, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(decrypted, message) {
			t.Errorf("Decrypted ciphertext != plaintext")
		}
	})

	t.Run("Rotate AESGCM key", func(t *testing.T) {
		// Encrypt message with key1
		bs := []byte(`{"providers":[{"aesgcm":{"keys":[{"name":"key1","secret":"c2VjcmV0IGlzIHNlY3VyZQ=="}]}}]}`)
		var config1 EncryptorConfiguration
		err := json.Unmarshal(bs, &config1)
		if err != nil {
			t.Fatal(err)
		}
		e1, err := BuildEncryptor(&config1)
		if err != nil {
			t.Fatal(err)
		}
		message := []byte("Hello")
		encrypted, err := e1.Encrypt(message, nil)
		if err != nil {
			t.Fatal(err)
		}

		// Now create a new encryptor with key2 as the new
		// encryption key and keep key1 for decryption
		bs = []byte(`{"providers":[{"aesgcm":{"keys":[{"name":"key2","secret":"dGhpcyBpcyBwYXNzd29yZA=="},{"name":"key1","secret":"c2VjcmV0IGlzIHNlY3VyZQ=="}]}}]}`)
		var config2 EncryptorConfiguration
		err = json.Unmarshal(bs, &config2)
		if err != nil {
			t.Fatal(err)
		}
		e2, err := BuildEncryptor(&config2)
		if err != nil {
			t.Fatal(err)
		}

		// Now decrypt and re-encrypt
		decrypted, err := e2.Decrypt(encrypted, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(decrypted, message) {
			t.Errorf("Decrypted ciphertext != plaintext")
		}
		encrypted, err = e2.Encrypt(decrypted, nil)
		if err != nil {
			t.Fatal(err)
		}

		// Now create a new encryptor with only key2 and
		// verify decryption
		bs = []byte(`{"providers":[{"aesgcm":{"keys":[{"name":"key2","secret":"dGhpcyBpcyBwYXNzd29yZA=="}]}}]}`)
		var config3 EncryptorConfiguration
		err = json.Unmarshal(bs, &config3)
		if err != nil {
			t.Fatal(err)
		}
		e3, err := BuildEncryptor(&config2)
		if err != nil {
			t.Fatal(err)
		}
		decrypted, err = e3.Decrypt(encrypted, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(decrypted, message) {
			t.Errorf("Decrypted ciphertext != plaintext")
		}
	})
}
