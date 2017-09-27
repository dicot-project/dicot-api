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
	"github.com/dicot-project/dicot-api/pkg/rest"
)

type DomainListRes struct {
	Domains []DomainInfo `json:"domains"`
}

type DomainInfo struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Enabled     bool          `json:"enabled"`
	Links       rest.LinkInfo `json:"links"`
}

type DomainCreateReq struct {
	Domain DomainInfo `json:"domain"`
}

type DomainUpdateReq struct {
	Domain DomainUpdateInfo `json:"domain"`
}

type DomainUpdateInfo struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Enabled     *bool   `json:"enabled"`
}

type DomainShowRes struct {
	Domain DomainInfo `json:"domain"`
}

func (svc *service) DomainList(c *gin.Context) {
	name := c.Query("name")

	clnt := identity.NewProjectClient(svc.RESTClient, v1.NamespaceSystem)

	projects, err := clnt.List()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	res := &DomainListRes{
		Domains: []DomainInfo{},
	}

	// XXX Links field
	for _, project := range projects.Items {
		if project.Spec.Parent != "" {
			continue
		}
		if name != "" && project.ObjectMeta.Name != name {
			continue
		}
		res.Domains = append(res.Domains, DomainInfo{
			ID:          string(project.ObjectMeta.UID),
			Name:        project.ObjectMeta.Name,
			Enabled:     project.Spec.Enabled,
			Description: project.Spec.Description,
		})
	}

	c.JSON(http.StatusOK, res)
}

func (svc *service) DomainCreate(c *gin.Context) {
	var req DomainCreateReq
	err := c.BindJSON(&req)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	clnt := identity.NewProjectClient(svc.RESTClient, v1.NamespaceSystem)

	exists, err := clnt.Exists(req.Domain.Name)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if exists {
		c.AbortWithStatus(http.StatusConflict)
		return
	}

	project := &v1.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name: req.Domain.Name,
		},
		Spec: v1.ProjectSpec{
			Enabled:     req.Domain.Enabled,
			Description: req.Domain.Description,
			Namespace:   identity.FormatDomainNamespace(req.Domain.Name),
		},
	}

	projectNS := &k8sv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: project.Spec.Namespace,
		},
	}

	project, err = clnt.Create(project)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	projectNS, err = svc.Clientset.Namespaces().Create(projectNS)
	if err != nil {
		clnt.Delete(project.ObjectMeta.Name, nil)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// XXX links
	res := DomainShowRes{
		Domain: DomainInfo{
			ID:          string(project.ObjectMeta.UID),
			Name:        project.ObjectMeta.Name,
			Enabled:     project.Spec.Enabled,
			Description: project.Spec.Description,
		},
	}

	c.JSON(http.StatusCreated, res)
}

func (svc *service) DomainShow(c *gin.Context) {
	domainID := c.Param("domainID")

	clnt := identity.NewProjectClient(svc.RESTClient, v1.NamespaceSystem)

	project, err := clnt.GetByUID(domainID)
	if err != nil {
		if errors.IsNotFound(err) {
			c.AbortWithError(http.StatusNotFound, err)
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		return
	}

	// XXX links
	res := DomainShowRes{
		Domain: DomainInfo{
			ID:          string(project.ObjectMeta.UID),
			Name:        project.ObjectMeta.Name,
			Enabled:     project.Spec.Enabled,
			Description: project.Spec.Description,
		},
	}

	c.JSON(http.StatusCreated, res)
}

func (svc *service) DomainUpdate(c *gin.Context) {
	var req DomainUpdateReq
	err := c.BindJSON(&req)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	domainID := c.Param("domainID")

	clnt := identity.NewProjectClient(svc.RESTClient, v1.NamespaceSystem)

	project, err := clnt.GetByUID(domainID)
	if err != nil {
		if errors.IsNotFound(err) {
			c.AbortWithError(http.StatusNotFound, err)
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		return
	}

	if req.Domain.Name != nil {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	if req.Domain.Enabled != nil {
		project.Spec.Enabled = *req.Domain.Enabled
	}
	if req.Domain.Description != nil {
		project.Spec.Description = *req.Domain.Description
	}

	project, err = clnt.Update(project)

	res := DomainShowRes{
		Domain: DomainInfo{
			ID:          string(project.ObjectMeta.UID),
			Name:        project.ObjectMeta.Name,
			Enabled:     project.Spec.Enabled,
			Description: project.Spec.Description,
		},
	}

	c.JSON(http.StatusOK, res)
}

func (svc *service) DomainDelete(c *gin.Context) {
	domainID := c.Param("domainID")

	clnt := identity.NewProjectClient(svc.RESTClient, v1.NamespaceSystem)

	project, err := clnt.GetByUID(domainID)
	if err != nil {
		if errors.IsNotFound(err) {
			c.AbortWithError(http.StatusNotFound, err)
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		return
	}

	err = clnt.Delete(project.ObjectMeta.Name, nil)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	_ = svc.Clientset.Namespaces().Delete(project.Spec.Namespace, nil)

	c.String(http.StatusNoContent, "")
}
