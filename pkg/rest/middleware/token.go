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

package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	k8srest "k8s.io/client-go/rest"

	"github.com/dicot-project/dicot-api/pkg/api/identity"
	"github.com/dicot-project/dicot-api/pkg/api/identity/v1"
	"github.com/dicot-project/dicot-api/pkg/auth"
)

type tokenHandler struct {
	TokenManager auth.TokenManager
	RESTClient   *k8srest.RESTClient
	AllowAnon    bool
}

func newTokenHandler(tokenManager auth.TokenManager, restClient *k8srest.RESTClient, allowAnon bool) Middleware {
	return &tokenHandler{
		TokenManager: tokenManager,
		RESTClient:   restClient,
		AllowAnon:    allowAnon,
	}
}

func NewTokenHandler(tokenManager auth.TokenManager, restClient *k8srest.RESTClient) Middleware {
	return newTokenHandler(tokenManager, restClient, false)
}

func NewTokenHandlerAllowAnon(tokenManager auth.TokenManager, restClient *k8srest.RESTClient) Middleware {
	return newTokenHandler(tokenManager, restClient, true)
}

func (h *tokenHandler) setToken(c *gin.Context, tok *auth.Token) error {
	glog.V(1).Infof("Set token %s", tok)

	userNS := identity.FormatDomainNamespace(tok.Subject.DomainName)
	userClnt := identity.NewUserClient(h.RESTClient, userNS)
	glog.V(1).Infof("Lookup '%s/%s'", userNS, tok.Subject.UserName)
	user, err := userClnt.Get(tok.Subject.UserName)
	if err != nil {
		glog.V(1).Info("Fail %s", err)
		return err
	}

	projectNS := identity.FormatDomainNamespace(tok.Scope.DomainName)
	projectClnt := identity.NewProjectClient(h.RESTClient, projectNS)
	glog.V(1).Infof("Lookup scope '%s/%s'", projectNS, tok.Scope.ProjectName)
	project, err := projectClnt.Get(tok.Scope.ProjectName)
	if err != nil {
		return err
	}

	glog.V(1).Infof("Set subject %s scope %s", user, project)
	c.Set("TokenSubject", user)
	c.Set("TokenScope", project)

	return nil
}

func GetTokenSubject(c *gin.Context) *v1.User {
	obj, ok := c.Get("TokenSubject")
	if ok {
		return nil
	}
	user, ok := obj.(*v1.User)
	if !ok {
		return nil
	}
	return user
}

func GetTokenScope(c *gin.Context) *v1.Project {
	obj, ok := c.Get("TokenScope")
	if ok {
		return nil
	}
	project, ok := obj.(*v1.Project)
	if !ok {
		return nil
	}
	return project
}

func (h *tokenHandler) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		toksig := c.GetHeader("X-Auth-Token")

		if toksig == "" {
			if !h.AllowAnon {
				c.AbortWithStatus(http.StatusUnauthorized)
			}
			return
		}

		token, err := h.TokenManager.ValidateToken(toksig)

		if err != nil {
			c.AbortWithError(http.StatusUnauthorized, err)
			return
		}

		err = h.setToken(c, token)
		if err != nil {
			c.AbortWithError(http.StatusUnauthorized, err)
			return
		}
	}
}
