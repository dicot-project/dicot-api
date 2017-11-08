/*
 * This file is part of the Dicot project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2017 Red Hat, Inc.
 *
 */

package crypto

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"testing"
)

var (
	rsaKeyPEM = "-----BEGIN RSA PRIVATE KEY-----\n" +
		"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\n" +
		"-----END RSA PRIVATE KEY-----\n"

	ecKeyPEM = "-----BEGIN EC PRIVATE KEY-----\n" +
		"MHcCAQEEID9nmvk5D5dG8MojMax7Qx3XiYJqCpHqQDNZ2ldWciYwoAoGCCqGSM49\n" +
		"AwEHoUQDQgAEO+g8Z1XdyRrzrMVi5s0RcheU0iVC7S7eM/8n3Aa6xSdGEnbnEKSc\n" +
		"HZVjfWAyTjAMtPj979+IpdAi5v26Lk0O2A==\n" +
		"-----END EC PRIVATE KEY-----\n"

	bogusKeyPEM = "-----BEGIN BOGUS PRIVATE KEY-----\n" +
		"MHcCAQEEID9nmvk5D5dG8MojMax7Qx3XiYJqCpHqQDNZ2ldWciYwoAoGCCqGSM49\n" +
		"-----END BOGUS PRIVATE KEY-----\n"
)

func TestLoadPEM(t *testing.T) {

	keyPEMs := []byte(rsaKeyPEM + ecKeyPEM)

	keys, err := LoadPEMKeys(keyPEMs)
	if err != nil {
		t.Errorf("Unable to load PEM keys %s", err)
		return
	}

	if len(keys) != 2 {
		t.Errorf("Expected to load 2 keys not %d", len(keys))
		return
	}

	_, ok := keys[0].(*rsa.PrivateKey)
	if !ok {
		t.Errorf("Expected an RSA key")
		return
	}

	_, ok = keys[1].(*ecdsa.PrivateKey)
	if !ok {
		t.Errorf("Expected an ECDSA key")
		return
	}
}

func TestLoadPEMFail(t *testing.T) {

	keyPEMs := []byte(rsaKeyPEM + ecKeyPEM + bogusKeyPEM)

	keys, err := LoadPEMKeys(keyPEMs)
	if err == nil {
		t.Errorf("Expected to see PEM load failure")
		return
	}

	if len(keys) != 0 {
		t.Errorf("Expected 0 keys, not %d", len(keys))
		return
	}

}
