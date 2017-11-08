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
	"strings"
	"testing"
)

func TestSSHKeyGen(t *testing.T) {
	mgr := NewSSHKeyManager()

	priv, pub, err := mgr.CreateKeyPair(AlgRSA, 2048)
	if err != nil {
		t.Fatalf("Unable to create RSA-2048 keypair", err)
		return
	}

	if !strings.HasPrefix(priv, "-----BEGIN RSA PRIVATE KEY-----") {
		t.Fatalf("Missing RSA private key header in %s", priv)
		return
	}
	if !strings.HasPrefix(pub, "ssh-rsa") {
		t.Fatalf("Missing RSA public key marker in %s", pub)
		return
	}
}

func TestSSHKeyFingerprint(t *testing.T) {
	mgr := NewSSHKeyManager()

	pub := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDI9jR1Q4qS" +
		"RtrQhKPQo8KIjh8B6krZboDTop3rnBQduFIIE6az+0svLC9" +
		"JFDg8aLXEhO2wx8h+EaOFk95uyTDUeDkTR46Kwz05tgwsSo" +
		"dwsMVje8wynNQylKYcJb5fwizImnrkAe89SF1pMov/iU+xm" +
		"bwVCpOn0FCvlNqZhcyPjbe82iNjmiCqugJdVHkHO5hhXgSL" +
		"A4IMcObobbSTR5Om4MZ0qICTXdVH0jUV3+olvBzXvlzewTx" +
		"PVJF+vRSBh2bRKFEl/csDFJlam3pcHnZt5nhhaiYinvefkR" +
		"rLHk/GGxfHYvRCpG4PG5yjfG+d1NXORgzk/Y3Fi5Ha+obx8x/t"

	fpr, err := mgr.FingerPrint(pub)
	if err != nil {
		t.Fatalf("Unable to create key fingerprint", err)
		return
	}

	expect := "22:8e:cf:bf:1f:bc:b1:c3:d2:9f:70:22:ce:17:39:92"

	if fpr != expect {
		t.Fatalf("Incorrect fingerprint '%s' expected '%s'", fpr, expect)
		return
	}
}
