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
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/scrypt"
)

const (
	SCRYPT_COST_FACTOR     = 16384
	SCRYPT_BLOCK_SIZE      = 8
	SCRYPT_PARALLELIZATION = 1
	SCRYPT_OUTPUT_SIZE     = 32
	SCRYPT_SALT_SIZE       = 32
)

func makeHash(password []byte, salt []byte) ([]byte, error) {
	return scrypt.Key(
		password, salt,
		SCRYPT_COST_FACTOR,
		SCRYPT_BLOCK_SIZE,
		SCRYPT_PARALLELIZATION,
		SCRYPT_OUTPUT_SIZE)
}

func HashPassword(password string) (string, error) {

	salt := make([]byte, SCRYPT_SALT_SIZE)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}

	pwHash, err := makeHash([]byte(password), salt)
	if err != nil {
		return "", err
	}

	pwHash64 := base64.StdEncoding.EncodeToString(pwHash)
	salt64 := base64.StdEncoding.EncodeToString(salt)

	return fmt.Sprintf("scrypt,%s,%s", salt64, pwHash64), nil
}

func CheckPassword(password string, hash string) (bool, error) {
	bits := strings.Split(hash, ",")
	if len(bits) != 3 {
		return false, fmt.Errorf("Expected 3 bits in hash")
	}

	if bits[0] != "scrypt" {
		return false, fmt.Errorf("Expected 'scrypt' scheme not '%s'", bits[0])
	}

	salt, err := base64.StdEncoding.DecodeString(bits[1])
	if err != nil {
		return false, err
	}

	gotHash, err := makeHash([]byte(password), salt)
	if err != nil {
		return false, err
	}

	wantHash, err := base64.StdEncoding.DecodeString(bits[2])
	if err != nil {
		return false, err
	}

	return bytes.Compare(gotHash, wantHash) == 0, nil
}
