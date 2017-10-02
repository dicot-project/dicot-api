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
	"time"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"

	"github.com/dicot-project/dicot-api/pkg/api/identity"
	"github.com/dicot-project/dicot-api/pkg/api/identity/v1"
	"github.com/dicot-project/dicot-api/pkg/auth"
	"github.com/dicot-project/dicot-api/pkg/crypto"
)

type AuthReq struct {
	Auth AuthInfo `json:"auth"`
}

type AuthInfo struct {
	Scope    AuthInfoScope    `json:"scope"`
	Identity AuthInfoIdentity `json:"identity"`
}

type AuthInfoScope struct {
	Project ProjectInfoRef `json:"project"`
}

type ProjectInfoRef struct {
	ID     string        `json:"id"`
	Name   string        `json:"name"`
	Domain DomainInfoRef `json:"domain"`
}

type DomainInfoRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AuthInfoIdentity struct {
	Methods  []string         `json:"methods"`
	Password AuthInfoPassword `json:"password"`
	Token    AuthInfoToken    `json:"token"`
}

type AuthInfoToken struct {
	ID string `json:"id"`
}

type AuthInfoPassword struct {
	User UserInfoRef `json:"user"`
}

type TokenRes struct {
	Token TokenInfo `json:"token"`
}

type TokenInfo struct {
	Methods   []string           `json:"methods"`
	Roles     []RoleInfo         `json:"roles"`
	ExpiresAt string             `json:"expires_at"`
	IssuedAt  string             `json:"issued_at"`
	Project   ProjectInfoRef     `json:"project"`
	IsDomain  bool               `json:"is_domain"`
	Catalogs  []TokenInfoCatalog `json:"catalog"`
	User      UserInfoRef        `json:"user"`
	AuditIDs  []string           `json:"audit_ids"`
	Extras    map[string]string  `json:"extras"`
}

type TokenInfoCatalog struct {
	ID        string              `json:"id"`
	Endpoints []TokenInfoEndpoint `json:"endpoints"`
	Type      string              `json:"type"`
	Name      string              `json:"name"`
}

type TokenInfoEndpoint struct {
	ID        string `json:"id"`
	Region    string `json:"region"`
	RegionID  string `json:"region_id"`
	URL       string `json:"url"`
	Interface string `json:"interface"`
}

type RoleInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type UserInfoRef struct {
	ID                string        `json:"id"`
	Name              string        `json:"name"`
	Domain            DomainInfoRef `json:"domain"`
	Password          string        `json:"password"`
	PasswordExpiresAt string        `json:"password_expires_at"`
}

func (svc *service) TokensPost(c *gin.Context) {
	var req AuthReq
	err := c.BindJSON(&req)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	pwAuth := false
	for _, val := range req.Auth.Identity.Methods {
		if val == "password" {
			pwAuth = true
			break
		}
	}
	if !pwAuth {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	domClnt := identity.NewProjectClient(svc.IdentityClient, v1.NamespaceSystem)
	var userDomain *v1.Project
	if req.Auth.Identity.Password.User.Domain.Name != "" {
		userDomain, err = domClnt.Get(req.Auth.Identity.Password.User.Domain.Name)
		if err != nil {
			c.AbortWithError(http.StatusUnauthorized, err)
			return
		}
	} else if req.Auth.Identity.Password.User.Domain.ID != "" {
		userDomain, err = domClnt.GetByUID(req.Auth.Identity.Password.User.Domain.ID)
		if err != nil {
			c.AbortWithError(http.StatusUnauthorized, err)
			return
		}
	} else {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	userNamespace := identity.FormatDomainNamespace(userDomain.ObjectMeta.Name)

	userClnt := identity.NewUserClient(svc.IdentityClient, userNamespace)

	var user *v1.User
	if req.Auth.Identity.Password.User.Name != "" {
		user, err = userClnt.Get(req.Auth.Identity.Password.User.Name)
	} else {
		user, err = userClnt.GetByUID(req.Auth.Identity.Password.User.ID)
	}
	if err != nil {
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	secret, err := svc.K8SClient.CoreV1().Secrets(userNamespace).Get(user.Spec.Password.SecretRef, metav1.GetOptions{})
	if err != nil {
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	allowed, err := crypto.CheckPassword(
		req.Auth.Identity.Password.User.Password,
		string(secret.Data["password"]))
	if err != nil {
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	}
	if !allowed {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	var projectDomain *v1.Project
	if req.Auth.Scope.Project.Domain.Name != "" {
		projectDomain, err = domClnt.Get(req.Auth.Scope.Project.Domain.Name)
		if err != nil {
			c.AbortWithError(http.StatusUnauthorized, err)
			return
		}
	} else if req.Auth.Scope.Project.Domain.ID != "" {
		projectDomain, err = domClnt.GetByUID(req.Auth.Scope.Project.Domain.ID)
		if err != nil {
			c.AbortWithError(http.StatusUnauthorized, err)
			return
		}
	} else {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	projectNamespace := identity.FormatDomainNamespace(projectDomain.ObjectMeta.Name)

	projectClnt := identity.NewProjectClient(svc.IdentityClient, projectNamespace)

	var project *v1.Project
	if req.Auth.Scope.Project.Name != "" {
		project, err = projectClnt.Get(req.Auth.Scope.Project.Name)
	} else {
		project, err = projectClnt.GetByUID(req.Auth.Scope.Project.ID)
	}
	if err != nil {
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	// XXX validate user's access to the requested project

	token := svc.TokenManager.NewToken()
	token.Subject = auth.TokenSubject{
		DomainName: userDomain.ObjectMeta.Name,
		UserName:   user.ObjectMeta.Name,
	}
	token.Scope = auth.TokenScope{
		DomainName:  projectDomain.ObjectMeta.Name,
		ProjectName: project.ObjectMeta.Name,
	}

	tokensig, err := svc.TokenManager.SignToken(token)
	if err != nil {
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	catalog := []TokenInfoCatalog{}

	interfaces := []string{
		"internal", "admin", "public",
	}

	for _, service := range svc.Services.Services {
		endpoints := []TokenInfoEndpoint{}

		for _, iface := range interfaces {
			endpoints = append(endpoints, TokenInfoEndpoint{
				ID:        "4e7639cf-f78f-4cd2-aa2a-131196e25974",
				URL:       "http://" + c.Request.Host + service.GetPrefix() + "/",
				Region:    "RegionOne",
				RegionID:  "d3fd5ef9-7eff-422a-8df1-f2bc523d3381",
				Interface: iface,
			})
		}

		catalog = append(catalog, TokenInfoCatalog{
			ID:        service.GetUID(),
			Type:      service.GetType(),
			Name:      service.GetName(),
			Endpoints: endpoints,
		})
	}

	res := &TokenRes{
		Token: TokenInfo{
			Methods: []string{"password"},
			Roles: []RoleInfo{
				RoleInfo{
					ID:   "f56be11a-94a7-11e7-9f6d-e4b318e0afce",
					Name: "admin",
				},
			},
			IssuedAt:  token.Issued.Format(time.RFC3339),
			ExpiresAt: token.Expiry.Format(time.RFC3339),
			IsDomain:  false,
			AuditIDs: []string{
				string(uuid.NewUUID()),
			},
			Project: ProjectInfoRef{
				Domain: DomainInfoRef{
					ID:   string(projectDomain.ObjectMeta.UID),
					Name: projectDomain.ObjectMeta.Name,
				},
				ID:   string(project.ObjectMeta.UID),
				Name: project.ObjectMeta.Name,
			},
			User: UserInfoRef{
				Domain: DomainInfoRef{
					ID:   string(userDomain.ObjectMeta.UID),
					Name: userDomain.ObjectMeta.Name,
				},
				ID:                string(user.ObjectMeta.UID),
				Name:              user.Spec.Name,
				PasswordExpiresAt: time.Now().Add(10 * time.Minute).Format(time.RFC3339),
			},
			Extras: map[string]string{
				"fish": "food",
			},
			Catalogs: catalog,
		},
	}
	c.Header("X-Subject-Token", tokensig)
	c.JSON(http.StatusOK, res)
}
