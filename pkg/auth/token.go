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

package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"fmt"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/rest"

	"github.com/dicot-project/dicot-api/pkg/api/identity"
	"github.com/dicot-project/dicot-api/pkg/api/identity/v1"
	"github.com/dicot-project/dicot-api/pkg/crypto"
)

const (
	ClaimSubject = "sub"
	ClaimIssuer  = "iss"
	ClaimIssued  = "iat"
	ClaimExpiry  = "exp"
	ClaimID      = "jti"

	ClaimScopeDomain  = "github.com/dicot-project/scope/domain"
	ClaimScopeProject = "github.com/dicot-project/scope/project"

	Issuer = "github.com/dicot-project/api"
)

type TokenManager interface {
	NewToken() *Token
	SignToken(tok *Token) (string, error)
	ValidateToken(toksig string) (*Token, error)
}

type Token struct {
	ID      string
	Issued  time.Time
	Expiry  time.Time
	Subject TokenSubject
	Scope   TokenScope
}

type TokenSubject struct {
	DomainName string
	UserName   string
}

type TokenScope struct {
	DomainName  string
	ProjectName string
}

type tokenManager struct {
	keys        []interface{}
	lifetime    time.Duration
	tokenClient *identity.RevokedTokenClient
}

func NewTokenManagerFromPEM(keyPEM string, lifetime time.Duration, cl rest.Interface) (TokenManager, error) {
	keys, err := crypto.LoadPEMKeys([]byte(keyPEM))
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return nil, fmt.Errorf("No keys found in PEM data")
	}

	return NewTokenManager(keys, lifetime, cl), nil
}

func NewTokenManager(keys []interface{}, lifetime time.Duration, cl rest.Interface) TokenManager {
	return &tokenManager{
		keys:        keys,
		lifetime:    lifetime,
		tokenClient: identity.NewRevokedTokenClient(cl, v1.NamespaceSystem),
	}
}

func (tm *tokenManager) NewToken() *Token {
	now := time.Now()
	return &Token{
		ID:     string(uuid.NewUUID()),
		Issued: now,
		Expiry: now.Add(tm.lifetime),
	}
}

func (tm *tokenManager) SignToken(tok *Token) (string, error) {
	claims := jwt.MapClaims{
		ClaimID:           tok.ID,
		ClaimIssued:       tok.Issued.Unix(),
		ClaimExpiry:       tok.Expiry.Unix(),
		ClaimIssuer:       Issuer,
		ClaimSubject:      tok.Subject.DomainName + "/" + tok.Subject.UserName,
		ClaimScopeDomain:  tok.Scope.DomainName,
		ClaimScopeProject: tok.Scope.ProjectName,
	}

	var jtok *jwt.Token
	switch key := tm.keys[0].(type) {
	case *rsa.PrivateKey:
		jtok = jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	case *ecdsa.PrivateKey:
		switch key.Curve {
		case elliptic.P256():
			jtok = jwt.NewWithClaims(jwt.SigningMethodES256, claims)
		case elliptic.P384():
			jtok = jwt.NewWithClaims(jwt.SigningMethodES384, claims)
		case elliptic.P521():
			jtok = jwt.NewWithClaims(jwt.SigningMethodES512, claims)
		default:
			return "", fmt.Errorf("Unknown elliptic curve type")
		}
	default:
		return "", fmt.Errorf("Unknown private key type")
	}

	return jtok.SignedString(tm.keys[0])
}

func validateTokenKey(toksig string, key interface{}) (*Token, error) {
	jtok, err := jwt.Parse(toksig, func(jtok *jwt.Token) (interface{}, error) {
		switch jtok.Method.(type) {
		case *jwt.SigningMethodRSA:
			privKey, ok := key.(*rsa.PrivateKey)
			if ok {
				return &privKey.PublicKey, nil
			}
			return nil, fmt.Errorf("Not an RSA key")
		case *jwt.SigningMethodECDSA:
			privKey, ok := key.(*ecdsa.PrivateKey)
			if ok {
				return &privKey.PublicKey, nil
			}
			return nil, fmt.Errorf("Not an ECDSA key")
		default:
			return nil, fmt.Errorf("Unknown key type")
		}
	})

	if err != nil {
		return nil, err
	}

	claims, ok := jtok.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("Unexpected claims type")
	}

	id, ok := claims[ClaimID].(string)
	if !ok {
		return nil, fmt.Errorf("Unexpected id claim type")
	}

	subject, ok := claims[ClaimSubject].(string)
	if !ok {
		return nil, fmt.Errorf("Unexpected subject claim type")
	}

	domain, ok := claims[ClaimScopeDomain].(string)
	if !ok {
		return nil, fmt.Errorf("Unexpected domain claim type")
	}

	project, ok := claims[ClaimScopeProject].(string)
	if !ok {
		return nil, fmt.Errorf("Unexpected project claim type")
	}

	subjectBits := strings.Split(subject, "/")
	if len(subjectBits) != 2 {
		return nil, fmt.Errorf("Unexpected subject format %s", subject)
	}

	return &Token{
		ID: id,
		Subject: TokenSubject{
			DomainName: subjectBits[0],
			UserName:   subjectBits[1],
		},
		Scope: TokenScope{
			DomainName:  domain,
			ProjectName: project,
		},
	}, nil
}

func (tm *tokenManager) ValidateToken(toksig string) (*Token, error) {
	var firstErr error
	for _, key := range tm.keys {
		tok, err := validateTokenKey(toksig, key)
		if tok != nil {
			_, err = tm.tokenClient.Get(tok.ID)
			if err == nil {
				return nil, fmt.Errorf("Token %s is revoked", tok.ID)
			}

			return tok, nil
		}
		if firstErr == nil {
			firstErr = err
		}
	}

	return nil, firstErr
}
