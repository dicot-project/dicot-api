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
	"strconv"

	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sv1 "k8s.io/client-go/pkg/api/v1"

	"github.com/dicot-project/dicot-api/pkg/api"
	"github.com/dicot-project/dicot-api/pkg/api/v1"
	"github.com/dicot-project/dicot-api/pkg/rest"
)

type ProjectListRes struct {
	Projects []ProjectInfo `json:"projects"`
}

type ProjectInfo struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Enabled     bool          `json:"enabled"`
	IsDomain    bool          `json:"is_domain"`
	ParentID    string        `json:"parent_id"`
	DomainID    string        `json:"domain_id"`
	Links       rest.LinkInfo `json:"links"`
}

type ProjectCreateReq struct {
	Project ProjectInfo `json:"project"`
}

type ProjectUpdateReq struct {
	Project ProjectUpdateInfo `json:"project"`
}

type ProjectUpdateInfo struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Enabled     *bool   `json:"enabled"`
}

type ProjectShowRes struct {
	Project ProjectInfo `json:"project"`
}

func (svc *service) ProjectList(c *gin.Context) {
	isDomStr := c.Query("is_domain")
	name := c.Query("name")
	parent := c.Query("parent_id")

	isDom, err := strconv.ParseBool(isDomStr)
	if err != nil {
		isDom = false
	}

	clnt := api.NewProjectClient(svc.RESTClient, k8sv1.NamespaceAll)

	projects, err := clnt.List()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	res := &ProjectListRes{
		Projects: []ProjectInfo{},
	}

	// XXX Links field
	for _, project := range projects.Items {
		if project.Spec.Parent == "" {
			if !isDom {
				continue
			}
		} else {
			if isDom {
				continue
			}
		}
		if name != "" && project.ObjectMeta.Name != name {
			continue
		}
		if parent != "" && project.Spec.Parent != parent {
			continue
		}
		res.Projects = append(res.Projects, ProjectInfo{
			ID:          string(project.ObjectMeta.UID),
			Name:        project.ObjectMeta.Name,
			Enabled:     project.Spec.Enabled,
			Description: project.Spec.Description,
			IsDomain:    isDom,
			ParentID:    project.Spec.Parent,
			DomainID:    project.Spec.Domain,
		})
	}

	c.JSON(http.StatusOK, res)
}

func (svc *service) ProjectCreate(c *gin.Context) {
	var req ProjectCreateReq
	err := c.BindJSON(&req)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// XXX pick right domain based on auth token project
	var parentID string
	var domainID string
	var namespace string
	var clnt *api.ProjectClient
	domClnt := api.NewProjectClient(svc.RESTClient, v1.NamespaceSystem)
	if req.Project.IsDomain {
		exists, err := domClnt.Exists(req.Project.Name)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		if exists {
			c.AbortWithStatus(http.StatusConflict)
			return
		}

		if req.Project.DomainID != "" || req.Project.ParentID != "" {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		clnt = domClnt
		namespace = api.FormatDomainNamespace(req.Project.Name)
	} else {
		var domainName string
		if req.Project.DomainID == "" && req.Project.ParentID == "" {
			// XXX pull 'default' from auth token
			domain, err := domClnt.Get("default")
			if err != nil {
				if errors.IsNotFound(err) {
					c.AbortWithStatus(http.StatusBadRequest)
				} else {
					c.AbortWithError(http.StatusInternalServerError, err)
				}
				return
			}
			domainName = domain.ObjectMeta.Name
			domainID = string(domain.ObjectMeta.UID)
			parentID = domainID
		} else {
			if req.Project.DomainID != "" {
				domain, err := domClnt.GetByUID(req.Project.DomainID)
				if err != nil {
					if errors.IsNotFound(err) {
						c.AbortWithError(http.StatusBadRequest, err)
					} else {
						c.AbortWithError(http.StatusInternalServerError, err)
					}
					return
				}

				domainName = domain.ObjectMeta.Name
				domainID = string(domain.ObjectMeta.UID)
			}
			if req.Project.ParentID == "" {
				parentID = domainID
			} else {
				allProjClnt := api.NewProjectClient(svc.RESTClient, k8sv1.NamespaceAll)

				parent, err := allProjClnt.GetByUID(req.Project.ParentID)
				if err != nil {
					if errors.IsNotFound(err) {
						c.AbortWithError(http.StatusBadRequest, err)
					} else {
						c.AbortWithError(http.StatusInternalServerError, err)
					}
					return
				}

				parentID = string(parent.ObjectMeta.UID)
				if domainID == "" {
					domain, err := domClnt.GetByUID(parent.Spec.Domain)
					if err != nil {
						if errors.IsNotFound(err) {
							c.AbortWithError(http.StatusBadRequest, err)
						} else {
							c.AbortWithError(http.StatusInternalServerError, err)
						}
						return
					}

					domainID = string(domain.ObjectMeta.UID)
					domainName = domain.ObjectMeta.Name
				} else if domainID != parent.Spec.Domain {
					c.AbortWithStatus(http.StatusBadRequest)
					return
				}
			}
		}

		clnt = api.NewProjectClient(svc.RESTClient, api.FormatDomainNamespace(domainName))
		namespace = api.FormatProjectNamespace(domainName, req.Project.Name)
	}

	project := &v1.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name: req.Project.Name,
		},
		Spec: v1.ProjectSpec{
			Enabled:     req.Project.Enabled,
			Description: req.Project.Description,
			Parent:      parentID,
			Domain:      domainID,
			Namespace:   namespace,
		},
	}

	projectNS := &k8sv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: project.Spec.Namespace,
		},
	}

	exists, err := clnt.Exists(req.Project.Name)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if exists {
		c.AbortWithStatus(http.StatusConflict)
		return
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
	res := ProjectShowRes{
		Project: ProjectInfo{
			ID:          string(project.ObjectMeta.UID),
			Name:        project.ObjectMeta.Name,
			Enabled:     project.Spec.Enabled,
			Description: project.Spec.Description,
			IsDomain:    req.Project.IsDomain,
			ParentID:    project.Spec.Parent,
			DomainID:    project.Spec.Domain,
		},
	}

	c.JSON(http.StatusCreated, res)
}

func (svc *service) ProjectShow(c *gin.Context) {
	id := c.Param("id")

	clnt := api.NewProjectClient(svc.RESTClient, k8sv1.NamespaceAll)

	project, err := clnt.GetByUID(id)
	if err != nil {
		if errors.IsNotFound(err) {
			c.AbortWithError(http.StatusBadRequest, err)
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		return
	}

	// XXX links
	res := ProjectShowRes{
		Project: ProjectInfo{
			ID:          string(project.ObjectMeta.UID),
			Name:        project.ObjectMeta.Name,
			Enabled:     project.Spec.Enabled,
			Description: project.Spec.Description,
			ParentID:    project.Spec.Parent,
			DomainID:    project.Spec.Domain,
		},
	}

	c.JSON(http.StatusCreated, res)
}

func (svc *service) ProjectUpdate(c *gin.Context) {
	var req ProjectUpdateReq
	err := c.BindJSON(&req)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	id := c.Param("id")

	clnt := api.NewProjectClient(svc.RESTClient, k8sv1.NamespaceAll)

	project, err := clnt.GetByUID(id)
	if err != nil {
		if errors.IsNotFound(err) {
			c.AbortWithError(http.StatusBadRequest, err)
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		return
	}

	clnt = api.NewProjectClient(svc.RESTClient, project.ObjectMeta.Namespace)

	if req.Project.Name != nil {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	if req.Project.Enabled != nil {
		project.Spec.Enabled = *req.Project.Enabled
	}
	if req.Project.Description != nil {
		project.Spec.Description = *req.Project.Description
	}

	project, err = clnt.Update(project)

	res := ProjectShowRes{
		Project: ProjectInfo{
			ID:          string(project.ObjectMeta.UID),
			Name:        project.ObjectMeta.Name,
			Enabled:     project.Spec.Enabled,
			Description: project.Spec.Description,
			ParentID:    project.Spec.Parent,
			DomainID:    project.Spec.Domain,
		},
	}

	c.JSON(http.StatusOK, res)
}

func (svc *service) ProjectDelete(c *gin.Context) {
	id := c.Param("id")

	clnt := api.NewProjectClient(svc.RESTClient, k8sv1.NamespaceAll)

	project, err := clnt.GetByUID(id)
	if err != nil {
		if errors.IsNotFound(err) {
			c.AbortWithError(http.StatusBadRequest, err)
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		return
	}

	clnt = api.NewProjectClient(svc.RESTClient, project.ObjectMeta.Namespace)

	err = clnt.Delete(project.ObjectMeta.Name, nil)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	_ = svc.Clientset.Namespaces().Delete(project.Spec.Namespace, nil)

	c.String(http.StatusNoContent, "")
}
