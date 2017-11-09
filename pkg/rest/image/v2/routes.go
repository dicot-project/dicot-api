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

package v2

import (
	"github.com/gin-gonic/gin"
	k8srest "k8s.io/client-go/rest"

	"github.com/dicot-project/dicot-api/pkg/auth"
	"github.com/dicot-project/dicot-api/pkg/rest"
	"github.com/dicot-project/dicot-api/pkg/rest/middleware"
)

type service struct {
	IdentityClient k8srest.Interface
	ImageClient    k8srest.Interface
	Prefix         string
	ServerID       string
	TokenManager   auth.TokenManager
	ImageRepo      string
}

func NewService(identityClient k8srest.Interface, imageClient k8srest.Interface, tm auth.TokenManager, imagerepo string, serverID string, prefix string) rest.Service {
	if prefix == "" {
		prefix = "/image"
	}
	return &service{
		IdentityClient: identityClient,
		ImageClient:    imageClient,
		Prefix:         prefix,
		ServerID:       serverID,
		TokenManager:   tm,
		ImageRepo:      imagerepo,
	}
}

func (svc *service) GetPrefix() string {
	return svc.Prefix
}

func (svc *service) GetName() string {
	return "dicot-image"
}

func (svc *service) GetType() string {
	return "image"
}

func (svc *service) GetUID() string {
	return "578c5644-ec4a-408c-b5a4-03dec9e88298"
}

func (svc *service) RegisterRoutes(router *gin.RouterGroup) {
	router.Use(middleware.NewTokenHandler(svc.TokenManager, svc.IdentityClient).Handler())

	router.GET("/", svc.IndexShow)
	router.GET("/versions", svc.VersionIndexShow)

	router.GET("/v2/images", svc.ImageList)
	router.POST("/v2/images", svc.ImageCreate)
	router.GET("/v2/images/:imageID", svc.ImageShow)
	router.DELETE("/v2/images/:imageID", svc.ImageDelete)
	router.PATCH("/v2/images/:imageID", rest.RequiresFormat("application/openstack-images-v2.1-json-patch"), svc.ImagePatch)
	router.POST("/v2/images/:imageID/actions/deactivate", svc.ImageDeactivate)
	router.POST("/v2/images/:imageID/actions/reactivate", svc.ImageReactivate)
	router.PUT("/v2/images/:imageID/tags/:tag", svc.ImageTagAdd)
	router.DELETE("/v2/images/:imageID/tags/:tag", svc.ImageTagDelete)
	router.PUT("/v2/images/:imageID/file", svc.ImageDataUpload)
	router.GET("/v2/images/:imageID/file", svc.ImageDataDownload)

	router.GET("/v2/schemas/image", svc.SchemaImageShow)
}
