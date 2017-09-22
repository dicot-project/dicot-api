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
	"github.com/gin-gonic/gin"
	k8s "k8s.io/client-go/kubernetes"
	k8srest "k8s.io/client-go/rest"

	"github.com/dicot-project/dicot-api/pkg/rest"
)

type service struct {
	RESTClient *k8srest.RESTClient
	Clientset  *k8s.Clientset
	Prefix     string
	Services   *rest.ServiceList
}

func NewService(cl *k8srest.RESTClient, cls *k8s.Clientset, svcs *rest.ServiceList, prefix string) rest.Service {
	if prefix == "" {
		prefix = "/identity/v3"
	}
	return &service{
		RESTClient: cl,
		Clientset:  cls,
		Prefix:     prefix,
		Services:   svcs,
	}
}

func (svc *service) GetPrefix() string {
	return svc.Prefix
}

func (svc *service) GetName() string {
	return "dicot-identity"
}

func (svc *service) GetType() string {
	return "identity"
}

func (svc *service) GetUID() string {
	return "f291d9c6-d70e-43a3-bde7-0051cd257f16"
}

func (svc *service) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/", svc.IndexGet)
	router.POST("/auth/tokens", svc.TokensPost)

	router.GET("/domains", svc.DomainList)
	router.POST("/domains", svc.DomainCreate)
	router.GET("/domains/:id", svc.DomainShow)
	router.PATCH("/domains/:id", svc.DomainUpdate)
	router.DELETE("/domains/:id", svc.DomainDelete)

	router.GET("/projects", svc.ProjectList)
	router.POST("/projects", svc.ProjectCreate)
	router.GET("/projects/:id", svc.ProjectShow)
	router.PATCH("/projects/:id", svc.ProjectUpdate)
	router.DELETE("/projects/:id", svc.ProjectDelete)

	router.GET("/users", svc.UserList)
	router.POST("/users", svc.UserCreate)
	router.GET("/users/:id", svc.UserShow)
	router.PATCH("/users/:id", svc.UserUpdate)
	router.DELETE("/users/:id", svc.UserDelete)
}
