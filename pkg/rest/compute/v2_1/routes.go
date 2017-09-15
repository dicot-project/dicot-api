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

	"github.com/dicot-project/dicot-api/pkg/rest"
)

type service struct {
	RESTClient *k8srest.RESTClient
	Clientset  *k8s.Clientset
	Prefix     string
	ServerID   string
}

func NewService(cl *k8srest.RESTClient, cls *k8s.Clientset, serverID string, prefix string) rest.Service {
	if prefix == "" {
		prefix = "/compute"
	}
	return &service{
		RESTClient: cl,
		Clientset:  cls,
		Prefix:     prefix,
		ServerID:   serverID,
	}
}

func (svc *service) GetPrefix() string {
	return svc.Prefix
}

func (svc *service) RegisterRoutes(router *gin.Engine) {
	router.GET(svc.Prefix+"/v2.1/flavors", svc.FlavorList)
	router.POST(svc.Prefix+"/v2.1/flavors", svc.FlavorCreate)
	router.DELETE(svc.Prefix+"/v2.1/flavors/:id", svc.FlavorDelete)
	//router.GET(svc.Prefix+"/v2.1/flavors/detail", svc.FlavorListDetail)
	router.GET(svc.Prefix+"/v2.1/flavors/:id", svc.FlavorShow)
	router.GET(svc.Prefix+"/v2.1/flavors/:id/os-extra_specs", svc.FlavorShowExtraSpecs)
	router.POST(svc.Prefix+"/v2.1/flavors/:id/os-extra_specs", svc.FlavorCreateExtraSpecs)
	router.GET(svc.Prefix+"/v2.1/flavors/:id/os-extra_specs/:key", svc.FlavorShowExtraSpec)
	router.POST(svc.Prefix+"/v2.1/flavors/:id/os-extra_specs/:key", svc.FlavorCreateExtraSpec)
	router.DELETE(svc.Prefix+"/v2.1/flavors/:id/os-extra_specs/:key", svc.FlavorDeleteExtraSpec)

	router.GET(svc.Prefix+"/v2.1/os-keypairs", svc.KeypairList)
	router.POST(svc.Prefix+"/v2.1/os-keypairs", svc.KeypairCreate)
	router.GET(svc.Prefix+"/v2.1/os-keypairs/:name", svc.KeypairShow)
	router.DELETE(svc.Prefix+"/v2.1/os-keypairs/:name", svc.KeypairDelete)
}
