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
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"golang.org/x/crypto/ssh"
	"strings"
)

type sshKeyManager struct {
}

func NewSSHKeyManager() KeyManager {
	return &sshKeyManager{}
}

func (k *sshKeyManager) CreateKeyPair(algorithm string, length int) (string, string, error) {
	var privKeyPEM []byte
	var pubKeyHex []byte
	if algorithm == AlgRSA {
		key, err := rsa.GenerateKey(rand.Reader, length)
		if err != nil {
			return "", "", err
		}
		privKeyBlock := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}
		privKeyPEM = pem.EncodeToMemory(privKeyBlock)

		pubKey, err := ssh.NewPublicKey(&key.PublicKey)
		if err != nil {
			return "", "", err
		}
		pubKeyHex = ssh.MarshalAuthorizedKey(pubKey)
	} else {
		return "", "", fmt.Errorf("Unsuported algorithm %s", algorithm)
	}

	return string(privKeyPEM), string(pubKeyHex), nil
}

func (k *sshKeyManager) FingerPrint(pubkey string) (string, error) {
	key, _, _, _, err := ssh.ParseAuthorizedKey([]byte(pubkey))
	if err != nil {
		return "", err
	}

	md5sum := md5.Sum(key.Marshal())
	hexarray := make([]string, len(md5sum))
	for i, c := range md5sum {
		hexarray[i] = hex.EncodeToString([]byte{c})
	}
	return strings.Join(hexarray, ":"), nil
}
