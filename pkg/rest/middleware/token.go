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

	"github.com/dicot-project/dicot-api/pkg/api/identity"
	"github.com/dicot-project/dicot-api/pkg/api/identity/v1"
	"github.com/dicot-project/dicot-api/pkg/auth"
)

type tokenHandler struct {
	TokenManager auth.TokenManager
	Client       identity.Interface
	AllowAnon    bool
}

func newTokenHandler(tokenManager auth.TokenManager, client identity.Interface, allowAnon bool) Middleware {
	return &tokenHandler{
		TokenManager: tokenManager,
		Client:       client,
		AllowAnon:    allowAnon,
	}
}

func NewTokenHandler(tokenManager auth.TokenManager, client identity.Interface) Middleware {
	return newTokenHandler(tokenManager, client, false)
}

func NewTokenHandlerAllowAnon(tokenManager auth.TokenManager, client identity.Interface) Middleware {
	return newTokenHandler(tokenManager, client, true)
}

func (h *tokenHandler) setToken(c *gin.Context, tok *auth.Token) error {
	userNS := identity.FormatDomainNamespace(tok.Subject.DomainName)
	userClnt := h.Client.Users(userNS)
	glog.V(1).Infof("Lookup subject user '%s/%s'", userNS, tok.Subject.UserName)
	user, err := userClnt.Get(tok.Subject.UserName)
	if err != nil {
		glog.V(1).Info("Fail %s", err)
		return err
	}

	domainClnt := h.Client.Projects(v1.NamespaceSystem)
	glog.V(1).Infof("Lookup scope domain '%s/%s'", v1.NamespaceSystem, tok.Scope.DomainName)
	domain, err := domainClnt.Get(tok.Scope.DomainName)
	if err != nil {
		return err
	}

	projectNS := identity.FormatDomainNamespace(tok.Scope.DomainName)
	projectClnt := h.Client.Projects(projectNS)
	glog.V(1).Infof("Lookup scope project '%s/%s'", projectNS, tok.Scope.ProjectName)
	project, err := projectClnt.Get(tok.Scope.ProjectName)
	if err != nil {
		return err
	}

	glog.V(1).Infof("Set user %s domain %s project %s", user, domain, project)
	c.Set("TokenSubjectUser", user)
	c.Set("TokenScopeDomain", domain)
	c.Set("TokenScopeProject", project)

	return nil
}

func GetTokenSubjectUser(c *gin.Context) *v1.User {
	obj, ok := c.Get("TokenSubjectUser")
	if !ok {
		return nil
	}
	user, ok := obj.(*v1.User)
	if !ok {
		return nil
	}
	return user
}

func RequiredTokenSubjectUser(c *gin.Context) *v1.User {
	user := GetTokenSubjectUser(c)
	if user == nil {
		panic("User is unexpectedly nil")
	}
	return user
}

func GetTokenScopeDomain(c *gin.Context) *v1.Project {
	obj, ok := c.Get("TokenScopeDomain")
	if !ok {
		return nil
	}
	domain, ok := obj.(*v1.Project)
	if !ok {
		return nil
	}
	return domain
}

func RequiredTokenScopeDomain(c *gin.Context) *v1.Project {
	proj := GetTokenScopeDomain(c)
	if proj == nil {
		panic("Domain is unexpectedly nil")
	}
	return proj
}

func GetTokenScopeProject(c *gin.Context) *v1.Project {
	obj, ok := c.Get("TokenScopeProject")
	if !ok {
		return nil
	}
	project, ok := obj.(*v1.Project)
	if !ok {
		return nil
	}
	return project
}

func RequiredTokenScopeProject(c *gin.Context) *v1.Project {
	proj := GetTokenScopeProject(c)
	if proj == nil {
		panic("Project is unexpectedly nil")
	}
	return proj
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
