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
	"strconv"
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

func HashPassword(password string) (string, error) {

	salt := make([]byte, SCRYPT_SALT_SIZE)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}

	pwHash, err := scrypt.Key(
		[]byte(password), salt,
		SCRYPT_COST_FACTOR,
		SCRYPT_BLOCK_SIZE,
		SCRYPT_PARALLELIZATION,
		SCRYPT_OUTPUT_SIZE)
	if err != nil {
		return "", err
	}

	pwHash64 := base64.StdEncoding.EncodeToString(pwHash)
	salt64 := base64.StdEncoding.EncodeToString(salt)

	return fmt.Sprintf("scrypt,%x,%x,%x,%s,%s",
		SCRYPT_COST_FACTOR,
		SCRYPT_BLOCK_SIZE,
		SCRYPT_PARALLELIZATION,
		salt64, pwHash64), nil
}

func CheckPassword(password string, hash string) (bool, error) {
	bits := strings.Split(hash, ",")
	if len(bits) != 6 {
		return false, fmt.Errorf("Expected 6 bits in hash")
	}

	if bits[0] != "scrypt" {
		return false, fmt.Errorf("Expected 'scrypt' scheme not '%s'", bits[0])
	}

	costFactor, err := strconv.ParseInt(bits[1], 16, 0)
	if err != nil {
		return false, err
	}
	blockSize, err := strconv.ParseInt(bits[2], 16, 0)
	if err != nil {
		return false, err
	}
	parallelization, err := strconv.ParseInt(bits[3], 16, 0)
	if err != nil {
		return false, err
	}

	salt, err := base64.StdEncoding.DecodeString(bits[4])
	if err != nil {
		return false, err
	}

	raw, _ := base64.StdEncoding.DecodeString(bits[5])
	pwHash, err := scrypt.Key(
		[]byte(password), salt, int(costFactor), int(blockSize), int(parallelization), len(raw))

	return bytes.Compare(raw, pwHash) == 0, nil
}
