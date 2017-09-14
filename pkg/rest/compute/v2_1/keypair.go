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

package v2_1

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sv1 "k8s.io/client-go/pkg/api/v1"

	"github.com/dicot-project/dicot-api/pkg/api"
	"github.com/dicot-project/dicot-api/pkg/api/v1"
	"github.com/dicot-project/dicot-api/pkg/crypto"
)

type KeypairListRes struct {
	Keypairs []KeypairInfo `json:"keypairs"`
	Links    []LinkInfo    `json:"keypair_links"`
}

type KeypairCreateReq struct {
	Keypair KeypairInfo `json:"keypair"`
}

type KeypairCreateRes struct {
	Keypair KeypairNewInfo `json:"keypair"`
}

type KeypairNewInfo struct {
	Fingerprint string `json:"fingerprint"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	PublicKey   string `json:"public_key"`
	PrivateKey  string `json:"private_key,omitempty"`
	UserID      string `json:"user_id"`
}

type KeypairShowRes struct {
	Keypair KeypairInfo `json:"keypair"`
}

type KeypairInfo struct {
	Fingerprint string  `json:"fingerprint"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	PublicKey   string  `json:"public_key"`
	PrivateKey  string  `json:"private_key,omitempty"`
	UserID      string  `json:"user_id"`
	Deleted     bool    `json:"deleted"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   *string `json:"updated_at"`
	DeletedAt   *string `json:"deleted_at"`
	ID          uint64  `json:"id"`
}

func (svc *service) KeypairList(c *gin.Context) {
	// XXX user id
	marker := c.Query("marker")
	filterLimit, limit := GetFilterUInt(c, "limit")

	clnt := api.NewKeypairClient(svc.Client, k8sv1.NamespaceDefault)

	keypairs, err := clnt.List()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	res := &KeypairListRes{
		Keypairs: []KeypairInfo{},
	}

	count := uint64(0)
	seenMarker := false
	if marker == "" {
		seenMarker = true
	}
	// XXX Links field
	for _, keypair := range keypairs.Items {
		if marker != "" {
			if marker == keypair.ObjectMeta.Name {
				seenMarker = true
				marker = ""
				continue
			}
		}
		if !seenMarker {
			continue
		}

		res.Keypairs = append(res.Keypairs, KeypairInfo{
			Name:        keypair.ObjectMeta.Name,
			Fingerprint: keypair.Spec.Fingerprint,
			Type:        keypair.Spec.Type,
			PublicKey:   keypair.Spec.PublicKey,
		})

		count = count + 1
		if filterLimit && count >= limit {
			break
		}
	}

	c.JSON(http.StatusOK, res)
}

func (svc *service) KeypairCreate(c *gin.Context) {
	req := KeypairCreateReq{
		Keypair: KeypairInfo{
			Type:   "ssh",
			UserID: "admin",
		},
	}
	err := c.BindJSON(&req)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	clnt := api.NewKeypairClient(svc.Client, k8sv1.NamespaceDefault)

	exists, err := clnt.Exists(req.Keypair.Name)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if exists {
		c.AbortWithStatus(http.StatusConflict)
		return
	}

	var keyManager crypto.KeyManager
	if req.Keypair.Type == "ssh" || req.Keypair.Type == "" {
		keyManager = crypto.NewSSHKeyManager()
	} else if req.Keypair.Type == "x509" {
		keyManager = crypto.NewX509KeyManager()
	} else {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	var privKey string
	if req.Keypair.PublicKey == "" {
		privKey, req.Keypair.PublicKey, err = keyManager.CreateKeyPair(crypto.AlgRSA, 2048)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

	}
	fingerprint, err := keyManager.FingerPrint(req.Keypair.PublicKey)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	keypair := &v1.Keypair{
		ObjectMeta: metav1.ObjectMeta{
			Name: req.Keypair.Name,
		},
		Spec: v1.KeypairSpec{
			Fingerprint: fingerprint,
			Type:        req.Keypair.Type,
			PublicKey:   req.Keypair.PublicKey,
			UserID:      req.Keypair.UserID,
			CreatedAt:   time.Now().Format(time.RFC3339),
		},
	}

	keypair, err = clnt.Create(keypair)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	res := KeypairCreateRes{
		Keypair: KeypairNewInfo{
			Name:        keypair.ObjectMeta.Name,
			Fingerprint: keypair.Spec.Fingerprint,
			Type:        keypair.Spec.Type,
			PublicKey:   keypair.Spec.PublicKey,
			PrivateKey:  privKey,
			UserID:      keypair.Spec.UserID,
		},
	}

	c.JSON(http.StatusCreated, res)
}

func (svc *service) KeypairShow(c *gin.Context) {
	name := c.Param("name")

	clnt := api.NewKeypairClient(svc.Client, k8sv1.NamespaceDefault)

	keypair, err := clnt.Get(name)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if keypair == nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	res := KeypairShowRes{
		Keypair: KeypairInfo{
			ID:          keypair.Spec.ID,
			Name:        keypair.ObjectMeta.Name,
			Fingerprint: keypair.Spec.Fingerprint,
			Type:        keypair.Spec.Type,
			PublicKey:   keypair.Spec.PublicKey,
			UserID:      keypair.Spec.UserID,
			CreatedAt:   keypair.Spec.CreatedAt,
			Deleted:     false,
		},
	}

	c.JSON(http.StatusOK, res)
}

func (svc *service) KeypairDelete(c *gin.Context) {
	name := c.Param("name")

	clnt := api.NewKeypairClient(svc.Client, k8sv1.NamespaceDefault)

	keypair, err := clnt.Get(name)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if keypair == nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	err = clnt.Delete(keypair.ObjectMeta.Name, nil)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.String(http.StatusOK, "")
}
