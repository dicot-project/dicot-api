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

	"github.com/dicot-project/dicot-api/pkg/auth"
	"github.com/dicot-project/dicot-api/pkg/rest"
	"github.com/dicot-project/dicot-api/pkg/rest/middleware"
)

type service struct {
	IdentityClient k8srest.Interface
	K8SClient      k8s.Interface
	Prefix         string
	Services       *rest.ServiceList
	TokenManager   auth.TokenManager
}

func NewService(identityClient k8srest.Interface, k8sClient k8s.Interface, tm auth.TokenManager, svcs *rest.ServiceList, prefix string) rest.Service {
	if prefix == "" {
		prefix = "/identity/v3"
	}
	return &service{
		IdentityClient: identityClient,
		K8SClient:      k8sClient,
		Prefix:         prefix,
		Services:       svcs,
		TokenManager:   tm,
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
	tokNoAnon := middleware.NewTokenHandler(svc.TokenManager, svc.IdentityClient).Handler()
	tokAllowAnon := middleware.NewTokenHandlerAllowAnon(svc.TokenManager, svc.IdentityClient).Handler()

	router.GET("/", svc.IndexGet)
	router.POST("/auth/tokens", tokAllowAnon, svc.TokensPost)

	router.GET("/domains", tokNoAnon, svc.DomainList)
	router.POST("/domains", tokNoAnon, svc.DomainCreate)
	router.GET("/domains/:domainID", tokNoAnon, svc.DomainShow)
	router.PATCH("/domains/:domainID", tokNoAnon, svc.DomainUpdate)
	router.DELETE("/domains/:domainID", tokNoAnon, svc.DomainDelete)

	router.GET("/projects", tokNoAnon, svc.ProjectList)
	router.POST("/projects", tokNoAnon, svc.ProjectCreate)
	router.GET("/projects/:projectID", tokNoAnon, svc.ProjectShow)
	router.PATCH("/projects/:projectID", tokNoAnon, svc.ProjectUpdate)
	router.DELETE("/projects/:projectID", tokNoAnon, svc.ProjectDelete)

	router.GET("/users", tokNoAnon, svc.UserList)
	router.POST("/users", tokNoAnon, svc.UserCreate)
	router.GET("/users/:userID", tokNoAnon, svc.UserShow)
	router.PATCH("/users/:userID", tokNoAnon, svc.UserUpdate)
	router.DELETE("/users/:userID", tokNoAnon, svc.UserDelete)

	router.GET("/groups", tokNoAnon, svc.GroupList)
	router.POST("/groups", tokNoAnon, svc.GroupCreate)
	router.GET("/groups/:groupID", tokNoAnon, svc.GroupShow)
	router.PATCH("/groups/:groupID", tokNoAnon, svc.GroupUpdate)
	router.DELETE("/groups/:groupID", tokNoAnon, svc.GroupDelete)
	router.GET("/groups/:groupID/users", tokNoAnon, svc.GroupUserList)
	router.PUT("/groups/:groupID/users/:userID", tokNoAnon, svc.GroupUserAdd)
	router.HEAD("/groups/:groupID/users/:userID", tokNoAnon, svc.GroupUserCheck)
	router.DELETE("/groups/:groupID/users/:userID", tokNoAnon, svc.GroupUserDelete)
}
