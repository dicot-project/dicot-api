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
	"testing"
)

func TestPasswordHash(t *testing.T) {
	passwd := "correct horse battery staple"

	hash, err := HashPassword(passwd)
	if err != nil {
		t.Errorf("Cannot hash password %s", err)
		return
	}

	// Valid password, should match
	match, err := CheckPassword(passwd, hash)
	if err != nil {
		t.Errorf("Cannot check password %s", err)
		return
	}

	if !match {
		t.Errorf("Password hash did not match")
		return
	}

	// Invalid password should not match
	match, err = CheckPassword(passwd+"!", hash)
	if err != nil {
		t.Errorf("Cannot check password %s", err)
		return
	}

	if match {
		t.Errorf("Password hash should not match")
		return
	}

	// Invalid hash scheme should raise error
	match, err = CheckPassword(passwd, "a"+hash[1:])
	if err == nil {
		t.Errorf("Should see error from bad method")
		return
	}

	if match {
		t.Errorf("Password hash should not match")
		return
	}

	// Corrupted base64 should raise error
	match, err = CheckPassword(passwd, hash[0:len(hash)-1])
	if err == nil {
		t.Errorf("SHould see error from corrupt base64")
		return
	}

	if match {
		t.Errorf("Password hash should not match")
		return
	}

	// Truncated base64 hash should not match
	match, err = CheckPassword(passwd, hash[0:len(hash)-4])
	if err != nil {
		t.Errorf("Cannot check password %s", err)
		return
	}

	if match {
		t.Errorf("Password hash should not match")
		return
	}
}
