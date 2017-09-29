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
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

const (
	ECPrivateKey  = "EC PRIVATE KEY"
	RSAPrivateKey = "RSA PRIVATE KEY"
)

func LoadPEMKeys(keyPEM []byte) ([]interface{}, error) {
	var keys []interface{}
	var block *pem.Block
	for {
		block, keyPEM = pem.Decode(keyPEM)
		if block == nil {
			break
		}

		switch block.Type {
		case ECPrivateKey:
			key, err := x509.ParseECPrivateKey(block.Bytes)
			if err != nil {
				return []interface{}{}, err
			}
			keys = append(keys, key)

		case RSAPrivateKey:
			key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
			if err != nil {
				return []interface{}{}, err
			}
			keys = append(keys, key)

		default:
			return []interface{}{}, fmt.Errorf("Unknown PEM block '%s'", block.Type)
		}
	}

	return keys, nil
}
