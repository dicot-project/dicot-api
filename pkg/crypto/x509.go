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
	"crypto/sha1"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
)

type x509KeyManager struct {
}

func NewX509KeyManager() KeyManager {
	return &x509KeyManager{}
}

func (k *x509KeyManager) CreateKeyPair(algorithm string, length int) (string, string, error) {
	// XXX fixme
	return "", "", fmt.Errorf("Unable to generate certificates")
}

func (k *x509KeyManager) FingerPrint(pubkey string) (string, error) {
	block, _ := pem.Decode([]byte(pubkey))
	if block == nil {
		return "", fmt.Errorf("Unable to decode PEM file")
	}

	if block.Type != "CERTIFICATE" {
		return "", fmt.Errorf("Unexpected PEM file type '%s'", block.Type)
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", err
	}

	hash := sha1.New()
	return hex.EncodeToString(hash.Sum(cert.Raw)), nil
}
