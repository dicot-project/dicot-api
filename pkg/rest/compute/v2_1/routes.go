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
	"github.com/gin-gonic/gin"
	k8s "k8s.io/client-go/kubernetes"
	k8srest "k8s.io/client-go/rest"

	"github.com/dicot-project/dicot-api/pkg/auth"
	"github.com/dicot-project/dicot-api/pkg/rest"
	"github.com/dicot-project/dicot-api/pkg/rest/middleware"
)

type service struct {
	IdentityClient *k8srest.RESTClient
	ComputeClient  *k8srest.RESTClient
	K8SClient      *k8s.Clientset
	Prefix         string
	ServerID       string
	TokenManager   auth.TokenManager
}

func NewService(identityClient *k8srest.RESTClient, computeClient *k8srest.RESTClient, k8sClient *k8s.Clientset, tm auth.TokenManager, serverID string, prefix string) rest.Service {
	if prefix == "" {
		prefix = "/compute/v2.1"
	}
	return &service{
		IdentityClient: identityClient,
		ComputeClient:  computeClient,
		K8SClient:      k8sClient,
		Prefix:         prefix,
		ServerID:       serverID,
		TokenManager:   tm,
	}
}

func (svc *service) GetPrefix() string {
	return svc.Prefix
}

func (svc *service) GetName() string {
	return "dicot-compute"
}

func (svc *service) GetType() string {
	return "compute"
}

func (svc *service) GetUID() string {
	return "f187c571-8a3d-455b-8846-1f373a2f6207"
}

func (svc *service) RegisterRoutes(router *gin.RouterGroup) {
	min := &middleware.MicroVersion{
		Major: 2,
		Micro: 53,
	}
	max := &middleware.MicroVersion{
		Major: 2,
		Micro: 53,
	}
	router.Use(middleware.NewMicroVersionHandler("compute", "X-OpenStack-Nova-API-Version", min, max).Handler())

	router.Use(middleware.NewTokenHandler(svc.TokenManager, svc.IdentityClient).Handler())

	//router.GET("/", svc.IndexShow)
	router.GET("/", svc.VersionIndexShow)

	router.GET("/flavors", svc.FlavorList)
	router.POST("/flavors", svc.FlavorCreate)
	router.DELETE("/flavors/:id", svc.FlavorDelete)
	//router.GET("/flavors/detail", svc.FlavorListDetail)
	router.GET("/flavors/:id", svc.FlavorShow)
	router.GET("/flavors/:id/os-extra_specs", svc.FlavorShowExtraSpecs)
	router.POST("/flavors/:id/os-extra_specs", svc.FlavorCreateExtraSpecs)
	router.GET("/flavors/:id/os-extra_specs/:key", svc.FlavorShowExtraSpec)
	router.POST("/flavors/:id/os-extra_specs/:key", svc.FlavorCreateExtraSpec)
	router.DELETE("/flavors/:id/os-extra_specs/:key", svc.FlavorDeleteExtraSpec)

	router.GET("/os-keypairs", svc.KeypairList)
	router.POST("/os-keypairs", svc.KeypairCreate)
	router.GET("/os-keypairs/:name", svc.KeypairShow)
	router.DELETE("/os-keypairs/:name", svc.KeypairDelete)

	router.GET("/os-hypervisors", svc.HypervisorList)
	//router.GET("/os-hypervisors/detail", svc.HypervisorList)
	router.GET("/os-hypervisors/:name", svc.HypervisorShow)

}
