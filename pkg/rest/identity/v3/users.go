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

package v3

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sv1 "k8s.io/client-go/pkg/api/v1"

	"github.com/dicot-project/dicot-api/pkg/api/identity"
	"github.com/dicot-project/dicot-api/pkg/api/identity/v1"
	"github.com/dicot-project/dicot-api/pkg/crypto"
	"github.com/dicot-project/dicot-api/pkg/rest"
	"github.com/dicot-project/dicot-api/pkg/rest/middleware"
)

type UserListRes struct {
	Users []UserInfo `json:"users"`
}

type UserInfo struct {
	ID                string        `json:"id"`
	Name              string        `json:"name"`
	Enabled           bool          `json:"enabled"`
	DomainID          string        `json:"domain_id"`
	Password          string        `json:"password,omitempty"`
	PasswordExpiresAt *string       `json:"password_expires_at"`
	DefaultProjectID  string        `json:"default_project_id,omitempty"`
	Links             rest.LinkInfo `json:"links"`
	Description       string        `json:"description,omitempty"`
	EMail             string        `json:"email,omitempty"`
}

type UserCreateReq struct {
	User UserInfo `json:"user"`
}

type UserUpdateReq struct {
	User UserUpdateInfo `json:"user"`
}

type UserUpdateInfo struct {
	Name             *string `json:"name"`
	Enabled          *bool   `json:"enabled"`
	DomainID         *string `json:"domain_id"`
	DefaultProjectID *string `json:"default_project_id"`
	Password         *string `json:"password"`
	Description      *string `json:"description"`
	EMail            *string `json:"email"`
}

type UserShowRes struct {
	User UserInfo `json:"user"`
}

func (svc *service) UserList(c *gin.Context) {
	name := c.Query("name")

	clnt := identity.NewUserClient(svc.IdentityClient, k8sv1.NamespaceAll)

	users, err := clnt.List()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	res := &UserListRes{
		Users: []UserInfo{},
	}

	// XXX Links field
	for _, user := range users.Items {
		if name != "" && user.ObjectMeta.Name != name {
			continue
		}
		info := UserInfo{
			ID:               string(user.ObjectMeta.UID),
			Name:             user.Spec.Name,
			Enabled:          user.Spec.Enabled,
			DomainID:         user.Spec.DomainID,
			DefaultProjectID: user.Spec.DefaultProjectID,
			Description:      user.Spec.Description,
			EMail:            user.Spec.EMail,
		}
		if user.Spec.Password.ExpiresAt != "" {
			info.PasswordExpiresAt = &user.Spec.Password.ExpiresAt
		}
		res.Users = append(res.Users, info)
	}

	c.JSON(http.StatusOK, res)
}

func (svc *service) UserCreate(c *gin.Context) {
	dom := middleware.GetTokenScopeDomain(c)
	var req UserCreateReq
	err := c.BindJSON(&req)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	domClnt := identity.NewProjectClient(svc.IdentityClient, v1.NamespaceSystem)

	var domNamespace string
	if req.User.DomainID != "" {
		dom, err := domClnt.GetByUID(req.User.DomainID)
		if err != nil {
			if errors.IsNotFound(err) {
				c.AbortWithError(http.StatusBadRequest, err)
			} else {
				c.AbortWithError(http.StatusInternalServerError, err)
			}
			return
		}
		domNamespace = dom.Spec.Namespace
	} else {
		req.User.DomainID = string(dom.ObjectMeta.UID)
		domNamespace = dom.Spec.Namespace
	}

	clnt := identity.NewUserClient(svc.IdentityClient, domNamespace)

	exists, err := clnt.Exists(req.User.Name)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if exists {
		c.AbortWithStatus(http.StatusConflict)
		return
	}

	pwHash, err := crypto.HashPassword(req.User.Password)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	pwSecret := &k8sv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "user-password-" + identity.SanitizeName(req.User.Name),
		},
		Data: map[string][]byte{
			"password": []byte(pwHash),
		},
	}

	user := &v1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: identity.SanitizeName(req.User.Name),
		},
		Spec: v1.UserSpec{
			Name:             req.User.Name,
			Enabled:          req.User.Enabled,
			DomainID:         req.User.DomainID,
			DefaultProjectID: req.User.DefaultProjectID,
			Description:      req.User.Description,
			EMail:            req.User.EMail,
			Password: v1.UserPassword{
				SecretRef: pwSecret.ObjectMeta.Name,
			},
		},
	}

	user, err = clnt.Create(user)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	pwSecret, err = svc.K8SClient.CoreV1().Secrets(domNamespace).Create(pwSecret)
	if err != nil {
		clnt.Delete(user.ObjectMeta.Name, nil)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// XXX links
	res := UserShowRes{
		User: UserInfo{
			ID:               string(user.ObjectMeta.UID),
			Name:             user.Spec.Name,
			Enabled:          user.Spec.Enabled,
			DomainID:         user.Spec.DomainID,
			DefaultProjectID: user.Spec.DefaultProjectID,
			Description:      user.Spec.Description,
			EMail:            user.Spec.EMail,
		},
	}
	if user.Spec.Password.ExpiresAt != "" {
		res.User.PasswordExpiresAt = &user.Spec.Password.ExpiresAt
	}

	c.JSON(http.StatusCreated, res)
}

func (svc *service) UserShow(c *gin.Context) {
	userID := c.Param("userID")

	clnt := identity.NewUserClient(svc.IdentityClient, k8sv1.NamespaceAll)

	user, err := clnt.GetByUID(userID)
	if err != nil {
		if errors.IsNotFound(err) {
			c.AbortWithError(http.StatusNotFound, err)
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		return
	}

	// XXX links
	res := UserShowRes{
		User: UserInfo{
			ID:               string(user.ObjectMeta.UID),
			Name:             user.Spec.Name,
			Enabled:          user.Spec.Enabled,
			DomainID:         user.Spec.DomainID,
			DefaultProjectID: user.Spec.DefaultProjectID,
			Description:      user.Spec.Description,
			EMail:            user.Spec.EMail,
		},
	}
	if user.Spec.Password.ExpiresAt != "" {
		res.User.PasswordExpiresAt = &user.Spec.Password.ExpiresAt
	}

	c.JSON(http.StatusCreated, res)
}

func (svc *service) UserUpdate(c *gin.Context) {
	var req UserUpdateReq
	err := c.BindJSON(&req)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	userID := c.Param("userID")

	clnt := identity.NewUserClient(svc.IdentityClient, k8sv1.NamespaceAll)

	user, err := clnt.GetByUID(userID)
	if err != nil {
		if errors.IsNotFound(err) {
			c.AbortWithError(http.StatusNotFound, err)
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		return
	}

	clnt = identity.NewUserClient(svc.IdentityClient, user.ObjectMeta.Namespace)

	if req.User.Name != nil {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	if req.User.DomainID != nil {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	if req.User.Enabled != nil {
		user.Spec.Enabled = *req.User.Enabled
	}

	if req.User.DefaultProjectID != nil {
		user.Spec.DefaultProjectID = *req.User.DefaultProjectID
	}

	if req.User.Description != nil {
		user.Spec.Description = *req.User.Description
	}

	if req.User.EMail != nil {
		user.Spec.EMail = *req.User.EMail
	}

	if req.User.Password != nil {
		pwHash, err := crypto.HashPassword(*req.User.Password)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		secret, err := svc.K8SClient.CoreV1().Secrets(user.ObjectMeta.Namespace).Get(
			user.Spec.Password.SecretRef, metav1.GetOptions{})
		if err != nil && !errors.IsNotFound(err) {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		if err != nil {
			secret := &k8sv1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: user.Spec.Password.SecretRef,
				},
				Data: map[string][]byte{
					"password": []byte(pwHash),
				},
			}
			secret, err = svc.K8SClient.CoreV1().Secrets(user.ObjectMeta.Namespace).Create(secret)
		} else {
			secret.Data["password"] = []byte(pwHash)
			secret, err = svc.K8SClient.CoreV1().Secrets(user.ObjectMeta.Namespace).Update(secret)
		}
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	user, err = clnt.Update(user)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	res := UserShowRes{
		User: UserInfo{
			ID:               string(user.ObjectMeta.UID),
			Name:             user.Spec.Name,
			Enabled:          user.Spec.Enabled,
			DomainID:         user.Spec.DomainID,
			DefaultProjectID: user.Spec.DefaultProjectID,
			Description:      user.Spec.Description,
			EMail:            user.Spec.EMail,
		},
	}
	if user.Spec.Password.ExpiresAt != "" {
		res.User.PasswordExpiresAt = &user.Spec.Password.ExpiresAt
	}

	c.JSON(http.StatusOK, res)
}

func (svc *service) UserDelete(c *gin.Context) {
	userID := c.Param("userID")

	clnt := identity.NewUserClient(svc.IdentityClient, k8sv1.NamespaceAll)

	user, err := clnt.GetByUID(userID)
	if err != nil {
		if errors.IsNotFound(err) {
			c.AbortWithError(http.StatusNotFound, err)
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		return
	}

	clnt = identity.NewUserClient(svc.IdentityClient, user.ObjectMeta.Namespace)

	err = svc.K8SClient.CoreV1().Secrets(user.ObjectMeta.Namespace).Delete(
		user.Spec.Password.SecretRef, nil)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	err = clnt.Delete(user.ObjectMeta.Name, nil)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.String(http.StatusNoContent, "")
}
