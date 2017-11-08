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

func xxxTestX509KeyGen(t *testing.T) {
	mgr := NewX509KeyManager()

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

func TestX509KeyFingerprint(t *testing.T) {
	mgr := NewX509KeyManager()

	pub := "-----BEGIN CERTIFICATE-----\n" +
		"MIIDlzCCAn+gAwIBAgIMWdSlWAJNWKeXgryHMA0GCSqGSIb3DQEBCwUAMBoxGDAW\n" +
		"BgNVBAMTD0RhbmllbCBCZXJyYW5nZTAeFw0xNzEwMDQwOTA5NDRaFw0xODEwMDQw\n" +
		"OTA5NDRaMDkxEjAQBgNVBAMTCWxvY2FsaG9zdDEjMCEGA1UEChMaTmFtZSAgb2Yg\n" +
		"eW91ciBvcmdhbml6YXRpb24wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIB\n" +
		"AQDV1PNjgt1GRx/TDu80Jri82E4J/Ryd8LJjTXDWu73/kT7S6ixDEfn967PHvyUU\n" +
		"ejIuyMtz4vJjTOjn7rhNNlsiu7F50P2pMPyntHND4B47f9lwBnOH3MFXoqT1h3Qk\n" +
		"PthI2I5SmaGxaP63iXCnaoH9Ea1jWS+rOHLxRKdOOgSqWM5CZscEyDeUk/z/0UeA\n" +
		"96c30bVoDcGdsmXNnMYDTb63Dvqe1jH3g8C1+ndSTvqGYmQU3yQRPFMlF2+2Zpwr\n" +
		"qt4+5X1hKwZ4DF+sPLa/FZDRhJ8NgdjVKpzs3s98m/w5rPTYBkYigl5bU6dUcxDk\n" +
		"RTx1RpRmPgr1mtpXtKjU6uMLAgMBAAGjgb0wgbowDAYDVR0TAQH/BAIwADBEBgNV\n" +
		"HREEPTA7ggpsb2NhbGhvc3Q2ghVsb2NhbGhvc3QubG9jYWxkb21haW6HBH8AAAGH\n" +
		"EAAAAAAAAAAAAAAAAAAAAAEwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDwYDVR0PAQH/\n" +
		"BAUDAwegADAdBgNVHQ4EFgQUor1GLDA08gjMpe2wG3DQJ8a9X/0wHwYDVR0jBBgw\n" +
		"FoAU8FHn9nygbm2i2RTS4JcDBNjsoYwwDQYJKoZIhvcNAQELBQADggEBAH0FVqhj\n" +
		"qjv756tlpVNNoJ1lpYDWiPLPCxSuYpHqh8GTh5iWlSDe5Ely46xX73M4NosV0vCV\n" +
		"w+EVYM7qmq7gmXK9BMiSuIM8ewc9sW8AhOzxy/pCAsU15GFPYlIQedFf1bnF4CBY\n" +
		"om7/axOF2eaUmzNmcHzkk8AqIl6z9X14RWEmX1d5wflawNfTC3yuyi0xRphZWdub\n" +
		"Ba98ZR2eqC72wwwJesoWjg9XmE8gXa6CvlRLuxbrDXYVDXaGDrnAk21tiyImIP/7\n" +
		"s1bL4DksmLiTydixSXU0CjMYm0N/6/ZTcHaAaWdeCUGnJDaPU7NqB07MDqUoORht\n" +
		"JhETrcxrlILRO6s=\n" +
		"-----END CERTIFICATE-----\n"

	fpr, err := mgr.FingerPrint(pub)
	if err != nil {
		t.Fatalf("Unable to create key fingerprint", err)
		return
	}

	expect := "308203973082027fa003020102020c59d4a558024d58a79782bc87300d06092a864886" +
		"f70d01010b0500301a311830160603550403130f44616e69656c2042657272616e6765" +
		"301e170d3137313030343039303934345a170d3138313030343039303934345a303931" +
		"123010060355040313096c6f63616c686f737431233021060355040a131a4e616d6520" +
		"206f6620796f7572206f7267616e697a6174696f6e30820122300d06092a864886f70d" +
		"01010105000382010f003082010a0282010100d5d4f36382dd46471fd30eef3426b8bc" +
		"d84e09fd1c9df0b2634d70d6bbbdff913ed2ea2c4311f9fdebb3c7bf25147a322ec8cb" +
		"73e2f2634ce8e7eeb84d365b22bbb179d0fda930fca7b47343e01e3b7fd970067387dc" +
		"c157a2a4f58774243ed848d88e5299a1b168feb78970a76a81fd11ad63592fab3872f1" +
		"44a74e3a04aa58ce4266c704c8379493fcffd14780f7a737d1b5680dc19db265cd9cc6" +
		"034dbeb70efa9ed631f783c0b5fa77524efa86626414df24113c5325176fb6669c2baa" +
		"de3ee57d612b06780c5fac3cb6bf1590d1849f0d81d8d52a9cecdecf7c9bfc39acf4d8" +
		"064622825e5b53a7547310e4453c754694663e0af59ada57b4a8d4eae30b0203010001" +
		"a381bd3081ba300c0603551d130101ff0402300030440603551d11043d303b820a6c6f" +
		"63616c686f73743682156c6f63616c686f73742e6c6f63616c646f6d61696e87047f00" +
		"000187100000000000000000000000000000000130130603551d25040c300a06082b06" +
		"010505070301300f0603551d0f0101ff0405030307a000301d0603551d0e04160414a2" +
		"bd462c3034f208cca5edb01b70d027c6bd5ffd301f0603551d23041830168014f051e7" +
		"f67ca06e6da2d914d2e0970304d8eca18c300d06092a864886f70d01010b0500038201" +
		"01007d0556a863aa3bfbe7ab65a5534da09d65a580d688f2cf0b14ae6291ea87c19387" +
		"98969520dee44972e3ac57ef7338368b15d2f095c3e11560ceea9aaee09972bd04c892" +
		"b8833c7b073db16f0084ecf1cbfa4202c535e4614f62521079d15fd5b9c5e02058a26e" +
		"ff6b1385d9e6949b3366707ce493c02a225eb3f57d784561265f5779c1f95ac0d7d30b" +
		"7caeca2d3146985959db9b05af7c651d9ea82ef6c30c097aca168e0f57984f205dae82" +
		"be544bbb16eb0d76150d76860eb9c0936d6d8b222620fffbb356cbe0392c98b893c9d8" +
		"b14975340a33189b437febf65370768069675e0941a724368f53b36a074ecc0ea52839" +
		"186d261113adcc6b9482d13babda39a3ee5e6b4b0d3255bfef95601890afd80709"

	if fpr != expect {
		t.Fatalf("Incorrect fingerprint '%s' expected '%s'", fpr, expect)
		return
	}
}
